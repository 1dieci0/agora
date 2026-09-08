package dms

import (
	"agora/internal/mentions"
	"agora/internal/notifications"
	"agora/internal/realtime"
	"agora/internal/users"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	repo          *Repository
	hub           *realtime.UserHub
	notifications *notifications.Repository
}

func NewHandler(
	repo *Repository,
	hub *realtime.UserHub,
	notificationsRepo *notifications.Repository,
) *Handler {
	return &Handler{
		repo:          repo,
		hub:           hub,
		notifications: notificationsRepo,
	}
}

type createConversationRequest struct {
	UserID int `json:"user_id"`
}

type createMessageRequest struct {
	Content string `json:"content"`
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request createConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if request.UserID <= 0 {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	conversationID, err := h.repo.GetOrCreateConversation(
		userID,
		request.UserID,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]int{
		"conversation_id": conversationID,
	})
}

func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := h.repo.GetConversations(userID)
	if err != nil {
		http.Error(
			w,
			"Could not get conversations",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(conversations)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID, err := strconv.Atoi(
		r.PathValue("id"),
	)
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation id", http.StatusBadRequest)
		return
	}

	messages, err := h.repo.GetMessages(
		conversationID,
		userID,
	)
	if err != nil {
		if errors.Is(err, ErrConversationNotFound) {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"Could not get messages",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(messages)
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID, err := strconv.Atoi(
		r.PathValue("id"),
	)
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation id", http.StatusBadRequest)
		return
	}

	var request createMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.Content) == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	message, err := h.repo.CreateMessage(
		conversationID,
		userID,
		request.Content,
	)
	if err != nil {
		if errors.Is(err, ErrConversationNotFound) {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"Could not create message",
			http.StatusInternalServerError,
		)
		return
	}

	otherUserID, err := h.repo.GetOtherUser(
		conversationID,
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"Could not get conversation recipient",
			http.StatusInternalServerError,
		)
		return
	}

	otherUsername, err := h.repo.GetUser(otherUserID)
	if err != nil {
		http.Error(
			w,
			"Could not get conversation recipient",
			http.StatusInternalServerError,
		)
		return
	}

	messageMentions := mentions.Extract(request.Content)

	mentionedRecipient := false

	for _, username := range messageMentions {
		if strings.EqualFold(username, otherUsername) {
			mentionedRecipient = true
			break
		}
	}

	if mentionedRecipient {
		notificationID, err := h.notifications.CreateDMMention(
			otherUserID,
			conversationID,
			userID,
		)

		if err != nil {
			http.Error(
				w,
				"Could not create mention notification",
				http.StatusInternalServerError,
			)
			return
		}

		conversationIDCopy := conversationID

		mentionEvent := realtime.Event{
			Type: "mention",
			Data: realtime.MentionNotification{
				ID:             notificationID,
				ServerID:       nil,
				ChannelID:      nil,
				ConversationID: &conversationIDCopy,
				MessageID:      nil,
				FromUserID:     userID,
			},
		}

		mentionData, err := json.Marshal(mentionEvent)
		if err != nil {
			http.Error(
				w,
				"Could not create mention realtime event",
				http.StatusInternalServerError,
			)
			return
		}

		h.hub.SendToUser(
			otherUserID,
			mentionData,
		)
	}

	event, err := json.Marshal(map[string]interface{}{
		"type": "dm_created",
		"data": message,
	})
	if err != nil {
		http.Error(
			w,
			"Could not create realtime event",
			http.StatusInternalServerError,
		)
		return
	}

	h.hub.SendToUser(otherUserID, event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(message)
}

func (h *Handler) MarkConversationRead(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	conversationID, err := strconv.Atoi(
		r.PathValue("id"),
	)
	if err != nil {
		http.Error(
			w,
			"Invalid conversation ID",
			http.StatusBadRequest,
		)
		return
	}

	var body struct {
		MessageID int `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if body.MessageID <= 0 {
		http.Error(
			w,
			"Invalid message ID",
			http.StatusBadRequest,
		)
		return
	}

	err = h.repo.MarkConversationRead(
		conversationID,
		userID,
		body.MessageID,
	)
	if errors.Is(err, ErrConversationNotFound) {
		http.Error(
			w,
			"Conversation not found",
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		http.Error(
			w,
			"Could not mark conversation as read",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
