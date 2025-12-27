package middleware

import (
	"net/http"
	"runtime/debug"

	"autera/internal/transport/http/response"

	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						zap.Any("panic", rec),
						zap.ByteString("stack", debug.Stack()),
					)
					response.Internal(w, "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
