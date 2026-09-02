package realtime

import (
	"sync"
)

type UserHub struct {
	mu sync.RWMutex

	clients map[int]map[*Client]struct{}
}

func NewUserHub() *UserHub {
	return &UserHub{
		clients: make(map[int]map[*Client]struct{}),
	}
}

func (h *UserHub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}

	h.clients[client.UserID][client] = struct{}{}
}

func (h *UserHub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.UserID]

	if clients == nil {
		return
	}

	delete(clients, client)

	if len(clients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *UserHub) SendToUser(userID int, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients[userID] {
		select {
		case client.send <- data:
		default:
			h.removeLocked(client)
		}
	}
}

func (h *UserHub) removeLocked(client *Client) {
	clients := h.clients[client.UserID]

	if clients == nil {
		return
	}

	if _, exists := clients[client]; !exists {
		return
	}

	delete(clients, client)

	close(client.send)

	if len(clients) == 0 {
		delete(h.clients, client.UserID)
	}
}
