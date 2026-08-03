package util

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetBody retrieves the validated body stored in the context by the
// EnsureJsonValidRequest middleware.
func GetBody[T any](c *gin.Context, key string) (T, error) {
	var zero T
	raw, ok := c.Get(key)
	if !ok {
		return zero, fmt.Errorf("%s not found in context", key)
	}
	body, ok := raw.(*T)
	if !ok {
		return zero, fmt.Errorf("invalid %s type in context", key)
	}
	return *body, nil
}
