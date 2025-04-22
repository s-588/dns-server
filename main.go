package main

import (
	"log/slog"
	"os"

	"github.com/prionis/dns-server/server"
)

func main() {
	s, err := server.NewServer(":53")
	if err != nil {
		slog.Error("critical error", "err", err)
		os.Exit(1)
	}
	slog.Error("critical error", "err", s.Start())
}
