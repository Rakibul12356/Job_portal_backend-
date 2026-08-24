package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedJobHandler struct {
	savedJobService service.SavedJobService
}

func NewSavedJobHandler(savedJobService service.SavedJobService) *SavedJobHandler {
	return &SavedJobHandler{
		savedJobService: savedJobService,
	}
}

func (h *SavedJobHandler) GetSavedJobs(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	jobs, err := h.savedJobService.ListSavedJobs(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, jobs)
}

func (h *SavedJobHandler) SaveJob(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	var input struct {
		JobID string `json:"jobId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	jobID, err := primitive.ObjectIDFromHex(input.JobID)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	err = h.savedJobService.SaveJob(c.Request.Context(), userID, jobID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Job bookmarked successfully", nil)
}

func (h *SavedJobHandler) UnsaveJob(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	jobIDStr := c.Param("jobId")
	jobID, err := primitive.ObjectIDFromHex(jobIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	err = h.savedJobService.UnsaveJob(c.Request.Context(), userID, jobID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

func (h *SavedJobHandler) handleValidationError(c *gin.Context, err error) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		response.Error(c, appErrors.NewValidationError("Invalid JSON request format"))
		return
	}

	var details []appErrors.ErrorDetail
	for _, fieldErr := range validationErrors {
		details = append(details, appErrors.ErrorDetail{
			Field:   fieldErr.Field(),
			Message: fieldErr.ActualTag(),
		})
	}

	response.Error(c, appErrors.NewValidationError("Validation failed", details...))
}
