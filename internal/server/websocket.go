package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket represents a WebSocket writer that can write messages to multiple WebSocket connections.
type WebSocket struct {
	mx    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

// NewWSWriter creates a new WebSocket writer.
func NewWSWriter() *WebSocket {
	return &WebSocket{
		conns: make(map[*websocket.Conn]struct{}),
	}
}

func (w *WebSocket) Write(msg []byte) (n int, err error) {
	for conn := range w.conns {
		err := conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			return n, err
		}
	}
	return n, err
}

// AddConn adds a WebSocket connection to the WebSocket writer.
func (w *WebSocket) AddConn(conn *websocket.Conn) {
	w.mx.Lock()
	defer w.mx.Unlock()
	w.conns[conn] = struct{}{}
}

// DeleteConn removes a WebSocket connection from the WebSocket writer.
func (w *WebSocket) DeleteConn(conn *websocket.Conn) {
	w.mx.Lock()
	defer w.mx.Unlock()
	delete(w.conns, conn)
}
