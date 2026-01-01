package http

import (
	"autera/internal/modules/users/domain"
	"net/http"
	"time"

	adsh "autera/internal/modules/ads/transport/http"
	insh "autera/internal/modules/inspections/transport/http"
	reph "autera/internal/modules/reports/transport/http"
	userh "autera/internal/modules/users/transport/http"

	"autera/internal/transport/http/middleware"
	"autera/pkg/auth"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type RouterDeps struct {
	Logger *zap.Logger
	JWT    *auth.JWT

	// нужно для Auth middleware: is_active + token_version
	UsersRepo domain.Repository

	UsersHandler *userh.Handler
	AdsHandler   *adsh.Handler
	InsHandler   *insh.Handler
	RepHandler   *reph.Handler
}

func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	// базовые middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Recovery(d.Logger))
	r.Use(middleware.Logging(d.Logger))
	r.Use(chimw.Timeout(60 * time.Second))

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		// PUBLIC
		userh.RegisterPublicRoutes(api, d.UsersHandler)
		adsh.RegisterPublicRoutes(api, d.AdsHandler)

		// AUTH group
		api.Group(func(authR chi.Router) {
			authR.Use(middleware.Auth(d.JWT, d.UsersRepo, d.Logger))

			// общие auth endpoints: logout/change_password
			userh.RegisterAuthRoutes(authR, d.UsersHandler)

			// SELLER
			authR.Route("/seller", func(seller chi.Router) {
				seller.Use(middleware.RBAC(d.Logger, domain.RoleSeller))
				adsh.RegisterSellerRoutes(seller, d.AdsHandler)
				insh.RegisterSellerRoutes(seller, d.InsHandler)
			})

			// INSPECTOR
			authR.Route("/inspector", func(ins chi.Router) {
				ins.Use(middleware.RBAC(d.Logger, domain.RoleInspector))
				insh.RegisterInspectorRoutes(ins, d.InsHandler)
			})

			// ADMIN
			authR.Route("/admin", func(admin chi.Router) {
				admin.Use(middleware.RBAC(d.Logger, domain.RoleAdmin))

				adsh.RegisterAdminRoutes(admin, d.AdsHandler)
				insh.RegisterAdminRoutes(admin, d.InsHandler)

				// admin может: block/unblock + назначать роли без admin/owner
				userh.RegisterAdminRoutes(admin, d.UsersHandler)
			})

			// OWNER
			authR.Route("/owner", func(owner chi.Router) {
				owner.Use(middleware.RBAC(d.Logger, domain.RoleOwner))

				reph.RegisterOwnerRoutes(owner, d.RepHandler)

				// owner может больше: включая назначение admin (но не owner)
				userh.RegisterOwnerRoutes(owner, d.UsersHandler)
			})

			// BUYER
			authR.Route("/buyer", func(buyer chi.Router) {
				buyer.Use(middleware.RBAC(d.Logger, domain.RoleBuyer))
				reph.RegisterBuyerRoutes(buyer, d.RepHandler)
			})
		})
	})

	return r
}
