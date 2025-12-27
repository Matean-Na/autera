package response

import "net/http"

type Error struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func BadRequest(w http.ResponseWriter, msg string, details any) {
	JSON(w, http.StatusBadRequest, Error{Error: msg, Details: details})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnauthorized, Error{Error: msg})
}

func Forbidden(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusForbidden, Error{Error: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Error{Error: msg})
}

func Internal(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, Error{Error: msg})
}
