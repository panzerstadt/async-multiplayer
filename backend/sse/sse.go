package sse

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

// SSEManager manages SSE connections and broadcasts messages.
type SSEManager struct {
	clients   map[chan string]bool
	clientsMu sync.RWMutex
}

// NewSSEManager creates a new SSEManager.
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[chan string]bool),
	}
}

// AddClient adds a new client to the SSE manager.
func (sm *SSEManager) AddClient(client chan string) {
	sm.clientsMu.Lock()
	defer sm.clientsMu.Unlock()
	sm.clients[client] = true
	log.Println("Client added. Total clients:", len(sm.clients))
}

// RemoveClient removes a client from the SSE manager.
func (sm *SSEManager) RemoveClient(client chan string) {
	sm.clientsMu.Lock()
	defer sm.clientsMu.Unlock()
	delete(sm.clients, client)
	close(client)
	log.Println("Client removed. Total clients:", len(sm.clients))
}

// BroadcastMessage sends a message to all connected clients.
func (sm *SSEManager) BroadcastMessage(eventType string, data interface{}) {
	sm.clientsMu.RLock()
	defer sm.clientsMu.RUnlock()

	messageBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling SSE data: %v", err)
		return
	}

	formattedMessage := "event: " + eventType + "\n" + "data: " + string(messageBytes) + "\n\n"

	for client := range sm.clients {
		select {
		case client <- formattedMessage:
		default:
			log.Println("Could not send to a client, channel is full or closed.")
		}
	}
}

// Broadcaster defines the interface for broadcasting SSE messages.
type Broadcaster interface {
	BroadcastMessage(eventType string, data interface{})
	AddClient(client chan string)
	RemoveClient(client chan string)
	Run()
}

// Run starts the SSE manager.
func (sm *SSEManager) Run() {
	// This can be used for periodic cleanup or other maintenance tasks.
	// For now, it does nothing.
}

// ServeSSE handles SSE connections.
func ServeSSE(sm Broadcaster, c *gin.Context) {
	clientChan := make(chan string)
	sm.AddClient(clientChan)
	defer sm.RemoveClient(clientChan)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Initial connection message
	c.Writer.Flush()

	for {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return // Channel closed
			}
			_, err := c.Writer.WriteString(msg)
			if err != nil {
				log.Printf("Error writing to SSE client: %v", err)
				return
			}
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
