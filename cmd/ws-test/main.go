package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func main() {
	ctx := context.Background()

	headers := http.Header{}
	headers.Set("Cookie", "session=01da146a52515038f0ea19dd3060469cb4deff820b87e62efee5f501315f8072")

	conn, _, err := websocket.Dial(
		ctx,
		"ws://localhost:8080/ws/channels/1",
		&websocket.DialOptions{
			HTTPHeader: headers,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close(websocket.StatusNormalClosure, "")

	fmt.Println("Connected to WebSocket")

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			log.Fatal(err)
		}

		var event Event

		if err := json.Unmarshal(data, &event); err != nil {
			log.Println("Invalid event:", err)
			continue
		}

		switch event.Type {
		case "message_created":
			fmt.Println("New message:", string(event.Data))

		default:
			fmt.Println("Unknown event:", event.Type)
		}
	}
}
