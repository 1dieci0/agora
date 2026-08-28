package voice

type VoiceMessage struct {
	Type string `json:"type"`

	// User who sent the message.
	UserID int `json:"user_id,omitempty"`

	// Target user for signaling messages.
	TargetUserID int `json:"target_user_id,omitempty"`

	// WebRTC SDP.
	SDP string `json:"sdp,omitempty"`

	// WebRTC ICE candidate.
	Candidate string `json:"candidate,omitempty"`
}
