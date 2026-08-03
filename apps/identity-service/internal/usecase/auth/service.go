// Package auth contains identity application use cases.
package auth

import (
	"context"
	"errors"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/port"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid session")
	ErrNotAuthenticated   = errors.New("not authenticated")
	ErrInvalidRequest     = errors.New("invalid request")
)

type User struct {
	ID    string
	Email string
}
type Service struct {
	kratos     port.KratosService
	hydra      port.HydraService
	assertions port.ActorIssuer
}

func New(kratos port.KratosService, hydra port.HydraService, assertions port.ActorIssuer) *Service {
	return &Service{kratos: kratos, hydra: hydra, assertions: assertions}
}
func (s *Service) Register(ctx context.Context, email, password string) (User, string, error) {
	if err := s.kratos.Register(ctx, email, password); err != nil {
		return User{}, "", err
	}
	return s.Login(ctx, email, password)
}
func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	token, err := s.kratos.Login(ctx, email, password)
	if err != nil {
		return User{}, "", ErrInvalidCredentials
	}
	user, err := s.resolve(ctx, token)
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
}
func (s *Service) ResolveSession(ctx context.Context, token, audience string) (User, string, error) {
	if token == "" || audience == "" {
		return User{}, "", ErrInvalidRequest
	}
	user, err := s.resolve(ctx, token)
	if err != nil {
		return User{}, "", err
	}
	assertion, err := s.assertions.Issue(ctx, port.Actor{Subject: user.ID, Email: user.Email, Audience: audience})
	if err != nil {
		return User{}, "", err
	}
	return user, assertion, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidRequest
	}
	if err := s.kratos.Logout(ctx, token); err != nil {
		return ErrInvalidSession
	}
	return nil
}
func (s *Service) HydraLogin(ctx context.Context, challenge, token string) (string, error) {
	subject, skip, err := s.hydra.LoginRequest(ctx, challenge)
	if err != nil {
		return "", err
	}
	if skip {
		return s.hydra.AcceptLogin(ctx, challenge, subject)
	}
	if token == "" {
		return "", ErrNotAuthenticated
	}
	user, err := s.resolve(ctx, token)
	if err != nil {
		return "", err
	}
	return s.hydra.AcceptLogin(ctx, challenge, user.ID)
}
func (s *Service) HydraConsent(ctx context.Context, challenge string) (string, error) {
	scopes, err := s.hydra.ConsentScopes(ctx, challenge)
	if err != nil {
		return "", err
	}
	return s.hydra.AcceptConsent(ctx, challenge, scopes)
}
func (s *Service) HydraLogout(ctx context.Context, challenge string) (string, error) {
	return s.hydra.AcceptLogout(ctx, challenge)
}
func (s *Service) resolve(ctx context.Context, token string) (User, error) {
	id, email, err := s.kratos.Whoami(ctx, token)
	if err != nil {
		return User{}, ErrInvalidSession
	}
	return User{ID: id, Email: email}, nil
}
