package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/response"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/usecase/auth"
)

// AuthMiddleware validates Kratos session tokens for protected endpoints.
type AuthMiddleware struct {
	kratos port.KratosService
}

func NewAuthMiddleware(kratos port.KratosService) *AuthMiddleware {
	return &AuthMiddleware{kratos: kratos}
}

// EnsureAuthenticated resolves the Bearer token against Kratos and stores the
// identity in the context as "user".
func (m *AuthMiddleware) EnsureAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logrus.Warn("no auth header")
			response.Unauthorized(c, "unauthorized", auth.ErrNotAuthenticated)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			logrus.Warn("invalid auth header")
			response.Unauthorized(c, "unauthorized", auth.ErrInvalidSession)
			c.Abort()
			return
		}

		token := strings.TrimSpace(parts[1])

		id, email, err := m.kratos.Whoami(c.Request.Context(), token)
		if err != nil {
			logrus.WithError(err).Warn("invalid session token")
			response.Unauthorized(c, "unauthorized", auth.ErrInvalidSession)
			c.Abort()
			return
		}

		c.Set("user", &auth.User{ID: id, Email: email})
		c.Set("accessToken", token)
		c.Next()
	}
}
