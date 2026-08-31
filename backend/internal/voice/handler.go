package voice

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"agora/internal/channels"
	"agora/internal/servers"
	"agora/internal/users"

	"github.com/coder/websocket"
)

type Handler struct {
	hub         *Hub
	channelRepo *channels.Repository
	serverRepo  *servers.Repository
}

const websocketWriteTimeout = 5 * time.Second

func NewHandler(
	hub *Hub,
	channelRepo *channels.Repository,
	serverRepo *servers.Repository,
) *Handler {
	return &Handler{
		hub:         hub,
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
	}
}

func (h *Handler) writeMessage(
	client *Client,
	message []byte,
) error {
	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			websocketWriteTimeout,
		)

	defer cancel()

	return client.Conn.Write(
		ctx,
		websocket.MessageText,
		message,
	)
}

func (h *Handler) Connect(
	w http.ResponseWriter,
	r *http.Request,
) {

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	channelID, err := strconv.Atoi(
		r.PathValue("channelID"),
	)
	if err != nil {
		http.Error(
			w,
			"Invalid channel ID",
			http.StatusBadRequest,
		)
		return
	}

	// Get the channel.
	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(
			w,
			"Channel not found",
			http.StatusNotFound,
		)
		return
	}

	// Make sure this is a voice channel.
	if channel.Type != "voice" {
		http.Error(
			w,
			"Not a voice channel",
			http.StatusBadRequest,
		)
		return
	}

	// Make sure the user belongs to the server.
	isMember, err := h.serverRepo.IsMember(
		userID,
		channel.ServerID,
	)
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	if !isMember {
		http.Error(
			w,
			"Forbidden",
			http.StatusForbidden,
		)
		return
	}

	// Upgrade to WebSocket.
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
		return
	}

	defer conn.Close(
		websocket.StatusNormalClosure,
		"",
	)

	client := &Client{
		UserID:    userID,
		ChannelID: channelID,
		Conn:      conn,
	}

	oldClient := h.hub.FindClient(
		channelID,
		userID,
	)

	if oldClient != nil {
		log.Printf(
			"VOICE: replacing existing connection user=%d channel=%d",
			userID,
			channelID,
		)

		_ = oldClient.Conn.Close(
			websocket.StatusNormalClosure,
			"replaced by newer connection",
		)
	}

	// Get the users already in the channel before adding
	// the new client.
	//existingClients := h.hub.Clients(channelID)

	// Add the new client.
	//h.hub.Add(client)

	previousClient, existingClients := h.hub.Join(client)

	if previousClient != nil {
		log.Printf(
			"VOICE: replacing existing connection user=%d channel=%d",
			userID,
			channelID,
		)

		_ = previousClient.Conn.Close(
			websocket.StatusNormalClosure,
			"replaced by newer connection",
		)
	}

	// Tell the new client about everyone already in the channel.
	for _, existingClient := range existingClients {
		err := h.writeMessage(
			client,
			mustMarshal(VoiceMessage{
				Type:   "user_joined",
				UserID: existingClient.UserID,
			}),
		)

		if err != nil {
			h.hub.Remove(client)
			return
		}
	}

	// Tell everyone else that this user joined.
	h.broadcastEvent(
		channelID,
		VoiceMessage{
			Type:   "user_joined",
			UserID: userID,
		},
		client,
	)

	// When the connection ends, remove the client
	// and notify everyone else.
	defer func() {
		removed := h.hub.Remove(client)

		if !removed {
			/*
			* This connection was already replaced by a newer
			* connection for the same user.
			 */
			return
		}

		h.broadcastEvent(
			channelID,
			VoiceMessage{
				Type:   "user_left",
				UserID: userID,
			},
			client,
		)
	}()

	// Read messages from this client.
	for {
		_, message, err := conn.Read(r.Context())
		if err != nil {
			break
		}

		h.broadcast(
			client,
			message,
		)
	}
}

func (h *Handler) broadcast(
	sender *Client,
	message []byte,
) {
	var voiceMessage VoiceMessage

	if err := json.Unmarshal(message, &voiceMessage); err != nil {
		log.Printf(
			"VOICE: invalid message from user %d: %v",
			sender.UserID,
			err,
		)
		return
	}

	log.Printf(
		"VOICE: received type=%s from=%d target=%d",
		voiceMessage.Type,
		sender.UserID,
		voiceMessage.TargetUserID,
	)

	if voiceMessage.TargetUserID != 0 {
		target := h.hub.FindClient(
			sender.ChannelID,
			voiceMessage.TargetUserID,
		)

		if target == nil {
			log.Printf(
				"VOICE: target NOT FOUND type=%s from=%d target=%d channel=%d",
				voiceMessage.Type,
				sender.UserID,
				voiceMessage.TargetUserID,
				sender.ChannelID,
			)

			return
		}

		log.Printf(
			"VOICE: forwarding type=%s from=%d to=%d",
			voiceMessage.Type,
			sender.UserID,
			target.UserID,
		)

		err := h.writeMessage(
			target,
			message,
		)

		if err != nil {
			log.Printf(
				"VOICE: write FAILED type=%s from=%d to=%d: %v",
				voiceMessage.Type,
				sender.UserID,
				target.UserID,
				err,
			)
		}

		return
	}

	clients := h.hub.Clients(
		sender.ChannelID,
	)

	for _, client := range clients {
		if client == sender {
			continue
		}

		log.Printf(
			"VOICE: broadcasting type=%s from=%d to=%d",
			voiceMessage.Type,
			sender.UserID,
			client.UserID,
		)

		err := h.writeMessage(
			client,
			message,
		)

		if err != nil {
			log.Printf(
				"VOICE: broadcast write FAILED type=%s to=%d: %v",
				voiceMessage.Type,
				client.UserID,
				err,
			)
		}
	}
}

func (h *Handler) broadcastEvent(
	channelID int,
	message VoiceMessage,
	exclude *Client,
) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	clients :=
		h.hub.Clients(channelID)

	for _, client := range clients {
		if client == exclude {
			continue
		}

		err := h.writeMessage(
			client,
			data,
		)

		if err != nil {
			log.Printf(
				"VOICE: event write FAILED type=%s to=%d: %v",
				message.Type,
				client.UserID,
				err,
			)
		}
	}
}

func mustMarshal(message VoiceMessage) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	return data
}
