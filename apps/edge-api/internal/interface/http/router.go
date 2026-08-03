// Package httpinterface wires Edge HTTP routes.
package httpinterface

import (
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/handler"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(identity port.IdentityService, authMW *middleware.AuthMiddleware, orders port.OrderService) *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "tu tu du du max verstappen"}) })
	h := handler.NewAuthHandler(identity, orders)
	g := r.Group("/auth")
	g.POST("/register", middleware.EnsureJsonValidRequest[handler.RegisterRequest](), h.Register)
	g.POST("/login", middleware.EnsureJsonValidRequest[handler.LoginRequest](), h.Login)
	g.POST("/logout", authMW.EnsureAuthenticated(), h.Logout)
	g.GET("/me", authMW.EnsureAuthenticated(), h.Me)
	g.GET("/order-caller", authMW.EnsureAuthenticated(), h.OrderCaller)
	return r
}
