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
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	initConfig()

	certificateFile := viper.GetString("TLS_CERT_FILE")
	keyFile := viper.GetString("TLS_KEY_FILE")
	certificateAuthorityFile := viper.GetString("TLS_CA_FILE")
	actorPublicKeyFile := viper.GetString("ACTOR_PUBLIC_KEY_FILE")
	port := viper.GetString("GRPC_PORT")

	credentials, err := grpcinterface.Credentials(certificateFile, keyFile, certificateAuthorityFile)
	if err != nil {
		log.Fatal(err)
	}

	verifier, err := identity.NewVerifier(
		actorPublicKeyFile,
		viper.GetString("ACTOR_ISSUER"),
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

func initConfig() {
	viper.SetDefault("TLS_CERT_FILE", "../../deployments/compose/certs/order-service.crt")
	viper.SetDefault("TLS_KEY_FILE", "../../deployments/compose/certs/order-service.key")
	viper.SetDefault("TLS_CA_FILE", "../../deployments/compose/certs/ca.crt")
	viper.SetDefault("ACTOR_PUBLIC_KEY_FILE", "../../deployments/compose/certs/identity-actor.pub")
	viper.SetDefault("ACTOR_ISSUER", "identity-service")
	viper.SetDefault("GRPC_PORT", "9082")

	viper.AutomaticEnv()
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