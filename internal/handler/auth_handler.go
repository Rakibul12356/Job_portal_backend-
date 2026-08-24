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

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterSeeker(c *gin.Context) {
	var input dto.RegisterSeekerDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err := h.authService.RegisterSeeker(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Registration successful", nil)
}

func (h *AuthHandler) RegisterEmployer(c *gin.Context) {
	var input dto.RegisterEmployerDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err := h.authService.RegisterEmployer(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Registration successful", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Set refresh token in HttpOnly cookie
	c.SetCookie("refreshToken", res.RefreshToken, 60*60*24*7, "/", "", false, true)

	response.JSON(c, http.StatusOK, "Login successful", res)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input dto.RefreshDTO
	var refreshToken string

	// 1. Try reading from JSON body
	if err := c.ShouldBindJSON(&input); err == nil && input.RefreshToken != "" {
		refreshToken = input.RefreshToken
	} else {
		// 2. Fallback to HttpOnly cookie
		cookieToken, err := c.Cookie("refreshToken")
		if err != nil {
			response.Error(c, appErrors.NewUnauthorizedError("Refresh token is required"))
			return
		}
		refreshToken = cookieToken
	}

	newAccess, newRefresh, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.SetCookie("refreshToken", newRefresh, 60*60*24*7, "/", "", false, true)

	response.OK(c, gin.H{
		"accessToken":  newAccess,
		"refreshToken": newRefresh,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the refresh token cookie
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	response.JSON(c, http.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
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

	res, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) handleValidationError(c *gin.Context, err error) {
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
