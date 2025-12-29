package http

import (
	"autera/internal/transport/http/middleware"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"autera/internal/modules/users/application"
	"autera/internal/transport/http/response"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{
		svc: svc,
	}
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

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in application.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}
	out, err := h.svc.Refresh(r.Context(), in)
	if err != nil {
		response.Unauthorized(w, "refresh failed")
		return
	}
	response.JSON(w, http.StatusOK, out)
}

// logout: отзывает refresh по device_id; опционально — revoke access jti (если включено)
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromCtx(r)
	if !ok || u == nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var in application.LogoutInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}

	claims, _ := middleware.ClaimsFromCtx(r) // может пригодиться для revoke access
	err := h.svc.Logout(r.Context(), u.ID, in, claims)
	if err != nil {
		response.BadRequest(w, "logout failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromCtx(r)
	if !ok || u == nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var in application.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}

	if err := h.svc.ChangePassword(r.Context(), u.ID, in); err != nil {
		response.BadRequest(w, "change password failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true, "changed_at": time.Now()})
}

// ADMIN/OWNER: set roles
func (h *Handler) SetRoles(w http.ResponseWriter, r *http.Request) {
	idStr := chiURLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id", err.Error())
		return
	}

	var in application.SetRolesInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.BadRequest(w, "invalid json", err.Error())
		return
	}

	if err := h.svc.SetRoles(r.Context(), id, in); err != nil {
		response.BadRequest(w, "set roles failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	idStr := chiURLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id", err.Error())
		return
	}
	if err := h.svc.SetActive(r.Context(), id, false); err != nil {
		response.BadRequest(w, "block failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	idStr := chiURLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id", err.Error())
		return
	}
	if err := h.svc.SetActive(r.Context(), id, true); err != nil {
		response.BadRequest(w, "unblock failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
