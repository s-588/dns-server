package export

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
	"github.com/prionis/dns-server/cmd/tui/util"
)

type ExportModel struct {
	// Width and height of the screen.
	width, height int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model

	// Export and cancel buttons
	buttons []string
	// Selected button.
	cursor int
	// Are buttons focused or the inputs.
	focusButtons bool
}

// Create new add model.
func NewExportModel(t *transport.Transport, w, h int) ExportModel {
	inputs := make([]textinput.Model, 0, 2)

	startDate := textinput.New()
	startDate.Placeholder = "Start date of the report ( for example 2025-07-29 or 12:47:48 )"
	startDate.Width = w / 3
	startDate.Focus()
	startDate.Validate = util.ValidateTimeFunc
	inputs = append(inputs, startDate)

	endDate := textinput.New()
	endDate.Placeholder = "End date of the report ( for example 2025-07-29 or 12:47:48 )"
	endDate.Validate = util.ValidateTimeFunc
	endDate.Width = w / 3
	inputs = append(inputs, endDate)

	return ExportModel{
		width:       w,
		height:      h,
		inputFields: inputs,
		buttons:     []string{"Export", "Cancel"},
	}
}

func (dm ExportModel) Init() tea.Cmd {
	return nil
}

// Handle messages that comming from the user interactions.
func (m ExportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
		case "left", "h":
			if m.focusButtons && m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.focusButtons && m.cursor < len(m.buttons)-1 {
				m.cursor++
			}
		case "ctrl+c", "esc":
			return m, func() tea.Msg {
				return ExportCancelMsg{}
			}
		case "enter", " ":
			if m.focusButtons {
				if m.cursor == 0 { // Yes button
					if m.inputFields[0].Err != nil {
						return m, popup.Error(m.inputFields[0].Err.Error(), "")
					}
					if m.inputFields[1].Err != nil {
						return m, popup.Error(m.inputFields[1].Err.Error(), "")
					}

					// Skip errors because ValidateFunc from inputs
					// already use ParseTime error checking.
					startTime, _ := util.ParseTime(m.inputFields[0].Value())
					endTime, _ := util.ParseTime(m.inputFields[1].Value())

					return m,
						func() tea.Msg {
							return ExportMsg{
								StartTime: startTime,
								EndTime:   endTime,
							}
						}

				} else { // No button
					return m, func() tea.Msg {
						return ExportCancelMsg{}
					}
				}
			} else {
				var cmd tea.Cmd
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
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

// Render page on the screen.
func (m ExportModel) View() string {
	s := strings.Builder{}
	s.WriteString("\n")

	for _, i := range m.inputFields {
		if i.Focused() {
			s.WriteString(style.FocusedInputStyle.Render(i.View()))
		} else {
			s.WriteString(style.BlurredInputStyle.Render(i.View()))
		}
		s.WriteRune('\n')
	}
	s.WriteRune('\n')

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

type ExportMsg struct {
	StartTime, EndTime time.Time
}

type ExportCancelMsg struct{}
