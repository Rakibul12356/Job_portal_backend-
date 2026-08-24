package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApplicationService interface {
	ApplyToJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID, file *multipart.FileHeader, coverMessage string) (*dto.SeekerApplicationResponseDTO, error)
	ListSeekerApplications(ctx context.Context, userID primitive.ObjectID, query bson.M, page, limit int) ([]dto.SeekerApplicationResponseDTO, int64, error)
	GetSeekerApplicationByID(ctx context.Context, userID primitive.ObjectID, appID primitive.ObjectID) (*dto.SeekerApplicationResponseDTO, error)
	WithdrawApplication(ctx context.Context, userID primitive.ObjectID, appID primitive.ObjectID) error
	ListCompanyApplicants(ctx context.Context, companyID primitive.ObjectID, query bson.M, experienceFilter string, page, limit int) ([]dto.CompanyApplicantResponseDTO, int64, error)
	GetCompanyApplicantByID(ctx context.Context, companyID primitive.ObjectID, appID primitive.ObjectID) (*dto.CompanyApplicantResponseDTO, error)
	UpdateApplicantStatus(ctx context.Context, companyID primitive.ObjectID, appID primitive.ObjectID, status string) error
}

type applicationService struct {
	appRepo        repository.ApplicationRepository
	jobRepo        repository.JobRepository
	companyRepo    repository.CompanyRepository
	profileRepo    repository.ProfileRepository
	userRepo       repository.UserRepository
	storageService StorageService
}

func NewApplicationService(
	appRepo repository.ApplicationRepository,
	jobRepo repository.JobRepository,
	companyRepo repository.CompanyRepository,
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	storageService StorageService,
) ApplicationService {
	return &applicationService{
		appRepo:        appRepo,
		jobRepo:        jobRepo,
		companyRepo:    companyRepo,
		profileRepo:    profileRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

func (s *applicationService) ApplyToJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID, file *multipart.FileHeader, coverMessage string) (*dto.SeekerApplicationResponseDTO, error) {
	// 1. Verify role is user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user.Role != domain.RoleUser {
		return nil, appErrors.NewForbiddenError("Only job seekers can apply to jobs")
	}

	// 2. Check for duplicate application
	existing, _ := s.appRepo.FindByJobIDAndUserID(ctx, jobID, userID)
	if existing != nil {
		return nil, appErrors.NewConflictError("You have already applied to this job listing")
	}

	// 3. Resolve job details and check active status
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Job listing not found")
	}

	if job.Status != domain.JobStatusActive && job.Status != domain.JobStatusExpiringSoon {
		return nil, appErrors.NewValidationError("You can only apply to active job listings")
	}

	// 4. Save the resume PDF
	if file == nil {
		return nil, appErrors.NewValidationError("Resume PDF is required to apply")
	}
	resumeURL, err := s.storageService.SaveUploadedFile(file, "resume", userID.Hex())
	if err != nil {
		return nil, appErrors.NewValidationError("Resume upload failed: " + err.Error())
	}

	app := &domain.Application{
		ID:             primitive.NewObjectID(),
		JobID:          jobID,
		CompanyID:      job.CompanyID,
		UserID:         userID,
		Status:         domain.AppStatusPending,
		CoverMessage:   coverMessage,
		ResumeURL:      resumeURL,
		ResumeFilename: file.Filename,
	}

	err = s.appRepo.Create(ctx, app)
	if err != nil {
		// Cleanup uploaded file on DB fail
		_ = s.storageService.DeleteFile(resumeURL)
		return nil, appErrors.NewInternalError("Failed to save job application")
	}

	// 5. Increment job applicantsCount
	job.ApplicantsCount++
	_ = s.jobRepo.Update(ctx, job)

	return s.mapToSeekerDTO(ctx, app)
}

func (s *applicationService) ListSeekerApplications(ctx context.Context, userID primitive.ObjectID, query bson.M, page, limit int) ([]dto.SeekerApplicationResponseDTO, int64, error) {
	skip := (page - 1) * limit
	apps, total, err := s.appRepo.FindByUserID(ctx, userID, query, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]dto.SeekerApplicationResponseDTO, 0)
	for _, app := range apps {
		dtoVal, err := s.mapToSeekerDTO(ctx, &app)
		if err == nil {
			dtos = append(dtos, *dtoVal)
		}
	}
	return dtos, total, nil
}

func (s *applicationService) GetSeekerApplicationByID(ctx context.Context, userID primitive.ObjectID, appID primitive.ObjectID) (*dto.SeekerApplicationResponseDTO, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Application not found")
	}

	if app.UserID != userID {
		return nil, appErrors.NewForbiddenError("You are not the owner of this application")
	}

	return s.mapToSeekerDTO(ctx, app)
}

func (s *applicationService) WithdrawApplication(ctx context.Context, userID primitive.ObjectID, appID primitive.ObjectID) error {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return appErrors.NewNotFoundError("Application not found")
	}

	if app.UserID != userID {
		return appErrors.NewForbiddenError("You are not the owner of this application")
	}

	if app.Status != domain.AppStatusPending && app.Status != domain.AppStatusShortlisted {
		return appErrors.NewValidationError("You can only withdraw pending or shortlisted applications")
	}

	app.Status = domain.AppStatusWithdrawn
	err = s.appRepo.Update(ctx, app)
	if err != nil {
		return appErrors.NewInternalError("Failed to withdraw application")
	}

	// Decrement job applicantsCount
	job, err := s.jobRepo.FindByID(ctx, app.JobID)
	if err == nil {
		job.ApplicantsCount--
		if job.ApplicantsCount < 0 {
			job.ApplicantsCount = 0
		}
		_ = s.jobRepo.Update(ctx, job)
	}

	return nil
}

func (s *applicationService) ListCompanyApplicants(ctx context.Context, companyID primitive.ObjectID, query bson.M, experienceFilter string, page, limit int) ([]dto.CompanyApplicantResponseDTO, int64, error) {
	skip := (page - 1) * limit
	apps, total, err := s.appRepo.FindByCompanyID(ctx, companyID, query, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]dto.CompanyApplicantResponseDTO, 0)
	for _, app := range apps {
		// Apply experience level filter in-memory if specified
		if experienceFilter != "" {
			profile, err := s.profileRepo.FindByUserID(ctx, app.UserID)
			if err != nil || profile.Title != experienceFilter { // title field is experienceLevel fallback in register seeker
				continue
			}
		}

		dtoVal, err := s.mapToCompanyDTO(ctx, &app)
		if err == nil {
			dtos = append(dtos, *dtoVal)
		}
	}
	return dtos, total, nil
}

func (s *applicationService) GetCompanyApplicantByID(ctx context.Context, companyID primitive.ObjectID, appID primitive.ObjectID) (*dto.CompanyApplicantResponseDTO, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Application not found")
	}

	if app.CompanyID != companyID {
		return nil, appErrors.NewForbiddenError("Forbidden: Listing belongs to another company")
	}

	return s.mapToCompanyDTO(ctx, app)
}

func (s *applicationService) UpdateApplicantStatus(ctx context.Context, companyID primitive.ObjectID, appID primitive.ObjectID, status string) error {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return appErrors.NewNotFoundError("Application not found")
	}

	if app.CompanyID != companyID {
		return appErrors.NewForbiddenError("Forbidden: Listing belongs to another company")
	}

	// Validate status transition
	// - pending -> shortlisted | rejected | interviewed
	// - shortlisted -> interviewed | rejected
	// - interviewed -> rejected | shortlisted
	valid := false
	switch app.Status {
	case domain.AppStatusPending:
		valid = status == domain.AppStatusShortlisted || status == domain.AppStatusRejected || status == domain.AppStatusInterviewed
	case domain.AppStatusShortlisted:
		valid = status == domain.AppStatusInterviewed || status == domain.AppStatusRejected
	case domain.AppStatusInterviewed:
		valid = status == domain.AppStatusRejected || status == domain.AppStatusShortlisted
	}

	if !valid {
		return appErrors.NewValidationError("Invalid application status transition from " + app.Status + " to " + status)
	}

	app.Status = status
	now := time.Now()
	app.ReviewedAt = &now

	err = s.appRepo.Update(ctx, app)
	if err != nil {
		return appErrors.NewInternalError("Failed to update status")
	}

	return nil
}

func (s *applicationService) mapToSeekerDTO(ctx context.Context, app *domain.Application) (*dto.SeekerApplicationResponseDTO, error) {
	job, err := s.jobRepo.FindByID(ctx, app.JobID)
	if err != nil {
		return nil, err
	}

	company, err := s.companyRepo.FindByID(ctx, app.CompanyID)
	companyName := "Unknown Company"
	companyLogo := ""
	if err == nil && company != nil {
		companyName = company.Name
		companyLogo = company.LogoURL
	}

	return &dto.SeekerApplicationResponseDTO{
		ID:             app.ID.Hex(),
		JobID:          app.JobID.Hex(),
		JobTitle:       job.Title,
		Company:        companyName,
		CompanyID:      app.CompanyID.Hex(),
		CompanyLogo:    companyLogo,
		Status:         app.Status,
		AppliedAt:      app.AppliedAt,
		ResumeURL:      app.ResumeURL,
		ResumeFilename: app.ResumeFilename,
		CoverMessage:   app.CoverMessage,
		Location:       job.Location,
		JobType:        job.JobType,
	}, nil
}

func (s *applicationService) mapToCompanyDTO(ctx context.Context, app *domain.Application) (*dto.CompanyApplicantResponseDTO, error) {
	job, err := s.jobRepo.FindByID(ctx, app.JobID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, app.UserID)
	if err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.FindByUserID(ctx, app.UserID)
	seekerTitle := ""
	seekerPhone := ""
	seekerAvatar := ""
	seekerSkills := []string{}
	if err == nil && profile != nil {
		seekerTitle = profile.Title
		seekerPhone = profile.Phone
		seekerAvatar = profile.AvatarURL
		seekerSkills = profile.Skills
	}

	return &dto.CompanyApplicantResponseDTO{
		ID:             app.ID.Hex(),
		JobID:          app.JobID.Hex(),
		JobTitle:       job.Title,
		UserID:         app.UserID.Hex(),
		SeekerName:     user.Name,
		SeekerEmail:    user.Email,
		SeekerTitle:    seekerTitle,
		SeekerPhone:    seekerPhone,
		SeekerAvatar:   seekerAvatar,
		SeekerSkills:   seekerSkills,
		Status:         app.Status,
		AppliedAt:      app.AppliedAt,
		ResumeURL:      app.ResumeURL,
		ResumeFilename: app.ResumeFilename,
		CoverMessage:   app.CoverMessage,
	}, nil
}
