package dms

type Participant struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

type Conversation struct {
	ID           int           `json:"id"`
	Participants []Participant `json:"participants"`
	CreatedAt    string        `json:"created_at"`
}

type Message struct {
	ID             int    `json:"id"`
	ConversationID int    `json:"conversation_id"`
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}
