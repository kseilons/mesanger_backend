package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/kseilons/messenger-backend/internal/models"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast channel for messages
	broadcast chan []byte

	// Room-based messaging
	rooms map[string]map[*Client]bool

	// User connections mapping
	userConnections map[string][]*Client

	// Mutex for thread safety
	mutex sync.RWMutex

	// Logger
	logger *slog.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan []byte),
		rooms:           make(map[string]map[*Client]bool),
		userConnections: make(map[string][]*Client),
		logger:          logger,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case <-ticker.C:
			h.pingClients()
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *Hub) BroadcastToAll(message []byte) {
	h.broadcast <- message
}

// BroadcastToRoom broadcasts a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(room, client)
				if len(room) == 0 {
					delete(h.rooms, roomID)
				}
			}
		}
	}
}

// BroadcastToUser broadcasts a message to all connections of a specific user
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if connections, exists := h.userConnections[userID]; exists {
		for _, client := range connections {
			select {
			case client.send <- message:
			default:
				// Remove disconnected client
				h.removeUserConnection(userID, client)
			}
		}
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.rooms[roomID] = true

	h.logger.Info("Client joined room", "client_id", client.ID, "room_id", roomID)
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if room, exists := h.rooms[roomID]; exists {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(client.rooms, roomID)

	h.logger.Info("Client left room", "client_id", client.ID, "room_id", roomID)
}

// GetRoomClients returns all clients in a room
func (h *Hub) GetRoomClients(roomID string) []*Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var clients []*Client
	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetUserConnections returns all connections for a user
func (h *Hub) GetUserConnections(userID string) []*Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if connections, exists := h.userConnections[userID]; exists {
		// Return a copy to avoid race conditions
		result := make([]*Client, len(connections))
		copy(result, connections)
		return result
	}
	return nil
}

// IsUserOnline checks if a user is online
func (h *Hub) IsUserOnline(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	connections, exists := h.userConnections[userID]
	return exists && len(connections) > 0
}

// GetOnlineUsers returns list of online user IDs
func (h *Hub) GetOnlineUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var onlineUsers []string
	for userID, connections := range h.userConnections {
		if len(connections) > 0 {
			onlineUsers = append(onlineUsers, userID)
		}
	}
	return onlineUsers
}

// private methods

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true

	// Add to user connections
	if client.UserID != "" {
		h.userConnections[client.UserID] = append(h.userConnections[client.UserID], client)
	}

	h.logger.Info("Client registered", "client_id", client.ID, "user_id", client.UserID)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}

	// Remove from all rooms
	for roomID := range client.rooms {
		if room, exists := h.rooms[roomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}

	// Remove from user connections
	if client.UserID != "" {
		h.removeUserConnection(client.UserID, client)
	}

	h.logger.Info("Client unregistered", "client_id", client.ID, "user_id", client.UserID)
}

func (h *Hub) removeUserConnection(userID string, client *Client) {
	if connections, exists := h.userConnections[userID]; exists {
		for i, conn := range connections {
			if conn == client {
				// Remove the connection
				h.userConnections[userID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}

		// If no connections left, remove user entry
		if len(h.userConnections[userID]) == 0 {
			delete(h.userConnections, userID)
		}
	}
}

func (h *Hub) broadcastToAll(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) pingClients() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	pingMessage := models.WebSocketMessage{
		Type:      "ping",
		Data:      nil,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(pingMessage)
	if err != nil {
		h.logger.Error("Failed to marshal ping message", "error", err)
		return
	}

	for client := range h.clients {
		select {
		case client.send <- messageBytes:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}
