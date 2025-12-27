package middleware

import (
	"context"
	"net/http"
	"strings"

	"autera/internal/transport/http/response"
	"autera/pkg/auth"

	"go.uber.org/zap"
)

type ctxKey string

const UserMetaKey ctxKey = "user_meta"

type UserMeta struct {
	UserID int64    `json:"user_id"`
	Roles  []string `json:"roles"`
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

			meta := UserMeta{
				UserID: claims.UserID,
				Roles:  claims.Roles,
			}
			ctx := context.WithValue(r.Context(), UserMetaKey, meta)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
