package repositories

import (
	"github.com/zen/shared/pkg/database"
	"chat-service/internal/models"
)

type ChatRepository interface {
	// Room operations
	CreateChatRoom(tenantID string, room *models.ChatRoom) error
	GetChatRoom(tenantID, roomID string) (*models.ChatRoom, error)
	GetUserChatRooms(tenantID, userID string) ([]*models.ChatRoom, error)
	
	// Room member operations
	AddRoomMember(tenantID, roomID, userID string) error
	RemoveRoomMember(tenantID, roomID, userID string) error
	IsRoomMember(tenantID, roomID, userID string) bool
	GetRoomMemberCount(tenantID, roomID string) (int64, error)
	
	// Message operations
	CreateChatMessage(tenantID string, message *models.ChatMessage) error
	GetChatMessages(tenantID, roomID string, limit, offset int) ([]*models.ChatMessage, error)
	GetRoomMessageCount(tenantID, roomID string) (int64, error)
	GetLatestMessage(tenantID, roomID string) (*models.ChatMessage, error)
}

type chatRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewChatRepository(tenantDBManager *database.TenantDatabaseManager) ChatRepository {
	return &chatRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *chatRepository) CreateChatRoom(tenantID string, room *models.ChatRoom) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// For now, store as JSON in a simple table structure
	// In a real implementation, you'd have proper chat_rooms table
	result := db.Exec(`
		INSERT INTO chat_rooms (id, name, description, type, is_private, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, room.ID, room.Name, room.Description, room.Type, room.IsPrivate, room.CreatedBy, room.CreatedAt, room.UpdatedAt)

	return result.Error
}

func (r *chatRepository) GetChatRoom(tenantID, roomID string) (*models.ChatRoom, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var room models.ChatRoom
	err = db.Raw(`
		SELECT id, name, description, type, is_private, created_by, created_at, updated_at
		FROM chat_rooms WHERE id = ?
	`, roomID).Scan(&room).Error

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *chatRepository) GetUserChatRooms(tenantID, userID string) ([]*models.ChatRoom, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var rooms []*models.ChatRoom
	err = db.Raw(`
		SELECT DISTINCT r.id, r.name, r.description, r.type, r.is_private, r.created_by, r.created_at, r.updated_at
		FROM chat_rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = ?
		ORDER BY r.updated_at DESC
	`, userID).Scan(&rooms).Error

	return rooms, err
}

func (r *chatRepository) AddRoomMember(tenantID, roomID, userID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO room_members (room_id, user_id, joined_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (room_id, user_id) DO NOTHING
	`, roomID, userID)

	return result.Error
}

func (r *chatRepository) RemoveRoomMember(tenantID, roomID, userID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM room_members WHERE room_id = ? AND user_id = ?
	`, roomID, userID)

	return result.Error
}

func (r *chatRepository) IsRoomMember(tenantID, roomID, userID string) bool {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return false
	}

	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM room_members WHERE room_id = ? AND user_id = ?
	`, roomID, userID).Scan(&count)

	return count > 0
}

func (r *chatRepository) GetRoomMemberCount(tenantID, roomID string) (int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Raw(`
		SELECT COUNT(*) FROM room_members WHERE room_id = ?
	`, roomID).Scan(&count).Error

	return count, err
}

func (r *chatRepository) CreateChatMessage(tenantID string, message *models.ChatMessage) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// Convert metadata to JSON
	metadataJSON := "{}"
	if message.Metadata != nil {
		// In a real implementation, properly marshal to JSON
		metadataJSON = "{}"
	}

	result := db.Exec(`
		INSERT INTO chat_messages (id, room_id, user_id, content, message_type, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, message.ID, message.RoomID, message.UserID, message.Content, message.MessageType, metadataJSON, message.CreatedAt)

	return result.Error
}

func (r *chatRepository) GetChatMessages(tenantID, roomID string, limit, offset int) ([]*models.ChatMessage, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var messages []*models.ChatMessage
	err = db.Raw(`
		SELECT id, room_id, user_id, content, message_type, created_at
		FROM chat_messages 
		WHERE room_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, roomID, limit, offset).Scan(&messages).Error

	return messages, err
}

func (r *chatRepository) GetRoomMessageCount(tenantID, roomID string) (int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Raw(`
		SELECT COUNT(*) FROM chat_messages WHERE room_id = ?
	`, roomID).Scan(&count).Error

	return count, err
}

func (r *chatRepository) GetLatestMessage(tenantID, roomID string) (*models.ChatMessage, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var message models.ChatMessage
	err = db.Raw(`
		SELECT id, room_id, user_id, content, message_type, created_at
		FROM chat_messages 
		WHERE room_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, roomID).Scan(&message).Error

	if err != nil {
		return nil, err
	}

	return &message, nil
}