package main

import (
	"log"

	"zcyp-im/internal/app"
)

func main() {
	server, err := app.NewWebSocketServer()
	if err != nil {
		log.Fatalf("init websocket server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("run websocket server: %v", err)
	}
}
