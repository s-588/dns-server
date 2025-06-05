package models

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/style"
)

type PopupModel struct {
	MsgChan chan PopupMsg
	msgs    []PopupMsg
}

type PopupMsg struct {
	Level    string
	Msg      string
	Duration time.Duration
	Timer    *time.Timer
}

type ClearPopupMsg struct {
	index int
}

func NewPopupModel() PopupModel {
	return PopupModel{
		MsgChan: make(chan PopupMsg, 1),
		msgs:    make([]PopupMsg, 0),
	}
}

func (p PopupModel) Init() tea.Cmd {
	return ListenForPopupMsg(p.MsgChan)
}

func (p PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case PopupMsg:
		for i, m := range p.msgs {
			if m.Msg == msg.Msg {
				m.Timer.Stop()
				p.msgs = append(p.msgs[:i], p.msgs[i+1:]...)
				break
			}
		}

		timer := time.NewTimer(msg.Duration)
		newMsg := PopupMsg{
			Level:    msg.Level,
			Msg:      msg.Msg,
			Timer:    timer,
			Duration: msg.Duration,
		}
		p.msgs = append(p.msgs, newMsg)
		cmds = append(cmds, tea.Tick(msg.Duration,
			func(t time.Time) tea.Msg { return ClearPopupMsg{len(p.msgs) - 1} }))
	case ClearPopupMsg:
		if msg.index >= 0 && msg.index < len(p.msgs) {
			p.msgs[msg.index].Timer.Stop()
			p.msgs = append(p.msgs[:msg.index], p.msgs[msg.index+1:]...)
		}
		cmds = append(cmds, ListenForPopupMsg(p.MsgChan))
	}

	return p, tea.Batch(cmds...)
}

func (p PopupModel) View() string {
	if len(p.msgs) == 0 {
		return ""
	}

	s := strings.Builder{}
	for _, msg := range p.msgs {
		s := strings.Builder{}
		switch msg.Level {
		case "INFO":
			s.WriteString("INFO: " + msg.Msg)
			return style.BlueBoarderStyle.Render(s.String())

		case "ERROR":
			s.WriteString("ERROR: " + msg.Msg)
			return style.RedBoarderStyle.Render(s.String())

		case "SUCCESS":
			s.WriteString("SUCCESS: " + msg.Msg)
			return style.GreenBoarderStyle.Render(s.String())
		}
	}
	return s.String()
}

func ListenForPopupMsg(msgs <-chan PopupMsg) tea.Cmd {
	return func() tea.Msg {
		return <-msgs
	}
}

func Info(msg string, duration time.Duration) PopupMsg {
	return PopupMsg{
		Level:    "INFO",
		Msg:      msg,
		Duration: duration,
	}
}

func Error(msg string, duration time.Duration) PopupMsg {
	return PopupMsg{
		Level:    "ERROR",
		Msg:      msg,
		Duration: duration,
	}
}

func Success(msg string, duration time.Duration) PopupMsg {
	return PopupMsg{
		Level:    "SUCCESS",
		Msg:      msg,
		Duration: duration,
	}
}
