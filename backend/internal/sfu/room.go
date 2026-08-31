package sfu

import (
	"sync"
)

type Room struct {
	mu sync.RWMutex

	ID int

	peers map[int]*Peer
}

func NewRoom(id int) *Room {
	return &Room{
		ID:    id,
		peers: make(map[int]*Peer),
	}
}

func (r *Room) AddPeer(peer *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.peers[peer.UserID] = peer
}

func (r *Room) RemovePeer(userID int) *Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer := r.peers[userID]

	delete(r.peers, userID)

	return peer
}

func (r *Room) GetPeer(userID int) *Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.peers[userID]
}

func (r *Room) Peers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Peer, 0, len(r.peers))

	for _, peer := range r.peers {
		result = append(result, peer)
	}

	return result
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.peers) == 0
}
