package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/devmuggs/youtube-music-continuity/backend/internal/auth"
	"github.com/devmuggs/youtube-music-continuity/backend/internal/database"
	"github.com/devmuggs/youtube-music-continuity/backend/internal/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from Chrome extensions and localhost
		origin := r.Header.Get("Origin")

		// log the current origin for debugging
		log.Printf("WebSocket connection from origin: %s", origin)

		return origin == "" || // WebSocket connections might not have origin
			len(origin) > 16 && origin[:16] == "chrome-extension" ||
			len(origin) > 16 && origin[:16] == "http://localhost" || origin == "https://music.youtube.com"
	},
}

// Client represents a WebSocket client
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int
}

// Hub maintains the set of active clients
type Hub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	userClients map[int][]*Client // Map user ID to their clients
	mutex       sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		userClients: make(map[int][]*Client),
	}
}

// Run starts the hub and handles client registration/unregistration
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			if h.userClients[client.userID] == nil {
				h.userClients[client.userID] = make([]*Client, 0)
			}
			h.userClients[client.userID] = append(h.userClients[client.userID], client)
			h.mutex.Unlock()
			log.Printf("Client registered for user %d", client.userID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove from user clients
				userClients := h.userClients[client.userID]
				for i, c := range userClients {
					if c == client {
						h.userClients[client.userID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}

				// Clean up empty user client list
				if len(h.userClients[client.userID]) == 0 {
					delete(h.userClients, client.userID)
				}
			}
			h.mutex.Unlock()
			log.Printf("Client unregistered for user %d", client.userID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToUser sends a message to all clients of a specific user
func (h *Hub) BroadcastToUser(userID int, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := h.userClients[userID]
	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *Hub, db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		// Get token from query parameter or first message
		token := r.URL.Query().Get("token")
		var userID int

		if token != "" {
			userID, err = auth.ValidateToken(token)
			if err != nil {
				log.Printf("Invalid token in WebSocket connection: %v", err)
				conn.Close()
				return
			}
		}

		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			userID: userID,
		}

		// If no token in query, wait for auth message
		if userID == 0 {
			// Read first message for authentication
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading auth message: %v", err)
				conn.Close()
				return
			}

			var authMsg models.WebSocketMessage
			if err := json.Unmarshal(messageBytes, &authMsg); err != nil {
				log.Printf("Error unmarshaling auth message: %v", err)
				conn.Close()
				return
			}

			if authMsg.Type != models.MessageTypeAuth {
				log.Printf("First message must be auth message")
				conn.Close()
				return
			}

			tokenData, ok := authMsg.Payload.(string)
			if !ok {
				log.Printf("Auth message payload must be a token string")
				conn.Close()
				return
			}

			userID, err = auth.ValidateToken(tokenData)
			if err != nil {
				log.Printf("Invalid token in auth message: %v", err)
				conn.Close()
				return
			}

			client.userID = userID
		}

		client.hub.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump(db)
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump(db *database.DB) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg models.WebSocketMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case models.MessageTypeSessionUpdate:
			if err := c.handleSessionUpdate(msg.Payload, db); err != nil {
				log.Printf("Error handling session update: %v", err)
			}
		}
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		}
	}
}

// handleSessionUpdate processes session update messages
func (c *Client) handleSessionUpdate(payload interface{}, db *database.DB) error {
	// Convert payload to SessionUpdateRequest
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var req models.SessionUpdateRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return err
	}

	// Update session in database
	session, err := db.UpsertSession(c.userID, req.VideoID, req.PlaybackTime, req.IsPlaying)
	if err != nil {
		return err
	}

	// Broadcast session update to all user's clients
	updateMsg := models.WebSocketMessage{
		Type:    models.MessageTypeSessionUpdate,
		Payload: session,
	}

	msgBytes, err := json.Marshal(updateMsg)
	if err != nil {
		return err
	}

	c.hub.BroadcastToUser(c.userID, msgBytes)
	return nil
}
