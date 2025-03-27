package server

import (
	"log/slog"
	"net"

	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
)

type Server struct {
	port string
	db   sqlite.DB
}

func NewServer(port string) (Server, error) {
	s := Server{}

	db, err := sqlite.NewDB()
	if err != nil {
		return s, err
	}
	slog.Info("connection with database established")

	s = Server{
		port: port,
		db:   db,
	}
	slog.Info("server created")
	return s, nil
}

func (s Server) Start() error {
	slog.Info("server started")

	addr, err := net.ResolveUDPAddr("udp", s.port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	for {
		data := make([]byte, 512)
		_, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			slog.Error("read from udp conn", "err", err)
			continue
		}
		go s.handleRequest(data, conn, addr)
	}
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
