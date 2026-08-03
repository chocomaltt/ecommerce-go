package port

import "context"

type User struct {
	ID    string
	Email string
}
type Authentication struct {
	User         User
	SessionToken string
}
type Session struct {
	User           User
	ActorAssertion string
}
type IdentityService interface {
	Register(context.Context, string, string) (Authentication, error)
	Login(context.Context, string, string) (Authentication, error)
	ResolveSession(context.Context, string, string) (Session, error)
	Logout(context.Context, string) error
}
