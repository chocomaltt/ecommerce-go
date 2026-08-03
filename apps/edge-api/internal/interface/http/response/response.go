// Package response standardizes the HTTP response envelope.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{Code: http.StatusCreated, Message: message, Data: data})
}

func BadRequest(c *gin.Context, message string, err error) {
	writeError(c, http.StatusBadRequest, message, err)
}

func Unauthorized(c *gin.Context, message string, err error) {
	writeError(c, http.StatusUnauthorized, message, err)
}

func InternalServerError(c *gin.Context, err error) {
	writeError(c, http.StatusInternalServerError, "internal server error", err)
}

func writeError(c *gin.Context, code int, message string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	c.JSON(code, gin.H{"code": code, "message": message, "error": detail})
}
