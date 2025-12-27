package middleware

import (
	"autera/internal/transport/http/response"
	"net/http"

	"autera/internal/modules/users/domain"
	"go.uber.org/zap"
)

func RBAC(logger *zap.Logger, allowed ...domain.Role) func(http.Handler) http.Handler {
	allowedSet := make(map[domain.Role]struct{}, len(allowed))
	for _, rr := range allowed {
		allowedSet[rr] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromCtx(r)
			if !ok || u == nil {
				// Обычно не Debug, а Warn (подозрительно: зашли в RBAC без пользователя)
				logger.Warn("rbac: missing user in context",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_ip", r.RemoteAddr),
				)
				response.Unauthorized(w, "unauthorized")
				return
			}

			for _, ur := range u.Roles {
				if _, ok := allowedSet[ur]; ok {
					// Чтобы не засорять логи, это лучше Debug (или вообще убрать)
					logger.Debug("rbac: allowed",
						zap.Int64("user_id", u.ID),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Strings("user_roles", rolesToStrings(u.Roles)),
						zap.Strings("required_roles", rolesToStrings(allowed)),
					)
					next.ServeHTTP(w, r)
					return
				}
			}

			// Forbidden — это полезно видеть
			logger.Info("rbac: forbidden",
				zap.Int64("user_id", u.ID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_ip", r.RemoteAddr),
				zap.Strings("user_roles", rolesToStrings(u.Roles)),
				zap.Strings("required_roles", rolesToStrings(allowed)),
			)

			response.Forbidden(w, "forbidden")
		})
	}
}

func rolesToStrings(rs []domain.Role) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, string(r))
	}
	return out
}
