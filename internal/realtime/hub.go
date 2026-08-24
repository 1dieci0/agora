package realtime

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	UserID    int
	ChannelID int
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

	delete(clients, client)

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
		err := client.Conn.Write(
			context.Background(),
			websocket.MessageText,
			data,
		)

		if err != nil {
			h.Remove(client)
		}
	}
}
