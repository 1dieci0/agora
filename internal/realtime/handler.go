package realtime

import (
	"database/sql"
	"net/http"
	"strconv"

	"agora/internal/users"

	"github.com/coder/websocket"
)

type Handler struct {
	db  *sql.DB
	hub *Hub
}

func NewHandler(db *sql.DB, hub *Hub) *Handler {
	return &Handler{
		db:  db,
		hub: hub,
	}
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
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

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:      conn,
		UserID:    userID,
		ChannelID: channelID,
		send:      make(chan []byte, 32),
	}

	h.hub.Add(client)

	go client.writeLoop()

	defer func() {
		h.hub.Remove(client)
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			return
		}
	}
}
