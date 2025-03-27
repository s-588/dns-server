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

func (s Server) handleRequest(data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	message, err := protocol.DecodeRequest(data)
	if err != nil {
		slog.Error("decode request", "err", err)
		return
	}

	slog.Info("message processed", "head", message.Head)

	for _, question := range message.Questions {
		slog.Info("searching for domain", "domain", question.Domain)
		answers, err := s.db.GetResourceRecord(question.Domain)
		if err != nil {
			slog.Error("get RR from db", "err", err)
			return
		}
		message.Answers = append(message.Answers, answers...)
	}

	answer, err := protocol.EncodeResponse(message)
	if err != nil {
		slog.Error("encode response", "err", err)
		return
	}
	slog.Info("response encoded")

	_, err = conn.WriteToUDP(answer, addr)
	if err != nil {
		slog.Error("write response to conn", "err", err)
		return
	}
	slog.Info("request writed to client")
}
