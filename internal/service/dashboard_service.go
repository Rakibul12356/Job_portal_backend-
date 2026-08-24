package service

import (
	"context"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DashboardService interface {
	GetSeekerDashboard(ctx context.Context, userID primitive.ObjectID) (*dto.SeekerDashboardResponseDTO, error)
	GetCompanyDashboard(ctx context.Context, companyID primitive.ObjectID) (*dto.CompanyDashboardResponseDTO, error)
}

type dashboardService struct {
	appRepo     repository.ApplicationRepository
	jobRepo     repository.JobRepository
	savedRepo   repository.SavedJobRepository
	profileRepo repository.ProfileRepository
	appService  ApplicationService
	jobService  JobService
}

func NewDashboardService(
	appRepo repository.ApplicationRepository,
	jobRepo repository.JobRepository,
	savedRepo repository.SavedJobRepository,
	profileRepo repository.ProfileRepository,
	appService ApplicationService,
	jobService JobService,
) DashboardService {
	return &dashboardService{
		appRepo:     appRepo,
		jobRepo:     jobRepo,
		savedRepo:   savedRepo,
		profileRepo: profileRepo,
		appService:  appService,
		jobService:  jobService,
	}
}

func (s *dashboardService) GetSeekerDashboard(ctx context.Context, userID primitive.ObjectID) (*dto.SeekerDashboardResponseDTO, error) {
	// 1. Fetch stats
	totalApps, err := s.appRepo.Count(ctx, bson.M{"userId": userID})
	if err != nil {
		totalApps = 0
	}
	pendingReviews, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusPending})
	shortlisted, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusShortlisted})
	rejected, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusRejected})

	savedJobsCount, err := s.savedRepo.FindByUserID(ctx, userID)
	var savedJobs int64
	if err == nil {
		savedJobs = int64(len(savedJobsCount))
	}

	stats := dto.SeekerDashboardStats{
		TotalApplications: totalApps,
		Shortlisted:       shortlisted,
		Rejected:          rejected,
		PendingReviews:    pendingReviews,
		SavedJobs:         savedJobs,
	}

	// 2. Recent applications (limit 5)
	recentApps, _, err := s.appService.ListSeekerApplications(ctx, userID, bson.M{}, 1, 5)
	if err != nil {
		recentApps = []dto.SeekerApplicationResponseDTO{}
	}

	// 3. Recommended jobs
	var recommended []dto.JobResponseDTO
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err == nil && profile != nil && (len(profile.Skills) > 0 || profile.Title != "") {
		// Try matching skills or title/experience fallback
		orQueries := []bson.M{}
		if len(profile.Skills) > 0 {
			orQueries = append(orQueries, bson.M{"skills": bson.M{"$in": profile.Skills}})
		}
		if profile.Title != "" {
			orQueries = append(orQueries, bson.M{"category": profile.Title})
			orQueries = append(orQueries, bson.M{"experienceLevel": profile.Title})
		}

		filter := bson.M{
			"status": bson.M{"$in": []string{domain.JobStatusActive, domain.JobStatusExpiringSoon}},
			"$or":   orQueries,
		}
		recommended, _, _ = s.jobService.ListPublicJobs(ctx, filter, "newest", 1, 5)
	}

	// Fallback to general newest active jobs if none recommended
	if len(recommended) == 0 {
		recommended, _, _ = s.jobService.ListPublicJobs(ctx, bson.M{}, "newest", 1, 5)
	}

	return &dto.SeekerDashboardResponseDTO{
		Stats:           stats,
		RecentApplied:   recentApps,
		RecommendedJobs: recommended,
	}, nil
}

func (s *dashboardService) GetCompanyDashboard(ctx context.Context, companyID primitive.ObjectID) (*dto.CompanyDashboardResponseDTO, error) {
	// 1. Stats
	activeJobs, _ := s.jobRepo.Count(ctx, bson.M{
		"companyId": companyID,
		"status":    bson.M{"$in": []string{domain.JobStatusActive, domain.JobStatusExpiringSoon}},
	})
	totalApplicants, _ := s.appRepo.Count(ctx, bson.M{"companyId": companyID})
	pendingReviews, _ := s.appRepo.Count(ctx, bson.M{"companyId": companyID, "status": domain.AppStatusPending})
	shortlisted, _ := s.appRepo.Count(ctx, bson.M{"companyId": companyID, "status": domain.AppStatusShortlisted})

	stats := dto.CompanyDashboardStats{
		ActiveJobs:      activeJobs,
		TotalApplicants: totalApplicants,
		PendingReviews:  pendingReviews,
		Shortlisted:     shortlisted,
	}

	// 2. Recent jobs (limit 5)
	recentJobs, _, err := s.jobService.ListCompanyJobs(ctx, companyID, bson.M{}, "newest", 1, 5)
	if err != nil {
		recentJobs = []dto.JobResponseDTO{}
	}

	// 3. Recent applicants (limit 5)
	recentApplicants, _, err := s.appService.ListCompanyApplicants(ctx, companyID, bson.M{}, "", 1, 5)
	if err != nil {
		recentApplicants = []dto.CompanyApplicantResponseDTO{}
	}

	return &dto.CompanyDashboardResponseDTO{
		Stats:            stats,
		RecentJobs:       recentJobs,
		RecentApplicants: recentApplicants,
	}, nil
}
