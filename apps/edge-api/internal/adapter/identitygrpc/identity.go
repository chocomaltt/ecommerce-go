// Package identitygrpc implements Edge's outbound identity port.
package identitygrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	identityv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	connection *grpc.ClientConn
	service    identityv1.IdentityServiceClient
}

func New(target, serverName, certFile, keyFile, caFile string) (*Client, error) {
	creds, err := credentialsFor(serverName, certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	connection, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create identity gRPC client: %w", err)
	}
	return &Client{connection: connection, service: identityv1.NewIdentityServiceClient(connection)}, nil
}
func (c *Client) Close() error { return c.connection.Close() }
func (c *Client) Register(ctx context.Context, email, password string) (port.Authentication, error) {
	out, err := c.service.Register(ctx, &identityv1.RegisterRequest{Email: email, Password: password})
	if err != nil {
		return port.Authentication{}, fmt.Errorf("register identity: %w", err)
	}
	return authentication(out.User, out.SessionToken), nil
}
func (c *Client) Login(ctx context.Context, email, password string) (port.Authentication, error) {
	out, err := c.service.Login(ctx, &identityv1.LoginRequest{Email: email, Password: password})
	if err != nil {
		return port.Authentication{}, fmt.Errorf("login identity: %w", err)
	}
	return authentication(out.User, out.SessionToken), nil
}
func (c *Client) ResolveSession(ctx context.Context, token, audience string) (port.Session, error) {
	out, err := c.service.ResolveSession(ctx, &identityv1.ResolveSessionRequest{SessionToken: token, Audience: audience})
	if err != nil {
		return port.Session{}, fmt.Errorf("resolve identity session: %w", err)
	}
	return port.Session{User: user(out.User), ActorAssertion: out.ActorAssertion}, nil
}
func (c *Client) Logout(ctx context.Context, token string) error {
	_, err := c.service.Logout(ctx, &identityv1.LogoutRequest{SessionToken: token})
	if err != nil {
		return fmt.Errorf("logout identity: %w", err)
	}
	return nil
}
func (c *Client) HydraLogin(ctx context.Context, challenge, token string) (string, error) {
	out, err := c.service.HydraLogin(ctx, &identityv1.HydraLoginRequest{Challenge: challenge, SessionToken: token})
	if err != nil {
		return "", fmt.Errorf("hydra login: %w", err)
	}
	return out.RedirectUrl, nil
}
func (c *Client) HydraConsent(ctx context.Context, challenge string) (string, error) {
	out, err := c.service.HydraConsent(ctx, &identityv1.HydraConsentRequest{Challenge: challenge})
	if err != nil {
		return "", fmt.Errorf("hydra consent: %w", err)
	}
	return out.RedirectUrl, nil
}
func (c *Client) HydraLogout(ctx context.Context, challenge string) (string, error) {
	out, err := c.service.HydraLogout(ctx, &identityv1.HydraLogoutRequest{Challenge: challenge})
	if err != nil {
		return "", fmt.Errorf("hydra logout: %w", err)
	}
	return out.RedirectUrl, nil
}
func authentication(u *identityv1.User, token string) port.Authentication {
	return port.Authentication{User: user(u), SessionToken: token}
}
func user(u *identityv1.User) port.User {
	if u == nil {
		return port.User{}
	}
	return port.User{ID: u.Id, Email: u.Email}
}
func credentialsFor(serverName, certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load edge certificate: %w", err)
	}
	pem, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("parse CA certificate")
	}
	return credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}, RootCAs: roots, ServerName: serverName}), nil
}
