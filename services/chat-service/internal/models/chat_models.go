package models

import (
	"encoding/json"
	"time"
)

type ChatRoom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	IsPrivate   bool      `json:"is_private"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID          string                 `json:"id"`
	RoomID      string                 `json:"room_id"`
	UserID      string                 `json:"user_id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ChatRoomResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	IsPrivate   bool      `json:"is_private"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     []string  `json:"members,omitempty"`
}

type ChatMessageResponse struct {
	ID          string                 `json:"id"`
	RoomID      string                 `json:"room_id"`
	UserID      string                 `json:"user_id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ToJSON converts a ChatMessageResponse to JSON bytes for WebSocket broadcasting
func (msg *ChatMessageResponse) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":         "chat_message",
		"room_id":      msg.RoomID,
		"user_id":      msg.UserID,
		"content":      msg.Content,
		"message_type": msg.MessageType,
		"metadata":     msg.Metadata,
		"timestamp":    msg.CreatedAt,
	})
}