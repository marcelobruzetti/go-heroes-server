// Package main starts the Go Heroes multiplayer server.
package main

import (
	"go-heroes-server/server"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := server.NewServer(":" + port)
	log.Fatal(srv.Start())
}
