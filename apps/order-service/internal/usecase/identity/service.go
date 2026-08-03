package identity

import (
	"context"
	"errors"

	"github.com/chocomaltt/ecommerce-go/apps/order-service/internal/port"
)

var ErrMissingActor = errors.New("missing authenticated actor")

type Service struct {
	context port.Context
}

func New(contexts port.Context) *Service {
	return &Service{context: contexts}
}

func (s *Service) Caller(ctx context.Context) (port.Actor, error) {
	actor, ok := s.context.Actor(ctx)
	if !ok {
		return port.Actor{}, ErrMissingActor
	}

	return actor, nil
}
