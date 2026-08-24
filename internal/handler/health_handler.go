package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rakib/job-portal-api/internal/database"
	"github.com/rakib/job-portal-api/internal/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, "API is healthy", gin.H{
		"status": "UP",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if database.MongoClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"status":  "DOWN",
			"error":   "Database connection is nil",
		})
		return
	}

	err := database.MongoClient.Ping(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"status":  "DOWN",
			"error":   err.Error(),
		})
		return
	}

	response.JSON(c, http.StatusOK, "Database connection is ready", gin.H{
		"status": "UP",
		"db":     "MongoDB Atlas connected",
	})
}
