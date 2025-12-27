package middleware

import (
	"autera/internal/modules/users/domain"
	"context"
	"net/http"
	"strings"

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

func Auth(jwt *auth.JWT, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				response.Unauthorized(w, "missing bearer token")
				return
			}

			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := jwt.Parse(token)
			if err != nil {
				log.Warn("jwt parse failed", zap.Error(err))
				response.Unauthorized(w, "invalid token")
				return
			}

			roles := make([]domain.Role, 0, len(claims.Roles))
			for _, rr := range claims.Roles {
				roles = append(roles, domain.Role(rr)) // если rr string
			}

			u := &domain.User{
				ID:    claims.UserID,
				Roles: roles,
			}

			r = WithUser(r, u)
			next.ServeHTTP(w, r)
		})
	}
}
