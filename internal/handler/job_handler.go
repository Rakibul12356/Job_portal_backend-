package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/pagination"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobHandler struct {
	jobService service.JobService
}

func NewJobHandler(jobService service.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

func (h *JobHandler) ListPublicJobs(c *gin.Context) {
	page, limit := pagination.GetParams(c)
	sort := c.DefaultQuery("sort", "newest")

	filter := h.buildJobFilter(c)

	jobs, total, err := h.jobService.ListPublicJobs(c.Request.Context(), filter, sort, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := pagination.NewPaginatedResult(jobs, total, page, limit)
	response.OK(c, result)
}

func (h *JobHandler) GetJobByID(c *gin.Context) {
	idStr := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	job, err := h.jobService.GetJobByID(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, job)
}

func (h *JobHandler) GetSimilarJobs(c *gin.Context) {
	idStr := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	jobs, err := h.jobService.GetSimilarJobs(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, jobs)
}

func (h *JobHandler) ReportJob(c *gin.Context) {
	// Dummy endpoint as per BACKEND.md section 9.3
	response.JSON(c, http.StatusOK, "Job reported successfully", nil)
}

func (h *JobHandler) ListCompanyJobs(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can manage company jobs"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	page, limit := pagination.GetParams(c)
	sort := c.DefaultQuery("sort", "newest")

	filter := bson.M{}
	if q := c.Query("q"); q != "" {
		filter["title"] = bson.M{"$regex": q, "$options": "i"}
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	jobs, total, err := h.jobService.ListCompanyJobs(c.Request.Context(), companyID, filter, sort, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := pagination.NewPaginatedResult(jobs, total, page, limit)
	response.OK(c, result)
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can post jobs"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	var input dto.CreateJobDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.jobService.CreateJob(c.Request.Context(), companyID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, "Job listed successfully", res)
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can update jobs"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	idStr := c.Param("id")
	jobID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	var input dto.UpdateJobDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.jobService.UpdateJob(c.Request.Context(), companyID, jobID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can delete jobs"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	idStr := c.Param("id")
	jobID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	err = h.jobService.DeleteJob(c.Request.Context(), companyID, jobID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

func (h *JobHandler) PublishJob(c *gin.Context) {
	h.setStatus(c, "publish")
}

func (h *JobHandler) CloseJob(c *gin.Context) {
	h.setStatus(c, "close")
}

func (h *JobHandler) ReactivateJob(c *gin.Context) {
	h.setStatus(c, "reactivate")
}

func (h *JobHandler) BulkAction(c *gin.Context) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can perform bulk actions"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	var input dto.BulkJobActionDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.handleValidationError(c, err)
		return
	}

	count, err := h.jobService.BulkAction(c.Request.Context(), companyID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Bulk action executed successfully", gin.H{
		"affectedCount": count,
	})
}

func (h *JobHandler) setStatus(c *gin.Context, action string) {
	companyIDStr, exists := c.Get("companyId")
	if !exists {
		response.Error(c, appErrors.NewForbiddenError("Forbidden: Only employer accounts can manage jobs status"))
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewForbiddenError("Invalid company identity"))
		return
	}

	idStr := c.Param("id")
	jobID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid Job ID format"))
		return
	}

	var serviceErr error
	switch action {
	case "publish":
		serviceErr = h.jobService.PublishJob(c.Request.Context(), companyID, jobID)
	case "close":
		serviceErr = h.jobService.CloseJob(c.Request.Context(), companyID, jobID)
	case "reactivate":
		serviceErr = h.jobService.ReactivateJob(c.Request.Context(), companyID, jobID)
	}

	if serviceErr != nil {
		response.Error(c, serviceErr)
		return
	}

	response.JSON(c, http.StatusOK, "Job status updated successfully", nil)
}

func (h *JobHandler) buildJobFilter(c *gin.Context) bson.M {
	filter := bson.M{}

	if q := c.Query("q"); q != "" {
		// Use MongoDB Text Search index, which checks title, description, skills
		filter["$text"] = bson.M{"$search": q}
	}

	if category := c.Query("category"); category != "" {
		filter["category"] = category
	}

	if location := c.Query("location"); location != "" {
		filter["location"] = bson.M{"$regex": location, "$options": "i"}
	}

	parseList := func(param string) []string {
		vals := c.QueryArray(param)
		if len(vals) == 0 {
			val := c.Query(param)
			if val != "" {
				vals = strings.Split(val, ",")
			}
		}
		var clean []string
		for _, v := range vals {
			if strings.TrimSpace(v) != "" {
				clean = append(clean, strings.TrimSpace(v))
			}
		}
		return clean
	}

	if jobTypes := parseList("jobType"); len(jobTypes) > 0 {
		filter["jobType"] = bson.M{"$in": jobTypes}
	}

	if workModes := parseList("workMode"); len(workModes) > 0 {
		filter["workMode"] = bson.M{"$in": workModes}
	}

	if levels := parseList("experienceLevel"); len(levels) > 0 {
		filter["experienceLevel"] = bson.M{"$in": levels}
	}

	if skills := parseList("skills"); len(skills) > 0 {
		filter["skills"] = bson.M{"$in": skills}
	}

	salaryMinStr := c.Query("salaryMin")
	salaryMaxStr := c.Query("salaryMax")

	salaryQuery := bson.M{}
	if salaryMinStr != "" {
		if val, err := strconv.Atoi(salaryMinStr); err == nil {
			salaryQuery["$gte"] = val
		}
	}
	if salaryMaxStr != "" {
		if val, err := strconv.Atoi(salaryMaxStr); err == nil {
			salaryQuery["$lte"] = val
		}
	}
	if len(salaryQuery) > 0 {
		// Filter jobs where salary range intersects or falls within criteria
		// We can filter by: jobs' salaryMin is >= requested min, or salaryMax is <= requested max,
		// or matching against the salaryMin/Max keys.
		// Let's check salaryMin/salaryMax in the DB.
		if val, exists := salaryQuery["$gte"]; exists {
			filter["salaryMin"] = bson.M{"$gte": val}
		}
		if val, exists := salaryQuery["$lte"]; exists {
			filter["salaryMax"] = bson.M{"$lte": val}
		}
	}

	return filter
}

func (h *JobHandler) handleValidationError(c *gin.Context, err error) {
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
