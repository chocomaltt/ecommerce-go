// Package httpinterface wires routes. Route definitions live here;
// handlers and middleware live in their own packages.
package httpinterface

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/handler"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/usecase/auth"
)

// New builds the HTTP router.
func New(authService *auth.AuthService, authMW *middleware.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "tu tu du du max verstappen"})
	})

	h := handler.NewAuthHandler(authService)

	g := r.Group("/auth")
	g.POST("/register", middleware.EnsureJsonValidRequest[handler.RegisterRequest](), h.Register)
	g.POST("/login", middleware.EnsureJsonValidRequest[handler.LoginRequest](), h.Login)
	g.POST("/logout", authMW.EnsureAuthenticated(), h.Logout)
	g.GET("/me", authMW.EnsureAuthenticated(), h.Me)

	// Hydra's login/consent app interface (called via redirect from Hydra).
	g.GET("/hydra/login", h.HydraLogin)
	g.GET("/hydra/consent", h.HydraConsent)
	g.GET("/hydra/logout", h.HydraLogout)

	return r
}
