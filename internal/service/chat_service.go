package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService interface {
	GetOrCreateRoom(ctx context.Context, jobID, seekerID primitive.ObjectID) (*domain.ChatRoom, error)
	ListUserRooms(ctx context.Context, userID primitive.ObjectID, role string) ([]dto.ChatRoomResponseDTO, error)
	GetRoomMessages(ctx context.Context, roomID, userID primitive.ObjectID, limit, offset int64) ([]domain.ChatMessage, error)
	HandleWebSocket(roomID, userID primitive.ObjectID, conn *websocket.Conn) error
}

type wsMessagePayload struct {
	Message string `json:"message"`
}

type Client struct {
	UserID primitive.ObjectID
	Conn   *websocket.Conn
}

type chatService struct {
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository
	jobRepo     repository.JobRepository
	appRepo     repository.ApplicationRepository
	companyRepo repository.CompanyRepository
	profileRepo repository.ProfileRepository

	// WebSocket connection management
	clientsMu         sync.RWMutex
	activeRoomClients map[primitive.ObjectID][]*Client
}

func NewChatService(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	appRepo repository.ApplicationRepository,
	companyRepo repository.CompanyRepository,
	profileRepo repository.ProfileRepository,
) ChatService {
	return &chatService{
		chatRepo:          chatRepo,
		userRepo:          userRepo,
		jobRepo:           jobRepo,
		appRepo:           appRepo,
		companyRepo:       companyRepo,
		profileRepo:       profileRepo,
		activeRoomClients: make(map[primitive.ObjectID][]*Client),
	}
}

func (s *chatService) GetOrCreateRoom(ctx context.Context, jobID, seekerID primitive.ObjectID) (*domain.ChatRoom, error) {
	// 1. Verify Job exists and get employer user ID
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Job listing not found")
	}

	company, err := s.companyRepo.FindByID(ctx, job.CompanyID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Company not found")
	}

	employerID := company.OwnerUserID

	// 2. Check if chat room already exists
	room, err := s.chatRepo.FindRoomByParticipants(ctx, jobID, seekerID, employerID)
	if err == nil && room != nil {
		return room, nil
	}

	// 3. Verify Seeker has an active job application for this job
	app, err := s.appRepo.FindByJobIDAndUserID(ctx, jobID, seekerID)
	if err != nil || app == nil {
		return nil, appErrors.NewForbiddenError("You can only chat with the employer if you have applied for this job")
	}

	// 4. Create new chat room
	newRoom := &domain.ChatRoom{
		ID:         primitive.NewObjectID(),
		JobID:      jobID,
		SeekerID:   seekerID,
		EmployerID: employerID,
	}

	err = s.chatRepo.CreateRoom(ctx, newRoom)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to initialize chat room: " + err.Error())
	}

	return newRoom, nil
}

func (s *chatService) ListUserRooms(ctx context.Context, userID primitive.ObjectID, role string) ([]dto.ChatRoomResponseDTO, error) {
	rooms, err := s.chatRepo.ListRoomsForUser(ctx, userID, role)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to list chat rooms")
	}

	responseDTOs := make([]dto.ChatRoomResponseDTO, 0, len(rooms))

	for _, room := range rooms {
		// Determine "other user"
		var otherUserID primitive.ObjectID
		if role == "company" {
			otherUserID = room.SeekerID
		} else {
			otherUserID = room.EmployerID
		}

		otherUser, err := s.userRepo.FindByID(ctx, otherUserID)
		if err != nil {
			continue // Skip if user not found
		}

		// Find other user name & avatar details
		otherName := otherUser.Name
		otherAvatar := ""

		if otherUser.Role == "company" {
			// Find company avatar (logo)
			if otherUser.CompanyID != nil {
				comp, err := s.companyRepo.FindByID(ctx, *otherUser.CompanyID)
				if err == nil {
					otherAvatar = comp.LogoURL
				}
			}
		} else {
			// Find seeker avatar
			prof, err := s.profileRepo.FindByUserID(ctx, otherUser.ID)
			if err == nil {
				otherAvatar = prof.AvatarURL
			}
		}

		// Last message details
		lastMsgSnippet := ""
		lastMsg, err := s.chatRepo.GetLastMessage(ctx, room.ID)
		if err == nil && lastMsg != nil {
			lastMsgSnippet = lastMsg.Message
		}

		// Unread count
		unreadCount, _ := s.chatRepo.GetUnreadCount(ctx, room.ID, userID)

		// Job details
		jobTitle := ""
		job, err := s.jobRepo.FindByID(ctx, room.JobID)
		if err == nil {
			jobTitle = job.Title
		}

		responseDTOs = append(responseDTOs, dto.ChatRoomResponseDTO{
			ID:              room.ID.Hex(),
			JobID:           room.JobID.Hex(),
			JobTitle:        jobTitle,
			OtherUserID:     otherUserID.Hex(),
			OtherUserName:   otherName,
			OtherUserAvatar: otherAvatar,
			LastMessage:     lastMsgSnippet,
			UnreadCount:     unreadCount,
			UpdatedAt:       room.UpdatedAt,
		})
	}

	return responseDTOs, nil
}

func (s *chatService) GetRoomMessages(ctx context.Context, roomID, userID primitive.ObjectID, limit, offset int64) ([]domain.ChatMessage, error) {
	// Verify user is participant in the room
	room, err := s.chatRepo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Chat room not found")
	}

	if room.SeekerID != userID && room.EmployerID != userID {
		return nil, appErrors.NewForbiddenError("Access denied to this chat room")
	}

	// Mark incoming messages as read
	_ = s.chatRepo.MarkMessagesAsRead(ctx, roomID, userID)

	messages, err := s.chatRepo.GetMessagesByRoom(ctx, roomID, limit, offset)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to fetch messages")
	}

	return messages, nil
}

func (s *chatService) HandleWebSocket(roomID, userID primitive.ObjectID, conn *websocket.Conn) error {
	// 1. Verify user is participant in this room
	ctx := context.Background()
	room, err := s.chatRepo.FindRoomByID(ctx, roomID)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "Chat room not found"})
		conn.Close()
		return errors.New("chat room not found")
	}

	if room.SeekerID != userID && room.EmployerID != userID {
		conn.WriteJSON(map[string]string{"error": "Access denied to chat room"})
		conn.Close()
		return errors.New("forbidden chat room access")
	}

	// 2. Register connection
	client := s.registerClient(roomID, userID, conn)
	defer s.unregisterClient(roomID, client)

	// 3. Mark messages as read on connect
	_ = s.chatRepo.MarkMessagesAsRead(ctx, roomID, userID)

	// 4. Message reading loop
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			// Disconnection / Read error
			break
		}

		var payload wsMessagePayload
		if err := json.Unmarshal(messageBytes, &payload); err != nil {
			conn.WriteJSON(map[string]string{"error": "Invalid payload format"})
			continue
		}

		if payload.Message == "" {
			continue
		}

		// Save message to database
		chatMsg := &domain.ChatMessage{
			ID:        primitive.NewObjectID(),
			RoomID:    roomID,
			SenderID:  userID,
			Message:   payload.Message,
			IsRead:    false,
			CreatedAt: time.Now(),
		}

		err = s.chatRepo.SaveMessage(ctx, chatMsg)
		if err != nil {
			log.Printf("Failed to save chat message: %v", err)
			conn.WriteJSON(map[string]string{"error": "Failed to save message"})
			continue
		}

		// Broadcast message to room
		s.broadcastToRoom(roomID, chatMsg)
	}

	return nil
}

func (s *chatService) registerClient(roomID, userID primitive.ObjectID, conn *websocket.Conn) *Client {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	client := &Client{
		UserID: userID,
		Conn:   conn,
	}

	s.activeRoomClients[roomID] = append(s.activeRoomClients[roomID], client)
	return client
}

func (s *chatService) unregisterClient(roomID primitive.ObjectID, client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clients := s.activeRoomClients[roomID]
	for i, c := range clients {
		if c == client {
			// Close WebSocket connection cleanly
			c.Conn.Close()
			// Remove from slice
			s.activeRoomClients[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(s.activeRoomClients[roomID]) == 0 {
		delete(s.activeRoomClients, roomID)
	}
}

func (s *chatService) broadcastToRoom(roomID primitive.ObjectID, msg *domain.ChatMessage) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	clients := s.activeRoomClients[roomID]
	for _, client := range clients {
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Failed to broadcast message to user %s: %v", client.UserID.Hex(), err)
		}
	}
}
