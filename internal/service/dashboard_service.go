package service

import (
	"context"
	"sync"

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

// GetSeekerDashboard assembles the seeker's dashboard by running all
// independent database queries concurrently. Without goroutines, the queries
// would run sequentially, meaning total latency = sum of all query latencies.
// With goroutines, total latency ≈ max(slowest query), giving us a significant
// speedup on every dashboard request.
func (s *dashboardService) GetSeekerDashboard(ctx context.Context, userID primitive.ObjectID) (*dto.SeekerDashboardResponseDTO, error) {
	var wg sync.WaitGroup

	// ─── Result holders ──────────────────────────────────────────────────────
	// Each goroutine writes to its own dedicated variable, so no mutex is needed
	// for the results themselves. Only error aggregation uses a mutex.
	var (
		totalApps      int64
		pendingReviews int64
		shortlisted    int64
		rejected       int64
		savedJobs      int64

		recentApps  []dto.SeekerApplicationResponseDTO
		recommended []dto.JobResponseDTO
	)

	// ─── Error collection (mutex-guarded) ───────────────────────────────────
	// Multiple goroutines may write an error simultaneously; the mutex ensures
	// we do not have a data race on the firstErr variable.
	var (
		mu      sync.Mutex
		firstErr error
	)
	setErr := func(err error) {
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}

	// =========================================================================
	// GOROUTINE 1 — Application Stats (5 independent Count queries)
	// These 5 counts have no data dependency on each other, so we fan them out
	// with a nested WaitGroup and collect results without any shared state risk.
	// =========================================================================
	wg.Add(1)
	go func() {
		defer wg.Done()

		var statsWg sync.WaitGroup

		statsWg.Add(1)
		go func() {
			defer statsWg.Done()
			n, err := s.appRepo.Count(ctx, bson.M{"userId": userID})
			if err == nil {
				totalApps = n
			}
		}()

		statsWg.Add(1)
		go func() {
			defer statsWg.Done()
			n, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusPending})
			pendingReviews = n
		}()

		statsWg.Add(1)
		go func() {
			defer statsWg.Done()
			n, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusShortlisted})
			shortlisted = n
		}()

		statsWg.Add(1)
		go func() {
			defer statsWg.Done()
			n, _ := s.appRepo.Count(ctx, bson.M{"userId": userID, "status": domain.AppStatusRejected})
			rejected = n
		}()

		statsWg.Add(1)
		go func() {
			defer statsWg.Done()
			saved, err := s.savedRepo.FindByUserID(ctx, userID)
			if err == nil {
				savedJobs = int64(len(saved))
			}
		}()

		// Block this goroutine until all 5 stat counts are done.
		statsWg.Wait()
	}()

	// =========================================================================
	// GOROUTINE 2 — Recent Applications (limit 5)
	// Runs completely independently of the stats queries above.
	// =========================================================================
	wg.Add(1)
	go func() {
		defer wg.Done()
		apps, _, err := s.appService.ListSeekerApplications(ctx, userID, bson.M{}, 1, 5)
		if err != nil {
			recentApps = []dto.SeekerApplicationResponseDTO{}
			return
		}
		recentApps = apps
	}()

	// =========================================================================
	// GOROUTINE 3 — Profile Fetch + Recommended Jobs
	// Fetches the profile and then derives a personalised recommendation filter.
	// Falls back to newest jobs if the profile has no skills/title.
	// This is one logical unit because the job query depends on the profile result.
	// =========================================================================
	wg.Add(1)
	go func() {
		defer wg.Done()

		profile, err := s.profileRepo.FindByUserID(ctx, userID)
		if err != nil || profile == nil || (len(profile.Skills) == 0 && profile.Title == "") {
			// No personalisation signal — fall back to newest active jobs.
			jobs, _, _ := s.jobService.ListPublicJobs(ctx, bson.M{}, "newest", 1, 5)
			recommended = jobs
			return
		}

		// Build an $or filter from the seeker's skills and job title.
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
			"$or":    orQueries,
		}
		jobs, _, err := s.jobService.ListPublicJobs(ctx, filter, "newest", 1, 5)
		if err != nil || len(jobs) == 0 {
			// Filter returned nothing — fall back to newest active jobs.
			jobs, _, _ = s.jobService.ListPublicJobs(ctx, bson.M{}, "newest", 1, 5)
			setErr(err)
		}
		recommended = jobs
	}()

	// ─── Synchronisation barrier ─────────────────────────────────────────────
	// Block until ALL 3 goroutines (and their nested goroutines) have finished.
	// This is the single point where sequential execution resumes.
	wg.Wait()

	// firstErr is intentionally not returned here — dashboard data is best-effort;
	// partial results are far more useful to the user than a hard 500 error.
	_ = firstErr

	return &dto.SeekerDashboardResponseDTO{
		Stats: dto.SeekerDashboardStats{
			TotalApplications: totalApps,
			Shortlisted:       shortlisted,
			Rejected:          rejected,
			PendingReviews:    pendingReviews,
			SavedJobs:         savedJobs,
		},
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
