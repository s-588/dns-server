// Package table contain table model for TUI.
package table

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// Descriptor describes the logical structure of a table.
type Descriptor struct {
	Columns []table.Column

	RefreshFn func(t *transport.Transport) ([]table.Row, error)

	FilterFn     func(inputs []textinput.Model, rows []table.Row) ([]table.Row, error)
	FilterFields []textinput.Model

	SortFn        func(column int, rows []table.Row, ascending bool) []table.Row
	SortedColumn  int
	SortAscending bool

	SearchFn func(query string, rows []table.Row) []table.Row
	DeleteFn func(t *transport.Transport, id int64) error

	AddFn        func(t *transport.Transport, inputs []textinput.Model) (table.Row, error)
	InputFields  []textinput.Model
	UpdateFn     func(t *transport.Transport, inputs []textinput.Model, id int64) (table.Row, error)
	UpdateFields []textinput.Model
}

// Model represents the state of a table.
type Model struct {
	Header string
	// Width and height of the screen.
	width, height int

	Buttons    []string
	Descriptor Descriptor
	// Selected button.
	cursor        int
	onButtonPress func(index int, table Model) (Model, tea.Cmd)

	Table         table.Model
	UnchangedRows []table.Row
}

// NewModel creates a new table model.
func NewModel(descriptor Descriptor,
	width, height int,
	buttons []string, rows []table.Row,
	buttonHandler func(int, Model) (Model, tea.Cmd),
	header string,
) Model {
	t := table.New(table.WithColumns(descriptor.Columns), table.WithRows(rows), table.WithHeight(height-20))
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(style.TextColor).
		Background(style.PurpleColor).
		Bold(true)
	t.SetStyles(s)
	return Model{
		Header:        header,
		width:         width,
		height:        height,
		Buttons:       buttons,
		Descriptor:    descriptor,
		onButtonPress: buttonHandler,
		Table:         t,
		UnchangedRows: rows,
	}
}

// Init initialize TableModel.
func (dm Model) Init() tea.Cmd {
	return nil
}

// Update handles messages that coming from the user interactions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "shift+tab", "k", "down", "tab", "j":
			m, cmd = processVerticalMoves(m, cmd, msg.String())
			cmds = append(cmds, cmd)

		case "left", "h":
			if !m.Table.Focused() && m.cursor > 0 {
				m.cursor--
			} else {
				m.Table, cmd = m.Table.Update(msg)
				cmds = append(cmds, cmd)
			}

		case "right", "l":
			if !m.Table.Focused() && m.cursor < len(m.Buttons)-1 {
				m.cursor++
			} else {
				m.Table, cmd = m.Table.Update(msg)
				cmds = append(cmds, cmd)
			}

		case "ctrl+c", "esc":
			return m, func() tea.Msg {
				return UnfocusMsg{}
			}

		case "enter", " ":
			if !m.Table.Focused() {
				m, cmd = m.onButtonPress(m.cursor, m)
				cmds = append(cmds, cmd)
			}

		default:
			if m.Table.Focused() {
				m.Table, cmd = m.Table.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func processVerticalMoves(m Model, cmd tea.Cmd, button string) (Model, tea.Cmd) {
	switch button {
	case "up", "shift+tab", "k":
		if m.Table.Focused() {
			m.Table, cmd = m.Table.Update(button)
		}
	case "down", "tab", "j":

		if !m.Table.Focused() {
			m.Table.Focus()
		} else {
			m.Table, cmd = m.Table.Update(button)
		}
	}
	return m, cmd
}

// View renders page on the screen.
func (m Model) View() string {
	s := strings.Builder{}

	styledButtons := make([]string, len(m.Buttons))
	for i, btn := range m.Buttons {
		if m.cursor == i {
			if !m.Table.Focused() {
				styledButtons[i] = style.SelectedButtonStyle.Render(btn)
			} else {
				styledButtons[i] = style.SecondarySelectedButtonStyle.Render(btn)
			}
		} else {
			styledButtons[i] = style.ButtonStyle.Render(btn)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, styledButtons...)))

	s.WriteRune('\n')
	if m.Table.Focused() {
		s.WriteString(style.SelectedBoarderStyle.Render(m.Table.View()))
	} else {
		s.WriteString(style.UnselectedBoarderStyle.Render(m.Table.View()))
	}

	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

// UpdateSize set w and h to width and height of a Model.
func (m Model) UpdateSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

func (m Model) UpdateRow(index int64, row table.Row) Model {
	rows := m.Table.Rows()
	rows[index] = row
	m.Table.SetRows(rows)
	m.Table.UpdateViewport()
	return m
}

type (
	UnfocusMsg struct{}
	FocusMsg   struct{}
)

type (
	FilterRequestMsg        struct{}
	SortRequestMsg          struct{}
	AddRequestMsg           struct{}
	DeleteRequestMsg        struct{}
	UpdateRequestMsg        struct{}
	ResetRequestMsg         struct{}
	RefreshRequestMsg       struct{}
	ExportToWordRequestMsg  struct{}
	ExportToExcelRequestMsg struct{}
)
