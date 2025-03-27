package main

import (
	"log"
	"os"

	"github.com/prionis/dns-server/server"
)

func main() {
	s, err := server.NewServer(":1053")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Fatal(s.Start())
}
