package voice

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type Client struct {
	UserID    int
	ChannelID int

	Conn *websocket.Conn

	send chan []byte

	closeOnce sync.Once
	done      chan struct{}
}

func NewClient(
	userID int,
	channelID int,
	conn *websocket.Conn,
) *Client {
	return &Client{
		UserID:    userID,
		ChannelID: channelID,
		Conn:      conn,

		send: make(chan []byte, 32),
		done: make(chan struct{}),
	}
}

func (c *Client) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-c.done:
			return

		case message := <-c.send:
			if err := c.Conn.Write(
				ctx,
				websocket.MessageText,
				message,
			); err != nil {
				return
			}
		}
	}
}

func (c *Client) Send(message []byte) bool {
	select {
	case <-c.done:
		return false

	case c.send <- message:
		return true

	default:
		return false
	}
}

func (c *Client) Close(
	status websocket.StatusCode,
	reason string,
) {
	c.closeOnce.Do(func() {
		close(c.done)

		_ = c.Conn.Close(
			status,
			reason,
		)
	})
}
