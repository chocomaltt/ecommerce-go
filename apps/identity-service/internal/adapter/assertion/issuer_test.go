package assertion

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/port"
)

func TestIssueCreatesSignedJWT(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	privateKeyFile := filepath.Join(t.TempDir(), "actor.key")
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateDER,
	})
	if err := os.WriteFile(privateKeyFile, privatePEM, 0600); err != nil {
		t.Fatal(err)
	}

	issuer, err := New(privateKeyFile, "identity-service", time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	token, err := issuer.Issue(context.Background(), port.Actor{
		Subject:  "user-123",
		Email:    "user@example.com",
		Audience: "order-service",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(strings.Split(token, ".")) != 3 {
		t.Fatalf("invalid JWT: %q", token)
	}
}
