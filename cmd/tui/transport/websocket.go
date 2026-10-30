package transport

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// EstablishWebsocketConnection setting the wsConn of the Transport.
func (t *Transport) EstablishWebsocketConnection(jar http.CookieJar, addr string) error {
	dialer := websocket.Dialer{Jar: jar}

	u := url.URL{Scheme: "ws", Path: "/api/logs/ws", Host: addr}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	t.WSConn = conn
	return nil
}

// LogMsg represents a log message received from the WebSocket connection.
type LogMsg struct {
	Time       time.Time
	Level, Msg string
}

// ListenWebSocket listens for incoming WebSocket messages and sends them to the provided channel.
func (t *Transport) ListenWebSocket(msgs chan LogMsg) {
	for {
		var msg LogMsg
		if err := t.WSConn.ReadJSON(&msg); err != nil {
			slog.Error("websocket listener stopped", "error", err)
			return
		}
		msgs <- msg
	}
}
