package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompanyHandler struct {
	companyService service.CompanyService
}

func NewCompanyHandler(companyService service.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
	}
}

func (h *CompanyHandler) GetOwnCompanyProfile(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can view company details"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	profile, err := h.companyService.GetPublicProfile(c.Request.Context(), companyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

func (h *CompanyHandler) GetCompanySettings(c *gin.Context) {
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

	settings, err := h.companyService.GetCompanySettings(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, settings)
}

func (h *CompanyHandler) UpdateCompanySettings(c *gin.Context) {
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

	var input dto.UpdateCompanySettingsDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	settings, err := h.companyService.UpdateCompanySettings(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, settings)
}

func (h *CompanyHandler) UploadLogo(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can upload logo"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	file, err := c.FormFile("logo")
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Logo image file is required"))
		return
	}

	logoURL, err := h.companyService.UploadLogo(c.Request.Context(), companyID, file)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Logo uploaded successfully", gin.H{
		"logoUrl": logoURL,
	})
}

func (h *CompanyHandler) RemoveLogo(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can remove logo"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	err = h.companyService.RemoveLogo(c.Request.Context(), companyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Logo removed successfully", nil)
}

func (h *CompanyHandler) GetPublicCompanyProfile(c *gin.Context) {
	idStr := c.Param("id")
	companyID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid company ID format"))
		return
	}

	profile, err := h.companyService.GetPublicProfile(c.Request.Context(), companyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

func (h *CompanyHandler) ListCompanies(c *gin.Context) {
	companies, err := h.companyService.ListCompanies(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, companies)
}

func (h *CompanyHandler) handleValidationError(c *gin.Context, err error) {
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
