package service

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileService interface {
	GetProfileByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.SeekerProfile, error)
	UpdateProfile(ctx context.Context, userID primitive.ObjectID, input dto.UpdateProfileDTO) (*domain.SeekerProfile, error)
	UploadAvatar(ctx context.Context, userID primitive.ObjectID, file *multipart.FileHeader) (string, error)
	RemoveAvatar(ctx context.Context, userID primitive.ObjectID) error
	UploadResume(ctx context.Context, userID primitive.ObjectID, file *multipart.FileHeader) (*domain.ResumeMetadata, error)
	RemoveResume(ctx context.Context, userID primitive.ObjectID) error
	AddExperience(ctx context.Context, userID primitive.ObjectID, input dto.ExperienceDTO) (*domain.Experience, error)
	UpdateExperience(ctx context.Context, userID primitive.ObjectID, expID string, input dto.ExperienceDTO) error
	DeleteExperience(ctx context.Context, userID primitive.ObjectID, expID string) error
	AddEducation(ctx context.Context, userID primitive.ObjectID, input dto.EducationDTO) (*domain.Education, error)
	UpdateEducation(ctx context.Context, userID primitive.ObjectID, eduID string, input dto.EducationDTO) error
	DeleteEducation(ctx context.Context, userID primitive.ObjectID, eduID string) error
}

type profileService struct {
	profileRepo    repository.ProfileRepository
	userRepo       repository.UserRepository
	storageService StorageService
}

func NewProfileService(
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	storageService StorageService,
) ProfileService {
	return &profileService{
		profileRepo:    profileRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

func (s *profileService) GetProfileByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.SeekerProfile, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		// If profile doesn't exist, create it dynamically
		profile = &domain.SeekerProfile{
			ID:         primitive.NewObjectID(),
			UserID:     userID,
			Skills:     []string{},
			Experience: []domain.Experience{},
			Education:  []domain.Education{},
		}
		err = s.profileRepo.Create(ctx, profile)
		if err != nil {
			return nil, appErrors.NewInternalError("Failed to initialize profile")
		}
	}
	return profile, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, input dto.UpdateProfileDTO) (*domain.SeekerProfile, error) {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("User identity not found")
	}

	// Update user info if email or name changes
	userUpdated := false
	if input.Name != "" && input.Name != user.Name {
		user.Name = input.Name
		nameParts := strings.Split(input.Name, " ")
		user.FirstName = nameParts[0]
		userUpdated = true
	}
	if input.Email != "" && input.Email != user.Email {
		user.Email = input.Email
		userUpdated = true
	}
	if userUpdated {
		err = s.userRepo.Update(ctx, user)
		if err != nil {
			return nil, appErrors.NewInternalError("Failed to update user identity: " + err.Error())
		}
	}

	// Update SeekerProfile fields
	if input.Phone != "" {
		profile.Phone = input.Phone
	}
	if input.Title != "" {
		profile.Title = input.Title
	}
	if input.Bio != "" {
		profile.Bio = input.Bio
	}

	profile.Location = domain.SeekerLocation{
		City:    input.City,
		State:   input.State,
		Country: input.Country,
		Zipcode: input.Zipcode,
	}

	profile.Social = domain.SeekerSocial{
		Linkedin:  input.Linkedin,
		Github:    input.Github,
		Portfolio: input.Portfolio,
	}

	if input.Skills != nil {
		profile.Skills = input.Skills
	}

	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to update profile")
	}

	return profile, nil
}

func (s *profileService) UploadAvatar(ctx context.Context, userID primitive.ObjectID, file *multipart.FileHeader) (string, error) {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	// Delete old avatar if present
	if profile.AvatarURL != "" {
		_ = s.storageService.DeleteFile(profile.AvatarURL)
	}

	avatarURL, err := s.storageService.SaveUploadedFile(file, "avatar", userID.Hex())
	if err != nil {
		return "", appErrors.NewValidationError("Failed to upload avatar: " + err.Error())
	}

	profile.AvatarURL = avatarURL
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		_ = s.storageService.DeleteFile(avatarURL)
		return "", appErrors.NewInternalError("Failed to save avatar URL to profile")
	}

	return avatarURL, nil
}

func (s *profileService) RemoveAvatar(ctx context.Context, userID primitive.ObjectID) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if profile.AvatarURL != "" {
		_ = s.storageService.DeleteFile(profile.AvatarURL)
		profile.AvatarURL = ""
		return s.profileRepo.Update(ctx, profile)
	}

	return nil
}

func (s *profileService) UploadResume(ctx context.Context, userID primitive.ObjectID, file *multipart.FileHeader) (*domain.ResumeMetadata, error) {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Delete old resume if present
	if profile.Resume != nil && profile.Resume.URL != "" {
		_ = s.storageService.DeleteFile(profile.Resume.URL)
	}

	resumeURL, err := s.storageService.SaveUploadedFile(file, "resume", userID.Hex())
	if err != nil {
		return nil, appErrors.NewValidationError("Failed to upload resume: " + err.Error())
	}

	meta := &domain.ResumeMetadata{
		URL:        resumeURL,
		Filename:   file.Filename,
		UploadedAt: time.Now(),
	}

	profile.Resume = meta
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		_ = s.storageService.DeleteFile(resumeURL)
		return nil, appErrors.NewInternalError("Failed to save resume URL to profile")
	}

	return meta, nil
}

func (s *profileService) RemoveResume(ctx context.Context, userID primitive.ObjectID) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if profile.Resume != nil {
		_ = s.storageService.DeleteFile(profile.Resume.URL)
		profile.Resume = nil
		return s.profileRepo.Update(ctx, profile)
	}

	return nil
}

func (s *profileService) AddExperience(ctx context.Context, userID primitive.ObjectID, input dto.ExperienceDTO) (*domain.Experience, error) {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	exp := domain.Experience{
		ID:          uuid.New().String(),
		Company:     input.Company,
		Title:       input.Title,
		Location:    input.Location,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Current:     input.Current,
		Description: input.Description,
	}

	profile.Experience = append(profile.Experience, exp)
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to add experience")
	}

	return &exp, nil
}

func (s *profileService) UpdateExperience(ctx context.Context, userID primitive.ObjectID, expID string, input dto.ExperienceDTO) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	foundIdx := -1
	for idx, val := range profile.Experience {
		if val.ID == expID {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return appErrors.NewNotFoundError("Experience item not found")
	}

	profile.Experience[foundIdx].Company = input.Company
	profile.Experience[foundIdx].Title = input.Title
	profile.Experience[foundIdx].Location = input.Location
	profile.Experience[foundIdx].StartDate = input.StartDate
	profile.Experience[foundIdx].EndDate = input.EndDate
	profile.Experience[foundIdx].Current = input.Current
	profile.Experience[foundIdx].Description = input.Description

	return s.profileRepo.Update(ctx, profile)
}

func (s *profileService) DeleteExperience(ctx context.Context, userID primitive.ObjectID, expID string) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	foundIdx := -1
	for idx, val := range profile.Experience {
		if val.ID == expID {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return appErrors.NewNotFoundError("Experience item not found")
	}

	// Remove item from slice
	profile.Experience = append(profile.Experience[:foundIdx], profile.Experience[foundIdx+1:]...)

	return s.profileRepo.Update(ctx, profile)
}

func (s *profileService) AddEducation(ctx context.Context, userID primitive.ObjectID, input dto.EducationDTO) (*domain.Education, error) {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	edu := domain.Education{
		ID:           uuid.New().String(),
		School:       input.School,
		Degree:       input.Degree,
		FieldOfStudy: input.FieldOfStudy,
		StartDate:    input.StartDate,
		EndDate:      input.EndDate,
		Description:  input.Description,
	}

	profile.Education = append(profile.Education, edu)
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to add education")
	}

	return &edu, nil
}

func (s *profileService) UpdateEducation(ctx context.Context, userID primitive.ObjectID, eduID string, input dto.EducationDTO) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	foundIdx := -1
	for idx, val := range profile.Education {
		if val.ID == eduID {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return appErrors.NewNotFoundError("Education item not found")
	}

	profile.Education[foundIdx].School = input.School
	profile.Education[foundIdx].Degree = input.Degree
	profile.Education[foundIdx].FieldOfStudy = input.FieldOfStudy
	profile.Education[foundIdx].StartDate = input.StartDate
	profile.Education[foundIdx].EndDate = input.EndDate
	profile.Education[foundIdx].Description = input.Description

	return s.profileRepo.Update(ctx, profile)
}

func (s *profileService) DeleteEducation(ctx context.Context, userID primitive.ObjectID, eduID string) error {
	profile, err := s.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	foundIdx := -1
	for idx, val := range profile.Education {
		if val.ID == eduID {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return appErrors.NewNotFoundError("Education item not found")
	}

	// Remove item from slice
	profile.Education = append(profile.Education[:foundIdx], profile.Education[foundIdx+1:]...)

	return s.profileRepo.Update(ctx, profile)
}
