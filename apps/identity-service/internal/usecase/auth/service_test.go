package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/chocomaltt/ecommerce-go/apps/identity-service/internal/port"
)

type fakeKratos struct {
	registerErr error
	loginErr    error
	whoamiErr   error
	logoutErr   error
}

func (f fakeKratos) Register(context.Context, string, string) error {
	return f.registerErr
}

func (f fakeKratos) Login(context.Context, string, string) (string, error) {
	return "session", f.loginErr
}

func (f fakeKratos) Whoami(context.Context, string) (string, string, error) {
	return "user-123", "user@example.com", f.whoamiErr
}

func (f fakeKratos) Logout(context.Context, string) error {
	return f.logoutErr
}

type fakeIssuer struct {
	actor port.Actor
}

func (f *fakeIssuer) Issue(_ context.Context, actor port.Actor) (string, error) {
	f.actor = actor
	return "assertion", nil
}

func TestResolveSessionIssuesTargetBoundActor(t *testing.T) {
	issuer := &fakeIssuer{}
	service := New(fakeKratos{}, issuer)

	user, assertion, err := service.ResolveSession(context.Background(), "session", "order-service")
	if err != nil {
		t.Fatal(err)
	}

	if user.ID != "user-123" || assertion != "assertion" {
		t.Fatalf("unexpected result: %#v %q", user, assertion)
	}
	if issuer.actor.Subject != user.ID || issuer.actor.Audience != "order-service" {
		t.Fatalf("unexpected actor: %#v", issuer.actor)
	}
}

func TestResolveSessionRejectsInvalidSession(t *testing.T) {
	service := New(fakeKratos{whoamiErr: errors.New("invalid")}, &fakeIssuer{})

	_, _, err := service.ResolveSession(context.Background(), "session", "order-service")
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("got %v", err)
	}
}
