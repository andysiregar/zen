package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"chat-service/internal/models"
	"chat-service/internal/repositories"
)

type ChatService interface {
	// Room management
	CreateChatRoom(userID, tenantID string, req *CreateChatRoomRequest) (*models.ChatRoomResponse, error)
	GetChatRoom(userID, tenantID, roomID string) (*models.ChatRoomResponse, error)
	ListChatRooms(userID, tenantID string) ([]*models.ChatRoomResponse, error)
	
	// Messaging
	SendChatMessage(userID, tenantID, roomID string, req *SendChatMessageRequest) (*models.ChatMessageResponse, error)
	GetChatMessages(userID, tenantID, roomID string, limit, offset int) ([]*models.ChatMessageResponse, error)
	
	// Statistics
	GetRoomStats(userID, tenantID, roomID string) (map[string]interface{}, error)
}

type CreateChatRoomRequest struct {
	Name        string   `json:"name" binding:"required,max=255"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // "direct", "group", "support"
	IsPrivate   bool     `json:"is_private"`
	Members     []string `json:"members"` // User IDs
}

type SendChatMessageRequest struct {
	Content     string                 `json:"content" binding:"required"`
	MessageType string                 `json:"message_type"` // "text", "file", "system"
	Metadata    map[string]interface{} `json:"metadata"`
}


type chatService struct {
	repo   repositories.ChatRepository
	logger *zap.Logger
}

func NewChatService(repo repositories.ChatRepository, logger *zap.Logger) ChatService {
	return &chatService{
		repo:   repo,
		logger: logger,
	}
}

func (s *chatService) CreateChatRoom(userID, tenantID string, req *CreateChatRoomRequest) (*models.ChatRoomResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, errors.New("room name is required")
	}

	// Set default type
	if req.Type == "" {
		req.Type = "group"
	}

	room := &models.ChatRoom{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.repo.CreateChatRoom(tenantID, room)
	if err != nil {
		s.logger.Error("Failed to create chat room", zap.Error(err))
		return nil, err
	}

	// Add creator as member
	if err := s.repo.AddRoomMember(tenantID, room.ID, userID); err != nil {
		s.logger.Warn("Failed to add creator as room member", zap.Error(err))
	}

	// Add other members
	for _, memberID := range req.Members {
		if memberID != userID {
			if err := s.repo.AddRoomMember(tenantID, room.ID, memberID); err != nil {
				s.logger.Warn("Failed to add room member", zap.String("member_id", memberID), zap.Error(err))
			}
		}
	}

	return s.roomToResponse(room), nil
}

func (s *chatService) GetChatRoom(userID, tenantID, roomID string) (*models.ChatRoomResponse, error) {
	// Check if user has access to the room
	if !s.userCanAccessRoom(userID, tenantID, roomID) {
		return nil, errors.New("access denied")
	}

	room, err := s.repo.GetChatRoom(tenantID, roomID)
	if err != nil {
		return nil, err
	}

	return s.roomToResponse(room), nil
}

func (s *chatService) ListChatRooms(userID, tenantID string) ([]*models.ChatRoomResponse, error) {
	rooms, err := s.repo.GetUserChatRooms(tenantID, userID)
	if err != nil {
		return nil, err
	}

	var responses []*models.ChatRoomResponse
	for _, room := range rooms {
		responses = append(responses, s.roomToResponse(room))
	}

	return responses, nil
}

func (s *chatService) SendChatMessage(userID, tenantID, roomID string, req *SendChatMessageRequest) (*models.ChatMessageResponse, error) {
	// Check if user has access to the room
	if !s.userCanAccessRoom(userID, tenantID, roomID) {
		return nil, errors.New("access denied")
	}

	// Set default message type
	messageType := req.MessageType
	if messageType == "" {
		messageType = "text"
	}

	message := &models.ChatMessage{
		ID:          uuid.New().String(),
		RoomID:      roomID,
		UserID:      userID,
		Content:     req.Content,
		MessageType: messageType,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateChatMessage(tenantID, message)
	if err != nil {
		s.logger.Error("Failed to create chat message", zap.Error(err))
		return nil, err
	}

	return s.messageToResponse(message), nil
}

func (s *chatService) GetChatMessages(userID, tenantID, roomID string, limit, offset int) ([]*models.ChatMessageResponse, error) {
	// Check if user has access to the room
	if !s.userCanAccessRoom(userID, tenantID, roomID) {
		return nil, errors.New("access denied")
	}

	messages, err := s.repo.GetChatMessages(tenantID, roomID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []*models.ChatMessageResponse
	for _, message := range messages {
		responses = append(responses, s.messageToResponse(message))
	}

	return responses, nil
}

func (s *chatService) GetRoomStats(userID, tenantID, roomID string) (map[string]interface{}, error) {
	// Check if user has access to the room
	if !s.userCanAccessRoom(userID, tenantID, roomID) {
		return nil, errors.New("access denied")
	}

	stats := make(map[string]interface{})

	// Get message count
	messageCount, err := s.repo.GetRoomMessageCount(tenantID, roomID)
	if err != nil {
		return nil, err
	}
	stats["message_count"] = messageCount

	// Get member count
	memberCount, err := s.repo.GetRoomMemberCount(tenantID, roomID)
	if err != nil {
		return nil, err
	}
	stats["member_count"] = memberCount

	// Get latest message timestamp
	latestMessage, err := s.repo.GetLatestMessage(tenantID, roomID)
	if err == nil && latestMessage != nil {
		stats["latest_message_at"] = latestMessage.CreatedAt
	}

	return stats, nil
}

// Helper methods
func (s *chatService) userCanAccessRoom(userID, tenantID, roomID string) bool {
	return s.repo.IsRoomMember(tenantID, roomID, userID)
}

func (s *chatService) roomToResponse(room *models.ChatRoom) *models.ChatRoomResponse {
	return &models.ChatRoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Type:        room.Type,
		IsPrivate:   room.IsPrivate,
		CreatedBy:   room.CreatedBy,
		CreatedAt:   room.CreatedAt,
		UpdatedAt:   room.UpdatedAt,
	}
}

func (s *chatService) messageToResponse(message *models.ChatMessage) *models.ChatMessageResponse {
	return &models.ChatMessageResponse{
		ID:          message.ID,
		RoomID:      message.RoomID,
		UserID:      message.UserID,
		Content:     message.Content,
		MessageType: message.MessageType,
		Metadata:    message.Metadata,
		CreatedAt:   message.CreatedAt,
	}
}

