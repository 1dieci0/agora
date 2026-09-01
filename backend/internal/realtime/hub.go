package realtime

import (
	"sync"
)

type VoiceState struct {
	UserID    int `json:"user_id"`
	ChannelID int `json:"channel_id"`
}

type Hub struct {
	mu sync.RWMutex

	// All realtime clients connected to each server.
	clients map[int]map[*Client]struct{}

	// Current voice presence:
	//
	// serverID
	//   channelID
	//      userID -> true
	voice map[int]map[int]map[int]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[*Client]struct{}),
		voice:   make(map[int]map[int]map[int]bool),
	}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.ServerID] == nil {
		h.clients[client.ServerID] = make(map[*Client]struct{})
	}

	h.clients[client.ServerID][client] = struct{}{}
}

func (h *Hub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.removeLocked(client)
}

func (h *Hub) removeLocked(client *Client) {
	clients := h.clients[client.ServerID]

	if clients == nil {
		return
	}

	if _, exists := clients[client]; !exists {
		return
	}

	delete(clients, client)

	close(client.send)

	if len(clients) == 0 {
		delete(h.clients, client.ServerID)
	}
}

func (h *Hub) BroadcastServer(
	serverID int,
	data []byte,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[serverID]

	for client := range clients {
		select {
		case client.send <- data:

		default:
			// Remove the client while the lock is already held.
			h.removeLocked(client)
		}
	}
}

func (h *Hub) JoinVoice(
	serverID int,
	channelID int,
	userID int,
) (bool, *VoiceState) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.voice[serverID] == nil {
		h.voice[serverID] = make(map[int]map[int]bool)
	}

	/*
	 * Check whether this user is already in a voice channel.
	 */
	var previous *VoiceState

	for existingChannelID, users := range h.voice[serverID] {
		if !users[userID] {
			continue
		}

		/*
		 * Already in the requested channel.
		 */
		if existingChannelID == channelID {
			return false, nil
		}

		/*
		 * User is moving from another voice channel.
		 */
		delete(users, userID)

		previous = &VoiceState{
			UserID:    userID,
			ChannelID: existingChannelID,
		}

		if len(users) == 0 {
			delete(
				h.voice[serverID],
				existingChannelID,
			)
		}

		break
	}

	if h.voice[serverID][channelID] == nil {
		h.voice[serverID][channelID] =
			make(map[int]bool)
	}

	h.voice[serverID][channelID][userID] = true

	return true, previous
}

func (h *Hub) LeaveVoice(
	serverID int,
	channelID int,
	userID int,
) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels := h.voice[serverID]

	if channels == nil {
		return false
	}

	users := channels[channelID]

	if users == nil {
		return false
	}

	if !users[userID] {
		return false
	}

	delete(users, userID)

	if len(users) == 0 {
		delete(channels, channelID)
	}

	if len(channels) == 0 {
		delete(h.voice, serverID)
	}

	return true
}

func (h *Hub) VoiceState(
	serverID int,
) []VoiceState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []VoiceState

	channels := h.voice[serverID]

	for channelID, users := range channels {
		for userID := range users {
			result = append(result, VoiceState{
				UserID:    userID,
				ChannelID: channelID,
			})
		}
	}

	return result
}

func (h *Hub) LeaveAllVoice(
	serverID int,
	userID int,
) []VoiceState {
	h.mu.Lock()
	defer h.mu.Unlock()

	var removed []VoiceState

	channels := h.voice[serverID]

	if channels == nil {
		return removed
	}

	for channelID, users := range channels {
		if !users[userID] {
			continue
		}

		delete(users, userID)

		removed = append(
			removed,
			VoiceState{
				UserID:    userID,
				ChannelID: channelID,
			},
		)

		if len(users) == 0 {
			delete(channels, channelID)
		}
	}

	if len(channels) == 0 {
		delete(h.voice, serverID)
	}

	return removed
}
