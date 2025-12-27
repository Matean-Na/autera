package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"autera/internal/app"
)

func main() {
	ctx := context.Background()

	// [config]
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// [logger]
	logger := app.NewZapLogger(cfg.Env)
	defer func() { _ = logger.Sync() }()

	// [application]
	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("app init failed", app.ZapErr(err))
	}

	srv := application.HTTPServer

	go func() {
		logger.Info("http server starting", app.ZapString("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("http server stopped", app.ZapErr(err))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", app.ZapErr(err))
	}
	logger.Info("bye")
}
