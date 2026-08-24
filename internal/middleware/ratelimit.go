package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rakib/job-portal-api/internal/config"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
)

type clientLimit struct {
	requests  int
	lastReset time.Time
}

var (
	clients = make(map[string]*clientLimit)
	mu      sync.Mutex
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limitRPM := config.AppConfig.RateLimitRPM
		if limitRPM <= 0 {
			c.Next()
			return
		}

		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		client, exists := clients[ip]
		if !exists || now.Sub(client.lastReset) > time.Minute {
			clients[ip] = &clientLimit{
				requests:  1,
				lastReset: now,
			}
			mu.Unlock()
			c.Next()
			return
		}

		if client.requests >= limitRPM {
			mu.Unlock()
			response.Error(c, appErrors.NewTooManyRequestsError("Rate limit exceeded. Please try again in a minute."))
			c.Abort()
			return
		}

		client.requests++
		mu.Unlock()
		c.Next()
	}
}
