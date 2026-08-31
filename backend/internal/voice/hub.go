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

func (h *Hub) Remove(
	client *Client,
) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients :=
		h.channels[client.ChannelID]

	if clients == nil {
		return false
	}

	current :=
		clients[client.UserID]

	if current != client {
		return false
	}

	delete(
		clients,
		client.UserID,
	)

	if len(clients) == 0 {
		delete(
			h.channels,
			client.ChannelID,
		)
	}

	return true
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

func (h *Hub) Join(
	client *Client,
) (
	previous *Client,
	existing []*Client,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[client.ChannelID] == nil {
		h.channels[client.ChannelID] =
			make(map[int]*Client)
	}

	clients :=
		h.channels[client.ChannelID]

	/*
	 * Check whether this user already has a connection.
	 */
	previous =
		clients[client.UserID]

	/*
	 * Get everyone else currently in the channel.
	 *
	 * Because clients are keyed by user ID, the new user
	 * isn't in the map yet, so everyone here is another user.
	 */
	existing =
		make(
			[]*Client,
			0,
			len(clients),
		)

	for userID, existingClient := range clients {

		if userID == client.UserID {
			continue
		}

		existing = append(
			existing,
			existingClient,
		)
	}

	/*
	 * The new connection becomes the active connection
	 * for this user.
	 */
	clients[client.UserID] =
		client

	return previous, existing
}
