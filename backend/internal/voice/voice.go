package voice

import "github.com/coder/websocket"

type Client struct {
	UserID    int
	ChannelID int

	Conn *websocket.Conn
}
