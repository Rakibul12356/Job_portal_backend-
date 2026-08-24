package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, appErrors.NewUnauthorizedError("Authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, appErrors.NewUnauthorizedError("Authorization header format must be Bearer <token>"))
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := utils.VerifyAccessToken(tokenStr)
		if err != nil {
			response.Error(c, appErrors.NewUnauthorizedError("Invalid or expired access token"))
			c.Abort()
			return
		}

		// Save claims to context
		c.Set("userId", claims.Subject)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("name", claims.Name)
		c.Set("firstName", claims.FirstName)
		if claims.CompanyID != nil {
			c.Set("companyId", *claims.CompanyID)
		}

		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenStr := parts[1]
		claims, err := utils.VerifyAccessToken(tokenStr)
		if err == nil {
			c.Set("userId", claims.Subject)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("name", claims.Name)
			c.Set("firstName", claims.FirstName)
			if claims.CompanyID != nil {
				c.Set("companyId", *claims.CompanyID)
			}
		}

		c.Next()
	}
}
