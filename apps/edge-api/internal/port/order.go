package port

import "context"

type Caller struct {
	UserID  string
	Email   string
	Service string
}

type OrderService interface {
	GetCaller(context.Context, string) (Caller, error)
}
