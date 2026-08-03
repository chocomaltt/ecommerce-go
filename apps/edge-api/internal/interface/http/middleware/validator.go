// Package middleware provides generic HTTP middlewares (validation).
package middleware

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"github.com/chocomaltt/ecommerce-go/apps/edge-api/internal/interface/http/response"
)

var Validator = validator.New()

type CustomValidatable interface {
	Validate() error
}

func init() {
	Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}
		return name
	})
}

// EnsureJsonValidRequest binds and validates a JSON body, storing it in the
// context under "body" for the handler to consume via util.GetBody[T].
func EnsureJsonValidRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		body := new(T)

		if err := c.ShouldBindJSON(body); err != nil {
			logrus.WithError(err).Warn("failed to bind JSON")
			response.BadRequest(c, "invalid request", err)
			c.Abort()
			return
		}

		if err := Validator.Struct(body); err != nil {
			var errStr string
			for i, e := range err.(validator.ValidationErrors) {
				if i > 0 {
					errStr += ", "
				}
				errStr += fmt.Sprintf("%s %s", e.Field(), e.Tag())
			}
			logrus.Warn("validation errors: ", errStr)
			response.BadRequest(c, "invalid request", errors.New(errStr))
			c.Abort()
			return
		}

		if custom, ok := any(body).(CustomValidatable); ok {
			if err := custom.Validate(); err != nil {
				logrus.WithError(err).Warn("custom validation failed")
				response.BadRequest(c, "invalid request", err)
				c.Abort()
				return
			}
		}

		c.Set("body", body)
		c.Next()
	}
}
