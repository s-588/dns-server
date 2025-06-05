package models

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/miekg/dns"
	"github.com/prionis/dns-server/cmd/tui/popup"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
	"github.com/prionis/dns-server/internal/database"
)

type UpdateModel struct {
	width, height int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model

	// Add, Clear and Cancel buttons.
	buttons      []string
	cursor       int
	focusButtons bool

	transport *transport.Transport
}

func NewUpdateModel(t *transport.Transport, row table.Row, w, h int) UpdateModel {
	inputs := make([]textinput.Model, 0, 6)

	idInput := textinput.New()
	idInput.Width = w / 3
	idInput.SetValue(row[0])
	inputs = append(inputs, idInput)

	domainInput := textinput.New()
	domainInput.Placeholder = "Domain name"
	domainInput.Width = w / 3
	domainInput.SetValue(row[1])
	domainInput.Focus()
	inputs = append(inputs, domainInput)

	dataInput := textinput.New()
	dataInput.Placeholder = "Data"
	dataInput.Width = w / 3
	dataInput.SetValue(row[2])
	inputs = append(inputs, dataInput)

	typeInput := textinput.New()
	typeInput.Placeholder = "Type"
	typeInput.Width = w / 3
	typeInput.SetValue(row[3])
	typeInput.SetSuggestions([]string{
		"A",
		"NS",
		"MD",
		"MF",
		"CNAME",
		"SOA",
		"MB",
		"MG",
		"MR",
		"NULL",
		"WKS",
		"PTR",
		"HINFO",
		"MINFO",
		"MX",
		"TXT",
	})
	inputs = append(inputs, typeInput)

	classInput := textinput.New()
	classInput.Placeholder = "Class"
	classInput.ShowSuggestions = true
	classInput.Width = w / 3
	classInput.SetValue(row[4])
	classInput.SetSuggestions([]string{
		"IN",
		"CS",
		"CH",
		"HS",
	})
	inputs = append(inputs, classInput)

	ttlInput := textinput.New()
	ttlInput.Placeholder = "Time to live"
	ttlInput.Width = w / 3
	ttlInput.SetValue(row[5])
	inputs = append(inputs, ttlInput)

	return UpdateModel{
		width:       w,
		height:      h,
		inputFields: inputs,
		buttons:     []string{"Yes", "No"},
		transport:   t,
	}
}

func (m UpdateModel) Init() tea.Cmd {
	return nil
}

func (m UpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					return AddCancelMsg{}
				},
			)
		case "enter", " ":
			if m.focusButtons {
				if m.cursor == 0 { // Yes button
					domain := m.inputFields[1].Value()
					dataStr := m.inputFields[2].Value()
					rrType := m.inputFields[3].Value()
					class := m.inputFields[4].Value()
					ttlStr := m.inputFields[5].Value()

					if ttlStr == "" {
						ttlStr = "0"
					}
					if domain == "" || dataStr == "" || rrType == "" || class == "" {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "All fields are required",
								Duration: 4 * time.Second,
							}
						}
					}

					data := net.ParseIP(dataStr)
					if data == nil {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid IP address",
								Duration: 4 * time.Second,
							}
						}
					}

					ttl, err := strconv.ParseInt(ttlStr, 10, 64)
					if err != nil {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid TTL",
								Duration: 4 * time.Second,
							}
						}
					}

					if _, ok := dns.StringToType[rrType]; !ok {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid record type",
								Duration: 4 * time.Second,
							}
						}
					}

					if _, ok := dns.StringToClass[class]; !ok {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid class",
								Duration: 4 * time.Second,
							}
						}
					}
					idS := m.inputFields[0].Value()
					id, err := strconv.ParseInt(idS, 10, 64)
					if err != nil {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      fmt.Sprintf("Failed to parse id: %v", err),
								Duration: 4 * time.Second,
							}
						}
					}
					err = m.update(id, domain, data.String(), rrType, class, ttl)
					if err != nil {
						return m, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      fmt.Sprintf("Failed to update record: %v", err),
								Duration: 4 * time.Second,
							}
						}
					}

					return m, tea.Batch(
						func() tea.Msg {
							return popup.PopupMsg{
								Level:    "SUCCESS",
								Msg:      fmt.Sprintf("Record %s updated", domain),
								Duration: 4 * time.Second,
							}
						},
						func() tea.Msg {
							return AddSuccessMsg{}
						},
					)
				} else { // No button
					return m, tea.Batch(
						func() tea.Msg {
							return popup.PopupMsg{
								Level:    "INFO",
								Msg:      "Addition canceled",
								Duration: 4 * time.Second,
							}
						},
						func() tea.Msg {
							return AddCancelMsg{}
						},
					)
				}
			} else {
				var cmd tea.Cmd
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		case "esc", "q":
			if m.focusButtons {
				m.focusButtons = false
				m.inputFields[m.focusIndex].Focus()
			} else {
				return m, tea.Batch(
					func() tea.Msg {
						return popup.PopupMsg{
							Level:    "INFO",
							Msg:      "Addition canceled",
							Duration: 4 * time.Second,
						}
					},
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

func (m UpdateModel) update(id int64, domain, data, t, class string, ttl int64) error {
	err := m.transport.UpdateRR(database.ResourceRecord{
		ID:     id,
		Domain: domain,
		Data:   data,
		Type:   t,
		Class:  class,
		TTL:    ttl,
	})
	if err != nil {
		return fmt.Errorf("can't add new record: %w", err)
	}
	return nil
}

func (m UpdateModel) View() string {
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

	// Center content vertically
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type UpdateSuccessMsg struct {
	Id int64
}

type UpdateCancelMsg struct{}
