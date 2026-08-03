// Package tests contains integration tests for the auth flow.
// They run against the local docker stack (kratos, hydra); skipped when down.
package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/hydra"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/adapter/kratos"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/usecase/auth"
)

const (
	kratosPublic = "http://localhost:4433"
	hydraPublic  = "http://localhost:4444"
	hydraAdmin   = "http://localhost:4445"
)

func newService(t *testing.T) *auth.AuthService {
	t.Helper()

	// Skip when the local stack is not running.
	if !reachable(kratosPublic + "/health/ready") {
		t.Skip("kratos not reachable, run: docker compose up")
	}
	if !reachable(hydraPublic + "/health/ready") {
		t.Skip("hydra not reachable, run: docker compose up")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	return auth.NewAuthService(
		&kratos.KratosAdapter{PublicURL: kratosPublic, HTTP: client},
		&hydra.HydraAdapter{AdminURL: hydraAdmin, HTTP: client},
	)
}

func reachable(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func TestRegisterLoginMe(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	email := "it-test-" + time.Now().Format("150405.000000000") + "@example.com"

	user, token, err := svc.Register(ctx, email, "Str0ngPass!1")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if token == "" {
		t.Fatal("register: empty session token")
	}
	if user.Email != email {
		t.Fatalf("register: got email %q, want %q", user.Email, email)
	}

	// Login with the same credentials.
	user2, token2, err := svc.Login(ctx, email, "Str0ngPass!1")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token2 == "" {
		t.Fatal("login: empty session token")
	}
	if user2.ID != user.ID {
		t.Fatalf("login: got id %q, want %q", user2.ID, user.ID)
	}

	// Whoami through the session token.
	me, err := svc.Me(ctx, token2)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if me.ID != user.ID {
		t.Fatalf("me: got id %q, want %q", me.ID, user.ID)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	svc := newService(t)

	_, _, err := svc.Login(context.Background(), "nobody@example.com", "wrong-password")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestHydraClientCredentialsIntrospect(t *testing.T) {
	newService(t) // gate on stack availability

	client := &http.Client{Timeout: 5 * time.Second}
	hydraAdapter := &hydra.HydraAdapter{AdminURL: hydraAdmin, HTTP: client}
	if err := hydraAdapter.EnsureClient(context.Background(), "edge-api", "edge-api-secret",
		[]string{"http://localhost:8080/auth/callback"}); err != nil {
		t.Fatalf("ensure client: %v", err)
	}

	token, err := clientCredentialsToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	active, subject, err := hydraAdapter.Introspect(context.Background(), token)
	if err != nil {
		t.Fatalf("introspect: %v", err)
	}
	if !active {
		t.Fatal("introspect: token not active")
	}
	if subject != "edge-api" {
		t.Fatalf("introspect: subject %q, want edge-api", subject)
	}

	if !reachable(hydraPublic + "/.well-known/openid-configuration") {
		t.Fatal("hydra discovery not reachable")
	}
}

func clientCredentialsToken() (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "edge-api")
	form.Set("client_secret", "edge-api-secret")
	form.Set("scope", "openid")

	req, _ := http.NewRequest(http.MethodPost, hydraPublic+"/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}
