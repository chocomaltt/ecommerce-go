package port

import "context"

// HydraService is the OAuth2/OIDC provider capability.
type HydraService interface {
	EnsureClient(ctx context.Context, clientID, secret string, redirectURIs []string) error
	// LoginRequest returns the subject when the login request is skippable.
	LoginRequest(ctx context.Context, challenge string) (subject string, skip bool, err error)
	AcceptLogin(ctx context.Context, challenge, subject string) (redirect string, err error)
	// ConsentScopes returns the scopes requested by the client.
	ConsentScopes(ctx context.Context, challenge string) ([]string, error)
	AcceptConsent(ctx context.Context, challenge string, scopes []string) (redirect string, err error)
	AcceptLogout(ctx context.Context, challenge string) (redirect string, err error)
	// Introspect reports whether an access token is active and its subject.
	Introspect(ctx context.Context, token string) (active bool, subject string, err error)
}
