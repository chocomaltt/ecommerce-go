package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	identityv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/identity/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/config"
	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/adapter/assertion"
	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/adapter/kratos"
	grpcinterface "github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/interface/grpc"
	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/usecase/auth"
)

func main() {
	log := logrus.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	kratosAdapter := &kratos.Adapter{PublicURL: cfg.Kratos.PublicURL, HTTP: httpClient}
	issuer, err := assertion.New(cfg.Actor.PrivateKeyFile, cfg.Actor.Issuer, time.Duration(cfg.Actor.TTLSeconds)*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	creds, err := grpcinterface.Credentials(cfg.GRPC.TLSCertFile, cfg.GRPC.TLSKeyFile, cfg.GRPC.TLSCAFile)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(grpcinterface.AuthorizeCaller(cfg.GRPC.TrustedCaller)))
	identityServer := grpcinterface.New(auth.New(kratosAdapter, issuer))
	identityv1.RegisterIdentityServiceServer(server, identityServer)
	go func() {
		<-ctx.Done()
		done := make(chan struct{})
		go func() { server.GracefulStop(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			server.Stop()
		}
	}()
	log.Infof("identity-service gRPC listening on :%s", cfg.GRPC.Port)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
