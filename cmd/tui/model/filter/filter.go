package filter

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
)

type FilterModel struct {
	width, height int

	// Focused input field
	focusedInput int
	// Domain, data, type, class and ttl inputs
	inputs []textinput.Model

	// Add, Clear and Cancel buttons.
	buttons      []string
	cursor       int
	focusButtons bool

	table *table.TableModel
}

func NewFilterModel(table *table.TableModel, w int) FilterModel {
	m := FilterModel{
		width: w,

		inputs: table.Descriptor.FilterFields,

		buttons: []string{"Filter", "Cancel"},
		table:   table,
	}
	return m
}

func (f FilterModel) Init() tea.Cmd {
	return nil
}

func (m FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if !m.focusButtons {
				if m.focusedInput > 0 {
					m.inputs[m.focusedInput].Blur()
					m.focusedInput--
					m.inputs[m.focusedInput].Focus()
				}
			} else {
				m.focusButtons = false
				m.focusedInput = len(m.inputs) - 1
				m.inputs[len(m.inputs)-1].Focus()
			}

		case "down", "tab":
			if !m.focusButtons {
				if m.focusedInput < len(m.inputs)-1 {
					m.inputs[m.focusedInput].Blur()
					m.focusedInput++
					m.inputs[m.focusedInput].Focus()
				} else {
					m.inputs[m.focusedInput].Blur()
					m.focusButtons = true
					m.cursor = 0
				}
			} else {
				m.focusButtons = false
				m.focusedInput = 0
				m.inputs[0].Focus()
			}
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
					return popup.PopupMsg{
						Level:    "INFO",
						Msg:      "Addition canceled",
						Duration: 4 * time.Second,
					}
				},
				func() tea.Msg {
					return FilterCancelMsg{}
				},
			)
		case "enter", " ":
			if m.focusButtons {
				if m.cursor == 0 { // Yes button
					rows, err := m.table.Descriptor.FilterFn(m.inputs, m.table.UnchangedRows)
					if err != nil {
						return m, popup.Error(err.Error(), "")
					}
					return m, func() tea.Msg {
						return FilterMsg{
							Rows: rows,
						}
					}
				} else { // No button
					return m, func() tea.Msg {
						return FilterCancelMsg{}
					}
				}
			} else {
				var cmd tea.Cmd
				m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
				cmds = append(cmds, cmd)
			}
		case "esc", "q":
			if m.focusButtons {
				m.focusButtons = false
				m.inputs[m.focusedInput].Focus()
			} else {
				return m, func() tea.Msg {
					return FilterCancelMsg{}
				}
			}
		default:
			if !m.focusButtons {
				var cmd tea.Cmd
				m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (f FilterModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Filter record"))
	s.WriteString("\n\n")

	for _, i := range f.inputs {
		if i.Focused() {
			s.WriteString(style.FocusedInputStyle.Render(i.View()))
		} else {
			s.WriteString(style.BlurredInputStyle.Render(i.View()))
		}
		s.WriteRune('\n')
	}
	s.WriteString("\n\n")

	styledButtons := make([]string, len(f.buttons))
	for i, btn := range f.buttons {
		if f.cursor == i {
			if f.focusButtons {
				styledButtons[i] = style.SelectedButtonStyle.Render(btn)
			} else {
				styledButtons[i] = style.SecondarySelectedButtonStyle.Render(btn)
			}
		} else {
			styledButtons[i] = style.ButtonStyle.Render(btn)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(f.width - 4).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, styledButtons...)))

	// Center content vertically
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type FilterMsg struct {
	Rows []bubbleTable.Row
}

type FilterCancelMsg struct{}
