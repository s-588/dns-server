package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	mx    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

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

func (w *WebSocket) AddConn(conn *websocket.Conn) {
	w.mx.Lock()
	defer w.mx.Unlock()
	w.conns[conn] = struct{}{}
}

func (w *WebSocket) DeleteConn(conn *websocket.Conn) {
	w.mx.Lock()
	defer w.mx.Unlock()
	delete(w.conns, conn)
}
