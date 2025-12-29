package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func chiURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
