package realtime

import (
	"encoding/json"
	"net/http"
	"strconv"

	"agora/internal/channels"
	"agora/internal/servers"
	"agora/internal/users"

	"github.com/coder/websocket"
)

type Handler struct {
	channelRepo *channels.Repository
	serverRepo  *servers.Repository
	hub         *Hub
}

func NewHandler(
	channelRepo *channels.Repository,
	serverRepo *servers.Repository,
	hub *Hub,
) *Handler {
	return &Handler{
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
		hub:         hub,
	}
}

func (h *Handler) Connect(
	w http.ResponseWriter,
	r *http.Request,
) {
	serverID, err := strconv.Atoi(
		r.PathValue("serverID"),
	)
	if err != nil {
		http.Error(
			w,
			"Invalid server ID",
			http.StatusBadRequest,
		)
		return
	}

	userID, ok := users.UserIDFromContext(
		r.Context(),
	)
	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	/*
	 * Make sure the user belongs to this server.
	 */
	member, err := h.serverRepo.IsMember(
		userID,
		serverID,
	)
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	if !member {
		http.Error(
			w,
			"You cannot access this server",
			http.StatusForbidden,
		)
		return
	}

	conn, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{
			OriginPatterns: []string{
				"localhost:5173",
			},
		},
	)
	if err != nil {
		http.Error(
			w,
			"Could not establish WebSocket connection",
			http.StatusBadRequest,
		)
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		ServerID: serverID,
		send:     make(chan []byte, 32),
	}

	h.hub.Add(client)

	go client.writeLoop()

	/*
	 * Send the current voice state to the newly
	 * connected client.
	 *
	 * This is why a user does NOT need the SFU to
	 * tell them who was already in voice.
	 */
	state := h.hub.VoiceState(serverID)

	data, err := json.Marshal(Event{
		Type: "voice_state",
		Data: state,
	})

	if err == nil {
		select {
		case client.send <- data:
		default:
		}
	}

	defer func() {
		/*
		 * If the user disconnects from the server,
		 * remove them from every voice channel.
		 */
		removed := h.hub.LeaveAllVoice(
			serverID,
			userID,
		)

		/*
		 * Tell everyone that the user left those
		 * voice channels.
		 */
		for _, state := range removed {
			data, err := json.Marshal(Event{
				Type: "voice_leave",
				Data: state,
			})

			if err != nil {
				continue
			}

			h.hub.BroadcastServer(
				serverID,
				data,
			)
		}

		h.hub.Remove(client)

		_ = conn.Close(
			websocket.StatusNormalClosure,
			"",
		)
	}()

	for {
		_, data, err := conn.Read(
			r.Context(),
		)

		if err != nil {
			return
		}

		h.handleMessage(
			client,
			data,
		)
	}
}

func (h *Handler) handleMessage(
	client *Client,
	raw []byte,
) {
	var event ClientEvent

	if err := json.Unmarshal(
		raw,
		&event,
	); err != nil {
		return
	}

	switch event.Type {
	case "voice_join":
		h.handleVoiceJoin(
			client,
			event.Data.ChannelID,
		)

	case "voice_leave":
		h.handleVoiceLeave(
			client,
			event.Data.ChannelID,
		)
	}
}

func (h *Handler) handleVoiceJoin(
	client *Client,
	channelID int,
) {
	/*
	 * Validate that this channel actually belongs
	 * to the server the user is connected to.
	 */
	channel, err := h.channelRepo.GetByID(channelID)

	if err != nil {
		return
	}

	if channel.ServerID != client.ServerID {
		return
	}

	changed, previous := h.hub.JoinVoice(
		client.ServerID,
		channelID,
		client.UserID,
	)

	if !changed {
		return
	}

	/*
	 * If the user was already in another voice
	 * channel, tell everyone they left it.
	 */
	if previous != nil {
		data, err := json.Marshal(Event{
			Type: "voice_leave",
			Data: *previous,
		})

		if err == nil {
			h.hub.BroadcastServer(
				client.ServerID,
				data,
			)
		}
	}

	/*
	 * Tell everyone on the server that this user
	 * joined the new voice channel.
	 */
	data, err := json.Marshal(Event{
		Type: "voice_join",
		Data: VoiceState{
			UserID:    client.UserID,
			ChannelID: channelID,
		},
	})

	if err != nil {
		return
	}

	h.hub.BroadcastServer(
		client.ServerID,
		data,
	)
}

func (h *Handler) handleVoiceLeave(
	client *Client,
	channelID int,
) {
	/*
	 * Validate the channel belongs to this server.
	 */
	channel, err := h.channelRepo.GetByID(channelID)

	if err != nil {
		return
	}

	if channel.ServerID != client.ServerID {
		return
	}

	changed := h.hub.LeaveVoice(
		client.ServerID,
		channelID,
		client.UserID,
	)

	if !changed {
		return
	}

	data, err := json.Marshal(Event{
		Type: "voice_leave",
		Data: VoiceState{
			UserID:    client.UserID,
			ChannelID: channelID,
		},
	})

	if err != nil {
		return
	}

	h.hub.BroadcastServer(
		client.ServerID,
		data,
	)
}
