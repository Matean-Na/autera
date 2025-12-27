package middleware

import (
	"net/http"

	"autera/internal/transport/http/response"

	"go.uber.org/zap"
)

func RBAC(log *zap.Logger, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metaAny := r.Context().Value(UserMetaKey)
			meta, ok := metaAny.(UserMeta)
			if !ok {
				response.Unauthorized(w, "unauthorized")
				return
			}

			for _, role := range meta.Roles {
				if role == requiredRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			log.Info("forbidden",
				zap.Int64("user_id", meta.UserID),
				zap.String("required_role", requiredRole),
			)
			response.Forbidden(w, "forbidden")
		})
	}
}
