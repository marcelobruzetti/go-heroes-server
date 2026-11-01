// Package main starts the Go Heroes multiplayer server.
package main

import (
	"go-heroes-server/server"
	"log"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	webSocketAddr := os.Getenv("WEBSOCKET_ADDR")
	if webSocketAddr == "" {
		webSocketAddr = ":8082"
	}

	var allowedOrigins []string
	for _, origin := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	srv := server.NewServer(":"+port, webSocketAddr, allowedOrigins)
	log.Fatal(srv.Start())
}
