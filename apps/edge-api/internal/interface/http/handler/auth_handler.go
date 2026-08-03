// Package handler maps Edge HTTP authentication endpoints to identity ports.
package handler

import (
	"net/http"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/middleware"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/response"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/port"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/shared/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
type authResponse struct {
	User         userResponse `json:"user"`
	SessionToken string       `json:"session_token"`
}
type AuthHandler struct {
	identity port.IdentityService
	orders   port.OrderService
}

func NewAuthHandler(identity port.IdentityService, orders port.OrderService) *AuthHandler {
	return &AuthHandler{identity: identity, orders: orders}
}
func (h *AuthHandler) Register(c *gin.Context) {
	req, err := util.GetBody[RegisterRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}
	result, err := h.identity.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.BadRequest(c, "registration failed", err)
		return
	}
	response.Created(c, "user registered successfully", authResponse{User: toUser(result.User), SessionToken: result.SessionToken})
}
func (h *AuthHandler) Login(c *gin.Context) {
	req, err := util.GetBody[LoginRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}
	result, err := h.identity.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, "unauthorized", err)
		return
	}
	response.OK(c, "user logged in successfully", authResponse{User: toUser(result.User), SessionToken: result.SessionToken})
}
func (h *AuthHandler) Logout(c *gin.Context) {
	token, ok := middleware.Token(c)
	if !ok {
		response.Unauthorized(c, "unauthorized", nil)
		return
	}
	if err := h.identity.Logout(c.Request.Context(), token); err != nil {
		response.Unauthorized(c, "unauthorized", nil)
		return
	}
	response.OK(c, "user logged out successfully", nil)
}
func (h *AuthHandler) Me(c *gin.Context) {
	session, ok := middleware.Session(c)
	if !ok {
		response.Unauthorized(c, "unauthorized", nil)
		return
	}
	response.OK(c, "user fetched successfully", toUser(session.User))
}
func (h *AuthHandler) OrderCaller(c *gin.Context) {
	session, ok := middleware.Session(c)
	if !ok {
		response.Unauthorized(c, "unauthorized", nil)
		return
	}
	caller, err := h.orders.GetCaller(c.Request.Context(), session.ActorAssertion)
	if err != nil {
		logrus.WithError(err).Error("order identity")
		response.InternalServerError(c, err)
		return
	}
	response.OK(c, "authenticated gRPC caller", gin.H{"user_id": caller.UserID, "email": caller.Email, "caller_service": caller.Service})
}
func (h *AuthHandler) HydraLogin(c *gin.Context) {
	challenge := c.Query("login_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", nil)
		return
	}
	token, _ := middleware.Token(c)
	redirect, err := h.identity.HydraLogin(c.Request.Context(), challenge, token)
	if err != nil {
		response.Unauthorized(c, "unauthorized", nil)
		return
	}
	c.Redirect(http.StatusFound, redirect)
}
func (h *AuthHandler) HydraConsent(c *gin.Context) {
	challenge := c.Query("consent_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", nil)
		return
	}
	redirect, err := h.identity.HydraConsent(c.Request.Context(), challenge)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	c.Redirect(http.StatusFound, redirect)
}
func (h *AuthHandler) HydraLogout(c *gin.Context) {
	challenge := c.Query("logout_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", nil)
		return
	}
	redirect, err := h.identity.HydraLogout(c.Request.Context(), challenge)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	c.Redirect(http.StatusFound, redirect)
}
func toUser(user port.User) userResponse { return userResponse{ID: user.ID, Email: user.Email} }
