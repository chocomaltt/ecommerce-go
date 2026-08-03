// Package handler contains HTTP handlers for the auth domain.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/response"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/shared/util"
	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/usecase/auth"
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
	User         *userResponse `json:"user"`
	SessionToken string        `json:"session_token"`
}

type AuthHandler struct {
	auth *auth.AuthService
}

func NewAuthHandler(auth *auth.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	req, err := util.GetBody[RegisterRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, token, err := h.auth.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		logrus.WithError(err).Warn("registration failed")
		response.BadRequest(c, "registration failed", err)
		return
	}

	response.Created(c, "user registered successfully", authResponse{
		User:         toUserResponse(user),
		SessionToken: token,
	})
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	req, err := util.GetBody[LoginRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		logrus.WithError(err).Warn("login failed")
		response.Unauthorized(c, "unauthorized", err)
		return
	}

	response.OK(c, "user logged in successfully", authResponse{
		User:         toUserResponse(user),
		SessionToken: token,
	})
}

// Logout godoc
// @Summary Logout user
// @Tags Auth
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Session tokens are not revocable client-side; the client drops it.
	response.OK(c, "user logged out successfully", nil)
}

// Me godoc
// @Summary Current user
// @Tags Auth
// @Security BearerAuth
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user, err := util.GetBody[auth.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}
	response.OK(c, "user fetched successfully", toUserResponse(&user))
}

// HydraLogin godoc
// @Summary Hydra login callback
// @Tags Auth
// @Router /auth/hydra/login [get]
func (h *AuthHandler) HydraLogin(c *gin.Context) {
	challenge := c.Query("login_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", auth.ErrInvalidRequest)
		return
	}

	redirect, err := h.auth.HydraLogin(c.Request.Context(), challenge, bearerToken(c))
	if err != nil {
		logrus.WithError(err).Error("hydra login")
		if err == auth.ErrNotAuthenticated || err == auth.ErrInvalidSession {
			response.Unauthorized(c, "unauthorized", err)
			return
		}
		response.InternalServerError(c, err)
		return
	}

	c.Redirect(http.StatusFound, redirect)
}

// HydraConsent godoc
// @Summary Hydra consent callback
// @Tags Auth
// @Router /auth/hydra/consent [get]
func (h *AuthHandler) HydraConsent(c *gin.Context) {
	challenge := c.Query("consent_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", auth.ErrInvalidRequest)
		return
	}

	redirect, err := h.auth.HydraConsent(c.Request.Context(), challenge)
	if err != nil {
		logrus.WithError(err).Error("hydra consent")
		response.InternalServerError(c, err)
		return
	}

	c.Redirect(http.StatusFound, redirect)
}

// HydraLogout godoc
// @Summary Hydra logout callback
// @Tags Auth
// @Router /auth/hydra/logout [get]
func (h *AuthHandler) HydraLogout(c *gin.Context) {
	challenge := c.Query("logout_challenge")
	if challenge == "" {
		response.BadRequest(c, "invalid request", auth.ErrInvalidRequest)
		return
	}

	redirect, err := h.auth.HydraLogout(c.Request.Context(), challenge)
	if err != nil {
		logrus.WithError(err).Error("hydra logout")
		response.InternalServerError(c, err)
		return
	}

	c.Redirect(http.StatusFound, redirect)
}

func toUserResponse(u *auth.User) *userResponse {
	return &userResponse{ID: u.ID, Email: u.Email}
}

func bearerToken(c *gin.Context) string {
	raw, ok := c.Get("accessToken")
	if !ok {
		return ""
	}
	token, _ := raw.(string)
	return token
}
