// Package crud contains the model for adding new resource records to the database.
package crud

import (
	"strings"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// AddModel represent page for adding new resource records to the database.
type AddModel struct {
	// Width and height of the screen.
	width int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model
	table       *table.Model

	// Add, Clear and Cancel buttons.
	buttons []string
	// Selected button.
	cursor int
	// Are buttons focused or the inputs.
	focusButtons bool

	// Pointer to the database connection.
	transport *transport.Transport
}

// NewAddModel creates a new add model.
func NewAddModel(table *table.Model, t *transport.Transport, width int) AddModel {
	return AddModel{
		width:       width,
		inputFields: table.Descriptor.InputFields,
		table:       table,
		buttons:     []string{"Yes", "No"},
		transport:   t,
	}
}

// Init initializes the add model.
func (m AddModel) Init() tea.Cmd {
	return nil
}

// Update handles messages that coming from the user interactions.
func (m AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "shift+tab", "down", "tab":
			m = processAddVerticalMoves(m, msg.String())
		case "left", "h":
			if m.focusButtons && m.cursor > 0 {
				m.cursor--
			} else {
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}

		case "right", "l":
			if m.focusButtons && m.cursor < len(m.buttons)-1 {
				m.cursor++
			} else {
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		case "enter", " ":
			var cmd tea.Cmd
			m, cmd = processAddEnterPress(m, msg)
			cmds = append(cmds, cmd)
		case "esc", "q", "ctrl+c":
			if m.focusButtons {
				m.focusButtons = false
				m.inputFields[m.focusIndex].Focus()
			} else {
				return m, tea.Batch(
					popup.Info("Addition canceled", ""),
					func() tea.Msg {
						return AddCancelMsg{}
					},
				)
			}
		default:
			if !m.focusButtons {
				var cmd tea.Cmd
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func processAddEnterPress(m AddModel, msg tea.KeyMsg) (AddModel, tea.Cmd) {
	if m.focusButtons {
		if m.cursor == 0 { // Yes button
			row, err := m.table.Descriptor.AddFn(m.transport, m.inputFields)
			if err != nil {
				return m, popup.Error(err.Error(), "")
			}
			return m, tea.Batch(
				popup.Success("Record was added", ""),
				func() tea.Msg {
					return AddSuccessMsg{
						Row: row,
					}
				},
			)
		} else { // No button
			return m, tea.Batch(
				popup.Info("Addition canceled", ""),
				func() tea.Msg {
					return AddCancelMsg{}
				},
			)
		}
	}
	var cmd tea.Cmd
	m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
	return m, cmd
}

func processAddVerticalMoves(m AddModel, button string) AddModel {
	switch button {
	case "up", "shift+tab":
		if !m.focusButtons {
			if m.focusIndex > 0 {
				m.inputFields[m.focusIndex].Blur()
				m.focusIndex--
				m.inputFields[m.focusIndex].Focus()
			}
		} else {
			m.focusButtons = false
			m.focusIndex = len(m.inputFields) - 1
			m.inputFields[len(m.inputFields)-1].Focus()
		}

	case "down", "tab":
		if !m.focusButtons {
			if m.focusIndex < len(m.inputFields)-1 {
				m.inputFields[m.focusIndex].Blur()
				m.focusIndex++
				m.inputFields[m.focusIndex].Focus()
			} else {
				m.inputFields[m.focusIndex].Blur()
				m.focusButtons = true
				m.cursor = 0
			}
		} else {
			m.focusButtons = false
			m.focusIndex = 0
			m.inputFields[0].Focus()
		}
	}
	return m
}

// View renders the add model.
func (m AddModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Add new record"))
	s.WriteString("\n\n")

	for _, i := range m.inputFields {
		if i.Focused() {
			s.WriteString(style.FocusedInputStyle.Render(i.View()))
		} else {
			s.WriteString(style.BlurredInputStyle.Render(i.View()))
		}
		s.WriteRune('\n')
	}
	s.WriteString("\n\n")

	styledButtons := make([]string, len(m.buttons))
	for i, btn := range m.buttons {
		if m.cursor == i {
			if m.focusButtons {
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

	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

// AddSuccessMsg is a message that indicates a successful addition of a new record.
type AddSuccessMsg struct {
	Row bubbleTable.Row
}

// AddCancelMsg is a message that indicates the addition of a new record was canceled.
type AddCancelMsg struct{}
