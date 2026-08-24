package middleware

import (
	"github.com/gin-gonic/gin"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
			c.Abort()
			return
		}

		userRole := roleVal.(string)
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		response.Error(c, appErrors.NewForbiddenError("Forbidden: Insufficient privileges for this role"))
		c.Abort()
	}
}
