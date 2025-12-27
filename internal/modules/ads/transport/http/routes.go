package http

import "github.com/go-chi/chi/v5"

func RegisterPublicRoutes(r chi.Router, h *Handler) {
	r.Get("/ads", h.ListPublic)
	r.Get("/ads/{id}", h.GetPublic)
}

func RegisterSellerRoutes(r chi.Router, h *Handler) {
	r.Post("/ads", h.CreateSeller)
	r.Post("/ads/{id}/submit", h.SubmitSeller)
}

func RegisterAdminRoutes(r chi.Router, h *Handler) {
	r.Post("/ads/{id}/moderate", h.ModerateAdmin)
}
