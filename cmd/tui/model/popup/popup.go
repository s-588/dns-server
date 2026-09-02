package popup

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/style"
)

type PopupModel struct {
	maxID   int
	MsgChan chan PopupMsg
	msgs    []PopupMsg
}

type PopupMsg struct {
	id       int
	Level    string
	Msg      string
	Duration time.Duration
	Timer    *time.Timer
}

type ClearPopupMsg struct {
	id int
}

func NewPopupModel() PopupModel {
	return PopupModel{
		MsgChan: make(chan PopupMsg, 1),
		msgs:    make([]PopupMsg, 0),
	}
}

func (m PopupModel) Init() tea.Cmd {
	return ListenForPopupMsg(m.MsgChan)
}

func (m PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case PopupMsg:
		timer := time.NewTimer(msg.Duration)
		newMsg := PopupMsg{
			id:       m.maxID,
			Level:    msg.Level,
			Msg:      msg.Msg,
			Timer:    timer,
			Duration: msg.Duration,
		}
		m.msgs = append(m.msgs, newMsg)
		cmds = append(cmds, tea.Tick(msg.Duration,
			func(t time.Time) tea.Msg {
				return ClearPopupMsg{newMsg.id}
			}))
		m.maxID++

	case ClearPopupMsg:
		idx := -1
		for i, message := range m.msgs {
			if message.id == msg.id {
				idx = i
			}
		}
		if idx != -1 {
			m.msgs = slices.Delete(m.msgs, idx, idx+1)
		}
		cmds = append(cmds, ListenForPopupMsg(m.MsgChan))
	}

	return m, tea.Batch(cmds...)
}

func (p PopupModel) View() string {
	if len(p.msgs) == 0 {
		return ""
	}

	s := strings.Builder{}
	for _, msg := range p.msgs {
		s := strings.Builder{}
		level := msg.Level[:1] + strings.ToLower(msg.Level[1:])
		s.WriteString(level + ": ")
		s.WriteString(msg.Msg)
		switch msg.Level {
		case "INFO":
			return style.BlueBoarderStyle.Render(s.String())

		case "ERROR":
			return style.RedBoarderStyle.Render(s.String())

		case "WARNING":
			return style.RedBoarderStyle.Render(s.String())

		case "SUCCESS":
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

func Warning(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return PopupMsg{
			Level:    "WARNING",
			Msg:      msg,
			Duration: d,
		}
	}
}

func Success(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return PopupMsg{
			Level:    "SUCCESS",
			Msg:      msg,
			Duration: d,
		}
	}
}

func Info(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return PopupMsg{
			Level:    "INFO",
			Msg:      msg,
			Duration: d,
		}
	}
}

func Error(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return PopupMsg{
			Level:    "ERROR",
			Msg:      msg,
			Duration: d,
		}
	}
}
