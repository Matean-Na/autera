package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"autera/internal/modules/users/domain"
	"autera/internal/transport/http/response"
	"autera/pkg/auth"

	"go.uber.org/zap"
)

type ctxKey int

const (
	userKey   ctxKey = 1
	claimsKey ctxKey = 2
)

func WithUser(r *http.Request, u *domain.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey, u))
}

func UserFromCtx(r *http.Request) (*domain.User, bool) {
	u, ok := r.Context().Value(userKey).(*domain.User)
	return u, ok
}

func WithClaims(r *http.Request, c *auth.Claims) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), claimsKey, c))
}

func ClaimsFromCtx(r *http.Request) (*auth.Claims, bool) {
	c, ok := r.Context().Value(claimsKey).(*auth.Claims)
	return c, ok
}

// TokenBlacklist опционален. Если nil — просто не проверяем revoke access.
type TokenBlacklist interface {
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

func Auth(jwt *auth.JWT, repo domain.Repository, bl TokenBlacklist, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				response.Unauthorized(w, "missing bearer token")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			if token == "" {
				response.Unauthorized(w, "missing bearer token")
				return
			}

			claims, err := jwt.Parse(token)
			if err != nil {
				log.Warn("jwt parse failed", zap.Error(err))
				response.Unauthorized(w, "invalid token")
				return
			}

			// 1) обязателен access
			if claims.TokenType != auth.TokenAccess {
				response.Unauthorized(w, "invalid token type")
				return
			}

			// 2) обязательный jti (Claims.ID)
			if claims.ID == "" {
				response.Unauthorized(w, "missing jti")
				return
			}

			// 3) опционально: blacklist для мгновенного revoke access
			if bl != nil {
				revoked, err := bl.IsRevoked(r.Context(), claims.ID)
				if err != nil {
					log.Warn("token blacklist check failed", zap.Error(err))
					response.Unauthorized(w, "invalid token")
					return
				}
				if revoked {
					response.Unauthorized(w, "token revoked")
					return
				}
			}

			// 4) user active?
			active, err := repo.IsActive(r.Context(), claims.UserID)
			if err != nil {
				log.Warn("is_active check failed", zap.Error(err))
				response.Unauthorized(w, "invalid token")
				return
			}
			if !active {
				response.Unauthorized(w, "user blocked")
				return
			}

			// 5) token_version must match (устаревшие роли/пароль/блок)
			tv, err := repo.GetTokenVersion(r.Context(), claims.UserID)
			if err != nil {
				log.Warn("token_version check failed", zap.Error(err))
				response.Unauthorized(w, "invalid token")
				return
			}
			if tv != claims.TokenVersion {
				response.Unauthorized(w, "token outdated")
				return
			}

			// roles из токена (они валидны пока token_version совпадает)
			roles := make([]domain.Role, 0, len(claims.Roles))
			for _, rr := range claims.Roles {
				roles = append(roles, domain.Role(rr))
			}

			u := &domain.User{
				ID:    claims.UserID,
				Roles: roles,
			}

			r = WithUser(r, u)
			r = WithClaims(r, claims)

			// TODO спросить grt
			_ = time.Until(claims.ExpiresAt.Time)

			next.ServeHTTP(w, r)
		})
	}
}
