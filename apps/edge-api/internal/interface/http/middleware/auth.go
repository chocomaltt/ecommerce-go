package middleware

import (
	"strings"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/response"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	"github.com/gin-gonic/gin"
)

const sessionKey = "session"

type AuthMiddleware struct {
	identity port.IdentityService
	audience string
}

func NewAuthMiddleware(identity port.IdentityService, audience string) *AuthMiddleware {
	return &AuthMiddleware{
		identity: identity,
		audience: audience,
	}
}

func (m *AuthMiddleware) EnsureAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearer(c.GetHeader("Authorization"))
		if !ok {
			response.Unauthorized(c, "unauthorized", nil)
			c.Abort()
			return
		}

		session, err := m.identity.ResolveSession(c.Request.Context(), token, m.audience)
		if err != nil {
			response.Unauthorized(c, "unauthorized", nil)
			c.Abort()
			return
		}

		c.Set(sessionKey, session)
		c.Next()
	}
}

func Session(c *gin.Context) (port.Session, bool) {
	value, ok := c.Get(sessionKey)
	if !ok {
		return port.Session{}, false
	}

	session, ok := value.(port.Session)
	return session, ok
}

func Token(c *gin.Context) (string, bool) {
	return bearer(c.GetHeader("Authorization"))
}

func bearer(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", false
	}

	return fields[1], fields[1] != ""
}
