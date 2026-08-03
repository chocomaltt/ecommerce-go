// Package hydra implements the Ory Hydra admin port.
package hydra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Adapter struct {
	AdminURL string
	HTTP     *http.Client
}
type client struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func (a *Adapter) EnsureClient(ctx context.Context, id, secret string, redirects []string) error {
	var ignored any
	err := a.get(ctx, "/admin/clients/"+id, &ignored)
	if err == nil {
		return nil
	}
	body, err := json.Marshal(client{ClientID: id, ClientSecret: secret, GrantTypes: []string{"authorization_code", "refresh_token", "client_credentials"}, ResponseTypes: []string{"code"}, Scope: "openid profile email offline_access", RedirectURIs: redirects, TokenEndpointAuthMethod: "client_secret_post"})
	if err != nil {
		return err
	}
	return a.post(ctx, "/admin/clients", body, &ignored, http.StatusCreated)
}
func (a *Adapter) LoginRequest(ctx context.Context, challenge string) (string, bool, error) {
	var out struct {
		Subject string `json:"subject"`
		Skip    bool   `json:"skip"`
	}
	err := a.get(ctx, "/admin/oauth2/auth/requests/login?login_challenge="+challenge, &out)
	return out.Subject, out.Skip, err
}
func (a *Adapter) AcceptLogin(ctx context.Context, challenge, subject string) (string, error) {
	var out struct {
		Redirect string `json:"redirect_to"`
	}
	body, _ := json.Marshal(map[string]string{"subject": subject})
	err := a.put(ctx, "/admin/oauth2/auth/requests/login/accept?login_challenge="+challenge, body, &out)
	return out.Redirect, err
}
func (a *Adapter) ConsentScopes(ctx context.Context, challenge string) ([]string, error) {
	var out struct {
		Scopes []string `json:"requested_scope"`
	}
	err := a.get(ctx, "/admin/oauth2/auth/requests/consent?consent_challenge="+challenge, &out)
	return out.Scopes, err
}
func (a *Adapter) AcceptConsent(ctx context.Context, challenge string, scopes []string) (string, error) {
	var out struct {
		Redirect string `json:"redirect_to"`
	}
	body, _ := json.Marshal(map[string]any{"grant_scope": scopes})
	err := a.put(ctx, "/admin/oauth2/auth/requests/consent/accept?consent_challenge="+challenge, body, &out)
	return out.Redirect, err
}
func (a *Adapter) AcceptLogout(ctx context.Context, challenge string) (string, error) {
	var out struct {
		Redirect string `json:"redirect_to"`
	}
	err := a.put(ctx, "/admin/oauth2/auth/requests/logout/accept?logout_challenge="+challenge, nil, &out)
	return out.Redirect, err
}
func (a *Adapter) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.AdminURL+path, nil)
	if err != nil {
		return err
	}
	res, err := a.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: status %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
func (a *Adapter) put(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, a.AdminURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := a.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: status %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
func (a *Adapter) post(ctx context.Context, path string, body []byte, out any, want int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.AdminURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := a.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != want {
		return fmt.Errorf("%s: status %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
