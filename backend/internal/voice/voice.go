package voice

import (
	"sync"

	"github.com/coder/websocket"
)

type Client struct {
	UserID    int
	ChannelID int

	Conn *websocket.Conn

	writeMu sync.Mutex

	closeMu sync.Mutex
	closed  bool
}

func (c *Client) close(
	status websocket.StatusCode,
	reason string,
) bool {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return false
	}

	c.closed = true

	_ = c.Conn.Close(
		status,
		reason,
	)

	return true
}
