// Package server provides the implementation of the HTTP and DNS server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prionis/dns-server/internal/database"
)

var (
	adminRights = []string{"admin"}
	userRights  = []string{"user", "admin"}
)

// Server represents the DNS and HTTP server.
type Server struct {
	dnsPort  string
	httpPort string
	https    bool
	db       database.Repository
	metrics  *Metrics
}

// NewServer creates a new Server instance with the provided options.
func NewServer(opts ...Option) (Server, error) {
	conf := options{
		dnsPort:  ":53",
		httpPort: ":8083",
	}
	for _, opt := range opts {
		opt.apply(&conf)
	}

	s := Server{
		dnsPort:  conf.dnsPort,
		httpPort: conf.httpPort,
		db:       conf.db,
		metrics:  NewMetrics(),
		https:    conf.https,
	}
	return s, nil
}

// Start starts the DNS and HTTP servers and listens for incoming requests.
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
			slog.Error(fmt.Sprintf("Can't create new user: %s", err.Error()))
		} else {
			slog.Info(fmt.Sprintf("New admin was created: %v", user))
		}
	} else {
		slog.Info("admin user exists")
	}

	go s.serveUDP()
	go s.serveTCP()

	go s.serveHTTP(ws)

	slog.Info("server listen DNS requests on " + s.dnsPort)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	slog.Info("Signal(" + sig.String() + ") received, terminating")
	return nil
}

func (s Server) serveHTTP(ws *WebSocket) {
	router := chi.NewRouter()
	router.Use(s.loggerMiddleware())
	router.Use(prometheusMiddleware(s.metrics))
	reg := prometheus.NewRegistry()
	s.metrics.RegisterAll(reg)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
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

	slog.Info("server listen HTTP requests on " + s.httpPort)
	err := server.ListenAndServe()
	if err != nil {
		slog.Error("can't start HTTP server: " + err.Error())
	}
}

func (s *Server) serveUDP() {
	addr, err := net.ResolveUDPAddr("udp", s.dnsPort)
	if err != nil {
		slog.Error("resolve UDP address", "error", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("listen UDP", "error", err)
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			slog.Error("can't close UDP connection", "error", err)
		}
	}()

	buf := make([]byte, 512)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		go func(data []byte, addr *net.UDPAddr) {
			resp := s.dnsHandler(data)
			if resp != nil {
				_, err := conn.WriteToUDP(resp, addr)
				if err != nil {
					slog.Error("can't write response for UDP client: " + err.Error())
				}
			}
		}(append([]byte(nil), buf[:n]...), clientAddr)
	}
}

func (s *Server) serveTCP() {
	addr, err := net.ResolveTCPAddr("tcp", s.dnsPort)
	if err != nil {
		slog.Error("resolve TCP addr", "error", err)
		return
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		slog.Error("listen TCP", "error", err)
		return
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			slog.Error("can't close TCP listener", "error", err)
		}
	}()

	buf := make([]byte, 512)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func(conn net.Conn) {
			_, err := conn.Read(buf)
			if err != nil {
				return
			}
			resp := s.dnsHandler(append([]byte(nil), buf...))
			if resp != nil {
				_, err := conn.Write(resp)
				if err != nil {
					slog.Error("can't write response for TCP client: " + err.Error())
				}
			}
		}(conn)
	}
}
