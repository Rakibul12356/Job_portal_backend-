package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("RequestID", reqID)
		c.Header("X-Request-ID", reqID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		userID, exists := c.Get("userId")
		userIDStr := "guest"
		if exists {
			userIDStr = userID.(string)
		}

		log.Printf("[API-LOG] ID:%s | %s %s | Status:%d | Latency:%v | User:%s",
			reqID, method, path, status, latency, userIDStr)
	}
}
