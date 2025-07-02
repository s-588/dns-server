package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miekg/dns"

	"github.com/prionis/dns-server/internal/database"
)

var (
	adminRights = []string{"admin"}
	userRights  = []string{"user", "admin"}
)

type Server struct {
	dnsPort  string
	httpPort string
	logger   Logger
	db       database.Repository
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

func (s Server) Start(ws *WebSocket) error {
	_, err := s.db.GetUser(context.Background(), "admin")
	if err != nil {
		user, err := s.db.AddUser(context.Background(), database.User{
			Login:     "admin",
			FirstName: "John",
			LastName:  "Doe",
			Role:      "admin",
		}, "admin")
		if err != nil {
			s.logger.Error(fmt.Sprintf("Can't create new user: %s", err.Error()))
		} else {
			s.logger.Info(fmt.Sprintf("New admin was created: %v", user))
		}
	} else {
		s.logger.Info("admin user exists")
	}

	dns.HandleFunc(".", s.dnsHandler)

	go s.serveDNS("udp")
	go s.serveDNS("tcp")

	go s.serveHTTP(ws)

	s.logger.Info("server listen DNS requests on " + s.dnsPort)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	s.logger.Info("Signal(" + sig.String() + ") recived, terminating")
	return nil
}

func (s Server) serveHTTP(ws *WebSocket) {
	router := chi.NewRouter()
	router.Use(s.loggerMiddleware())

	router.Route("/auth", func(r chi.Router) {
		r.Use(s.timeoutMiddleware(10 * time.Second))
		r.Post("/login", s.loginHandler)
		r.Post("/register", s.registerHandler)
	})

	router.Route("/api", func(r chi.Router) {
		r.Use(s.authenticationMiddleware)

		r.Route("/users", func(r chi.Router) {
			r.Use(s.authorizationMiddleware(adminRights))
			r.Get("/all", s.getAllUsersHandler)
			r.Get("/{id}", s.getUserHandler)
			r.Delete("/", s.deleteUserHandler)
			r.Patch("/", s.patchUserHandler)
		})

		r.Route("/rrs", func(r chi.Router) {
			r.Use(s.authorizationMiddleware(userRights))
			r.Get("/all", s.getAllRecordsHandler)
			r.Get("/{id}", s.getRecordHandler)
			r.Delete("/{id}", s.deleteRRHandler)
			r.Post("/", s.postRRHandler)
			r.Patch("/{id}", s.patchRRHandler)
		})

		r.Route("/logs", func(r chi.Router) {
			r.Use(s.authorizationMiddleware(userRights))
			r.HandleFunc("/all", s.getAllLogsHandler)
			r.HandleFunc("/ws", s.websocketHandler(ws))
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
