package tui

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Socket struct {
	Conns map[net.Conn]struct{}
}

func NewUISocket() (*Socket, error) {
	err := clearTemp()
	if err != nil {
		return nil, err
	}

	err = prepareSocketFolder()
	if err != nil {
		return nil, err
	}

	conns := make(map[net.Conn]struct{}, 0)

	sock := &Socket{conns}
	go sock.handleSocket()

	return sock, nil
}

func clearTemp() error {
	tempDir := os.TempDir()
	err := filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.Contains(filepath.Base(path), "DNS-server") {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
		return nil
	},
	)
	return err
}

func prepareSocketFolder() error {
	err := os.Mkdir(os.TempDir()+"/DNS-server", 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func (sock *Socket) handleSocket() {
	socket, err := net.Listen("unix", os.TempDir()+"/DNS-server/ui.sock")
	if err != nil {
	}
	defer socket.Close()

	for {
		conn, err := socket.Accept()
		if err != nil {
		}
		sock.Conns[conn] = struct{}{}
	}
}

func (sock *Socket) Write(p []byte) (int, error) {
	retry := make(map[net.Conn]int, 0)
	for conn := range sock.Conns {
		_, err := conn.Write(p)
		if err != nil {
			if retry[conn] >= 3 {
				continue
			}
			sock.Conns[conn] = struct{}{}
			retry[conn]++
		}
	}
	return 0, nil
}

func MakeSockConn() (net.Conn, error) {
	conn, err := net.Dial("unix", os.TempDir()+"/DNS-server"+"/ui.sock")
	if err != nil {
		return nil, fmt.Errorf("can't connect to server unix socket: %w", err)
	}
	return conn, nil
}
