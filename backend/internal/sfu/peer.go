package sfu

import (
	"context"
	"sync"

	"github.com/coder/websocket"
	"github.com/pion/webrtc/v4"
)

type Peer struct {
	UserID int
	RoomID int

	Conn *websocket.Conn
	PC   *webrtc.PeerConnection

	Context context.Context

	mu sync.Mutex

	signalingMu sync.Mutex

	negotiationMu sync.Mutex

	outgoingTracks map[int]*webrtc.TrackLocalStaticRTP
}
