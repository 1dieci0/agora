package realtime

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type ClientEvent struct {
	Type string `json:"type"`
	Data struct {
		ChannelID int `json:"channel_id"`
	} `json:"data"`
}
