package identity

import (
	"context"
	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/port"
)

type contextKey struct{}
type Context struct{}

func (Context) WithActor(ctx context.Context, actor port.Actor) context.Context {
	return context.WithValue(ctx, contextKey{}, actor)
}
func (Context) Actor(ctx context.Context) (port.Actor, bool) {
	actor, ok := ctx.Value(contextKey{}).(port.Actor)
	return actor, ok
}
