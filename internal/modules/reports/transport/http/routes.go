package http

import "github.com/go-chi/chi/v5"

func RegisterBuyerRoutes(r chi.Router, h *Handler) {
	r.Get("/ads/{ad_id}/report", h.GetBuyerReport)
}

func RegisterOwnerRoutes(r chi.Router, h *Handler) {
	r.Get("/dashboard", h.OwnerDashboard)
}
