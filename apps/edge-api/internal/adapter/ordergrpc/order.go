// Package ordergrpc implements Edge's outbound Order port.
package ordergrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	orderv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	connection *grpc.ClientConn
	service    orderv1.OrderIdentityServiceClient
}

func New(target, serverName, certFile, keyFile, caFile string) (*Client, error) {
	creds, err := credentialsFor(serverName, certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	connection, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create order gRPC client: %w", err)
	}
	return &Client{connection: connection, service: orderv1.NewOrderIdentityServiceClient(connection)}, nil
}
func (c *Client) Close() error { return c.connection.Close() }
func (c *Client) GetCaller(ctx context.Context, assertion string) (port.Caller, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+assertion)
	out, err := c.service.GetCaller(ctx, &orderv1.GetCallerRequest{})
	if err != nil {
		return port.Caller{}, fmt.Errorf("get caller from order-service: %w", err)
	}
	return port.Caller{UserID: out.UserId, Email: out.Email, Service: out.CallerService}, nil
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
