// Package kratos implements port.KratosService against Ory Kratos.
package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type KratosAdapter struct {
	PublicURL string
	AdminURL  string
	HTTP      *http.Client
}

func (a *KratosAdapter) Login(ctx context.Context, email, password string) (string, error) {
	flow, err := a.startFlow(ctx, "/self-service/login/api")
	if err != nil {
		return "", err
	}
	var out struct {
		SessionToken string `json:"session_token"`
	}
	if err := a.submitFlow(ctx, "/self-service/login", flow, credentialsRequest{
		Method:     "password",
		Identifier: email,
		Password:   password,
		CSRF:       "",
	}, &out); err != nil {
		return "", err
	}
	if out.SessionToken == "" {
		return "", fmt.Errorf("login: no session token in response")
	}
	return out.SessionToken, nil
}

func (a *KratosAdapter) Register(ctx context.Context, email, password string) error {
	flow, err := a.startFlow(ctx, "/self-service/registration/api")
	if err != nil {
		return err
	}
	return a.submitFlow(ctx, "/self-service/registration", flow, credentialsRequest{
		Method:   "password",
		Password: password,
		Traits:   map[string]string{"email": email},
		CSRF:     "",
	}, &struct {
		Identity identity `json:"identity"`
	}{})
}

func (a *KratosAdapter) Whoami(ctx context.Context, token string) (string, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.PublicURL+"/sessions/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	res, err := a.HTTP.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("whoami: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("whoami: status %d", res.StatusCode)
	}

	var s struct {
		Identity identity `json:"identity"`
	}
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return "", "", fmt.Errorf("whoami: decode: %w", err)
	}
	return s.Identity.ID, s.Identity.Traits.Email, nil
}

type identity struct {
	ID     string `json:"id"`
	Traits struct {
		Email string `json:"email"`
	} `json:"traits"`
}

type credentialsRequest struct {
	Method     string `json:"method"`
	Identifier string `json:"identifier,omitempty"`
	Password   string `json:"password"`
	Traits     any    `json:"traits,omitempty"`
	CSRF       string `json:"csrf_token"`
}

func (a *KratosAdapter) startFlow(ctx context.Context, path string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.PublicURL+path, nil)
	req.Header.Set("Accept", "application/json")

	res, err := a.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: status %d", path, res.StatusCode)
	}

	var f struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&f); err != nil {
		return "", fmt.Errorf("%s: decode: %w", path, err)
	}
	return f.ID, nil
}

func (a *KratosAdapter) submitFlow(ctx context.Context, path, flow string, body any, out any) error {
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		a.PublicURL+path+"?flow="+flow, bytes.NewReader(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := a.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("submit %s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("submit %s: status %d: %s", path, res.StatusCode, string(b))
	}
	return json.NewDecoder(res.Body).Decode(out)
}
