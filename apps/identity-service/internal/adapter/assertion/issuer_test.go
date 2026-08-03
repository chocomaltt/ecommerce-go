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
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "actor.key")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	issuer, err := New(path, "identity-service", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := issuer.Issue(context.Background(), port.Actor{Subject: "user-123", Email: "user@example.com", Audience: "order-service"})
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Fatalf("invalid JWT: %q", token)
	}
}
