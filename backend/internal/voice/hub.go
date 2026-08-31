package voice

import (
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	channels map[int]map[int]*Client
}

func NewHub() *Hub {
	return &Hub{
		channels: make(map[int]map[int]*Client),
	}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[client.ChannelID] == nil {
		h.channels[client.ChannelID] = make(map[int]*Client)
	}

	h.channels[client.ChannelID][client.UserID] = client
}

func (h *Hub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.channels[client.ChannelID]

	if clients == nil {
		return
	}

	// Only remove this client if it is still the
	// active connection for this user.
	if current := clients[client.UserID]; current == client {
		delete(clients, client.UserID)
	}

	if len(clients) == 0 {
		delete(h.channels, client.ChannelID)
	}
}

func (h *Hub) Clients(channelID int) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[channelID]

	result := make([]*Client, 0, len(clients))

	for _, client := range clients {
		result = append(result, client)
	}

	return result
}

func (h *Hub) FindClient(
	channelID int,
	userID int,
) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[channelID]

	if clients == nil {
		return nil
	}

	return clients[userID]
}
