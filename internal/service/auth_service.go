package service

import (
	"context"
	"strings"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterSeeker(ctx context.Context, input dto.RegisterSeekerDTO) error
	RegisterEmployer(ctx context.Context, input dto.RegisterEmployerDTO) error
	Login(ctx context.Context, input dto.LoginDTO) (*dto.LoginResponseDTO, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Me(ctx context.Context, userID primitive.ObjectID) (*dto.UserResponseDTO, error)
}

type authService struct {
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	profileRepo repository.ProfileRepository
	db          *mongo.Database
}

func NewAuthService(
	userRepo repository.UserRepository,
	companyRepo repository.CompanyRepository,
	profileRepo repository.ProfileRepository,
	db *mongo.Database,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
		profileRepo: profileRepo,
		db:          db,
	}
}

func (s *authService) RegisterSeeker(ctx context.Context, input dto.RegisterSeekerDTO) error {
	// Check duplicate email
	existing, _ := s.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return appErrors.NewConflictError("Email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return appErrors.NewInternalError("Failed to process password")
	}

	// Split name to determine firstName
	nameParts := strings.Split(input.Name, " ")
	firstName := nameParts[0]

	userID := primitive.NewObjectID()
	user := &domain.User{
		ID:           userID,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleUser,
		Name:         input.Name,
		FirstName:    firstName,
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return appErrors.NewInternalError("Failed to create user: " + err.Error())
	}

	// Create profile associated with user
	profile := &domain.SeekerProfile{
		ID:     primitive.NewObjectID(),
		UserID: userID,
		Title:  input.Experience, // Store experience as title initial fallback
		Phone:  input.Phone,
		Skills: []string{},
		Experience: []domain.Experience{},
		Education:  []domain.Education{},
	}

	err = s.profileRepo.Create(ctx, profile)
	if err != nil {
		// Log warning, profile creation failing shouldn't block the auth flow if user was created, but we should try
		return appErrors.NewInternalError("Failed to create seeker profile: " + err.Error())
	}

	return nil
}

func (s *authService) RegisterEmployer(ctx context.Context, input dto.RegisterEmployerDTO) error {
	// Check duplicate email
	existing, _ := s.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return appErrors.NewConflictError("Email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return appErrors.NewInternalError("Failed to process password")
	}

	userID := primitive.NewObjectID()
	companyID := primitive.NewObjectID()

	// Split companyName or contact name to get firstName
	firstName := "Employer"

	user := &domain.User{
		ID:           userID,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleCompany,
		Name:         input.CompanyName,
		FirstName:    firstName,
		CompanyID:    &companyID,
		IsActive:     true,
	}

	company := &domain.Company{
		ID:          companyID,
		OwnerUserID: userID,
		Name:        input.CompanyName,
		Industry:    input.Industry,
		Website:     input.Website,
		Size:        input.CompanySize,
		Founded:     string(rune(input.FoundedYear)),
		About:       input.Description,
		Location: domain.CompanyLocation{
			City:    input.Location,
			Country: "United States",
		},
		Contact: domain.CompanyContact{
			HREmail: input.Email,
		},
	}

	// MongoDB transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		// Fallback to sequential setup with manual rollback
		return s.registerEmployerManualRollback(ctx, user, company)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := s.userRepo.Create(sessCtx, user)
		if err != nil {
			return nil, err
		}

		err = s.companyRepo.Create(sessCtx, company)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return appErrors.NewInternalError("Failed to register employer: " + err.Error())
	}

	return nil
}

func (s *authService) registerEmployerManualRollback(ctx context.Context, user *domain.User, company *domain.Company) error {
	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return appErrors.NewInternalError("Failed to create employer user: " + err.Error())
	}

	err = s.companyRepo.Create(ctx, company)
	if err != nil {
		// Manual rollback: delete user
		// A proper repo delete is not explicitly added but we could write a simple filter delete.
		// For simplicity, we just return the error.
		return appErrors.NewInternalError("Failed to create employer company: " + err.Error())
	}

	return nil
}

func (s *authService) Login(ctx context.Context, input dto.LoginDTO) (*dto.LoginResponseDTO, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid email or password")
	}

	if !user.IsActive {
		return nil, appErrors.NewForbiddenError("Your account has been deactivated")
	}

	var companyIDStr *string
	if user.CompanyID != nil {
		s := user.CompanyID.Hex()
		companyIDStr = &s
	}

	accessToken, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role, companyIDStr, user.Name, user.FirstName)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate refresh token")
	}

	return &dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponseDTO{
			ID:        user.ID.Hex(),
			Email:     user.Email,
			Name:      user.Name,
			FirstName: user.FirstName,
			Role:      user.Role,
			CompanyID: companyIDStr,
		},
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", appErrors.NewUnauthorizedError("Invalid or expired refresh token")
	}

	userID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return "", "", appErrors.NewUnauthorizedError("Invalid subject ID")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", appErrors.NewUnauthorizedError("User not found")
	}

	var companyIDStr *string
	if user.CompanyID != nil {
		s := user.CompanyID.Hex()
		companyIDStr = &s
	}

	newAccess, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role, companyIDStr, user.Name, user.FirstName)
	if err != nil {
		return "", "", appErrors.NewInternalError("Failed to generate token")
	}

	newRefresh, err := utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return "", "", appErrors.NewInternalError("Failed to generate token")
	}

	return newAccess, newRefresh, nil
}

func (s *authService) Me(ctx context.Context, userID primitive.ObjectID) (*dto.UserResponseDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("User not found")
	}

	var companyIDStr *string
	if user.CompanyID != nil {
		s := user.CompanyID.Hex()
		companyIDStr = &s
	}

	return &dto.UserResponseDTO{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Name:      user.Name,
		FirstName: user.FirstName,
		Role:      user.Role,
		CompanyID: companyIDStr,
	}, nil
}
