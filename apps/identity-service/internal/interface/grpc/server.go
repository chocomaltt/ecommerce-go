// Package grpc exposes identity use cases over private mTLS gRPC.
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

func New(auth *auth.Service) *Server { return &Server{auth: auth} }
func (s *Server) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	user, token, err := s.auth.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.RegisterResponse{User: userProto(user), SessionToken: token}, nil
}
func (s *Server) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	user, token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.LoginResponse{User: userProto(user), SessionToken: token}, nil
}
func (s *Server) ResolveSession(ctx context.Context, req *identityv1.ResolveSessionRequest) (*identityv1.ResolveSessionResponse, error) {
	user, assertion, err := s.auth.ResolveSession(ctx, req.SessionToken, req.Audience)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.ResolveSessionResponse{User: userProto(user), ActorAssertion: assertion}, nil
}
func (s *Server) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	if err := s.auth.Logout(ctx, req.SessionToken); err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.LogoutResponse{}, nil
}
func (s *Server) HydraLogin(ctx context.Context, req *identityv1.HydraLoginRequest) (*identityv1.HydraLoginResponse, error) {
	redirect, err := s.auth.HydraLogin(ctx, req.Challenge, req.SessionToken)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.HydraLoginResponse{RedirectUrl: redirect}, nil
}
func (s *Server) HydraConsent(ctx context.Context, req *identityv1.HydraConsentRequest) (*identityv1.HydraConsentResponse, error) {
	redirect, err := s.auth.HydraConsent(ctx, req.Challenge)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.HydraConsentResponse{RedirectUrl: redirect}, nil
}
func (s *Server) HydraLogout(ctx context.Context, req *identityv1.HydraLogoutRequest) (*identityv1.HydraLogoutResponse, error) {
	redirect, err := s.auth.HydraLogout(ctx, req.Challenge)
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.HydraLogoutResponse{RedirectUrl: redirect}, nil
}
func userProto(user auth.User) *identityv1.User {
	return &identityv1.User{Id: user.ID, Email: user.Email}
}
func toStatus(err error) error {
	switch err {
	case auth.ErrInvalidCredentials, auth.ErrInvalidSession, auth.ErrNotAuthenticated:
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
	pem, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("parse CA certificate")
	}
	return credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}, ClientCAs: pool, ClientAuth: tls.RequireAndVerifyClientCert}), nil
}
func AuthorizeCaller(expected string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing peer")
		}
		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok || len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
			return nil, status.Error(codes.Unauthenticated, "unverified client certificate")
		}
		if err := tlsInfo.State.VerifiedChains[0][0].VerifyHostname(expected); err != nil {
			return nil, status.Error(codes.PermissionDenied, "caller not authorized")
		}
		return handler(ctx, req)
	}
}
