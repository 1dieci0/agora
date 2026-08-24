package messages

import (
	"agora/internal/realtime"
	"agora/internal/users"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	repo *Repository
	hub  *realtime.Hub
}

func NewHandler(repo *Repository, hub *realtime.Hub) *Handler {
	return &Handler{
		repo: repo,
		hub:  hub,
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

	// Check that the user is a member of the server
	// that this channel belongs to.
	member, err := h.repo.IsMember(userID, channelID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You cannot access this channel", http.StatusForbidden)
		return
	}

	// Create the message.
	messageID, err := h.repo.Create(
		channelID,
		userID,
		data.Content,
	)

	if err != nil {
		http.Error(w, "Could not create message", http.StatusInternalServerError)
		return
	}

	// Get the complete message, including the username and timestamp.
	message, err := h.repo.GetByID(int(messageID))

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create the realtime event.
	event := realtime.Event{
		Type: "message_created",
		Data: message,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send the event to everyone connected to this channel.
	h.hub.Broadcast(channelID, eventData)

	// Return the created message to the HTTP client.
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

	member, err := h.repo.IsMember(userID, channelID)
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

	// Get the message.
	message, err := h.repo.GetByID(messageID)

	if err == sql.ErrNoRows {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Only the author can delete the message.
	if message.UserID != userID {
		http.Error(w, "You cannot delete this message", http.StatusForbidden)
		return
	}

	// Delete the message.
	if err := h.repo.Delete(messageID); err != nil {
		http.Error(w, "Could not delete message", http.StatusInternalServerError)
		return
	}

	// Tell every connected client in this channel.
	event := realtime.Event{
		Type: "message_deleted",
		Data: struct {
			ID int `json:"id"`
		}{
			ID: message.ID,
		},
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(message.ChannelID, eventData)

	w.WriteHeader(http.StatusNoContent)
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

	// Get the message.
	message, err := h.repo.GetByID(messageID)

	if err == sql.ErrNoRows {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Only the author can edit the message.
	if message.UserID != userID {
		http.Error(w, "You cannot edit this message", http.StatusForbidden)
		return
	}

	// Update the message.
	if err := h.repo.Update(messageID, data.Content); err != nil {
		http.Error(w, "Could not update message", http.StatusInternalServerError)
		return
	}

	// Update the copy we already have so we can
	// send the updated message to the clients.
	message.Content = data.Content

	// Broadcast the updated message.
	event := realtime.Event{
		Type: "message_updated",
		Data: message,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(message.ChannelID, eventData)

	// Return the updated message.
	response := MessageResponse{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
