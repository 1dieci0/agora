package realtime

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type ClientEvent struct {
	Type string `json:"type"`
	Data struct {
		ChannelID int  `json:"channel_id"`
		Muted     bool `json:"muted"`
		Deafened  bool `json:"deafened"`
	} `json:"data"`
}

type UnreadUpdate struct {
	ServerID    int `json:"server_id"`
	ChannelID   int `json:"channel_id"`
	MessageID   int `json:"message_id"`
	UserID      int `json:"user_id"`
	UnreadCount int `json:"unread_count"`
}

type MentionNotification struct {
	ID         int `json:"id"`
	ServerID   int `json:"server_id"`
	ChannelID  int `json:"channel_id"`
	MessageID  int `json:"message_id"`
	FromUserID int `json:"from_user_id"`
}
