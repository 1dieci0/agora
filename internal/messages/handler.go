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
	db  *sql.DB
	hub *realtime.Hub
}

func NewHandler(db *sql.DB, hub *realtime.Hub) *Handler {
	return &Handler{
		db:  db,
		hub: hub,
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

	var member bool

	err = h.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members sm
			JOIN channels c ON c.server_id = sm.server_id
			WHERE sm.user_id = ?
			AND c.id = ?
		)`,
		userID,
		channelID,
	).Scan(&member)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You cannot access this channel", http.StatusForbidden)
		return
	}

	result, err := h.db.Exec(
		`INSERT INTO messages (channel_id, user_id, content)
		 VALUES (?, ?, ?)`,
		channelID,
		userID,
		data.Content,
	)

	if err != nil {
		http.Error(w, "Could not create message", http.StatusInternalServerError)
		return
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get message ID", http.StatusInternalServerError)
		return
	}

	var message Message

	err = h.db.QueryRow(
		`SELECT
			m.id,
			m.channel_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = ?`,
		messageID,
	).Scan(
		&message.ID,
		&message.ChannelID,
		&message.UserID,
		&message.Username,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert the message to JSON for WebSocket clients.
	messageData, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send the message to everyone currently connected
	// to this channel.
	h.hub.Broadcast(channelID, messageData)

	response := MessageResponse{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
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

	var member bool

	err = h.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members sm
			JOIN channels c ON c.server_id = sm.server_id
			WHERE sm.user_id = ?
			AND c.id = ?
		)`,
		userID,
		channelID,
	).Scan(&member)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You cannot access this channel", http.StatusForbidden)
		return
	}

	rows, err := h.db.Query(
		`SELECT
			m.id,
			m.channel_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.channel_id = ?
		ORDER BY m.id DESC
		LIMIT 100`,
		channelID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var message Message

		if err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.UserID,
			&message.Username,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := MessagesResponse{
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
