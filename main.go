package main

import (
	"log/slog"
	"os"

	"github.com/prionis/dns-server/server"
)

func main() {
	s, err := server.NewServer(server.SetPort("1053"), server.WithLogger(slog.Default()))
	if err != nil {
		slog.Error("critical error", "err", err)
		os.Exit(1)
	}
	slog.Error("critical error", "err", s.Start())
}
