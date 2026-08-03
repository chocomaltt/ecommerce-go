// Package assertion issues signed, short-lived internal actor assertions.
package assertion

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/port"
)

type Issuer struct {
	privateKey ed25519.PrivateKey
	issuer     string
	ttl        time.Duration
}
type claims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Audience  string `json:"aud"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	ID        string `json:"jti"`
}

func New(privateKeyFile, issuer string, ttl time.Duration) (*Issuer, error) {
	raw, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read actor private key: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("decode actor private key PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse actor private key: %w", err)
	}
	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("actor private key is not Ed25519")
	}
	return &Issuer{privateKey: key, issuer: issuer, ttl: ttl}, nil
}
func (i *Issuer) Issue(ctx context.Context, actor port.Actor) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if actor.Subject == "" || actor.Audience == "" {
		return "", fmt.Errorf("actor subject and audience are required")
	}
	now := time.Now()
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", fmt.Errorf("generate assertion id: %w", err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`))
	payload, err := json.Marshal(claims{Subject: actor.Subject, Email: actor.Email, Audience: actor.Audience, Issuer: i.issuer, IssuedAt: now.Unix(), ExpiresAt: now.Add(i.ttl).Unix(), ID: base64.RawURLEncoding.EncodeToString(id)})
	if err != nil {
		return "", fmt.Errorf("encode assertion: %w", err)
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(i.privateKey, []byte(header+"."+body))
	return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
