package realtime

import (
	"sync"
)

type VoiceState struct {
	UserID    int  `json:"user_id"`
	ChannelID int  `json:"channel_id"`
	Muted     bool `json:"muted"`
	Deafened  bool `json:"deafened"`
}

type Hub struct {
	mu sync.RWMutex

	// All realtime clients connected to each server.
	clients map[int]map[*Client]struct{}

	// Current voice presence:
	//
	// serverID
	//   channelID
	//      userID -> VoiceState
	voice map[int]map[int]map[int]VoiceState
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[*Client]struct{}),
		voice:   make(map[int]map[int]map[int]VoiceState),
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
		h.voice[serverID] = make(map[int]map[int]VoiceState)
	}

	/*
	 * Check whether this user is already in a voice channel.
	 */
	var previous *VoiceState

	for existingChannelID, users := range h.voice[serverID] {
		state, exists := users[userID]

		if !exists {
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
			UserID:    state.UserID,
			ChannelID: state.ChannelID,
			Muted:     state.Muted,
			Deafened:  state.Deafened,
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
			make(map[int]VoiceState)
	}

	h.voice[serverID][channelID][userID] = VoiceState{
		UserID:    userID,
		ChannelID: channelID,
		Muted:     false,
		Deafened:  false,
	}

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

	if _, exists := users[userID]; !exists {
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

	result := make([]VoiceState, 0)

	channels := h.voice[serverID]

	for _, users := range channels {
		for _, state := range users {
			result = append(result, state)
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
		state, exists := users[userID]

		if !exists {
			continue
		}

		delete(users, userID)

		removed = append(
			removed,
			VoiceState{
				UserID:    state.UserID,
				ChannelID: channelID,
				Muted:     state.Muted,
				Deafened:  state.Deafened,
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

func (h *Hub) UpdateVoiceState(
	serverID int,
	channelID int,
	userID int,
	muted bool,
	deafened bool,
) (*VoiceState, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels := h.voice[serverID]

	if channels == nil {
		return nil, false
	}

	users := channels[channelID]

	if users == nil {
		return nil, false
	}

	state, exists := users[userID]

	if !exists {
		return nil, false
	}

	state.Muted = muted
	state.Deafened = deafened

	users[userID] = state

	return &state, true
}
