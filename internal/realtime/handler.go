package realtime

import (
	"database/sql"
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

func NewHandler(channelRepo *channels.Repository, serverRepo *servers.Repository, hub *Hub) *Handler {
	return &Handler{
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
		hub:         hub,
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

	channel, err := h.channelRepo.GetByID(channelID)
	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	member, err := h.serverRepo.IsMember(userID, channel.ServerID)
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
