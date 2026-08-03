package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/config"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/hydra"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/kratos"
	httpinterface "github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/usecase/auth"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if level, err := logrus.ParseLevel(cfg.Server.LogLevel); err == nil {
		log.SetLevel(level)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	kratosAdapter := &kratos.KratosAdapter{
		PublicURL: cfg.Kratos.PublicURL,
		AdminURL:  cfg.Kratos.AdminURL,
		HTTP:      &http.Client{Timeout: 10 * time.Second},
	}
	hydraAdapter := &hydra.HydraAdapter{
		AdminURL: cfg.Hydra.AdminURL,
		HTTP:     &http.Client{Timeout: 10 * time.Second},
	}

	// Idempotent bootstrap of the dev OIDC client.
	if err := hydraAdapter.EnsureClient(ctx, cfg.Hydra.ClientID, cfg.Hydra.ClientSecret, cfg.Hydra.RedirectURIs); err != nil {
		log.WithError(err).Fatal("bootstrap hydra client")
	}

	authService := auth.NewAuthService(kratosAdapter, hydraAdapter)
	authMW := middleware.NewAuthMiddleware(kratosAdapter)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: httpinterface.New(authService, authMW),
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	log.WithFields(logrus.Fields{
		"port":   cfg.Server.Port,
		"kratos": cfg.Kratos.PublicURL,
		"hydra":  cfg.Hydra.PublicURL,
	}).Info("edge-api listening")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
