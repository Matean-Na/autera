package http

import (
	"encoding/json"
	"net/http"

	"autera/internal/modules/users/application"
	"autera/internal/transport/http/response"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in application.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	id, err := h.svc.Register(r.Context(), in)
	if err != nil {
		response.BadRequest(w, "register failed", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"user_id": id})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in application.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	token, err := h.svc.Login(r.Context(), in)
	if err != nil {
		response.Unauthorized(w, "login failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"token": token})
}
