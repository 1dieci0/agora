package channels

import (
	"agora/internal/users"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("serverID"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data CreateChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data.Name = strings.TrimSpace(data.Name)

	if len(data.Name) < 1 {
		http.Error(w, "Channel name is required", http.StatusBadRequest)
		return
	}

	if len(data.Name) > 50 {
		http.Error(w, "Channel name too long", http.StatusBadRequest)
		return
	}

	if data.Type != "text" && data.Type != "voice" {
		http.Error(w, "Invalid channel type", http.StatusBadRequest)
		return
	}

	var member bool

	err = h.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members
			WHERE server_id = ?
			AND user_id = ?
		)`,
		serverID,
		userID,
	).Scan(&member)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "You are not a member of this server", http.StatusForbidden)
		return
	}

	result, err := h.db.Exec(
		`INSERT INTO channels (server_id, name, type)
		 VALUES (?, ?, ?)`,
		serverID,
		data.Name,
		data.Type,
	)

	if err != nil {
		http.Error(w, "Could not create channel", http.StatusInternalServerError)
		return
	}

	channelID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get channel ID", http.StatusInternalServerError)
		return
	}

	channel := Channel{
		ID:       int(channelID),
		ServerID: serverID,
		Name:     data.Name,
		Type:     data.Type,
	}

	response := ChannelResponse{
		Message: "Channel created",
		Channel: channel,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("serverID"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
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
			FROM server_members
			WHERE server_id = ?
			AND user_id = ?
		)`,
		serverID,
		userID,
	).Scan(&member)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	rows, err := h.db.Query(
		`SELECT id, server_id, name, type
		 FROM channels
		 WHERE server_id = ?
		 ORDER BY id`,
		serverID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	channels := make([]Channel, 0)

	for rows.Next() {
		var channel Channel

		err := rows.Scan(
			&channel.ID,
			&channel.ServerID,
			&channel.Name,
			&channel.Type,
		)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		channels = append(channels, channel)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ChannelsResponse{
		Channels: channels,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
