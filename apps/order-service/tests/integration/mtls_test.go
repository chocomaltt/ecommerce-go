package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/identity"
	grpcinterface "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/interface/grpc"
	identityusecase "github.com/chocomaltt/ecommerce-go/apps/order-service/internal/usecase/identity"
	orderv1 "github.com/chocomaltt/ecommerce-go/common-rpc/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func TestMTLSEdgeAssertionToOrder(t *testing.T) {
	certificates := filepath.Join("..", "..", "..", "..", "deployments", "compose", "certs")

	serverCredentials, err := grpcinterface.Credentials(
		filepath.Join(certificates, "order-service.crt"),
		filepath.Join(certificates, "order-service.key"),
		filepath.Join(certificates, "ca.crt"),
	)
	if err != nil {
		t.Fatalf("server credentials: %v; run certs/generate.sh", err)
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	publicKeyFile := filepath.Join(t.TempDir(), "actor.pub")
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})
	if err := os.WriteFile(publicKeyFile, publicPEM, 0600); err != nil {
		t.Fatal(err)
	}

	verifier, err := identity.NewVerifier(publicKeyFile, "identity-service", "order-service")
	if err != nil {
		t.Fatal(err)
	}

	contexts := identity.Context{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.Creds(serverCredentials),
		grpc.UnaryInterceptor(grpcinterface.Authenticate("edge-api.internal", verifier, contexts)),
	)
	orderv1.RegisterOrderIdentityServiceServer(server, grpcinterface.New(identityusecase.New(contexts)))

	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	clientCredentials, err := edgeCredentials(
		filepath.Join(certificates, "edge-api.crt"),
		filepath.Join(certificates, "edge-api.key"),
		filepath.Join(certificates, "ca.crt"),
	)
	if err != nil {
		t.Fatal(err)
	}

	connection, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(clientCredentials))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = connection.Close()
	})

	assertion, err := signedAssertion(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+assertion)
	response, err := orderv1.NewOrderIdentityServiceClient(connection).GetCaller(ctx, &orderv1.GetCallerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if response.UserId != "user-123" || response.Email != "user@example.com" || response.CallerService != "edge-api" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func edgeCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certificateAuthority, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certificateAuthority) {
		return nil, os.ErrInvalid
	}

	return credentials.NewTLS(&tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      roots,
		ServerName:   "order-service.internal",
	}), nil
}

func signedAssertion(privateKey ed25519.PrivateKey) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`))

	payload, err := json.Marshal(map[string]any{
		"sub":   "user-123",
		"email": "user@example.com",
		"iss":   "identity-service",
		"aud":   "order-service",
		"exp":   time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		return "", err
	}

	body := base64.RawURLEncoding.EncodeToString(payload)
	signedPayload := header + "." + body
	signature := ed25519.Sign(privateKey, []byte(signedPayload))

	return signedPayload + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
