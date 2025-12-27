package http

import "github.com/go-chi/chi/v5"

func RegisterSellerRoutes(r chi.Router, h *Handler) {
	r.Post("/ads/{ad_id}/inspection/request", h.RequestSeller)
}

func RegisterAdminRoutes(r chi.Router, h *Handler) {
	r.Post("/inspections/{id}/assign", h.AssignAdmin)
}

func RegisterInspectorRoutes(r chi.Router, h *Handler) {
	r.Get("/inspections", h.ListAssignedInspector)
	r.Post("/inspections/{id}/submit", h.SubmitInspector)
}
