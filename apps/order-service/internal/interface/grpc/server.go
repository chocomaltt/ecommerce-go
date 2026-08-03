package grpcinterface

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/port"
	identityusecase "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/usecase/identity"
	orderv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	orderv1.UnimplementedOrderIdentityServiceServer
	identity *identityusecase.Service
}

func New(identityService *identityusecase.Service) *Server {
	return &Server{identity: identityService}
}

func (s *Server) GetCaller(ctx context.Context, _ *orderv1.GetCallerRequest) (*orderv1.GetCallerResponse, error) {
	actor, err := s.identity.Caller(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing actor")
	}

	return &orderv1.GetCallerResponse{
		UserId:        actor.UserID,
		Email:         actor.Email,
		CallerService: actor.CallerService,
	}, nil
}

func Credentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load server certificate: %w", err)
	}

	certificateAuthority, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}

	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(certificateAuthority) {
		return nil, fmt.Errorf("parse CA certificate")
	}

	return credentials.NewTLS(&tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    clientCAs,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}

func Authenticate(expected string, verifier port.ActorVerifier, contexts port.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		peer, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing peer")
		}

		tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
		if !ok || len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
			return nil, status.Error(codes.Unauthenticated, "unverified client certificate")
		}
		if err := tlsInfo.State.VerifiedChains[0][0].VerifyHostname(expected); err != nil {
			return nil, status.Error(codes.PermissionDenied, "caller not authorized")
		}

		metadata, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing actor assertion")
		}

		values := metadata.Get("authorization")
		if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "missing actor assertion")
		}

		actor, err := verifier.Verify(strings.TrimPrefix(values[0], "Bearer "))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid actor assertion")
		}

		actor.CallerService = "edge-api"
		return handler(contexts.WithActor(ctx, actor), request)
	}
}
