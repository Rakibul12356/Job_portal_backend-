package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rakib/job-portal-api/internal/dto"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
	"github.com/rakib/job-portal-api/internal/pkg/response"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
	"github.com/rakib/job-portal-api/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for API compatibility
		return true
	},
}

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) CreateOrOpenRoom(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	var input dto.CreateChatRoomDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, appErrors.NewValidationError("Job ID is required"))
		return
	}

	jobID, err := primitive.ObjectIDFromHex(input.JobID)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid job ID format"))
		return
	}

	room, err := h.chatService.GetOrCreateRoom(c.Request.Context(), jobID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, "Chat room retrieved successfully", room)
}

func (h *ChatHandler) ListUserRooms(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	role, exists := c.Get("role")
	if !exists {
		role = "user"
	}

	rooms, err := h.chatService.ListUserRooms(c.Request.Context(), userID, role.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, rooms)
}

func (h *ChatHandler) GetRoomMessages(c *gin.Context) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("Authentication required"))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(c, appErrors.NewUnauthorizedError("Invalid user identity"))
		return
	}

	roomIDStr := c.Param("roomId")
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid room ID format"))
		return
	}

	// Parse pagination query params
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		offset = 0
	}

	messages, err := h.chatService.GetRoomMessages(c.Request.Context(), roomID, userID, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, messages)
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID format"})
		return
	}

	// Read token from query parameter since WebSockets don't easily allow headers in browsers
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token is required"})
		return
	}

	claims, err := utils.VerifyAccessToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user identity"})
		return
	}

	// Upgrade HTTP connection to WebSocket protocol
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket connection: %v", err)
		return
	}

	err = h.chatService.HandleWebSocket(roomID, userID, conn)
	if err != nil {
		log.Printf("WebSocket connection ended: %v", err)
	}
}
