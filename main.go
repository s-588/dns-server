package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":1053")
	if err != nil {
		slog.Error("create port", "err", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("create listener", "err", err)
		os.Exit(1)
	}
	for {
		data := make([]byte, 512)
		n, err := conn.Read(data)
		if err != nil {
			slog.Error("read from connection", "err", err)
		} else {
			slog.Info("readed data from connection", "n", n)
		}
		header, questions, err := DecodeRequest(data)
		if err != nil {
			slog.Error("decode request", "err", err)
		} else {
			slog.Info("readed data from connection",
				"header", fmt.Sprintf("%v", *header),
				"questions", fmt.Sprintf("%v", questions))
		}
	}
}
