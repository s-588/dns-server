package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
)

func waitForLogMsg(msgs chan map[string]any) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-msgs
		if !ok {
			return nil
		}
		return logMsg(msg)
	}
}

func (m model) readSocket() {
	buf := make([]byte, 1024)
	log := make(map[string]any)

	for {
		n, err := m.sockConn.Read(buf)
		if err != nil {
			close(m.logMsgChan)
			return
		}
		err = json.Unmarshal(buf[:n], &log)
		m.logMsgChan <- log
	}
}

type logMsg map[string]any
