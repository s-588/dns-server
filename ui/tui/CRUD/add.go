package crud

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
	"github.com/prionis/dns-server/ui/tui/popup"
	"github.com/prionis/dns-server/ui/tui/style"
)

type AddModel struct {
	width, height int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model

	// Add, Clear and Cancel buttons.
	buttons      []string
	cursor       int
	focusButtons bool

	db *sqlite.DB
}

func NewAddModel(db *sqlite.DB, w, h int) AddModel {
	inputs := make([]textinput.Model, 0, 5)

	domainInput := textinput.New()
	domainInput.Placeholder = "Domain name"
	domainInput.Width = w / 3
	domainInput.Focus()
	inputs = append(inputs, domainInput)

	dataInput := textinput.New()
	dataInput.Placeholder = "Data"
	dataInput.Width = w / 3
	inputs = append(inputs, dataInput)

	typeInput := textinput.New()
	typeInput.Placeholder = "Type"
	typeInput.Width = w / 3
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
	inputs = append(inputs, ttlInput)

	return AddModel{
		width:       w,
		height:      h,
		inputFields: inputs,
		buttons:     []string{"Yes", "No"},
		db:          db,
	}
}

func (dm AddModel) Init() tea.Cmd {
	return nil
}

func (d AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if !d.focusButtons {
				if d.focusIndex > 0 {
					d.inputFields[d.focusIndex].Blur()
					d.focusIndex--
					d.inputFields[d.focusIndex].Focus()
				}
			} else {
				d.focusButtons = false
				d.focusIndex = len(d.inputFields) - 1
				d.inputFields[len(d.inputFields)-1].Focus()
			}

		case "down", "tab":
			if !d.focusButtons {
				if d.focusIndex < len(d.inputFields)-1 {
					d.inputFields[d.focusIndex].Blur()
					d.focusIndex++
					d.inputFields[d.focusIndex].Focus()
				} else {
					d.inputFields[d.focusIndex].Blur()
					d.focusButtons = true
					d.cursor = 0
				}
			} else {
				d.focusButtons = false
				d.focusIndex = 0
				d.inputFields[0].Focus()
			}
		case "left", "h":
			if d.focusButtons && d.cursor > 0 {
				d.cursor--
			}
		case "right", "l":
			if d.focusButtons && d.cursor < len(d.buttons)-1 {
				d.cursor++
			}
		case "ctrl+c":
			return d, tea.Batch(
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
			if d.focusButtons {
				if d.cursor == 0 { // Yes button
					domain := d.inputFields[0].Value()
					dataStr := d.inputFields[1].Value()
					rrType := d.inputFields[2].Value()
					class := d.inputFields[3].Value()
					ttlStr := d.inputFields[4].Value()

					if ttlStr == "" {
						ttlStr = "0"
					}
					if domain == "" || dataStr == "" || rrType == "" || class == "" {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "All fields are required",
								Duration: 4 * time.Second,
							}
						}
					}

					data := net.ParseIP(dataStr)
					if data == nil {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid IP address",
								Duration: 4 * time.Second,
							}
						}
					}

					ttl, err := strconv.ParseInt(ttlStr, 10, 64)
					if err != nil {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid TTL",
								Duration: 4 * time.Second,
							}
						}
					}

					if _, ok := protocol.Types[rrType]; !ok {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid record type",
								Duration: 4 * time.Second,
							}
						}
					}

					if _, ok := protocol.Classes[class]; !ok {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      "Invalid class",
								Duration: 4 * time.Second,
							}
						}
					}

					err = d.add(domain, data, rrType, class, ttl)
					if err != nil {
						return d, func() tea.Msg {
							return popup.PopupMsg{
								Level:    "ERROR",
								Msg:      fmt.Sprintf("Failed to add record: %v", err),
								Duration: 4 * time.Second,
							}
						}
					}

					return d, tea.Batch(
						func() tea.Msg {
							return popup.PopupMsg{
								Level:    "SUCCESS",
								Msg:      fmt.Sprintf("Record %s added", domain),
								Duration: 4 * time.Second,
							}
						},
						func() tea.Msg {
							return AddSuccessMsg{}
						},
					)
				} else { // No button
					return d, tea.Batch(
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
				d.inputFields[d.focusIndex], cmd = d.inputFields[d.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		case "esc", "q":
			if d.focusButtons {
				d.focusButtons = false
				d.inputFields[d.focusIndex].Focus()
			} else {
				return d, tea.Batch(
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
			if !d.focusButtons {
				var cmd tea.Cmd
				d.inputFields[d.focusIndex], cmd = d.inputFields[d.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return d, tea.Batch(cmds...)
}

func (m AddModel) add(domain string, data net.IP, t, class string, ttl int64) error {
	err := m.db.AddRR(t, class, domain, data, ttl)
	if err != nil {
		return fmt.Errorf("can't add new record: %w", err)
	}
	return nil
}

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

	// Center content vertically
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type AddSuccessMsg struct {
	Id int64
}

type AddCancelMsg struct{}
