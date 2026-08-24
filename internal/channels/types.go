package channels

type Channel struct {
	ID       int    `json:"id"`
	ServerID int    `json:"server_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type CreateChannelRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ChannelResponse struct {
	Message string  `json:"message"`
	Channel Channel `json:"channel"`
}

type ChannelsResponse struct {
	Channels []Channel `json:"channels"`
}
