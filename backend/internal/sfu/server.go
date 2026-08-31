package sfu

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/coder/websocket"
	"github.com/pion/webrtc/v4"
)

type Server struct {
	api *webrtc.API

	mu sync.RWMutex

	rooms map[int]*Room
}

func NewServer() *Server {
	mediaEngine := &webrtc.MediaEngine{}

	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
	)

	return &Server{
		api:   api,
		rooms: make(map[int]*Room),
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/sfu/{roomID}",
		s.HandleHTTP,
	)

	log.Printf("SFU: listening on %s", addr)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) getOrCreateRoom(
	roomID int,
) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.rooms[roomID]

	if room == nil {
		room = NewRoom(roomID)
		s.rooms[roomID] = room

		log.Printf(
			"SFU: created room=%d",
			roomID,
		)
	}

	return room
}

func (s *Server) removeRoomIfEmpty(
	roomID int,
	room *Room,
) {
	if !room.IsEmpty() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	/*
	 * Make sure another goroutine hasn't replaced
	 * the room while we weren't holding the lock.
	 */
	if s.rooms[roomID] == room {
		delete(s.rooms, roomID)

		log.Printf(
			"SFU: removed empty room=%d",
			roomID,
		)
	}
}

func (s *Server) HandleHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	roomID, err := strconv.Atoi(
		r.PathValue("roomID"),
	)
	if err != nil {
		http.Error(
			w,
			"invalid room ID",
			http.StatusBadRequest,
		)
		return
	}

	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(
			w,
			"missing SFU token",
			http.StatusUnauthorized,
		)
		return
	}

	claims, err := ParseSFUToken(token)
	if err != nil {
		http.Error(
			w,
			"invalid SFU token",
			http.StatusUnauthorized,
		)
		return
	}

	if claims.ChannelID != roomID {
		http.Error(
			w,
			"token channel mismatch",
			http.StatusForbidden,
		)
		return
	}

	userID := claims.UserID

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

	defer conn.Close(
		websocket.StatusNormalClosure,
		"",
	)

	room := s.getOrCreateRoom(roomID)

	pc, err := s.api.NewPeerConnection(
		webrtc.Configuration{},
	)
	if err != nil {
		log.Printf(
			"SFU: create PeerConnection failed user=%d: %v",
			userID,
			err,
		)
		return
	}

	peer := &Peer{
		UserID:  userID,
		RoomID:  roomID,
		Conn:    conn,
		PC:      pc,
		Context: r.Context(),

		outgoingTracks: make(map[int]*webrtc.TrackLocalStaticRTP),
	}

	room.AddPeer(peer)

	defer func() {
		/*
		* Remove the peer from the room first.
		 */
		removedPeer := room.RemovePeer(userID)

		if removedPeer == nil {
			return
		}

		/*
		* Remove this user's forwarded tracks
		* from every remaining subscriber.
		 */
		s.removeForwardedTracks(
			room,
			userID,
		)

		/*
		* Now close the publisher's PeerConnection.
		 */
		if err := pc.Close(); err != nil {
			log.Printf(
				"SFU: PeerConnection close failed user=%d: %v",
				userID,
				err,
			)
		}

		s.removeRoomIfEmpty(
			roomID,
			room,
		)

		log.Printf(
			"SFU: peer left room=%d user=%d",
			roomID,
			userID,
		)
	}()

	log.Printf(
		"SFU: peer joined room=%d user=%d",
		roomID,
		userID,
	)

	s.setupPeerHandlers(peer)

	s.readSignaling(peer)
}

func (s *Server) setupPeerHandlers(
	peer *Peer,
) {
	peer.PC.OnConnectionStateChange(
		func(state webrtc.PeerConnectionState) {
			log.Printf(
				"SFU: user=%d room=%d connection=%s",
				peer.UserID,
				peer.RoomID,
				state,
			)
		},
	)

	peer.PC.OnICEConnectionStateChange(
		func(state webrtc.ICEConnectionState) {
			log.Printf(
				"SFU: user=%d room=%d ICE=%s",
				peer.UserID,
				peer.RoomID,
				state,
			)
		},
	)

	peer.PC.OnTrack(
		func(
			track *webrtc.TrackRemote,
			receiver *webrtc.RTPReceiver,
		) {
			log.Printf(
				"SFU: user=%d published track=%s codec=%s",
				peer.UserID,
				track.Kind(),
				track.Codec().MimeType,
			)

			room := s.getOrCreateRoom(
				peer.RoomID,
			)

			if err := s.forwardTrack(
				room,
				peer,
				track,
			); err != nil {
				log.Printf(
					"SFU: failed forwarding track user=%d: %v",
					peer.UserID,
					err,
				)

				return
			}
		},
	)
}

func (s *Server) readSignaling(
	peer *Peer,
) {
	for {
		_, data, err := peer.Conn.Read(
			peer.Context,
		)
		if err != nil {
			return
		}

		var message SignalMessage

		if err := json.Unmarshal(
			data,
			&message,
		); err != nil {
			log.Printf(
				"SFU: invalid signaling message user=%d: %v",
				peer.UserID,
				err,
			)
			continue
		}

		switch message.Type {
		case "offer":
			if err := s.handleOffer(
				peer,
				message,
			); err != nil {
				log.Printf(
					"SFU: offer failed user=%d: %v",
					peer.UserID,
					err,
				)

				return
			}

		case "answer":
			if err := s.handleAnswer(
				peer,
				message,
			); err != nil {
				log.Printf(
					"SFU: answer failed user=%d: %v",
					peer.UserID,
					err,
				)

				return
			}

		default:
			log.Printf(
				"SFU: unknown signal type=%q user=%d",
				message.Type,
				peer.UserID,
			)
		}
	}
}

func (s *Server) handleOffer(
	peer *Peer,
	message SignalMessage,
) error {
	err := peer.PC.SetRemoteDescription(
		webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  message.SDP,
		},
	)
	if err != nil {
		return err
	}

	answer, err := peer.PC.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err := peer.PC.SetLocalDescription(answer); err != nil {
		return err
	}

	localDescription := peer.PC.LocalDescription()
	if localDescription == nil {
		return nil
	}

	return s.sendSignal(
		peer,
		SignalMessage{
			Type: "answer",
			SDP:  localDescription.SDP,
		},
	)
}

func (s *Server) forwardTrack(
	room *Room,
	publisher *Peer,
	remoteTrack *webrtc.TrackRemote,
) error {
	codec := remoteTrack.Codec()

	log.Printf(
		"SFU: forwarding user=%d track=%s codec=%s",
		publisher.UserID,
		remoteTrack.Kind(),
		codec.MimeType,
	)

	for _, subscriber := range room.Peers() {
		/*
		 * Never send someone's track back to themselves.
		 */
		if subscriber.UserID == publisher.UserID {
			continue
		}

		localTrack, err :=
			webrtc.NewTrackLocalStaticRTP(
				codec.RTPCodecCapability,
				"audio",
				"agora",
			)

		if err != nil {
			return err
		}

		_, err = subscriber.PC.AddTrack(
			localTrack,
		)
		if err != nil {
			return err
		}

		subscriber.mu.Lock()

		subscriber.outgoingTracks[publisher.UserID] = localTrack

		subscriber.mu.Unlock()

		log.Printf(
			"SFU: user=%d subscribed to user=%d",
			subscriber.UserID,
			publisher.UserID,
		)

		if err := s.renegotiate(subscriber); err != nil {
			return err
		}

		go s.copyRTP(
			remoteTrack,
			localTrack,
			publisher,
		)
	}

	return nil
}

func (s *Server) copyRTP(
	remoteTrack *webrtc.TrackRemote,
	localTrack *webrtc.TrackLocalStaticRTP,
	publisher *Peer,
) {
	for {
		packet, _, err := remoteTrack.ReadRTP()

		if err != nil {
			log.Printf(
				"SFU: RTP ended user=%d: %v",
				publisher.UserID,
				err,
			)

			return
		}

		if err := localTrack.WriteRTP(packet); err != nil {
			log.Printf(
				"SFU: RTP forwarding failed user=%d: %v",
				publisher.UserID,
				err,
			)

			return
		}
	}
}

func (s *Server) sendSignal(
	peer *Peer,
	message SignalMessage,
) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	peer.signalingMu.Lock()
	defer peer.signalingMu.Unlock()

	return peer.Conn.Write(
		peer.Context,
		websocket.MessageText,
		data,
	)
}

func (s *Server) renegotiate(
	peer *Peer,
) error {
	peer.negotiationMu.Lock()
	defer peer.negotiationMu.Unlock()

	offer, err := peer.PC.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err := peer.PC.SetLocalDescription(offer); err != nil {
		return err
	}

	localDescription := peer.PC.LocalDescription()

	if localDescription == nil {
		return nil
	}

	log.Printf(
		"SFU: sending renegotiation offer user=%d",
		peer.UserID,
	)

	return s.sendSignal(
		peer,
		SignalMessage{
			Type: "offer",
			SDP:  localDescription.SDP,
		},
	)
}

func (s *Server) handleAnswer(
	peer *Peer,
	message SignalMessage,
) error {
	return peer.PC.SetRemoteDescription(
		webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  message.SDP,
		},
	)
}

func (s *Server) removeForwardedTracks(
	room *Room,
	publisherID int,
) {
	for _, subscriber := range room.Peers() {
		if subscriber.UserID == publisherID {
			continue
		}

		subscriber.mu.Lock()

		localTrack, exists :=
			subscriber.outgoingTracks[publisherID]

		if exists {
			delete(
				subscriber.outgoingTracks,
				publisherID,
			)
		}

		subscriber.mu.Unlock()

		if !exists {
			continue
		}

		log.Printf(
			"SFU: removing forwarded track publisher=%d subscriber=%d",
			publisherID,
			subscriber.UserID,
		)

		/*
		 * Find the sender associated with this
		 * TrackLocalStaticRTP and remove it.
		 */
		for _, sender := range subscriber.PC.GetSenders() {
			if sender.Track() == localTrack {
				if err := subscriber.PC.RemoveTrack(sender); err != nil {
					log.Printf(
						"SFU: failed removing track publisher=%d subscriber=%d: %v",
						publisherID,
						subscriber.UserID,
						err,
					)
				}

				break
			}
		}

		/*
		 * Renegotiate after removing the track.
		 */
		if err := s.renegotiate(subscriber); err != nil {
			log.Printf(
				"SFU: renegotiation after publisher leave failed publisher=%d subscriber=%d: %v",
				publisherID,
				subscriber.UserID,
				err,
			)
		}
	}
}
