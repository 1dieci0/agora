package main

import (
	"log"

	"agora/internal/sfu"
)

func main() {
	server := sfu.NewServer()

	if err := server.Start(":9000"); err != nil {
		log.Fatal(err)
	}
}
