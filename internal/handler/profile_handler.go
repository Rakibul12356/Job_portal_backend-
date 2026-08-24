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

type ProfileHandler struct {
	profileService service.ProfileService
}

func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) GetProfileMe(c *gin.Context) {
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

	profile, err := h.profileService.GetProfileByUserID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

func (h *ProfileHandler) UpdateProfileMe(c *gin.Context) {
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

	var input dto.UpdateProfileDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	profile, err := h.profileService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
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

	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Avatar image file is required"))
		return
	}

	avatarURL, err := h.profileService.UploadAvatar(c.Request.Context(), userID, file)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Avatar uploaded successfully", gin.H{
		"avatarUrl": avatarURL,
	})
}

func (h *ProfileHandler) RemoveAvatar(c *gin.Context) {
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

	err = h.profileService.RemoveAvatar(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Avatar removed successfully", nil)
}

func (h *ProfileHandler) UploadResume(c *gin.Context) {
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

	file, err := c.FormFile("resume")
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Resume PDF file is required"))
		return
	}

	meta, err := h.profileService.UploadResume(c.Request.Context(), userID, file)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Resume uploaded successfully", meta)
}

func (h *ProfileHandler) RemoveResume(c *gin.Context) {
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

	err = h.profileService.RemoveResume(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Resume removed successfully", nil)
}

func (h *ProfileHandler) AddExperience(c *gin.Context) {
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

	var input dto.ExperienceDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	exp, err := h.profileService.AddExperience(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Experience added successfully", exp)
}

func (h *ProfileHandler) UpdateExperience(c *gin.Context) {
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

	expID := c.Param("expId")
	if expID == "" {
		response.Error(c, appErrors.NewValidationError("Experience ID is required"))
		return
	}

	var input dto.ExperienceDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err = h.profileService.UpdateExperience(c.Request.Context(), userID, expID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Experience updated successfully", nil)
}

func (h *ProfileHandler) DeleteExperience(c *gin.Context) {
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

	expID := c.Param("expId")
	if expID == "" {
		response.Error(c, appErrors.NewValidationError("Experience ID is required"))
		return
	}

	err = h.profileService.DeleteExperience(c.Request.Context(), userID, expID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

func (h *ProfileHandler) AddEducation(c *gin.Context) {
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

	var input dto.EducationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	edu, err := h.profileService.AddEducation(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Education added successfully", edu)
}

func (h *ProfileHandler) UpdateEducation(c *gin.Context) {
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

	eduID := c.Param("eduId")
	if eduID == "" {
		response.Error(c, appErrors.NewValidationError("Education ID is required"))
		return
	}

	var input dto.EducationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err = h.profileService.UpdateEducation(c.Request.Context(), userID, eduID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Education updated successfully", nil)
}

func (h *ProfileHandler) DeleteEducation(c *gin.Context) {
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

	eduID := c.Param("eduId")
	if eduID == "" {
		response.Error(c, appErrors.NewValidationError("Education ID is required"))
		return
	}

	err = h.profileService.DeleteEducation(c.Request.Context(), userID, eduID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

func (h *ProfileHandler) handleValidationError(c *gin.Context, err error) {
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
