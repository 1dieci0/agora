package realtime

import (
	"sync"

	"github.com/coder/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	UserID    int
	ChannelID int

	send chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[*Client]struct{}),
	}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.ChannelID] == nil {
		h.clients[client.ChannelID] = make(map[*Client]struct{})
	}

	h.clients[client.ChannelID][client] = struct{}{}
}

func (h *Hub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.ChannelID]

	if clients == nil {
		return
	}

	if _, exists := clients[client]; !exists {
		return
	}

	delete(clients, client)

	close(client.send)

	if len(clients) == 0 {
		delete(h.clients, client.ChannelID)
	}
}

func (h *Hub) Broadcast(channelID int, data []byte) {
	h.mu.RLock()

	clients := make([]*Client, 0, len(h.clients[channelID]))

	for client := range h.clients[channelID] {
		clients = append(clients, client)
	}

	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			// Client isn't consuming messages fast enough.
			h.Remove(client)
		}
	}
}
