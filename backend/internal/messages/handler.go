package messages

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"agora/internal/channels"
	"agora/internal/realtime"
	"agora/internal/servers"
	"agora/internal/unread"
	"agora/internal/users"
)

type Handler struct {
	repo        *Repository
	channelRepo *channels.Repository
	serverRepo  *servers.Repository
	hub         *realtime.Hub
	userHub     *realtime.UserHub
	unreadRepo  *unread.Repository
}

func NewHandler(
	repo *Repository,
	channelRepo *channels.Repository,
	serverRepo *servers.Repository,
	unreadRepo *unread.Repository,
	hub *realtime.Hub,
	userHub *realtime.UserHub,
) *Handler {
	return &Handler{
		repo:        repo,
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
		unreadRepo:  unreadRepo,
		hub:         hub,
		userHub:     userHub,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var data CreateMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data.Content = strings.TrimSpace(data.Content)

	if data.Content == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	if len(data.Content) > 2000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	// Get the channel so we know which server it belongs to.

	channel, err := h.channelRepo.GetByID(channelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	serverID := channel.ServerID

	// Make sure the user belongs to the server.
	member, err := h.serverRepo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You cannot access this channel", http.StatusForbidden)
		return
	}

	messageID, err := h.repo.Create(
		channelID,
		userID,
		data.Content,
	)

	if err != nil {
		http.Error(w, "Could not create message", http.StatusInternalServerError)
		return
	}

	message, err := h.repo.GetByID(int(messageID))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	event := realtime.Event{
		Type: "message_created",
		Data: message,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.hub.BroadcastServer(serverID, eventData)

	members, err := h.serverRepo.GetMembers(serverID)
	if err != nil {
		log.Printf(
			"Could not get server members for unread update: %v",
			err,
		)
	} else {
		for _, member := range members {
			if member.ID == userID {
				continue
			}

			unreadCount, err := h.unreadRepo.GetChannelUnreadCount(
				member.ID,
				channelID,
			)
			if err != nil {
				log.Printf(
					"Could not get unread count: user=%d channel=%d: %v",
					member.ID,
					channelID,
					err,
				)
				continue
			}

			unreadEvent := realtime.Event{
				Type: "unread_update",
				Data: realtime.UnreadUpdate{
					ServerID:    serverID,
					ChannelID:   channelID,
					MessageID:   int(message.ID),
					UserID:      userID,
					UnreadCount: unreadCount,
				},
			}

			unreadData, err := json.Marshal(unreadEvent)
			if err != nil {
				log.Printf(
					"Could not marshal unread update: %v",
					err,
				)
				continue
			}

			h.userHub.SendToUser(member.ID, unreadData)
		}
	}

	response := MessageResponse{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	serverID := channel.ServerID

	member, err := h.serverRepo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You cannot access this channel", http.StatusForbidden)
		return
	}

	messages, err := h.repo.GetByChannelID(channelID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := MessagesResponse{
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	messageID, err := strconv.Atoi(r.PathValue("messageID"))
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data UpdateMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data.Content = strings.TrimSpace(data.Content)

	if data.Content == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	if len(data.Content) > 2000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	message, err := h.repo.GetByID(messageID)

	if err == sql.ErrNoRows {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if message.UserID != userID {
		http.Error(w, "You cannot edit this message", http.StatusForbidden)
		return
	}

	channel, err := h.channelRepo.GetByID(message.ChannelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	serverID := channel.ServerID

	if err := h.repo.Update(messageID, data.Content); err != nil {
		http.Error(w, "Could not update message", http.StatusInternalServerError)
		return
	}

	message.Content = data.Content

	event := realtime.Event{
		Type: "message_updated",
		Data: message,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.hub.BroadcastServer(serverID, eventData)

	response := MessageResponse{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	messageID, err := strconv.Atoi(r.PathValue("messageID"))
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	message, err := h.repo.GetByID(messageID)

	if err == sql.ErrNoRows {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if message.UserID != userID {
		http.Error(w, "You cannot delete this message", http.StatusForbidden)
		return
	}

	channel, err := h.channelRepo.GetByID(message.ChannelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	serverID := channel.ServerID

	if err := h.repo.Delete(messageID); err != nil {
		http.Error(w, "Could not delete message", http.StatusInternalServerError)
		return
	}

	event := realtime.Event{
		Type: "message_deleted",
		Data: struct {
			ID        int `json:"id"`
			ChannelID int `json:"channel_id"`
		}{
			ID:        message.ID,
			ChannelID: message.ChannelID,
		},
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.hub.BroadcastServer(serverID, eventData)

	w.WriteHeader(http.StatusNoContent)
}
