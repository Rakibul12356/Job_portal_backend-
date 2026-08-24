package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic details with stack trace
				log.Printf("[PANIC RECOVERED] %v\nStack: %s", err, debug.Stack())

				appErr := appErrors.NewInternalError("An unexpected internal error occurred")
				response.Error(c, appErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}
