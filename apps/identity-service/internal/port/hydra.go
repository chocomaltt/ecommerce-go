package port

import "context"

type HydraService interface {
	EnsureClient(context.Context, string, string, []string) error
	LoginRequest(context.Context, string) (subject string, skip bool, err error)
	AcceptLogin(context.Context, string, string) (string, error)
	ConsentScopes(context.Context, string) ([]string, error)
	AcceptConsent(context.Context, string, []string) (string, error)
	AcceptLogout(context.Context, string) (string, error)
}
