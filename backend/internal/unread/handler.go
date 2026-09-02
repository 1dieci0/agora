package unread

import (
	"encoding/json"
	"net/http"
	"strconv"

	"agora/internal/channels"
	"agora/internal/servers"
	"agora/internal/users"
)

type Handler struct {
	repo        *Repository
	channelRepo *channels.Repository
	serverRepo  *servers.Repository
}

func NewHandler(
	repo *Repository,
	channelRepo *channels.Repository,
	serverRepo *servers.Repository,
) *Handler {
	return &Handler{
		repo:        repo,
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
	}
}

type UnreadResponse struct {
	Channels []ChannelUnread `json:"channels"`
}

type MarkReadRequest struct {
	MessageID int `json:"message_id"`
}

func (h *Handler) GetUnread(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channels, err := h.repo.GetUnreadChannels(userID)
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	response := UnreadResponse{
		Channels: channels,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) MarkChannelRead(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.Atoi(r.PathValue("channelID"))
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.channelRepo.GetByID(channelID)

	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	/*
	 * The channel must contain a message before there is
	 * anything meaningful to mark as read.
	 */

	var data MarkReadRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if data.MessageID <= 0 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	/*
	 * Make sure the message actually belongs to this channel.
	 */
	exists, err := h.repo.MessageExists(
		data.MessageID,
		channelID,
	)

	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	if !exists {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	member, err := h.serverRepo.IsMember(
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

	if !member {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.repo.MarkChannelRead(
		userID,
		channelID,
		data.MessageID,
	); err != nil {
		http.Error(
			w,
			"Could not mark channel as read",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
