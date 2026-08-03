package port

import "context"

type KratosService interface {
	Login(context.Context, string, string) (string, error)
	Register(context.Context, string, string) error
	Whoami(context.Context, string) (id, email string, err error)
	Logout(context.Context, string) error
}
