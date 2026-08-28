package voice

import (
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	channels map[int]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		channels: make(map[int]map[*Client]bool),
	}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[client.ChannelID] == nil {
		h.channels[client.ChannelID] = make(map[*Client]bool)
	}

	h.channels[client.ChannelID][client] = true
}

func (h *Hub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.channels[client.ChannelID]

	delete(clients, client)

	if len(clients) == 0 {
		delete(h.channels, client.ChannelID)
	}
}

func (h *Hub) Clients(channelID int) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[channelID]

	result := make([]*Client, 0, len(clients))

	for client := range clients {
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

	for client := range h.channels[channelID] {
		if client.UserID == userID {
			return client
		}
	}

	return nil
}
