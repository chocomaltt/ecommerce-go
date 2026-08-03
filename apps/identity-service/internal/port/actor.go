package port

import "context"

type Actor struct {
	Subject  string
	Email    string
	Audience string
}

type ActorIssuer interface {
	Issue(context.Context, Actor) (string, error)
}
