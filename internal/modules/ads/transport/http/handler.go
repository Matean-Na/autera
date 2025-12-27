package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"autera/internal/modules/ads/application"
	"autera/internal/modules/ads/domain"
	"autera/internal/transport/http/middleware"
	"autera/internal/transport/http/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	f := domain.ListFilter{Limit: 20, Offset: 0}
	items, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		response.Internal(w, "list failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	ad, err := h.svc.Get(r.Context(), id)
	if err != nil {
		response.NotFound(w, "not found")
		return
	}
	response.JSON(w, http.StatusOK, ad)
}

func (h *Handler) CreateSeller(w http.ResponseWriter, r *http.Request) {
	meta := r.Context().Value(middleware.UserMetaKey).(middleware.UserMeta)

	var in application.CreateAdInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	in.SellerID = meta.UserID

	id, err := h.svc.Create(r.Context(), in)
	if err != nil {
		response.BadRequest(w, "create failed", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"ad_id": id})
}

func (h *Handler) SubmitSeller(w http.ResponseWriter, r *http.Request) {
	meta := r.Context().Value(middleware.UserMetaKey).(middleware.UserMeta)
	adID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.svc.SubmitToModeration(r.Context(), adID, meta.UserID); err != nil {
		response.BadRequest(w, "submit failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"status": "moderation"})
}

func (h *Handler) ModerateAdmin(w http.ResponseWriter, r *http.Request) {
	adID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var body struct {
		Decision string `json:"decision"` // approve / reject
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	if err := h.svc.Moderate(r.Context(), adID, body.Decision); err != nil {
		response.BadRequest(w, "moderate failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
