package main

import (
	"log"

	"zcyp-im/internal/app"
)

func main() {
	server, err := app.NewServer()
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
