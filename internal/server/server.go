package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"

	"github.com/prionis/dns-server/internal/database"
)

type Server struct {
	dnsPort  string
	httpPort string
	logger   Logger
	db       *database.Repository
}

func NewServer(opts ...Option) (Server, error) {
	s := Server{}

	conf := options{
		dnsPort:  ":53",
		logger:   slog.Default(),
	}
	for _, opt := range opts {
		opt.apply(&conf)
	}

	s = Server{
		dnsPort:  conf.dnsPort,
		httpPort: ":8083",
		db:       conf.db,
		logger:   conf.logger,
	}
	return s, nil
}

func (s Server) Start() error {
	dns.HandleFunc(".", s.dnsHandler)

	go s.serveDNS("udp")
	go s.serveDNS("tcp")

	s.logger.Info("server listen DNS requests on " + s.dnsPort)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	s.logger.Info("Signal(" + sig.String() + ") recived, terminating")
	return nil
}

func (s Server) serveDNS(net string) {
	dnsServer := dns.Server{
		Net:  net,
		Addr: s.dnsPort,
	}
	dnsServer.ListenAndServe()
}
