package server

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"

	"github.com/prionis/dns-server/internal/database"
)

type Server struct {
	dnsPort  string
	httpPort string
	logger   Logger
	db       database.Repository
	wsConns  []*websocket.Conn
}

func NewServer(opts ...Option) (Server, error) {
	s := Server{}

	conf := options{
		dnsPort:  ":53",
		httpPort: ":8083",
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

	go s.serveHTTP()

	s.logger.Info("server listen DNS requests on " + s.dnsPort)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	s.logger.Info("Signal(" + sig.String() + ") recived, terminating")
	return nil
}

func (s Server) serveHTTP() {
	router := chi.NewRouter()

	router.Route("/auth", func(r chi.Router) {
		r.Use(s.timeoutMiddleware(10 * time.Second))
		r.Post("/login", s.loginHandler)
		r.Post("/register", s.registerHandler)
	})

	router.Route("/api", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/all", s.getAllUsersHandler)
			r.Get("/{login}", s.getUserHandler)
			r.Delete("/", s.deleteUserHandler)
			r.Patch("/", s.patchUserHandler)
		})

		r.Route("/rrs", func(r chi.Router) {
			r.Get("/", s.getResourceRecordsHandler)
			r.Delete("/{id}", s.deleteRRHandler)
			r.Post("/", s.postRRHandler)
			r.Patch("/{id}", s.patchRRHandler)
		})

		r.Route("/logs", func(r chi.Router) {
			r.Use(s.authenticationMiddleware)
			r.Use(s.authorizationMiddleware("user"))
			r.HandleFunc("/", s.websocketHandler)
		})
	})

	server := http.Server{
		Addr:         s.httpPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.logger.Info("server listen HTTP requests on " + s.httpPort)
	server.ListenAndServe()
}

func (s Server) serveDNS(net string) {
	dnsServer := dns.Server{
		Net:  net,
		Addr: s.dnsPort,
	}
	dnsServer.ListenAndServe()
}
