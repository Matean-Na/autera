package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"autera/internal/modules/inspections/application"
	"autera/internal/transport/http/middleware"
	"autera/internal/transport/http/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RequestSeller(w http.ResponseWriter, r *http.Request) {
	meta := r.Context().Value(middleware.UserMetaKey).(middleware.UserMeta)
	adID, _ := strconv.ParseInt(chi.URLParam(r, "ad_id"), 10, 64)

	id, err := h.svc.Request(r.Context(), adID, meta.UserID)
	if err != nil {
		response.BadRequest(w, "request failed", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"inspection_id": id})
}

func (h *Handler) AssignAdmin(w http.ResponseWriter, r *http.Request) {
	inspectionID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		InspectorID int64 `json:"inspector_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	if err := h.svc.Assign(r.Context(), inspectionID, body.InspectorID); err != nil {
		response.BadRequest(w, "assign failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ListAssignedInspector(w http.ResponseWriter, r *http.Request) {
	meta := r.Context().Value(middleware.UserMetaKey).(middleware.UserMeta)
	items, err := h.svc.ListAssigned(r.Context(), meta.UserID)
	if err != nil {
		response.Internal(w, "list failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) SubmitInspector(w http.ResponseWriter, r *http.Request) {
	meta := r.Context().Value(middleware.UserMetaKey).(middleware.UserMeta)
	inspectionID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.svc.Submit(r.Context(), inspectionID, meta.UserID); err != nil {
		response.BadRequest(w, "submit failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"status": "submitted"})
}
