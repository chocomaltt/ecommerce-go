// Package httpinterface wires Edge HTTP routes.
package httpinterface

import (
	"net/http"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/handler"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	"github.com/gin-gonic/gin"
)

func New(identity port.IdentityService, authMiddleware *middleware.AuthMiddleware, orders port.OrderService) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "tu tu du du max verstappen"})
	})

	authHandler := handler.NewAuthHandler(identity, orders)
	auth := router.Group("/auth")
	auth.POST("/register", middleware.EnsureJsonValidRequest[handler.RegisterRequest](), authHandler.Register)
	auth.POST("/login", middleware.EnsureJsonValidRequest[handler.LoginRequest](), authHandler.Login)
	auth.POST("/logout", authMiddleware.EnsureAuthenticated(), authHandler.Logout)
	auth.GET("/me", authMiddleware.EnsureAuthenticated(), authHandler.Me)
	auth.GET("/order-caller", authMiddleware.EnsureAuthenticated(), authHandler.OrderCaller)

	return router
}
