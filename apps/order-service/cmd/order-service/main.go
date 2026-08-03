package main

import (
	"context"
	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/identity"
	grpcinterface "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/interface/grpc"
	identityusecase "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/usecase/identity"
	orderv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/order/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cert := env("TLS_CERT_FILE", "../../deployments/compose/certs/order-service.crt")
	key := env("TLS_KEY_FILE", "../../deployments/compose/certs/order-service.key")
	ca := env("TLS_CA_FILE", "../../deployments/compose/certs/ca.crt")
	publicKey := env("ACTOR_PUBLIC_KEY_FILE", "../../deployments/compose/certs/identity-actor.pub")
	creds, err := grpcinterface.Credentials(cert, key, ca)
	if err != nil {
		log.Fatal(err)
	}
	verifier, err := identity.NewVerifier(publicKey, env("ACTOR_ISSUER", "identity-service"), "order-service")
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", ":"+env("GRPC_PORT", "9082"))
	if err != nil {
		log.Fatal(err)
	}
	contexts := identity.Context{}
	server := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(grpcinterface.Authenticate("edge-api.internal", verifier, contexts)))
	orderServer := grpcinterface.New(identityusecase.New(contexts))
	orderv1.RegisterOrderIdentityServiceServer(server, orderServer)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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
	log.Println("order-service gRPC listening on :9082 with mTLS and signed actor assertions")
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
