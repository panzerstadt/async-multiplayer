package sse

import (
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

// SSEClient represents a single SSE client connection
type SSEClient struct {
	ID   string
	Send chan []byte
}

// SSEManager manages SSE client connections and broadcasting
type SSEManager struct {
	clients    map[string]*SSEClient
	broadcast  chan []byte
	register   chan *SSEClient
	unregister chan *SSEClient
}

// NewSSEManager creates a new SSEManager
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients:    make(map[string]*SSEClient),
		broadcast:  make(chan []byte),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
	}
}

// Run starts the SSEManager
func (manager *SSEManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.clients[client.ID] = client
			log.Printf("SSE: Client %s registered. Total clients: %d", client.ID, len(manager.clients))
		case client := <-manager.unregister:
			if _, ok := manager.clients[client.ID]; ok {
				delete(manager.clients, client.ID)
				close(client.Send)
				log.Printf("SSE: Client %s unregistered. Total clients: %d", client.ID, len(manager.clients))
			}
		case message := <-manager.broadcast:
			for _, client := range manager.clients {
				select {
				case client.Send <- message:
					// Message sent
				default:
					// Client channel is blocked, unregister client
					close(client.Send)
					delete(manager.clients, client.ID)
					log.Printf("SSE: Client %s channel blocked, unregistered. Total clients: %d", client.ID, len(manager.clients))
				}
			}
		}
	}
}

// BroadcastMessage sends a message to all connected SSE clients
func (manager *SSEManager) BroadcastMessage(event string, data []byte) {
	// SSE format:
	// event: <event>
	// data: <data>

	// For simplicity, we'll just send data for now.
	// If you need event types or IDs, you'll need to format the message accordingly.
	formattedMessage := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
	manager.broadcast <- formattedMessage
}

// ServeSSE handles new SSE connections
func ServeSSE(manager *SSEManager, c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin")) // Reflect origin for CORS
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	client := &SSEClient{
		ID:   c.ClientIP(), // Simple ID, consider a more robust unique ID
		Send: make(chan []byte),
	}
	manager.register <- client

	defer func() {
		manager.unregister <- client
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-client.Send:
			_, err := w.Write(msg)
			if err != nil {
				log.Printf("SSE: Error writing to client %s: %v", client.ID, err)
				return false // Close connection
			}
			return true // Keep connection open
		case <-c.Request.Context().Done():
			log.Printf("SSE: Client %s disconnected (context done)", client.ID)
			return false // Close connection
		}
	})
}
