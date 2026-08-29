package service

import (
	"context"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService interface {
	Create(ctx context.Context, userID primitive.ObjectID, title, message, notifType string) error
	GetMyNotifications(ctx context.Context, userID primitive.ObjectID) ([]dto.NotificationResponseDTO, error)
	MarkAsRead(ctx context.Context, notifID primitive.ObjectID) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
	}
}

func (s *notificationService) Create(ctx context.Context, userID primitive.ObjectID, title, message, notifType string) error {
	notif := &domain.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		IsRead:  false,
	}
	err := s.notifRepo.Create(ctx, notif)
	if err != nil {
		return appErrors.NewInternalError("Failed to create notification: " + err.Error())
	}
	return nil
}

func (s *notificationService) GetMyNotifications(ctx context.Context, userID primitive.ObjectID) ([]dto.NotificationResponseDTO, error) {
	notifs, err := s.notifRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to retrieve notifications: " + err.Error())
	}

	response := make([]dto.NotificationResponseDTO, 0)
	for _, notif := range notifs {
		response = append(response, dto.NotificationResponseDTO{
			ID:        notif.ID.Hex(),
			UserID:    notif.UserID.Hex(),
			Title:     notif.Title,
			Message:   notif.Message,
			Type:      notif.Type,
			IsRead:    notif.IsRead,
			CreatedAt: notif.CreatedAt,
		})
	}

	return response, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, notifID primitive.ObjectID) error {
	err := s.notifRepo.MarkAsRead(ctx, notifID)
	if err != nil {
		return appErrors.NewInternalError("Failed to update notification read status: " + err.Error())
	}
	return nil
}
