package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func chiURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func parseIDParam(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}
