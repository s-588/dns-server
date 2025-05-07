package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type msgPopup struct {
	msgChan chan string
	level   string
	msg     string
	show    bool
}

type (
	popupMsg      string
	clearPopupMsg struct{}
)

func listenForPopupMsg(msgs <-chan string) tea.Cmd {
	return func() tea.Msg {
		return popupMsg(<-msgs)
	}
}

func (m *model) hidePopup(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearPopupMsg(struct{}{})
	})
}

func (m *model) popupInfo(msg string) {
	m.msgPopup.level = "INFO"
	m.msgPopup.show = true
	m.msgPopup.msgChan <- msg
}

func (m *model) popupError(msg string) {
	m.msgPopup.level = "ERROR"
	m.msgPopup.show = true
	m.msgPopup.msgChan <- msg
}

func (m *model) popupSuccess(msg string) {
	m.msgPopup.level = "SUCCESS"
	m.msgPopup.show = true
	m.msgPopup.msgChan <- msg
}

func (m *model) popupClear() {
	m.msgPopup.level = ""
	m.msgPopup.show = false
}
