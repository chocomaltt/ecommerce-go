// Package auth is the domain for authentication: aggregates and invariants.
package auth

// User is the authenticated identity (sourced from Kratos).
type User struct {
	ID    string
	Email string
}
