package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from clients
	broadcast chan []byte

	// Room-based messaging
	rooms map[string]map[*Client]bool

	logger *zap.Logger
}

// Client represents a WebSocket connection
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel for outbound messages
	send chan []byte

	// User information
	UserID   string
	TenantID string
	Rooms    map[string]bool

	hub *Hub
}

// Message represents a chat message
type Message struct {
	Type      string                 `json:"type"`
	RoomID    string                 `json:"room_id,omitempty"`
	UserID    string                 `json:"user_id"`
	TenantID  string                 `json:"tenant_id"`
	Content   string                 `json:"content,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[string]map[*Client]bool),
		logger:     logger,
	}
}

// Run starts the hub and handles client registration/deregistration
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("Client registered", 
				zap.String("user_id", client.UserID),
				zap.String("tenant_id", client.TenantID))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				
				// Remove from all rooms
				for roomID := range client.Rooms {
					h.leaveRoom(client, roomID)
				}
				
				close(client.send)
				h.logger.Info("Client unregistered", 
					zap.String("user_id", client.UserID),
					zap.String("tenant_id", client.TenantID))
			}

		case message := <-h.broadcast:
			// Broadcast to all clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, roomID string) {
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.Rooms[roomID] = true
	
	h.logger.Info("Client joined room", 
		zap.String("user_id", client.UserID),
		zap.String("room_id", roomID))
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.leaveRoom(client, roomID)
}

func (h *Hub) leaveRoom(client *Client, roomID string) {
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], client)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(client.Rooms, roomID)
	
	h.logger.Info("Client left room", 
		zap.String("user_id", client.UserID),
		zap.String("room_id", roomID))
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			select {
			case client.send <- message:
			default:
				delete(h.clients, client)
				delete(room, client)
				close(client.send)
			}
		}
	}
}

// GetRoomClients returns the number of clients in a room
func (h *Hub) GetRoomClients(roomID string) int {
	if room, exists := h.rooms[roomID]; exists {
		return len(room)
	}
	return 0
}

// GetRegisterChannel returns the register channel for client registration
func (h *Hub) GetRegisterChannel() chan<- *Client {
	return h.register
}

// ReadPump handles reading messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			c.hub.logger.Error("Failed to parse message", zap.Error(err))
			continue
		}

		// Set user and tenant info
		msg.UserID = c.UserID
		msg.TenantID = c.TenantID
		msg.Timestamp = time.Now()

		// Re-encode the message
		updatedMessage, err := json.Marshal(msg)
		if err != nil {
			c.hub.logger.Error("Failed to encode message", zap.Error(err))
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "join_room":
			if msg.RoomID != "" {
				c.hub.JoinRoom(c, msg.RoomID)
			}
		case "leave_room":
			if msg.RoomID != "" {
				c.hub.LeaveRoom(c, msg.RoomID)
			}
		case "chat_message":
			if msg.RoomID != "" {
				// Broadcast to room
				c.hub.BroadcastToRoom(msg.RoomID, updatedMessage)
			} else {
				// Broadcast to all
				c.hub.broadcast <- updatedMessage
			}
		}
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, userID, tenantID string) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   userID,
		TenantID: tenantID,
		Rooms:    make(map[string]bool),
		hub:      hub,
	}
}