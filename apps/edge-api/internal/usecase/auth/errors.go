package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotAuthenticated   = errors.New("not authenticated")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidRequest     = errors.New("invalid request")
)
