package http

import "github.com/go-chi/chi/v5"

func RegisterPublicRoutes(r chi.Router, h *Handler) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
}

func RegisterAuthRoutes(r chi.Router, h *Handler) {
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/change_password", h.ChangePassword)
}

func RegisterAdminRoutes(r chi.Router, h *Handler) {
	r.Post("/users/{id}/roles", h.SetRoles)
	r.Post("/users/{id}/block", h.BlockUser)
	r.Post("/users/{id}/unblock", h.UnblockUser)
}
