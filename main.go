package main

import (
	"go-heroes-server/server"
	"log"
)

func main() {
	srv := server.NewServer(":3000")
	log.Fatal(srv.Start())
}
