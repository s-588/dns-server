package main

import (
	"log/slog"
	"os"

	"github.com/prionis/dns-server/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdin, nil))
	config := []server.Option{
		server.SetPort("1053"),
		server.WithLogger(logger),
	}
	s, err := server.NewServer(config...)
	if err != nil {
		logger.Error("can't create new server", "config", config, "error", err)
		os.Exit(1)
	}

	if err := s.Start(); err != nil {
		logger.Error("can't start server", "config", config, "error", err)
	}
}
