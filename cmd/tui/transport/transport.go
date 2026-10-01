// Package transport provides the transport layer for the TUI, including HTTP client and WebSocket connection.
package transport

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/publicsuffix"
)

// Transport represents the transport layer for the TUI.
type Transport struct {
	Addr       string
	httpAddr   string
	HTTPClient http.Client
	WSConn     *websocket.Conn
}

// New creates a new Transport instance with the specified address and initializes the HTTP client with a cookie jar.
func New(addr string) (*Transport, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("can't create cookie jar for http client: %w", err)
	}
	client := http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
	t := &Transport{
		Addr:       addr,
		httpAddr:   "http://" + addr,
		HTTPClient: client,
	}
	return t, nil
}
