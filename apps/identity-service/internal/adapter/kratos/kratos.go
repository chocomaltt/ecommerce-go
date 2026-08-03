// Package kratos implements the Ory Kratos port.
package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Adapter struct {
	PublicURL string
	HTTP      *http.Client
}

func (a *Adapter) Login(ctx context.Context, email, password string) (string, error) {
	flow, err := a.startFlow(ctx, "/self-service/login/api")
	if err != nil {
		return "", err
	}

	var response struct {
		SessionToken string `json:"session_token"`
	}
	if err := a.submitFlow(ctx, "/self-service/login", flow, request{
		Method:     "password",
		Identifier: email,
		Password:   password,
	}, &response); err != nil {
		return "", err
	}
	if response.SessionToken == "" {
		return "", fmt.Errorf("login: no session token in response")
	}

	return response.SessionToken, nil
}

func (a *Adapter) Register(ctx context.Context, email, password string) error {
	flow, err := a.startFlow(ctx, "/self-service/registration/api")
	if err != nil {
		return err
	}

	return a.submitFlow(ctx, "/self-service/registration", flow, request{
		Method:   "password",
		Password: password,
		Traits:   map[string]string{"email": email},
	}, &struct{}{})
}

func (a *Adapter) Whoami(ctx context.Context, token string) (string, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, a.PublicURL+"/sessions/whoami", nil)
	if err != nil {
		return "", "", fmt.Errorf("whoami request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, err := a.HTTP.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("whoami: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("whoami: status %d", response.StatusCode)
	}

	var session struct {
		Identity identity `json:"identity"`
	}
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		return "", "", fmt.Errorf("whoami decode: %w", err)
	}

	return session.Identity.ID, session.Identity.Traits.Email, nil
}

func (a *Adapter) Logout(ctx context.Context, token string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, a.PublicURL+"/self-service/logout/api", nil)
	if err != nil {
		return fmt.Errorf("logout request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, err := a.HTTP.Do(request)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout: status %d", response.StatusCode)
	}

	return nil
}

type identity struct {
	ID     string `json:"id"`
	Traits struct {
		Email string `json:"email"`
	} `json:"traits"`
}

type request struct {
	Method     string `json:"method"`
	Identifier string `json:"identifier,omitempty"`
	Password   string `json:"password"`
	Traits     any    `json:"traits,omitempty"`
}

func (a *Adapter) startFlow(ctx context.Context, path string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, a.PublicURL+path, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")

	response, err := a.HTTP.Do(request)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: status %d", path, response.StatusCode)
	}

	var flow struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&flow); err != nil {
		return "", fmt.Errorf("%s decode: %w", path, err)
	}

	return flow.ID, nil
}

func (a *Adapter) submitFlow(ctx context.Context, path, flow string, body, output any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.PublicURL+path+"?flow="+flow,
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := a.HTTP.Do(request)
	if err != nil {
		return fmt.Errorf("submit %s: %w", path, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("submit %s: status %d: %s", path, response.StatusCode, body)
	}

	return json.NewDecoder(response.Body).Decode(output)
}
