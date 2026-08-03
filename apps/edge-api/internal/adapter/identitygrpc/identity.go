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
	credentials, err := credentialsFor(serverName, certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}

	connection, err := grpc.NewClient(target, grpc.WithTransportCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("create identity gRPC client: %w", err)
	}

	return &Client{
		connection: connection,
		service:    identityv1.NewIdentityServiceClient(connection),
	}, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

func (c *Client) Register(ctx context.Context, email, password string) (port.Authentication, error) {
	response, err := c.service.Register(ctx, &identityv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return port.Authentication{}, fmt.Errorf("register identity: %w", err)
	}

	return authentication(response.User, response.SessionToken), nil
}

func (c *Client) Login(ctx context.Context, email, password string) (port.Authentication, error) {
	response, err := c.service.Login(ctx, &identityv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return port.Authentication{}, fmt.Errorf("login identity: %w", err)
	}

	return authentication(response.User, response.SessionToken), nil
}

func (c *Client) ResolveSession(ctx context.Context, token, audience string) (port.Session, error) {
	response, err := c.service.ResolveSession(ctx, &identityv1.ResolveSessionRequest{
		SessionToken: token,
		Audience:     audience,
	})
	if err != nil {
		return port.Session{}, fmt.Errorf("resolve identity session: %w", err)
	}

	return port.Session{
		User:           user(response.User),
		ActorAssertion: response.ActorAssertion,
	}, nil
}

func (c *Client) Logout(ctx context.Context, token string) error {
	_, err := c.service.Logout(ctx, &identityv1.LogoutRequest{SessionToken: token})
	if err != nil {
		return fmt.Errorf("logout identity: %w", err)
	}

	return nil
}

func authentication(userProto *identityv1.User, token string) port.Authentication {
	return port.Authentication{
		User:         user(userProto),
		SessionToken: token,
	}
}

func user(userProto *identityv1.User) port.User {
	if userProto == nil {
		return port.User{}
	}

	return port.User{
		ID:    userProto.Id,
		Email: userProto.Email,
	}
}

func credentialsFor(serverName, certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load edge certificate: %w", err)
	}

	certificateAuthority, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certificateAuthority) {
		return nil, fmt.Errorf("parse CA certificate")
	}

	return credentials.NewTLS(&tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      roots,
		ServerName:   serverName,
	}), nil
}
