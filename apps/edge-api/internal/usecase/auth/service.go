package auth

import (
	"context"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
)

// AuthService implements the application use cases, depending only on ports.
type AuthService struct {
	kratos port.KratosService
	hydra  port.HydraService
}

func NewAuthService(kratos port.KratosService, hydra port.HydraService) *AuthService {
	return &AuthService{kratos: kratos, hydra: hydra}
}

// Register creates the identity and returns an authenticated session.
// Kratos does not issue a session on registration -> log in right after.
func (s *AuthService) Register(ctx context.Context, email, password string) (*User, string, error) {
	if err := s.kratos.Register(ctx, email, password); err != nil {
		return nil, "", err
	}
	token, err := s.kratos.Login(ctx, email, password)
	if err != nil {
		return nil, "", err
	}
	user, err := s.Me(ctx, token)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, error) {
	token, err := s.kratos.Login(ctx, email, password)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	user, err := s.Me(ctx, token)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Me(ctx context.Context, token string) (*User, error) {
	id, email, err := s.kratos.Whoami(ctx, token)
	if err != nil {
		return nil, ErrInvalidSession
	}
	return &User{ID: id, Email: email}, nil
}

// HydraLogin accepts the login request. sessionToken may be empty when
// Hydra's login request is skippable.
func (s *AuthService) HydraLogin(ctx context.Context, challenge, sessionToken string) (string, error) {
	subject, skip, err := s.hydra.LoginRequest(ctx, challenge)
	if err != nil {
		return "", err
	}
	if skip {
		return s.hydra.AcceptLogin(ctx, challenge, subject)
	}
	if sessionToken == "" {
		return "", ErrNotAuthenticated
	}
	id, _, err := s.kratos.Whoami(ctx, sessionToken)
	if err != nil {
		return "", ErrInvalidSession
	}
	return s.hydra.AcceptLogin(ctx, challenge, id)
}

// HydraConsent auto-grants the requested scopes (headless).
func (s *AuthService) HydraConsent(ctx context.Context, challenge string) (string, error) {
	scopes, err := s.hydra.ConsentScopes(ctx, challenge)
	if err != nil {
		return "", err
	}
	return s.hydra.AcceptConsent(ctx, challenge, scopes)
}

func (s *AuthService) HydraLogout(ctx context.Context, challenge string) (string, error) {
	return s.hydra.AcceptLogout(ctx, challenge)
}
