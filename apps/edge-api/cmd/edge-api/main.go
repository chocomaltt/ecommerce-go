package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/config"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/identitygrpc"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/ordergrpc"
	httpinterface "github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if level, err := logrus.ParseLevel(cfg.Server.LogLevel); err == nil {
		log.SetLevel(level)
	}

	identity, err := identitygrpc.New(
		cfg.Identity.GRPCTarget,
		cfg.Identity.TLSServerName,
		cfg.Identity.TLSCertFile,
		cfg.Identity.TLSKeyFile,
		cfg.Identity.TLSCAFile,
	)
	if err != nil {
		log.WithError(err).Fatal("configure identity gRPC client")
	}
	defer identity.Close()

	orders, err := ordergrpc.New(
		cfg.Order.GRPCTarget,
		cfg.Order.TLSServerName,
		cfg.Order.TLSCertFile,
		cfg.Order.TLSKeyFile,
		cfg.Order.TLSCAFile,
	)
	if err != nil {
		log.WithError(err).Fatal("configure order gRPC client")
	}
	defer orders.Close()

	handler := httpinterface.New(
		identity,
		middleware.NewAuthMiddleware(identity, "order-service"),
		orders,
	)
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go shutdownOnSignal(ctx, server)

	log.WithField("port", cfg.Server.Port).Info("edge-api listening")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownOnSignal(ctx context.Context, server *http.Server) {
	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdown)
}
