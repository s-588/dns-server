package crud

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/sqlite"
	"github.com/prionis/dns-server/ui/tui/popup"
	"github.com/prionis/dns-server/ui/tui/style"
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

	logTable *table.Model
	rrTable  *table.Model
}

func NewFilterModel(db *sqlite.DB, rrTable, logTable *table.Model, w, h int) FilterModel {
	m := FilterModel{
		width:  w,
		height: h,

		inputs: make([]textinput.Model, 0, 3),

		buttons: []string{"Filter", "Cancel"},

		rrTable:  rrTable,
		logTable: logTable,
	}

	if logTable != nil {
		for i := range 3 {
			t := textinput.New()
			t.CharLimit = 32
			t.Prompt = "> "
			t.Width = w / 3

			switch i {
			case 0:
				t.Focus()
				t.Placeholder = "Enter date(e.g., 2025-05-21 or 12:47:58)"
			case 1:
				t.Placeholder = "Enter date(e.g., 2025-05-21 or 12:47:58)"
			case 2:
				t.Placeholder = "Enter level(e.g., ERROR)"
			}
			m.inputs = append(m.inputs, t)
		}
	} else {
		for i := range 2 {
			t := textinput.New()
			t.CharLimit = 32
			t.Prompt = "> "
			t.Width = w / 3

			switch i {
			case 0:
				t.Focus()
				t.Placeholder = "Enter type(e.g. A)"
			case 1:
				t.Placeholder = "Enter class(e.g. IN)"
			}
			m.inputs = append(m.inputs, t)
		}
	}
	return m
}

func (f FilterModel) Init() tea.Cmd {
	return nil
}

func (f FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if !f.focusButtons {
				if f.focusedInput > 0 {
					f.inputs[f.focusedInput].Blur()
					f.focusedInput--
					f.inputs[f.focusedInput].Focus()
				}
			} else {
				f.focusButtons = false
				f.focusedInput = len(f.inputs) - 1
				f.inputs[len(f.inputs)-1].Focus()
			}

		case "down", "tab":
			if !f.focusButtons {
				if f.focusedInput < len(f.inputs)-1 {
					f.inputs[f.focusedInput].Blur()
					f.focusedInput++
					f.inputs[f.focusedInput].Focus()
				} else {
					f.inputs[f.focusedInput].Blur()
					f.focusButtons = true
					f.cursor = 0
				}
			} else {
				f.focusButtons = false
				f.focusedInput = 0
				f.inputs[0].Focus()
			}
		case "left", "h":
			if f.focusButtons && f.cursor > 0 {
				f.cursor--
			}
		case "right", "l":
			if f.focusButtons && f.cursor < len(f.buttons)-1 {
				f.cursor++
			}
		case "ctrl+c":
			return f, tea.Batch(
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
			if f.focusButtons {
				if f.cursor == 0 { // Yes button
					rows := make([]table.Row, 0)
					if f.logTable != nil {
						var startDate, endDate time.Time
						var err error
						if f.inputs[0].Value() != "" {
							startDate, err = time.Parse(time.DateTime, f.inputs[0].Value())
							if err != nil {
								return f, func() tea.Msg {
									return popup.PopupMsg{
										Level:    "ERROR",
										Msg:      fmt.Sprintf("incorrect format of the date: %s", f.inputs[0].Value()),
										Duration: 4 * time.Second,
									}
								}
							}
						}
						if f.inputs[1].Value() != "" {
							endDate, err = time.Parse(time.DateTime, f.inputs[1].Value())
							if err != nil {
								return f, func() tea.Msg {
									return popup.PopupMsg{
										Level:    "ERROR",
										Msg:      fmt.Sprintf("incorrect format of the date: %s", f.inputs[1].Value()),
										Duration: 4 * time.Second,
									}
								}
							}
						}
						for _, r := range f.logTable.Rows() {
							logDate, err := time.Parse(time.DateTime, r[0])
							if err != nil {
								return f, func() tea.Msg {
									return popup.PopupMsg{
										Level:    "ERROR",
										Msg:      fmt.Sprintf("wrong date format IN LOG O_O : %s", f.inputs[1].Value()),
										Duration: 4 * time.Second,
									}
								}
							}
							if !startDate.IsZero() {
								if !logDate.After(startDate) {
									continue
								}
							}
							if !endDate.IsZero() {
								if !logDate.Before(startDate) {
									continue
								}
							}

							if f.inputs[2].Value() != "" {
								if f.inputs[2].Value() != r[1] {
									continue
								}
							}
							rows = append(rows, r)
						}
						return f, tea.Batch(
							func() tea.Msg {
								return popup.PopupMsg{
									Level:    "SUCCESS",
									Msg:      fmt.Sprintf("%d records was filtered", len(f.logTable.Rows())-len(rows)),
									Duration: 4 * time.Second,
								}
							},
							func() tea.Msg {
								return FilterMsg{
									Rows: rows,
								}
							},
						)

					}
					// TODO: filter records table
					return f, tea.Batch(
						func() tea.Msg {
							return popup.PopupMsg{
								Level:    "SUCCESS",
								Msg:      fmt.Sprintf("%d records was filtered", len(f.logTable.Rows())-len(rows)),
								Duration: 4 * time.Second,
							}
						},
						func() tea.Msg {
							return FilterMsg{
								Rows: rows,
							}
						},
					)
				} else { // No button
					return f, tea.Batch(
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
				f.inputs[f.focusedInput], cmd = f.inputs[f.focusedInput].Update(msg)
				cmds = append(cmds, cmd)
			}
		case "esc", "q":
			if f.focusButtons {
				f.focusButtons = false
				f.inputs[f.focusedInput].Focus()
			} else {
				return f, tea.Batch(
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
			if !f.focusButtons {
				var cmd tea.Cmd
				f.inputs[f.focusedInput], cmd = f.inputs[f.focusedInput].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return f, tea.Batch(cmds...)
}

func (f FilterModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Add new record"))
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
	Rows []table.Row
}

type FilterCancelMsg struct{}
