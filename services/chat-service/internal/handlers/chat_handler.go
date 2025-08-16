package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/utils"
	"chat-service/internal/services"
	wsHub "chat-service/internal/websocket"
)

type ChatHandler struct {
	chatService services.ChatService
	hub         *wsHub.Hub
	upgrader    websocket.Upgrader
	logger      *zap.Logger
}

func NewChatHandler(chatService services.ChatService, hub *wsHub.Hub, logger *zap.Logger) *ChatHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}

	return &ChatHandler{
		chatService: chatService,
		hub:         hub,
		upgrader:    upgrader,
		logger:      logger,
	}
}

// Helper function to get user and tenant context
func (h *ChatHandler) getUserAndTenantContext(c *gin.Context) (string, string, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return "", "", fmt.Errorf("user ID not found in context")
	}

	tenantContext, err := middleware.GetTenantContext(c)
	if err != nil {
		return "", "", fmt.Errorf("tenant context not found: %w", err)
	}

	return userID, tenantContext.TenantID, nil
}

// HandleWebSocket upgrades HTTP connections to WebSocket for real-time chat
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := wsHub.NewClient(conn, h.hub, userID, tenantID)
	
	// Register the client with the hub
	select {
	case h.hub.GetRegisterChannel() <- client:
	default:
		h.logger.Error("Failed to register client - channel full")
		conn.Close()
		return
	}

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.WritePump()
	go client.ReadPump()
}

// CreateChatRoom handles POST /rooms
func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var req services.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	room, err := h.chatService.CreateChatRoom(userID, tenantID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create chat room")
		return
	}

	utils.CreatedResponse(c, room, "Chat room created successfully")
}

// ListChatRooms handles GET /rooms
func (h *ChatHandler) ListChatRooms(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	rooms, err := h.chatService.ListChatRooms(userID, tenantID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to list chat rooms")
		return
	}

	utils.SuccessResponse(c, rooms, "Chat rooms retrieved successfully")
}

// GetChatRoom handles GET /rooms/:id
func (h *ChatHandler) GetChatRoom(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		utils.BadRequestResponse(c, "Room ID is required")
		return
	}

	room, err := h.chatService.GetChatRoom(userID, tenantID, roomID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.NotFoundResponse(c, "Chat room not found")
		return
	}

	utils.SuccessResponse(c, room, "Chat room retrieved successfully")
}

// GetChatMessages handles GET /rooms/:id/messages
func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		utils.BadRequestResponse(c, "Room ID is required")
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}

	messages, err := h.chatService.GetChatMessages(userID, tenantID, roomID, limit, offset)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get chat messages")
		return
	}

	utils.SuccessResponse(c, messages, "Chat messages retrieved successfully")
}

// SendChatMessage handles POST /rooms/:id/messages
func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		utils.BadRequestResponse(c, "Room ID is required")
		return
	}

	var req services.SendChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	message, err := h.chatService.SendChatMessage(userID, tenantID, roomID, &req)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to send chat message")
		return
	}

	// Broadcast the message to WebSocket connections
	if messageBytes, err := message.ToJSON(); err == nil {
		h.hub.BroadcastToRoom(roomID, messageBytes)
	}

	utils.CreatedResponse(c, message, "Message sent successfully")
}

// GetRoomStats handles GET /rooms/:id/stats
func (h *ChatHandler) GetRoomStats(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		utils.BadRequestResponse(c, "Room ID is required")
		return
	}

	// Get online count from WebSocket hub
	onlineCount := h.hub.GetRoomClients(roomID)

	// Get other stats from service
	stats, err := h.chatService.GetRoomStats(userID, tenantID, roomID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get room stats")
		return
	}

	// Add online count to stats
	stats["online_users"] = onlineCount

	utils.SuccessResponse(c, stats, "Room statistics retrieved successfully")
}