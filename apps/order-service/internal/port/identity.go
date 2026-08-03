package port

import "context"

type Actor struct {
	UserID        string
	Email         string
	CallerService string
}
type ActorVerifier interface{ Verify(string) (Actor, error) }
type Context interface {
	WithActor(context.Context, Actor) context.Context
	Actor(context.Context) (Actor, bool)
}
