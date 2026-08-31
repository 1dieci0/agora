package sfu

type SignalMessage struct {
	Type string `json:"type"`

	UserID       int `json:"user_id,omitempty"`
	TargetUserID int `json:"target_user_id,omitempty"`

	SDP string `json:"sdp,omitempty"`

	Candidate *ICECandidate `json:"candidate,omitempty"`
}

type ICECandidate struct {
	Candidate        string  `json:"candidate"`
	SDPMid           *string `json:"sdp_mid,omitempty"`
	SDPMLineIndex    *uint16 `json:"sdp_mline_index,omitempty"`
	UsernameFragment *string `json:"username_fragment,omitempty"`
}
