package transport

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// EstablishWebsocketConnection setting the wsConn of the Transport.
func (t *Transport) EstablishWebsocketConnection(jar http.CookieJar, addr string) error {
	dialer := websocket.DefaultDialer
	dialer.Jar = jar

	u := url.URL{Scheme: "ws", Path: "/api/logs/ws", Host: addr}
	conn, r, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	if header := r.Header.Values("Content-Type"); len(header) != 0 {
		return errors.New(header[0])
	}
	t.WSConn = conn
	return nil
}

type LogMsg struct {
	Time       time.Time
	Level, Msg string
}

func (t *Transport) ListenWebSocket(msgs chan LogMsg) {
	for {
		var msg LogMsg
		err := t.WSConn.ReadJSON(&msg)
		if err == nil {
			msgs <- msg
		}
	}
}
