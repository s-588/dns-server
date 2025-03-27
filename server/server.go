package server

import (
	"log/slog"
	"net"

	_ "modernc.org/sqlite"

	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
)

type Server struct {
	port    string
	udpConn *net.UDPConn
	db      sqlite.DB
}

func NewServer(port string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	db, err := sqlite.NewDB()
	if err != nil {
		return nil, err
	}

	return &Server{
		port:    port,
		udpConn: conn,
		db:      db,
	}, nil
}

func (s *Server) Start() error {
	for {
		data := make([]byte, 512)
		_, err := s.udpConn.Read(data)
		if err != nil {
			slog.Error("read from connection", "err", err)
		}

		message, err := protocol.DecodeRequest(data)
		if err != nil {
			slog.Error("decode request", "err", err)
		}

		slog.Info("message processed", "message", message)

	}
}
