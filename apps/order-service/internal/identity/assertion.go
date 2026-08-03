// Package identity validates trusted actor assertions.
package identity

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/port"
	"os"
	"strings"
	"time"
)

type claims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Audience  string `json:"aud"`
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
}
type Verifier struct {
	publicKey        ed25519.PublicKey
	issuer, audience string
}

func NewVerifier(publicKeyFile, issuer, audience string) (*Verifier, error) {
	raw, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read actor public key: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("decode actor public key PEM")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse actor public key: %w", err)
	}
	key, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("actor public key is not Ed25519")
	}
	return &Verifier{publicKey: key, issuer: issuer, audience: audience}, nil
}
func (v *Verifier) Verify(token string) (port.Actor, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return port.Actor{}, fmt.Errorf("assertion must have three segments")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return port.Actor{}, fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(v.publicKey, []byte(parts[0]+"."+parts[1]), signature) {
		return port.Actor{}, fmt.Errorf("invalid assertion signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return port.Actor{}, fmt.Errorf("decode payload: %w", err)
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return port.Actor{}, fmt.Errorf("decode claims: %w", err)
	}
	if c.Subject == "" || c.Issuer != v.issuer || c.Audience != v.audience || time.Now().Unix() >= c.ExpiresAt {
		return port.Actor{}, fmt.Errorf("invalid assertion claims")
	}
	return port.Actor{UserID: c.Subject, Email: c.Email}, nil
}
