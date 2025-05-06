package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type errorPopup struct {
	message string
	show    bool
}

type errorMsg string

func listenForError(errors <-chan string) tea.Cmd {
	return func() tea.Msg {
		return errorMsg(<-errors)
	}
}

func hideError(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return errorMsg("")
	})
}
