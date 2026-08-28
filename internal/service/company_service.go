package service

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompanyService interface {
	GetCompanyByOwnerID(ctx context.Context, ownerID primitive.ObjectID) (*domain.Company, error)
	GetCompanySettings(ctx context.Context, ownerID primitive.ObjectID) (*dto.CompanySettingsResponseDTO, error)
	UpdateCompanySettings(ctx context.Context, ownerID primitive.ObjectID, input dto.UpdateCompanySettingsDTO) (*dto.CompanySettingsResponseDTO, error)
	UploadLogo(ctx context.Context, companyID primitive.ObjectID, file *multipart.FileHeader) (string, error)
	RemoveLogo(ctx context.Context, companyID primitive.ObjectID) error
	GetPublicProfile(ctx context.Context, companyID primitive.ObjectID) (*dto.PublicCompanyProfileResponseDTO, error)
	ListCompanies(ctx context.Context) ([]dto.PublicCompanyProfileResponseDTO, error)
}

type companyService struct {
	companyRepo    repository.CompanyRepository
	jobRepo        repository.JobRepository
	userRepo       repository.UserRepository
	storageService StorageService
}

func NewCompanyService(
	companyRepo repository.CompanyRepository,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	storageService StorageService,
) CompanyService {
	return &companyService{
		companyRepo:    companyRepo,
		jobRepo:        jobRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

func (s *companyService) GetCompanyByOwnerID(ctx context.Context, ownerID primitive.ObjectID) (*domain.Company, error) {
	company, err := s.companyRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Company not found for this user account")
	}
	return company, nil
}

func (s *companyService) GetCompanySettings(ctx context.Context, ownerID primitive.ObjectID) (*dto.CompanySettingsResponseDTO, error) {
	comp, err := s.GetCompanyByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	return &dto.CompanySettingsResponseDTO{
		CompanyName:  comp.Name,
		Industry:     comp.Industry,
		CompanySize:  comp.Size,
		CompanyType:  comp.Type,
		Website:      comp.Website,
		Founded:      comp.Founded,
		About:        comp.About,
		City:         comp.Location.City,
		State:        comp.Location.State,
		Country:      comp.Location.Country,
		Phone:        comp.Contact.Phone,
		HREmail:      comp.Contact.HREmail,
		SupportEmail: comp.Contact.SupportEmail,
		Linkedin:     comp.Social.Linkedin,
		Twitter:      comp.Social.Twitter,
		Facebook:     comp.Social.Facebook,
		Instagram:    comp.Social.Instagram,
		Github:       comp.Social.Github,
	}, nil
}

func (s *companyService) UpdateCompanySettings(ctx context.Context, ownerID primitive.ObjectID, input dto.UpdateCompanySettingsDTO) (*dto.CompanySettingsResponseDTO, error) {
	comp, err := s.GetCompanyByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, ownerID)
	if err == nil && input.CompanyName != "" && input.CompanyName != user.Name {
		user.Name = input.CompanyName
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, appErrors.NewInternalError("Failed to update user identity: " + err.Error())
		}
	}

	comp.Name = input.CompanyName
	comp.Industry = input.Industry
	comp.Size = input.CompanySize
	comp.Type = input.CompanyType
	comp.Website = input.Website
	comp.Founded = input.Founded
	comp.About = input.About

	comp.Location = domain.CompanyLocation{
		City:    input.City,
		State:   input.State,
		Country: input.Country,
	}

	comp.Contact = domain.CompanyContact{
		Phone:        input.Phone,
		HREmail:      input.HREmail,
		SupportEmail: input.SupportEmail,
	}

	comp.Social = domain.CompanySocial{
		Linkedin:  input.Linkedin,
		Twitter:   input.Twitter,
		Facebook:  input.Facebook,
		Instagram: input.Instagram,
		Github:    input.Github,
	}

	err = s.companyRepo.Update(ctx, comp)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to update company settings")
	}

	return &dto.CompanySettingsResponseDTO{
		CompanyName:  comp.Name,
		Industry:     comp.Industry,
		CompanySize:  comp.Size,
		CompanyType:  comp.Type,
		Website:      comp.Website,
		Founded:      comp.Founded,
		About:        comp.About,
		City:         comp.Location.City,
		State:        comp.Location.State,
		Country:      comp.Location.Country,
		Phone:        comp.Contact.Phone,
		HREmail:      comp.Contact.HREmail,
		SupportEmail: comp.Contact.SupportEmail,
		Linkedin:     comp.Social.Linkedin,
		Twitter:      comp.Social.Twitter,
		Facebook:     comp.Social.Facebook,
		Instagram:    comp.Social.Instagram,
		Github:       comp.Social.Github,
	}, nil
}

func (s *companyService) UploadLogo(ctx context.Context, companyID primitive.ObjectID, file *multipart.FileHeader) (string, error) {
	comp, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return "", appErrors.NewNotFoundError("Company not found")
	}

	// Delete old logo if present
	if comp.LogoURL != "" {
		_ = s.storageService.DeleteFile(comp.LogoURL)
	}

	logoURL, err := s.storageService.SaveUploadedFile(file, "logo", companyID.Hex())
	if err != nil {
		return "", appErrors.NewValidationError("Failed to upload logo: " + err.Error())
	}

	comp.LogoURL = logoURL
	err = s.companyRepo.Update(ctx, comp)
	if err != nil {
		_ = s.storageService.DeleteFile(logoURL)
		return "", appErrors.NewInternalError("Failed to save logo URL to company profile")
	}

	return logoURL, nil
}

func (s *companyService) RemoveLogo(ctx context.Context, companyID primitive.ObjectID) error {
	comp, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return appErrors.NewNotFoundError("Company not found")
	}

	if comp.LogoURL != "" {
		_ = s.storageService.DeleteFile(comp.LogoURL)
		comp.LogoURL = ""
		return s.companyRepo.Update(ctx, comp)
	}

	return nil
}

func (s *companyService) GetPublicProfile(ctx context.Context, companyID primitive.ObjectID) (*dto.PublicCompanyProfileResponseDTO, error) {
	comp, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Company profile not found")
	}

	// Find open active jobs for this company
	jobsFilter := bson.M{
		"companyId": companyID,
		"status":    bson.M{"$in": []string{domain.JobStatusActive, domain.JobStatusExpiringSoon}},
	}
	jobsList, _, err := s.jobRepo.FindAll(ctx, jobsFilter, "newest", 0, 50)
	if err != nil {
		jobsList = []domain.Job{}
	}

	openJobs := make([]dto.JobResponseDTO, 0)
	for _, job := range jobsList {
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

		var compLocParts []string
		if comp.Location.City != "" {
			compLocParts = append(compLocParts, comp.Location.City)
		}
		if comp.Location.State != "" {
			compLocParts = append(compLocParts, comp.Location.State)
		}
		if comp.Location.Country != "" {
			compLocParts = append(compLocParts, comp.Location.Country)
		}
		compLocStr := strings.Join(compLocParts, ", ")
		if compLocStr == "" {
			compLocStr = job.Location
		}

		compFounded := comp.Founded
		if compFounded != "" && !strings.HasPrefix(strings.ToLower(compFounded), "founded") {
			compFounded = "Founded in " + compFounded
		}

		openJobs = append(openJobs, dto.JobResponseDTO{
			ID:              job.ID.Hex(),
			Title:           job.Title,
			Company:         comp.Name,
			CompanyID:       companyID.Hex(),
			LogoURL:         comp.LogoURL,
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
			CompanyInfo: &dto.CompanyInfoDTO{
				Name:        comp.Name,
				CompanyName: comp.Name,
				Industry:    comp.Industry,
				About:       comp.About,
				Website:     comp.Website,
				Location:    compLocStr,
				Employees:   comp.Size,
				CompanySize: comp.Size,
				CompanyType: comp.Type,
				Founded:     compFounded,
				LogoURL:     comp.LogoURL,
			},
		})
	}

	return &dto.PublicCompanyProfileResponseDTO{
		ID:        comp.ID.Hex(),
		Name:      comp.Name,
		Industry:  comp.Industry,
		Website:   comp.Website,
		Size:      comp.Size,
		Type:      comp.Type,
		Founded:   comp.Founded,
		About:     comp.About,
		LogoURL:   comp.LogoURL,
		City:      comp.Location.City,
		State:     comp.Location.State,
		Country:   comp.Location.Country,
		Linkedin:  comp.Social.Linkedin,
		Twitter:   comp.Social.Twitter,
		Facebook:  comp.Social.Facebook,
		Instagram: comp.Social.Instagram,
		Github:    comp.Social.Github,
		OpenJobs:  openJobs,
	}, nil
}

func (s *companyService) ListCompanies(ctx context.Context) ([]dto.PublicCompanyProfileResponseDTO, error) {
	companies, err := s.companyRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.PublicCompanyProfileResponseDTO, 0)
	for _, comp := range companies {
		dtos = append(dtos, dto.PublicCompanyProfileResponseDTO{
			ID:        comp.ID.Hex(),
			Name:      comp.Name,
			Industry:  comp.Industry,
			Website:   comp.Website,
			Size:      comp.Size,
			Type:      comp.Type,
			Founded:   comp.Founded,
			About:     comp.About,
			LogoURL:   comp.LogoURL,
			City:      comp.Location.City,
			State:     comp.Location.State,
			Country:   comp.Location.Country,
			Linkedin:  comp.Social.Linkedin,
			Twitter:   comp.Social.Twitter,
			Facebook:  comp.Social.Facebook,
			Instagram: comp.Social.Instagram,
			Github:    comp.Social.Github,
		})
	}
	return dtos, nil
}
