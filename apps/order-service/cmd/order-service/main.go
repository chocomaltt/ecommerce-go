package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/identity"
	grpcinterface "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/interface/grpc"
	identityusecase "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/usecase/identity"
	orderv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/order/v1"
	"google.golang.org/grpc"
)

func main() {
	certificateFile := env("TLS_CERT_FILE", "../../deployments/compose/certs/order-service.crt")
	keyFile := env("TLS_KEY_FILE", "../../deployments/compose/certs/order-service.key")
	certificateAuthorityFile := env("TLS_CA_FILE", "../../deployments/compose/certs/ca.crt")
	actorPublicKeyFile := env("ACTOR_PUBLIC_KEY_FILE", "../../deployments/compose/certs/identity-actor.pub")
	port := env("GRPC_PORT", "9082")

	credentials, err := grpcinterface.Credentials(certificateFile, keyFile, certificateAuthorityFile)
	if err != nil {
		log.Fatal(err)
	}

	verifier, err := identity.NewVerifier(
		actorPublicKeyFile,
		env("ACTOR_ISSUER", "identity-service"),
		"order-service",
	)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	contexts := identity.Context{}
	server := grpc.NewServer(
		grpc.Creds(credentials),
		grpc.UnaryInterceptor(grpcinterface.Authenticate("edge-api.internal", verifier, contexts)),
	)
	orderv1.RegisterOrderIdentityServiceServer(server, grpcinterface.New(identityusecase.New(contexts)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go shutdownOnSignal(ctx, server)

	log.Printf("order-service gRPC listening on :%s with mTLS and signed actor assertions", port)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func shutdownOnSignal(ctx context.Context, server *grpc.Server) {
	<-ctx.Done()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		server.Stop()
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
