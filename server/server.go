package server

import (
	"log/slog"
	"net"

	"github.com/prionis/dns-server/database"
	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
)

type Server struct {
	port   string
	logger Logger
	db     database.DB
}

func NewServer(opts ...Option) (Server, error) {
	s := Server{}

	defDB, err := sqlite.NewDB()
	if err != nil {
		return s, err
	}
	conf := options{
		port:   "53",
		logger: slog.Default(),
		db:     defDB,
	}
	for _, opt := range opts {
		opt.apply(&conf)
	}

	s = Server{
		port:   conf.port,
		db:     conf.db,
		logger: conf.logger,
	}
	s.logger.Info("server created", "config", conf)
	return s, nil
}

func (s Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("server started", "available on", conn.LocalAddr().String())

	for {
		data := make([]byte, 512)
		_, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			s.logger.Error("read from udp conn", "err", err)
			continue
		}
		go s.handleRequest(data, conn, addr)
	}
}

func (s Server) handleRequest(data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	message, err := protocol.DecodeRequest(data)
	if err != nil {
		s.logger.Error("decode request", "err", err)
		return
	}

	for _, question := range message.Questions {
		answers, err := s.db.GetRRs(question.Domain)
		if err != nil {
			s.logger.Error("get RR from db", "err", err)
			return
		}
		message.Answers = append(message.Answers, answers...)
	}

	answer, err := protocol.EncodeResponse(message)
	if err != nil {
		s.logger.Error("encode response", "err", err)
		return
	}

	s.logger.Info("request handled",
		"client addr", addr.String(),
		"header", *message.Head,
		"questions", rrSliceToLog(message.Questions),
		"answers", rrSliceToLog(message.Answers),
		"authorities", rrSliceToLog(message.Authorities),
		"additions", rrSliceToLog(message.Additionals),
	)

	_, err = conn.WriteToUDP(answer, addr)
	if err != nil {
		s.logger.Error("write response to conn", "err", err)
		return
	}
}
