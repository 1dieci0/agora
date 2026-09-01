package realtime

import (
	"context"

	"github.com/coder/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserID   int
	ServerID int

	send chan []byte
}

func (c *Client) writeLoop() {
	for {
		message, ok := <-c.send

		if !ok {
			return
		}

		err := c.Conn.Write(
			context.Background(),
			websocket.MessageText,
			message,
		)

		if err != nil {
			return
		}
	}
}
