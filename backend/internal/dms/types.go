package dms

type Conversation struct {
	ID           int     `json:"id"`
	UserID       int     `json:"user_id"`
	Username     string  `json:"username"`
	AvatarURL    *string `json:"avatar_url"`
	CreatedAt    string  `json:"created_at"`
	UnreadCount  int     `json:"unread_count"`
	MentionCount int     `json:"mention_count"`
}
type DirectMessage struct {
	ID             int     `json:"id"`
	ConversationID int     `json:"conversation_id"`
	UserID         int     `json:"user_id"`
	Username       string  `json:"username"`
	AvatarURL      *string `json:"avatar_url"`
	Content        string  `json:"content"`
	CreatedAt      string  `json:"created_at"`
}
