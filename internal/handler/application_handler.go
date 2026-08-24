package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rakib/job-portal-api/internal/config"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/pagination"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApplicationHandler struct {
	appService service.ApplicationService
}

func NewApplicationHandler(appService service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
	}
}

func (h *ApplicationHandler) ApplyToJob(c *gin.Context) {
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

	jobIDStr := c.Param("id")
	jobID, err := primitive.ObjectIDFromHex(jobIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	coverMessage := c.PostForm("coverMessage")
	if len(coverMessage) > 500 {
		response.Error(c, appErrors.NewValidationError("Cover message must not exceed 500 characters"))
		return
	}

	file, err := c.FormFile("resume")
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Resume PDF file is required"))
		return
	}

	res, err := h.appService.ApplyToJob(c.Request.Context(), userID, jobID, file, coverMessage)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Applied successfully", res)
}

func (h *ApplicationHandler) ListSeekerApplications(c *gin.Context) {
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

	page, limit := pagination.GetParams(c)

	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	dateRange := c.Query("date")
	if dateRange != "" && dateRange != "all" {
		var d time.Duration
		switch dateRange {
		case "7d":
			d = 7 * 24 * time.Hour
		case "30d":
			d = 30 * 24 * time.Hour
		case "3m":
			d = 90 * 24 * time.Hour
		default:
			d = 0
		}
		if d > 0 {
			filter["appliedAt"] = bson.M{"$gte": time.Now().Add(-d)}
		}
	}

	apps, total, err := h.appService.ListSeekerApplications(c.Request.Context(), userID, filter, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := pagination.NewPaginatedResult(apps, total, page, limit)
	response.OK(c, result)
}

func (h *ApplicationHandler) GetSeekerApplicationByID(c *gin.Context) {
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

	appIDStr := c.Param("id")
	appID, err := primitive.ObjectIDFromHex(appIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid application ID format"))
		return
	}

	res, err := h.appService.GetSeekerApplicationByID(c.Request.Context(), userID, appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *ApplicationHandler) WithdrawApplication(c *gin.Context) {
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

	appIDStr := c.Param("id")
	appID, err := primitive.ObjectIDFromHex(appIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid application ID format"))
		return
	}

	err = h.appService.WithdrawApplication(c.Request.Context(), userID, appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Application withdrawn successfully", nil)
}

func (h *ApplicationHandler) ListCompanyApplicants(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can manage applicants"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	page, limit := pagination.GetParams(c)

	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if jobIDStr := c.Query("jobId"); jobIDStr != "" {
		if joid, err := primitive.ObjectIDFromHex(jobIDStr); err == nil {
			filter["jobId"] = joid
		}
	}

	dateRange := c.Query("date")
	if dateRange != "" && dateRange != "all" {
		var d time.Duration
		switch dateRange {
		case "24h":
			d = 24 * time.Hour
		case "7d":
			d = 7 * 24 * time.Hour
		case "30d":
			d = 30 * 24 * time.Hour
		default:
			d = 0
		}
		if d > 0 {
			filter["appliedAt"] = bson.M{"$gte": time.Now().Add(-d)}
		}
	}

	expFilter := c.Query("experienceLevel")

	apps, total, err := h.appService.ListCompanyApplicants(c.Request.Context(), companyID, filter, expFilter, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := pagination.NewPaginatedResult(apps, total, page, limit)
	response.OK(c, result)
}

func (h *ApplicationHandler) GetCompanyApplicantByID(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can view applicant details"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	appIDStr := c.Param("id")
	appID, err := primitive.ObjectIDFromHex(appIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid application ID format"))
		return
	}

	res, err := h.appService.GetCompanyApplicantByID(c.Request.Context(), companyID, appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *ApplicationHandler) UpdateApplicantStatus(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can review applicants"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	appIDStr := c.Param("id")
	appID, err := primitive.ObjectIDFromHex(appIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid application ID format"))
		return
	}

	var input dto.UpdateApplicationStatusDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err = h.appService.UpdateApplicantStatus(c.Request.Context(), companyID, appID, input.Status)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Applicant status updated successfully", nil)
}

func (h *ApplicationHandler) DownloadResume(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can download resumes"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	appIDStr := c.Param("id")
	appID, err := primitive.ObjectIDFromHex(appIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid application ID format"))
		return
	}

	res, err := h.appService.GetCompanyApplicantByID(c.Request.Context(), companyID, appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Read local file path and write as download stream if local file
	if strings.Contains(res.ResumeURL, "/uploads/") {
		parts := strings.Split(res.ResumeURL, "/uploads/")
		localPath := filepath.Join(config.AppConfig.UploadDir, parts[1])

		file, err := os.Open(localPath)
		if err != nil {
			response.Error(c, appErrors.NewNotFoundError("Resume file not found on disk"))
			return
		}
		defer file.Close()

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", res.ResumeFilename))
		c.Header("Content-Type", "application/pdf")
		_, _ = io.Copy(c.Writer, file)
		return
	}

	// Fallback redirect
	c.Redirect(http.StatusFound, res.ResumeURL)
}

func (h *ApplicationHandler) handleValidationError(c *gin.Context, err error) {
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
