package handler

import (
	"github.com/gin-gonic/gin"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DashboardHandler struct {
	dashboardService service.DashboardService
}

func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

func (h *DashboardHandler) GetSeekerDashboard(c *gin.Context) {
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

	res, err := h.dashboardService.GetSeekerDashboard(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *DashboardHandler) GetCompanyDashboard(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can access company dashboard"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	res, err := h.dashboardService.GetCompanyDashboard(c.Request.Context(), companyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}
