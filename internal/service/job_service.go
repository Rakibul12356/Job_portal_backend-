package service

import (
	"context"
	"strings"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobService interface {
	CreateJob(ctx context.Context, companyID primitive.ObjectID, input dto.CreateJobDTO) (*dto.JobResponseDTO, error)
	UpdateJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID, input dto.UpdateJobDTO) (*dto.JobResponseDTO, error)
	DeleteJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error
	GetJobByID(ctx context.Context, jobID primitive.ObjectID) (*dto.JobResponseDTO, error)
	ListPublicJobs(ctx context.Context, query bson.M, sort string, page, limit int) ([]dto.JobResponseDTO, int64, error)
	ListCompanyJobs(ctx context.Context, companyID primitive.ObjectID, query bson.M, sort string, page, limit int) ([]dto.JobResponseDTO, int64, error)
	GetSimilarJobs(ctx context.Context, jobID primitive.ObjectID) ([]dto.JobResponseDTO, error)
	PublishJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error
	CloseJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error
	ReactivateJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error
	BulkAction(ctx context.Context, companyID primitive.ObjectID, input dto.BulkJobActionDTO) (int64, error)
}

type jobService struct {
	jobRepo     repository.JobRepository
	companyRepo repository.CompanyRepository
}

func NewJobService(jobRepo repository.JobRepository, companyRepo repository.CompanyRepository) JobService {
	return &jobService{
		jobRepo:     jobRepo,
		companyRepo: companyRepo,
	}
}

func (s *jobService) CreateJob(ctx context.Context, companyID primitive.ObjectID, input dto.CreateJobDTO) (*dto.JobResponseDTO, error) {
	status := domain.JobStatusActive
	if input.Status != "" {
		status = input.Status
	}

	if status == domain.JobStatusActive {
		if err := s.validatePublishFields(input); err != nil {
			return nil, err
		}
	}

	job := &domain.Job{
		ID:              primitive.NewObjectID(),
		CompanyID:       companyID,
		Title:           input.Title,
		Status:          status,
		JobType:         input.JobType,
		WorkMode:        input.WorkMode,
		Category:        input.Category,
		ExperienceLevel: input.ExperienceLevel,
		Location:        input.Location,
		SalaryMin:       input.SalaryMin,
		SalaryMax:       input.SalaryMax,
		SalaryPeriod:    input.SalaryPeriod,
		Description:     input.Description,
		Requirements:    input.Requirements,
		Benefits:        input.Benefits,
		Skills:          input.Skills,
		Vacancies:       input.Vacancies,
		Deadline:        input.Deadline,
		ApplicantsCount: 0,
	}

	if status == domain.JobStatusActive {
		now := time.Now()
		job.PublishedAt = &now
	}

	err := s.jobRepo.Create(ctx, job)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to save job")
	}

	return s.mapToDTO(ctx, job)
}

func (s *jobService) UpdateJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID, input dto.UpdateJobDTO) (*dto.JobResponseDTO, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Job not found")
	}

	if job.CompanyID != companyID {
		return nil, appErrors.NewForbiddenError("You are not the owner of this job listing")
	}

	if input.Title != "" {
		job.Title = input.Title
	}
	if input.JobType != "" {
		job.JobType = input.JobType
	}
	if input.WorkMode != "" {
		job.WorkMode = input.WorkMode
	}
	if input.Category != "" {
		job.Category = input.Category
	}
	if input.ExperienceLevel != "" {
		job.ExperienceLevel = input.ExperienceLevel
	}
	if input.Location != "" {
		job.Location = input.Location
	}
	if input.SalaryMin != nil {
		job.SalaryMin = input.SalaryMin
	}
	if input.SalaryMax != nil {
		job.SalaryMax = input.SalaryMax
	}
	if input.SalaryPeriod != "" {
		job.SalaryPeriod = input.SalaryPeriod
	}
	if input.Description != "" {
		job.Description = input.Description
	}
	if input.Requirements != "" {
		job.Requirements = input.Requirements
	}
	if input.Benefits != "" {
		job.Benefits = input.Benefits
	}
	if len(input.Skills) > 0 {
		job.Skills = input.Skills
	}
	if input.Vacancies > 0 {
		job.Vacancies = input.Vacancies
	}
	if !input.Deadline.IsZero() {
		job.Deadline = input.Deadline
	}

	if input.Status != "" {
		if input.Status == domain.JobStatusActive && job.Status != domain.JobStatusActive {
			createDTO := dto.CreateJobDTO{
				Title:           job.Title,
				JobType:         job.JobType,
				WorkMode:        job.WorkMode,
				Category:        job.Category,
				ExperienceLevel: job.ExperienceLevel,
				Location:        job.Location,
				Description:     job.Description,
				Requirements:    job.Requirements,
				Skills:          job.Skills,
				Vacancies:       job.Vacancies,
				Deadline:        job.Deadline,
			}
			if err := s.validatePublishFields(createDTO); err != nil {
				return nil, err
			}
			now := time.Now()
			job.PublishedAt = &now
		}
		job.Status = input.Status
	}

	err = s.jobRepo.Update(ctx, job)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to update job")
	}

	return s.mapToDTO(ctx, job)
}

func (s *jobService) DeleteJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return appErrors.NewNotFoundError("Job not found")
	}

	if job.CompanyID != companyID {
		return appErrors.NewForbiddenError("You are not the owner of this job listing")
	}

	err = s.jobRepo.Delete(ctx, jobID)
	if err != nil {
		return appErrors.NewInternalError("Failed to delete job")
	}

	return nil
}

func (s *jobService) GetJobByID(ctx context.Context, jobID primitive.ObjectID) (*dto.JobResponseDTO, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Job not found")
	}

	// Update status dynamically on read if expired
	s.checkAndSetExpired(ctx, job)

	return s.mapToDTO(ctx, job)
}

func (s *jobService) ListPublicJobs(ctx context.Context, query bson.M, sort string, page, limit int) ([]dto.JobResponseDTO, int64, error) {
	// Public jobs must be active or expiring soon
	query["status"] = bson.M{"$in": []string{domain.JobStatusActive, domain.JobStatusExpiringSoon}}

	skip := (page - 1) * limit
	jobs, total, err := s.jobRepo.FindAll(ctx, query, sort, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.mapListToDTO(ctx, jobs), total, nil
}

func (s *jobService) ListCompanyJobs(ctx context.Context, companyID primitive.ObjectID, query bson.M, sort string, page, limit int) ([]dto.JobResponseDTO, int64, error) {
	query["companyId"] = companyID

	skip := (page - 1) * limit
	jobs, total, err := s.jobRepo.FindAll(ctx, query, sort, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.mapListToDTO(ctx, jobs), total, nil
}

func (s *jobService) GetSimilarJobs(ctx context.Context, jobID primitive.ObjectID) ([]dto.JobResponseDTO, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Job not found")
	}

	jobs, err := s.jobRepo.FindSimilar(ctx, job, 4)
	if err != nil {
		return nil, err
	}

	return s.mapListToDTO(ctx, jobs), nil
}

func (s *jobService) PublishJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error {
	_, err := s.UpdateJob(ctx, companyID, jobID, dto.UpdateJobDTO{Status: domain.JobStatusActive})
	return err
}

func (s *jobService) CloseJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error {
	_, err := s.UpdateJob(ctx, companyID, jobID, dto.UpdateJobDTO{Status: domain.JobStatusClosed})
	return err
}

func (s *jobService) ReactivateJob(ctx context.Context, companyID primitive.ObjectID, jobID primitive.ObjectID) error {
	// Reactivate closed job -> active
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return appErrors.NewNotFoundError("Job not found")
	}

	if job.Deadline.Before(time.Now()) {
		// Automatically extend deadline by 14 days if expired
		job.Deadline = time.Now().AddDate(0, 0, 14)
	}

	_, err = s.UpdateJob(ctx, companyID, jobID, dto.UpdateJobDTO{
		Status:   domain.JobStatusActive,
		Deadline: job.Deadline,
	})
	return err
}

func (s *jobService) BulkAction(ctx context.Context, companyID primitive.ObjectID, input dto.BulkJobActionDTO) (int64, error) {
	objectIDs := make([]primitive.ObjectID, 0)
	for _, idStr := range input.JobIDs {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}

	if len(objectIDs) == 0 {
		return 0, appErrors.NewValidationError("No valid job IDs provided")
	}

	switch input.Action {
	case "activate":
		return s.jobRepo.BulkUpdateStatus(ctx, objectIDs, companyID, domain.JobStatusActive)
	case "deactivate":
		return s.jobRepo.BulkUpdateStatus(ctx, objectIDs, companyID, domain.JobStatusDraft)
	case "delete":
		return s.jobRepo.BulkDelete(ctx, objectIDs, companyID)
	default:
		return 0, appErrors.NewValidationError("Invalid bulk action")
	}
}

func (s *jobService) validatePublishFields(input dto.CreateJobDTO) error {
	var missing []string
	if input.Title == "" {
		missing = append(missing, "title")
	}
	if input.JobType == "" {
		missing = append(missing, "jobType")
	}
	if input.WorkMode == "" {
		missing = append(missing, "workMode")
	}
	if input.Category == "" {
		missing = append(missing, "category")
	}
	if input.ExperienceLevel == "" {
		missing = append(missing, "experienceLevel")
	}
	if input.Location == "" {
		missing = append(missing, "location")
	}
	if input.Description == "" {
		missing = append(missing, "description")
	}
	if len(input.Skills) == 0 {
		missing = append(missing, "skills")
	}
	if input.Deadline.IsZero() {
		missing = append(missing, "deadline")
	}

	if len(missing) > 0 {
		var details []appErrors.ErrorDetail
		for _, m := range missing {
			details = append(details, appErrors.ErrorDetail{
				Field:   m,
				Message: "field is required to publish job",
			})
		}
		return appErrors.NewValidationError("Validation failed for publishing job", details...)
	}

	return nil
}

func (s *jobService) checkAndSetExpired(ctx context.Context, job *domain.Job) {
	if job.Status == domain.JobStatusClosed || job.Status == domain.JobStatusDraft {
		return
	}

	now := time.Now()
	if job.Deadline.Before(now) {
		job.Status = domain.JobStatusClosed
		_ = s.jobRepo.Update(ctx, job)
		return
	}

	// Compute expiring soon: deadline is within 7 days
	if job.Deadline.Sub(now) <= 7*24*time.Hour {
		if job.Status != domain.JobStatusExpiringSoon {
			job.Status = domain.JobStatusExpiringSoon
			_ = s.jobRepo.Update(ctx, job)
		}
	}
}

func formatCompanyInfo(jobLocation string, comp *domain.Company) (companyName, logoURL string, compInfo *dto.CompanyInfoDTO) {
	if comp == nil {
		return "Unknown Company", "", &dto.CompanyInfoDTO{
			Name:        "Unknown Company",
			CompanyName: "Unknown Company",
			Industry:    "Technology & Software",
			About:       "No company bio provided yet.",
			Website:     "",
			Location:    jobLocation,
			Employees:   "1-50 employees",
			CompanySize: "1-50 employees",
			CompanyType: "",
			Founded:     "",
			LogoURL:     "",
		}
	}

	companyName = comp.Name
	if companyName == "" {
		companyName = "Unknown Company"
	}
	logoURL = comp.LogoURL

	var locParts []string
	if comp.Location.City != "" {
		locParts = append(locParts, comp.Location.City)
	}
	if comp.Location.State != "" {
		locParts = append(locParts, comp.Location.State)
	}
	if comp.Location.Country != "" {
		locParts = append(locParts, comp.Location.Country)
	}
	companyLocation := strings.Join(locParts, ", ")
	if companyLocation == "" {
		companyLocation = jobLocation
	}

	industry := comp.Industry
	if industry == "" {
		industry = "Technology & Software"
	}
	about := comp.About
	if about == "" {
		about = "No company bio provided yet."
	}
	employees := comp.Size
	if employees == "" {
		employees = "1-50 employees"
	}
	founded := comp.Founded
	if founded != "" && !strings.HasPrefix(strings.ToLower(founded), "founded") {
		founded = "Founded in " + founded
	}

	compInfo = &dto.CompanyInfoDTO{
		Name:        companyName,
		CompanyName: companyName,
		Industry:    industry,
		About:       about,
		Website:     comp.Website,
		Location:    companyLocation,
		Employees:   employees,
		CompanySize: comp.Size,
		CompanyType: comp.Type,
		Founded:     founded,
		LogoURL:     comp.LogoURL,
	}
	return companyName, logoURL, compInfo
}

func (s *jobService) mapToDTO(ctx context.Context, job *domain.Job) (*dto.JobResponseDTO, error) {
	company, _ := s.companyRepo.FindByID(ctx, job.CompanyID)
	companyName, logoURL, companyInfo := formatCompanyInfo(job.Location, company)

	postedAt := job.CreatedAt
	if job.PublishedAt != nil {
		postedAt = *job.PublishedAt
	}

	tags := make([]string, 0)
	// Capitalize tag representations for beautiful UI compatibility
	if job.JobType != "" {
		tags = append(tags, strings.Title(job.JobType))
	}
	if job.WorkMode != "" {
		tags = append(tags, strings.Title(job.WorkMode))
	}
	if job.ExperienceLevel != "" {
		tags = append(tags, strings.Title(job.ExperienceLevel)+" Level")
	}

	salaryLabel := utils.FormatSalary(job.SalaryMin, job.SalaryMax, job.SalaryPeriod)
	postedLabel := utils.FormatRelativeTime(postedAt)

	return &dto.JobResponseDTO{
		ID:              job.ID.Hex(),
		Title:           job.Title,
		Company:         companyName,
		CompanyID:       job.CompanyID.Hex(),
		LogoURL:         logoURL,
		Location:        job.Location,
		PostedAt:        postedAt,
		PostedLabel:     postedLabel,
		Category:        job.Category,
		Description:     job.Description,
		Requirements:    job.Requirements,
		Benefits:        job.Benefits,
		Tags:            tags,
		Salary:          salaryLabel,
		SalaryMin:       job.SalaryMin,
		SalaryMax:       job.SalaryMax,
		SalaryPeriod:    job.SalaryPeriod,
		Applicants:      job.ApplicantsCount,
		JobType:         job.JobType,
		WorkMode:        job.WorkMode,
		ExperienceLevel: job.ExperienceLevel,
		Deadline:        job.Deadline,
		Status:          job.Status,
		Skills:          job.Skills,
		Vacancies:       job.Vacancies,
		CompanyInfo:     companyInfo,
	}, nil
}

func (s *jobService) mapListToDTO(ctx context.Context, jobs []domain.Job) []dto.JobResponseDTO {
	// Pre-load company details to avoid N+1 queries
	companyIDs := make([]primitive.ObjectID, 0)
	for _, j := range jobs {
		companyIDs = append(companyIDs, j.CompanyID)
	}

	companiesMap := make(map[string]*domain.Company)
	for _, cid := range companyIDs {
		if _, ok := companiesMap[cid.Hex()]; !ok {
			comp, err := s.companyRepo.FindByID(ctx, cid)
			if err == nil && comp != nil {
				companiesMap[cid.Hex()] = comp
			}
		}
	}

	dtos := make([]dto.JobResponseDTO, 0)
	for _, job := range jobs {
		// Update status dynamically if expired
		s.checkAndSetExpired(ctx, &job)

		postedAt := job.CreatedAt
		if job.PublishedAt != nil {
			postedAt = *job.PublishedAt
		}

		tags := make([]string, 0)
		if job.JobType != "" {
			tags = append(tags, strings.Title(job.JobType))
		}
		if job.WorkMode != "" {
			tags = append(tags, strings.Title(job.WorkMode))
		}
		if job.ExperienceLevel != "" {
			tags = append(tags, strings.Title(job.ExperienceLevel)+" Level")
		}

		salaryLabel := utils.FormatSalary(job.SalaryMin, job.SalaryMax, job.SalaryPeriod)
		postedLabel := utils.FormatRelativeTime(postedAt)

		comp := companiesMap[job.CompanyID.Hex()]
		companyName, logoURL, companyInfo := formatCompanyInfo(job.Location, comp)

		dtos = append(dtos, dto.JobResponseDTO{
			ID:              job.ID.Hex(),
			Title:           job.Title,
			Company:         companyName,
			CompanyID:       job.CompanyID.Hex(),
			LogoURL:         logoURL,
			Location:        job.Location,
			PostedAt:        postedAt,
			PostedLabel:     postedLabel,
			Category:        job.Category,
			Description:     job.Description,
			Tags:            tags,
			Salary:          salaryLabel,
			SalaryMin:       job.SalaryMin,
			SalaryMax:       job.SalaryMax,
			SalaryPeriod:    job.SalaryPeriod,
			Applicants:      job.ApplicantsCount,
			JobType:         job.JobType,
			WorkMode:        job.WorkMode,
			ExperienceLevel: job.ExperienceLevel,
			Deadline:        job.Deadline,
			Status:          job.Status,
			Skills:          job.Skills,
			Vacancies:       job.Vacancies,
			CompanyInfo:     companyInfo,
		})
	}
	return dtos
}
