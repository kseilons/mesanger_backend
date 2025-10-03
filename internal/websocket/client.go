package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a websocket client
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Hub reference
	hub *Hub

	// Unique client ID
	ID string

	// User ID if authenticated
	UserID string

	// User information
	Username string

	// Rooms this client is subscribed to
	rooms map[string]bool

	// Mutex for thread safety
	mutex sync.RWMutex

	// Logger
	logger *slog.Logger

	// Last activity timestamp
	lastActivity time.Time

	// Ping/pong handling
	pongWait       time.Duration
	pingPeriod     time.Duration
	writeWait      time.Duration
	maxMessageSize int64
}

// NewClient creates a new websocket client
func NewClient(conn *websocket.Conn, hub *Hub, logger *slog.Logger) *Client {
	return &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		hub:            hub,
		ID:             uuid.New().String(),
		rooms:          make(map[string]bool),
		logger:         logger,
		lastActivity:   time.Now(),
		pongWait:       60 * time.Second,
		pingPeriod:     54 * time.Second,
		writeWait:      10 * time.Second,
		maxMessageSize: 1024 * 1024, // 1MB
	}
}

// SetUser sets the user information for the client
func (c *Client) SetUser(userID, username string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.UserID = userID
	c.Username = username
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	// Set read limits
	c.conn.SetReadLimit(c.maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))

	// Set pong handler
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket error", "error", err, "client_id", c.ID)
			}
			break
		}

		c.lastActivity = time.Now()
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to this client
func (c *Client) SendMessage(message []byte) {
	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}

// JoinRoom joins a room
func (c *Client) JoinRoom(roomID string) {
	c.hub.JoinRoom(c, roomID)
}

// LeaveRoom leaves a room
func (c *Client) LeaveRoom(roomID string) {
	c.hub.LeaveRoom(c, roomID)
}

// IsInRoom checks if client is in a specific room
func (c *Client) IsInRoom(roomID string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.rooms[roomID]
	return exists
}

// GetRooms returns all rooms this client is in
func (c *Client) GetRooms() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	rooms := make([]string, 0, len(c.rooms))
	for roomID := range c.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// IsActive checks if client is still active
func (c *Client) IsActive() bool {
	return time.Since(c.lastActivity) < c.pongWait
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var wsMessage struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &wsMessage); err != nil {
		c.logger.Error("Failed to unmarshal WebSocket message", "error", err)
		c.sendError("Invalid message format")
		return
	}

	switch wsMessage.Type {
	case "join_room":
		c.handleJoinRoom(wsMessage.Data)
	case "leave_room":
		c.handleLeaveRoom(wsMessage.Data)
	case "typing":
		c.handleTyping(wsMessage.Data)
	case "stop_typing":
		c.handleStopTyping(wsMessage.Data)
	case "ping":
		c.handlePing()
	default:
		c.logger.Warn("Unknown message type", "type", wsMessage.Type)
		c.sendError("Unknown message type: " + wsMessage.Type)
	}
}

func (c *Client) handleJoinRoom(data json.RawMessage) {
	var request struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		c.sendError("Invalid join room request")
		return
	}

	c.JoinRoom(request.RoomID)
	c.logger.Info("Client joined room", "client_id", c.ID, "room_id", request.RoomID)
}

func (c *Client) handleLeaveRoom(data json.RawMessage) {
	var request struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		c.sendError("Invalid leave room request")
		return
	}

	c.LeaveRoom(request.RoomID)
	c.logger.Info("Client left room", "client_id", c.ID, "room_id", request.RoomID)
}

func (c *Client) handleTyping(data json.RawMessage) {
	var request struct {
		RoomID    string  `json:"room_id"`
		ChannelID *string `json:"channel_id"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		c.sendError("Invalid typing request")
		return
	}

	// Broadcast typing status to room
	typingMessage := map[string]interface{}{
		"type": "user_typing",
		"data": map[string]interface{}{
			"user_id":    c.UserID,
			"username":   c.Username,
			"room_id":    request.RoomID,
			"channel_id": request.ChannelID,
			"is_typing":  true,
			"timestamp":  time.Now(),
		},
	}

	messageBytes, _ := json.Marshal(typingMessage)
	c.hub.BroadcastToRoom(request.RoomID, messageBytes)
}

func (c *Client) handleStopTyping(data json.RawMessage) {
	var request struct {
		RoomID    string  `json:"room_id"`
		ChannelID *string `json:"channel_id"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		c.sendError("Invalid stop typing request")
		return
	}

	// Broadcast stop typing status to room
	stopTypingMessage := map[string]interface{}{
		"type": "user_typing",
		"data": map[string]interface{}{
			"user_id":    c.UserID,
			"username":   c.Username,
			"room_id":    request.RoomID,
			"channel_id": request.ChannelID,
			"is_typing":  false,
			"timestamp":  time.Now(),
		},
	}

	messageBytes, _ := json.Marshal(stopTypingMessage)
	c.hub.BroadcastToRoom(request.RoomID, messageBytes)
}

func (c *Client) handlePing() {
	pongMessage := map[string]interface{}{
		"type": "pong",
		"data": map[string]interface{}{
			"timestamp": time.Now(),
		},
	}

	messageBytes, _ := json.Marshal(pongMessage)
	c.SendMessage(messageBytes)
}

func (c *Client) sendError(message string) {
	errorMessage := map[string]interface{}{
		"type": "error",
		"data": map[string]interface{}{
			"message":   message,
			"timestamp": time.Now(),
		},
	}

	messageBytes, _ := json.Marshal(errorMessage)
	c.SendMessage(messageBytes)
}
