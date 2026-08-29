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
	HandleGlobalWebSocket(userID primitive.ObjectID, conn *websocket.Conn) error
}

type wsInboundPayload struct {
	Type     string `json:"type"`     // "message" | "typing" | "seen"
	Message  string `json:"message"`  // for type "message"
	IsTyping bool   `json:"isTyping"` // for type "typing"
}

type wsOutboundPayload struct {
	Type       string               `json:"type"`                 // "message" | "message_status" | "typing" | "user_status"
	Message    *domain.ChatMessage  `json:"message,omitempty"`    // for type "message"
	RoomID     string               `json:"roomId,omitempty"`     // for type "message_status"
	Status     string               `json:"status,omitempty"`     // for type "message_status" / "user_status"
	SeenBy     string               `json:"seenBy,omitempty"`     // for type "message_status"
	UserID     string               `json:"userId,omitempty"`     // for type "user_status"
	SenderID   string               `json:"senderId,omitempty"`   // for type "typing"
	SenderName string               `json:"senderName,omitempty"` // for type "typing"
	IsTyping   bool                 `json:"isTyping,omitempty"`   // for type "typing"
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
	clientsMu           sync.RWMutex
	activeRoomClients   map[primitive.ObjectID][]*Client
	activeGlobalClients map[primitive.ObjectID][]*Client // UserID -> []*Client
	activeUsers         map[primitive.ObjectID]int       // UserID -> connection count
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
		chatRepo:            chatRepo,
		userRepo:            userRepo,
		jobRepo:             jobRepo,
		appRepo:             appRepo,
		companyRepo:         companyRepo,
		profileRepo:         profileRepo,
		activeRoomClients:   make(map[primitive.ObjectID][]*Client),
		activeGlobalClients: make(map[primitive.ObjectID][]*Client),
		activeUsers:         make(map[primitive.ObjectID]int),
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

	// 3. Mark messages as read on connect (only other user's messages to this user)
	_ = s.chatRepo.MarkMessagesAsRead(ctx, roomID, userID)

	// Broadcast seen status to the room and globally
	seenPayload := wsOutboundPayload{
		Type:   "message_status",
		RoomID: roomID.Hex(),
		Status: "seen",
		SeenBy: userID.Hex(),
	}
	s.broadcastEventToRoom(roomID, seenPayload)
	s.broadcastEventGlobally(room.SeekerID, room.EmployerID, seenPayload)

	// Update B's pending messages to B (this user) as delivered in all of B's rooms
	affectedRooms, err := s.chatRepo.MarkAllMessagesAsDelivered(ctx, userID)
	if err == nil && len(affectedRooms) > 0 {
		for _, rID := range affectedRooms {
			rDetails, err := s.chatRepo.FindRoomByID(ctx, rID)
			if err == nil && rDetails != nil {
				delivPayload := wsOutboundPayload{
					Type:   "message_status",
					RoomID: rID.Hex(),
					Status: "delivered",
				}
				s.broadcastEventToRoom(rID, delivPayload)
				s.broadcastEventGlobally(rDetails.SeekerID, rDetails.EmployerID, delivPayload)
			}
		}
	}

	// Send initial status of the other user to the connecting client
	var otherUserID primitive.ObjectID
	if room.SeekerID == userID {
		otherUserID = room.EmployerID
	} else {
		otherUserID = room.SeekerID
	}
	initialStatus := "offline"
	if s.isUserOnline(otherUserID) {
		initialStatus = "online"
	}
	_ = conn.WriteJSON(wsOutboundPayload{
		Type:   "user_status",
		UserID: otherUserID.Hex(),
		Status: initialStatus,
	})

	// 4. Message reading loop
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var payload wsInboundPayload
		if err := json.Unmarshal(messageBytes, &payload); err != nil {
			conn.WriteJSON(map[string]string{"error": "Invalid payload format"})
			continue
		}

		// Support backward compatibility (if type is empty but message is not)
		if payload.Type == "" && payload.Message != "" {
			payload.Type = "message"
		}

		switch payload.Type {
		case "message":
			if payload.Message == "" {
				continue
			}

			// Determine recipient ID
			var recipientID primitive.ObjectID
			if room.SeekerID == userID {
				recipientID = room.EmployerID
			} else {
				recipientID = room.SeekerID
			}

			// Determine initial message status
			status := "sent"
			if s.isUserOnline(recipientID) {
				status = "delivered"
			}

			// Save message to database
			chatMsg := &domain.ChatMessage{
				ID:        primitive.NewObjectID(),
				RoomID:    roomID,
				SenderID:  userID,
				Message:   payload.Message,
				IsRead:    false,
				Status:    status,
				CreatedAt: time.Now(),
			}

			err = s.chatRepo.SaveMessage(ctx, chatMsg)
			if err != nil {
				log.Printf("Failed to save chat message: %v", err)
				conn.WriteJSON(map[string]string{"error": "Failed to save message"})
				continue
			}

			// Broadcast message
			msgPayload := wsOutboundPayload{
				Type:    "message",
				Message: chatMsg,
			}

			if status == "delivered" {
				// Broadcast to room (both participants)
				s.broadcastEventToRoom(roomID, msgPayload)

				// Reaches the recipient globally if they are online but not in the active room
				if !s.isUserInRoom(roomID, recipientID) {
					s.sendToGlobalClient(recipientID, msgPayload)
				}
				// Reaches the sender globally (if they have multiple connections)
				if !s.isUserInRoom(roomID, userID) {
					s.sendToGlobalClient(userID, msgPayload)
				}
			} else {
				// Recipient is offline - only send/broadcast to the sender
				// Send to sender's active room connection
				s.sendToRoomClient(roomID, userID, msgPayload)
				// Send to sender's global connections (if not in room)
				if !s.isUserInRoom(roomID, userID) {
					s.sendToGlobalClient(userID, msgPayload)
				}
			}

		case "typing":
			// Fetch sender's name to display
			user, err := s.userRepo.FindByID(ctx, userID)
			senderName := "Someone"
			if err == nil && user != nil {
				senderName = user.Name
			}

			typingPayload := wsOutboundPayload{
				Type:       "typing",
				SenderID:   userID.Hex(),
				SenderName: senderName,
				IsTyping:   payload.IsTyping,
			}
			// Broadcast typing indicator to other clients in the room
			s.clientsMu.RLock()
			clients := s.activeRoomClients[roomID]
			for _, client := range clients {
				if client.UserID != userID {
					_ = client.Conn.WriteJSON(typingPayload)
				}
			}
			s.clientsMu.RUnlock()

		case "seen":
			// Mark messages as read / seen in database (only other user's messages to reader B)
			_ = s.chatRepo.MarkMessagesAsRead(ctx, roomID, userID)

			// Broadcast seen status to room and globally
			seenPayload := wsOutboundPayload{
				Type:   "message_status",
				RoomID: roomID.Hex(),
				Status: "seen",
				SeenBy: userID.Hex(),
			}
			s.broadcastEventToRoom(roomID, seenPayload)
			s.broadcastEventGlobally(room.SeekerID, room.EmployerID, seenPayload)
		}
	}

	return nil
}

func (s *chatService) HandleGlobalWebSocket(userID primitive.ObjectID, conn *websocket.Conn) error {
	ctx := context.Background()

	// 1. Register global connection
	client := s.registerGlobalClient(userID, conn)
	defer s.unregisterGlobalClient(userID, client)

	// 2. delivered synchronization scan on connect
	affectedRooms, err := s.chatRepo.MarkAllMessagesAsDelivered(ctx, userID)
	if err == nil && len(affectedRooms) > 0 {
		for _, rID := range affectedRooms {
			rDetails, err := s.chatRepo.FindRoomByID(ctx, rID)
			if err == nil && rDetails != nil {
				delivPayload := wsOutboundPayload{
					Type:   "message_status",
					RoomID: rID.Hex(),
					Status: "delivered",
				}
				s.broadcastEventToRoom(rID, delivPayload)
				s.broadcastEventGlobally(rDetails.SeekerID, rDetails.EmployerID, delivPayload)
			}
		}
	}

	// 3. Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
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

	s.activeUsers[userID]++
	if s.activeUsers[userID] == 1 {
		go s.broadcastUserStatus(userID, "online")
	}

	return client
}

func (s *chatService) unregisterClient(roomID primitive.ObjectID, client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clients := s.activeRoomClients[roomID]
	for i, c := range clients {
		if c == client {
			c.Conn.Close()
			s.activeRoomClients[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(s.activeRoomClients[roomID]) == 0 {
		delete(s.activeRoomClients, roomID)
	}

	s.activeUsers[client.UserID]--
	if s.activeUsers[client.UserID] <= 0 {
		delete(s.activeUsers, client.UserID)
		go s.broadcastUserStatus(client.UserID, "offline")
	}
}

func (s *chatService) registerGlobalClient(userID primitive.ObjectID, conn *websocket.Conn) *Client {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	client := &Client{
		UserID: userID,
		Conn:   conn,
	}

	s.activeGlobalClients[userID] = append(s.activeGlobalClients[userID], client)

	s.activeUsers[userID]++
	if s.activeUsers[userID] == 1 {
		go s.broadcastUserStatus(userID, "online")
	}

	return client
}

func (s *chatService) unregisterGlobalClient(userID primitive.ObjectID, client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clients := s.activeGlobalClients[userID]
	for i, c := range clients {
		if c == client {
			c.Conn.Close()
			s.activeGlobalClients[userID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(s.activeGlobalClients[userID]) == 0 {
		delete(s.activeGlobalClients, userID)
	}

	s.activeUsers[userID]--
	if s.activeUsers[userID] <= 0 {
		delete(s.activeUsers, userID)
		go s.broadcastUserStatus(userID, "offline")
	}
}

func (s *chatService) isUserOnline(userID primitive.ObjectID) bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.activeUsers[userID] > 0
}

func (s *chatService) isUserInRoom(roomID, userID primitive.ObjectID) bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	clients := s.activeRoomClients[roomID]
	for _, c := range clients {
		if c.UserID == userID {
			return true
		}
	}
	return false
}

func (s *chatService) broadcastUserStatus(userID primitive.ObjectID, status string) {
	ctx := context.Background()
	rooms, err := s.chatRepo.ListRoomsForUser(ctx, userID, "company")
	if err == nil {
		s.broadcastStatusToRooms(rooms, userID, status)
	}
	roomsUser, err := s.chatRepo.ListRoomsForUser(ctx, userID, "user")
	if err == nil {
		s.broadcastStatusToRooms(roomsUser, userID, status)
	}
}

func (s *chatService) broadcastStatusToRooms(rooms []domain.ChatRoom, userID primitive.ObjectID, status string) {
	payload := wsOutboundPayload{
		Type:   "user_status",
		UserID: userID.Hex(),
		Status: status,
	}
	for _, room := range rooms {
		s.broadcastEventToRoom(room.ID, payload)
		s.broadcastEventGlobally(room.SeekerID, room.EmployerID, payload)
	}
}

func (s *chatService) broadcastEventToRoom(roomID primitive.ObjectID, payload wsOutboundPayload) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	clients := s.activeRoomClients[roomID]
	for _, client := range clients {
		_ = client.Conn.WriteJSON(payload)
	}
}

func (s *chatService) sendToRoomClient(roomID, userID primitive.ObjectID, payload wsOutboundPayload) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	clients := s.activeRoomClients[roomID]
	for _, client := range clients {
		if client.UserID == userID {
			_ = client.Conn.WriteJSON(payload)
		}
	}
}

func (s *chatService) broadcastEventGlobally(seekerID, employerID primitive.ObjectID, payload wsOutboundPayload) {
	s.sendToGlobalClient(seekerID, payload)
	s.sendToGlobalClient(employerID, payload)
}

func (s *chatService) sendToGlobalClient(userID primitive.ObjectID, payload wsOutboundPayload) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	globalConns := s.activeGlobalClients[userID]
	for _, client := range globalConns {
		_ = client.Conn.WriteJSON(payload)
	}
}
