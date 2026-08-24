package service

import (
	"context"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedJobService interface {
	SaveJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error
	UnsaveJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error
	ListSavedJobs(ctx context.Context, userID primitive.ObjectID) ([]dto.JobResponseDTO, error)
}

type savedJobService struct {
	savedRepo   repository.SavedJobRepository
	jobRepo     repository.JobRepository
	jobService  JobService
	companyRepo repository.CompanyRepository
}

func NewSavedJobService(
	savedRepo repository.SavedJobRepository,
	jobRepo repository.JobRepository,
	jobService JobService,
	companyRepo repository.CompanyRepository,
) SavedJobService {
	return &savedJobService{
		savedRepo:   savedRepo,
		jobRepo:     jobRepo,
		jobService:  jobService,
		companyRepo: companyRepo,
	}
}

func (s *savedJobService) SaveJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error {
	// 1. Verify job exists and is active
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return appErrors.NewNotFoundError("Job listing not found")
	}

	if job.Status != domain.JobStatusActive && job.Status != domain.JobStatusExpiringSoon {
		return appErrors.NewValidationError("Cannot bookmark inactive job listings")
	}

	// 2. Check if already bookmarked
	saved, _ := s.savedRepo.IsSaved(ctx, userID, jobID)
	if saved {
		return appErrors.NewConflictError("You have already bookmarked this job")
	}

	savedJob := &domain.SavedJob{
		ID:     primitive.NewObjectID(),
		UserID: userID,
		JobID:  jobID,
	}

	err = s.savedRepo.Create(ctx, savedJob)
	if err != nil {
		return appErrors.NewInternalError("Failed to save bookmarked job")
	}

	return nil
}

func (s *savedJobService) UnsaveJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error {
	err := s.savedRepo.Delete(ctx, userID, jobID)
	if err != nil {
		return appErrors.NewInternalError("Failed to remove bookmarked job")
	}
	return nil
}

func (s *savedJobService) ListSavedJobs(ctx context.Context, userID primitive.ObjectID) ([]dto.JobResponseDTO, error) {
	bookmarks, err := s.savedRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	jobs := make([]dto.JobResponseDTO, 0)
	for _, mark := range bookmarks {
		jobDTO, err := s.jobService.GetJobByID(ctx, mark.JobID)
		if err == nil && jobDTO != nil {
			jobs = append(jobs, *jobDTO)
		}
	}

	return jobs, nil
}
