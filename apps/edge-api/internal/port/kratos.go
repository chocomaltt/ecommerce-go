// Package port defines outbound ports (interfaces) the domain depends on.
package port

import "context"

// KratosService is the identity provider capability.
type KratosService interface {
	// Login authenticates and returns a session token.
	Login(ctx context.Context, email, password string) (string, error)
	// Register creates an identity.
	Register(ctx context.Context, email, password string) error
	// Whoami resolves a session token, returning identity id and email.
	Whoami(ctx context.Context, token string) (id, email string, err error)
}
