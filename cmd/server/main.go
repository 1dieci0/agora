package main

import (
	"log"
	"net/http"

	"agora/internal/app"
)

func main() {
	application, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	log.Println("Agora server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", application.Router()); err != nil {
		log.Fatal(err)
	}
}
