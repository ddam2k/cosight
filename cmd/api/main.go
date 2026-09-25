package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/config"
	"github.com/wasming/cosight/internal/database"
	"github.com/wasming/cosight/internal/httpapi"
	"github.com/wasming/cosight/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	databaseContext, cancelDatabase := context.WithTimeout(ctx, 10*time.Second)
	db, err := database.OpenPostgreSQL(databaseContext, cfg.PostgreSQL)
	cancelDatabase()
	if err != nil {
		slog.Error("initialize PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	verifier, err := auth.NewVerifier(ctx, cfg.Keycloak.Issuer(), cfg.Keycloak.ClientID)
	if err != nil {
		slog.Error("initialize keycloak verifier", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr: cfg.HTTPAddress,
		Handler: httpapi.NewRouter(cfg, verifier, httpapi.Dependencies{
			Users:    repository.NewSQLXUserRepository(db),
			Projects: repository.NewSQLXProjectRepository(db),
		}),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("cosight api listening", "address", cfg.HTTPAddress, "issuer", cfg.Keycloak.Issuer())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("api server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
