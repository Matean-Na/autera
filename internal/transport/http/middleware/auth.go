package middleware

import (
	"context"
	"net/http"
	"strings"

	"autera/internal/modules/users/domain"
	"autera/internal/transport/http/response"
	"autera/pkg/auth"

	"go.uber.org/zap"
)

type ctxKey int

const userKey ctxKey = 1

func WithUser(r *http.Request, u *domain.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey, u))
}

func UserFromCtx(r *http.Request) (*domain.User, bool) {
	u, ok := r.Context().Value(userKey).(*domain.User)
	return u, ok
}

func Auth(jwt *auth.JWT, repo domain.Repository, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				response.Unauthorized(w, "missing bearer token")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			claims, err := jwt.Parse(token)
			if err != nil {
				log.Warn("jwt parse failed", zap.Error(err))
				response.Unauthorized(w, "invalid token")
				return
			}

			if claims.TokenType != auth.TokenAccess {
				response.Unauthorized(w, "invalid token type")
				return
			}

			active, err := repo.IsActive(r.Context(), claims.UserID)
			if err != nil || !active {
				response.Unauthorized(w, "user blocked")
				return
			}

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

			roles := make([]domain.Role, 0, len(claims.Roles))
			for _, rr := range claims.Roles {
				roles = append(roles, domain.Role(rr))
			}

			u := &domain.User{ID: claims.UserID, Roles: roles}
			r = WithUser(r, u)
			next.ServeHTTP(w, r)
		})
	}
}
