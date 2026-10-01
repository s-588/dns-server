package crud

import (
	"strings"
	"time"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// UpdateModel represents the model for updating a resource record in the database.
type UpdateModel struct {
	width int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model
	table       table.Model

	// Add, Clear and Cancel buttons.
	buttons      []string
	cursor       int
	focusButtons bool

	id int64

	transport *transport.Transport
}

// NewUpdateModel creates a new update model.
func NewUpdateModel(t *transport.Transport, table *table.Model, id int64, width int) UpdateModel {
	data := table.Table.SelectedRow()[1:]
	inputs := make([]textinput.Model, len(table.Descriptor.InputFields))
	copy(inputs, table.Descriptor.InputFields)
	for i, cell := range data {
		inputs[i].Reset()
		inputs[i].SetValue(cell)
	}

	return UpdateModel{
		width:       width,
		table:       *table,
		inputFields: inputs,
		buttons:     []string{"Yes", "No"},
		transport:   t,
		id:          id,
	}
}

// Init initializes the update model.
func (m UpdateModel) Init() tea.Cmd {
	return nil
}

// Update processes user interactions.
func (m UpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "shift+tab", "down", "tab":
			m = processUpdateVerticalMoves(m, msg.String())
		case "left", "h":
			if m.focusButtons && m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.focusButtons && m.cursor < len(m.buttons)-1 {
				m.cursor++
			}
		case "ctrl+c":
			return m, tea.Batch(
				func() tea.Msg {
					return popup.Msg{
						Level:    "INFO",
						Msg:      "Addition canceled",
						Duration: 4 * time.Second,
					}
				},
				func() tea.Msg {
					return AddCancelMsg{}
				},
			)
		case "enter", " ":
			var cmd tea.Cmd
			m, cmd = processUpdateEnterPress(m, msg)
			cmds = append(cmds, cmd)
		case "esc", "q":
			if m.focusButtons {
				m.focusButtons = false
				m.inputFields[m.focusIndex].Focus()
			} else {
				return m, tea.Batch(
					popup.Info("Addition canceled", ""),
					func() tea.Msg {
						return UpdateCancelMsg{}
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

func processUpdateEnterPress(m UpdateModel, msg tea.KeyMsg) (UpdateModel, tea.Cmd) {
	if m.focusButtons {
		if m.cursor == 0 { // Yes button
			row, err := m.table.Descriptor.UpdateFn(m.transport, m.inputFields, m.id)
			if err != nil {
				return m, popup.Error(err.Error(), "")
			}

			return m, tea.Batch(func() tea.Msg {
				return UpdateSuccessMsg{
					int64(m.table.Table.Cursor()),
					row,
				}
			},
				popup.Success("Updated!", "2s"),
			)
		} else { // No button
			return m, tea.Batch(
				popup.Info("Addition canceled", ""),
				func() tea.Msg {
					return UpdateCancelMsg{}
				},
			)
		}
	}
	var cmd tea.Cmd
	m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
	return m, cmd
}

func processUpdateVerticalMoves(m UpdateModel, button string) UpdateModel {
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

// View renders the update model to the terminal.
func (m UpdateModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Update record"))
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

	// Center content vertically
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

// UpdateSuccessMsg is a message indicating that a record has been successfully updated.
type UpdateSuccessMsg struct {
	Index int64
	Row   bubbleTable.Row
}

// UpdateCancelMsg is a message indicating that the update operation has been canceled.
type UpdateCancelMsg struct{}
