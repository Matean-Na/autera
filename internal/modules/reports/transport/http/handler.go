package http

import (
	"net/http"
	"strconv"

	"autera/internal/modules/reports/application"
	"autera/internal/transport/http/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) GetBuyerReport(w http.ResponseWriter, r *http.Request) {
	adID, _ := strconv.ParseInt(chi.URLParam(r, "ad_id"), 10, 64)
	rep, err := h.svc.GetByAdID(r.Context(), adID)
	if err != nil {
		response.NotFound(w, "report not found")
		return
	}
	response.JSON(w, http.StatusOK, rep)
}

func (h *Handler) OwnerDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.Dashboard(r.Context())
	if err != nil {
		response.Internal(w, "dashboard failed")
		return
	}
	response.JSON(w, http.StatusOK, data)
}
