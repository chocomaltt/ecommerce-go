// Package grpcinterface exposes identity use cases over private mTLS gRPC.
package grpcinterface

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/usecase/auth"
	identityv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	auth *auth.Service
}

func New(authService *auth.Service) *Server {
	return &Server{auth: authService}
}

func (s *Server) Register(ctx context.Context, request *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	user, token, err := s.auth.Register(ctx, request.Email, request.Password)
	if err != nil {
		return nil, toStatus(err)
	}

	return &identityv1.RegisterResponse{
		User:         userProto(user),
		SessionToken: token,
	}, nil
}

func (s *Server) Login(ctx context.Context, request *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	user, token, err := s.auth.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, toStatus(err)
	}

	return &identityv1.LoginResponse{
		User:         userProto(user),
		SessionToken: token,
	}, nil
}

func (s *Server) ResolveSession(ctx context.Context, request *identityv1.ResolveSessionRequest) (*identityv1.ResolveSessionResponse, error) {
	user, assertion, err := s.auth.ResolveSession(ctx, request.SessionToken, request.Audience)
	if err != nil {
		return nil, toStatus(err)
	}

	return &identityv1.ResolveSessionResponse{
		User:           userProto(user),
		ActorAssertion: assertion,
	}, nil
}

func (s *Server) Logout(ctx context.Context, request *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	if err := s.auth.Logout(ctx, request.SessionToken); err != nil {
		return nil, toStatus(err)
	}

	return &identityv1.LogoutResponse{}, nil
}

func userProto(user auth.User) *identityv1.User {
	return &identityv1.User{
		Id:    user.ID,
		Email: user.Email,
	}
}

func toStatus(err error) error {
	switch err {
	case auth.ErrInvalidCredentials, auth.ErrInvalidSession:
		return status.Error(codes.Unauthenticated, "authentication failed")
	case auth.ErrInvalidRequest:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "identity operation failed")
	}
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

func AuthorizeCaller(expected string) grpc.UnaryServerInterceptor {
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

		return handler(ctx, request)
	}
}
