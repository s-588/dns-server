// Package popup provides a model for managing popup messages in a TUI application.
package popup

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/style"
)

// Model represents the model for managing popup messages in the TUI.
type Model struct {
	maxID int   // Unique identifier for the next popup message
	msgs  []Msg // Slice of currently active popup messages
}

// Msg represents a popup message with an ID, level, message content, duration, and timer.
type Msg struct {
	id       int           // Unique identifier for the popup message
	Level    string        // Level of the message (INFO, ERROR, WARNING, SUCCESS)
	Msg      string        // Content of the popup message
	Duration time.Duration // Duration for which the popup message should be displayed
	Timer    *time.Timer   // Timer for managing the popup message duration
}

// ClearMsg is a message used to clear a popup message after its duration has elapsed.
type ClearMsg struct {
	id int // Unique identifier of the popup message to be cleared
}

// NewModel creates a new instance of PopupModel.
func NewModel() Model {
	return Model{
		msgs: make([]Msg, 0),
	}
}

// Init initializes the popup model and starts listening for incoming popup messages.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the popup model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case Msg:
		timer := time.NewTimer(msg.Duration)
		newMsg := Msg{
			id:       m.maxID,
			Level:    msg.Level,
			Msg:      msg.Msg,
			Timer:    timer,
			Duration: msg.Duration,
		}
		m.msgs = append(m.msgs, newMsg)
		cmds = append(cmds, tea.Tick(msg.Duration,
			func(t time.Time) tea.Msg {
				return ClearMsg{newMsg.id}
			}))
		m.maxID++

	case ClearMsg:
		idx := -1
		for i, message := range m.msgs {
			if message.id == msg.id {
				idx = i
			}
		}
		if idx != -1 {
			m.msgs = slices.Delete(m.msgs, idx, idx+1)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the popup messages on the screen.
func (p Model) View() string {
	if len(p.msgs) == 0 {
		return ""
	}

	var out strings.Builder
	for _, msg := range p.msgs {
		if msg.Level == "" {
			continue
		}
		level := msg.Level[:1] + strings.ToLower(msg.Level[1:])
		var body strings.Builder
		body.WriteString(level)
		body.WriteString(": ")
		body.WriteString(msg.Msg)

		switch msg.Level {
		case "INFO":
			out.WriteString(style.BlueBoarderStyle.Render(body.String()))
		case "ERROR":
			out.WriteString(style.RedBoarderStyle.Render(body.String()))
		case "WARNING":
			out.WriteString(style.RedBoarderStyle.Render(body.String()))
		case "SUCCESS":
			out.WriteString(style.GreenBoarderStyle.Render(body.String()))
		}
		out.WriteRune('\n')
	}
	return out.String()
}

// Warning creates a warning popup message with the specified message and duration.
func Warning(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return Msg{
			Level:    "WARNING",
			Msg:      msg,
			Duration: d,
		}
	}
}

// Success creates a tea.Cmd that sends a PopupMsg with level "SUCCESS" and the provided message and duration.
func Success(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return Msg{
			Level:    "SUCCESS",
			Msg:      msg,
			Duration: d,
		}
	}
}

// Info creates a PopupMsg with level "INFO" and the provided message and duration.
func Info(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return Msg{
			Level:    "INFO",
			Msg:      msg,
			Duration: d,
		}
	}
}

// Error creates a PopupMsg with level "ERROR" and the provided message and duration.
func Error(msg string, t string) tea.Cmd {
	return func() tea.Msg {
		d, err := time.ParseDuration(t)
		if err != nil && t == "" {
			d = 4 * time.Second
		}
		return Msg{
			Level:    "ERROR",
			Msg:      msg,
			Duration: d,
		}
	}
}
