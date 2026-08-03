package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifierAcceptsCorrectAudience(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "public.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600); err != nil {
		t.Fatal(err)
	}
	verifier, err := NewVerifier(path, "identity-service", "order-service")
	if err != nil {
		t.Fatal(err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`))
	payload, err := json.Marshal(map[string]any{"sub": "user-123", "email": "user@example.com", "iss": "identity-service", "aud": "order-service", "exp": time.Now().Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	token := header + "." + body + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(header+"."+body)))
	actor, err := verifier.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if actor.UserID != "user-123" || actor.Email != "user@example.com" {
		t.Fatalf("unexpected actor: %#v", actor)
	}
}
