// Package hydra implements port.HydraService against Ory Hydra's admin API.
// Note: v26 client CRUD lives at /admin/clients (no oauth2 prefix).
package hydra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type HydraAdapter struct {
	AdminURL string
	HTTP     *http.Client
}

type oauth2Client struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func (a *HydraAdapter) EnsureClient(ctx context.Context, clientID, secret string, redirectURIs []string) error {
	get, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		a.AdminURL+"/admin/clients/"+clientID, nil)
	get.Header.Set("Accept", "application/json")
	res, err := a.HTTP.Do(get)
	if err != nil {
		return fmt.Errorf("get client: %w", err)
	}
	res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}

	body, _ := json.Marshal(oauth2Client{
		ClientID:                clientID,
		ClientSecret:            secret,
		GrantTypes:              []string{"authorization_code", "refresh_token", "client_credentials"},
		ResponseTypes:           []string{"code"},
		Scope:                   "openid profile email offline_access",
		RedirectURIs:            redirectURIs,
		TokenEndpointAuthMethod: "client_secret_post",
	})
	post, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		a.AdminURL+"/admin/clients", bytes.NewReader(body))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Accept", "application/json")

	res, err = a.HTTP.Do(post)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("create client: status %d", res.StatusCode)
	}
	return nil
}

func (a *HydraAdapter) LoginRequest(ctx context.Context, challenge string) (string, bool, error) {
	var lr struct {
		Subject string `json:"subject"`
		Skip    bool   `json:"skip"`
	}
	if err := a.get(ctx, "/admin/oauth2/auth/requests/login?login_challenge="+challenge, &lr); err != nil {
		return "", false, err
	}
	return lr.Subject, lr.Skip, nil
}

func (a *HydraAdapter) AcceptLogin(ctx context.Context, challenge, subject string) (string, error) {
	var out struct {
		RedirectTo string `json:"redirect_to"`
	}
	body, _ := json.Marshal(map[string]string{"subject": subject})
	if err := a.put(ctx, "/admin/oauth2/auth/requests/login/accept?login_challenge="+challenge, body, &out); err != nil {
		return "", err
	}
	return out.RedirectTo, nil
}

func (a *HydraAdapter) ConsentScopes(ctx context.Context, challenge string) ([]string, error) {
	var req struct {
		RequestedScopes []string `json:"requested_scope"`
	}
	if err := a.get(ctx, "/admin/oauth2/auth/requests/consent?consent_challenge="+challenge, &req); err != nil {
		return nil, err
	}
	return req.RequestedScopes, nil
}

func (a *HydraAdapter) AcceptConsent(ctx context.Context, challenge string, scopes []string) (string, error) {
	var out struct {
		RedirectTo string `json:"redirect_to"`
	}
	body, _ := json.Marshal(map[string]any{"grant_scope": scopes})
	if err := a.put(ctx, "/admin/oauth2/auth/requests/consent/accept?consent_challenge="+challenge, body, &out); err != nil {
		return "", err
	}
	return out.RedirectTo, nil
}

func (a *HydraAdapter) AcceptLogout(ctx context.Context, challenge string) (string, error) {
	var out struct {
		RedirectTo string `json:"redirect_to"`
	}
	if err := a.put(ctx, "/admin/oauth2/auth/requests/logout/accept?logout_challenge="+challenge, nil, &out); err != nil {
		return "", err
	}
	return out.RedirectTo, nil
}

func (a *HydraAdapter) Introspect(ctx context.Context, token string) (bool, string, error) {
	form := url.Values{}
	form.Set("token", token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		a.AdminURL+"/admin/oauth2/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := a.HTTP.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("introspect: %w", err)
	}
	defer res.Body.Close()

	var out struct {
		Active  bool   `json:"active"`
		Subject string `json:"sub"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return false, "", fmt.Errorf("introspect: decode: %w", err)
	}
	return out.Active, out.Subject, nil
}

func (a *HydraAdapter) get(ctx context.Context, path string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.AdminURL+path, nil)
	req.Header.Set("Accept", "application/json")
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

func (a *HydraAdapter) put(ctx context.Context, path string, body []byte, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, a.AdminURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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
