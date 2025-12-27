package app

import (
	"context"
	"database/sql"
	"net/http"

	transport "autera/internal/transport/http"

	adsapp "autera/internal/modules/ads/application"
	adsinfra "autera/internal/modules/ads/infrastructure"
	adstr "autera/internal/modules/ads/transport/http"

	insapp "autera/internal/modules/inspections/application"
	insinfra "autera/internal/modules/inspections/infrastructure"
	instr "autera/internal/modules/inspections/transport/http"

	repapp "autera/internal/modules/reports/application"
	repinfra "autera/internal/modules/reports/infrastructure"
	reptr "autera/internal/modules/reports/transport/http"

	userapp "autera/internal/modules/users/application"
	userinfra "autera/internal/modules/users/infrastructure"
	usertr "autera/internal/modules/users/transport/http"

	"autera/pkg/auth"

	"go.uber.org/zap"
)

type Application struct {
	DB         *sql.DB
	HTTPServer *http.Server
}

func New(ctx context.Context, cfg *Config, logger *zap.Logger) (*Application, error) {
	db, err := ConnectPostgres(ctx, cfg.DB)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db, cfg.Migrations.URL); err != nil {
		_ = db.Close()
		return nil, err
	}

	jwtSvc := auth.NewJWT(auth.JWTConfig{
		Secret: []byte(cfg.JWT.Secret),
		Issuer: cfg.JWT.Issuer,
		TTLMin: cfg.JWT.TTLMin,
	})

	// Users
	usersRepo := userinfra.NewPostgresRepo(db)
	usersSvc := userapp.NewService(usersRepo, jwtSvc)

	// Ads
	adsRepo := adsinfra.NewPostgresRepo(db)
	adsSvc := adsapp.NewService(adsRepo)

	// Inspections
	insRepo := insinfra.NewPostgresRepo(db)
	insSvc := insapp.NewService(insRepo)

	// Reports
	repRepo := repinfra.NewPostgresRepo(db)
	repSvc := repapp.NewService(repRepo)

	router := transport.NewRouter(transport.RouterDeps{
		Logger: logger,
		JWT:    jwtSvc,

		UsersHandler: usertr.NewHandler(usersSvc),
		AdsHandler:   adstr.NewHandler(adsSvc),
		InsHandler:   instr.NewHandler(insSvc),
		RepHandler:   reptr.NewHandler(repSvc),
	})

	srv := NewHTTPServer(cfg.HTTP.Addr, router)

	return &Application{
		DB:         db,
		HTTPServer: srv,
	}, nil
}
