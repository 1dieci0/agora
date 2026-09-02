package realtime

import (
	"net/http"

	"agora/internal/users"

	"github.com/coder/websocket"
)

type UserHandler struct {
	hub *UserHub
}

func NewUserHandler(hub *UserHub) *UserHandler {
	return &UserHandler{
		hub: hub,
	}
}

func (h *UserHandler) Connect(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		ServerID: 0,
		send:     make(chan []byte, 32),
	}

	h.hub.Add(client)
	go client.writeLoop()

	defer func() {
		h.hub.Remove(client)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			return
		}
	}
}
