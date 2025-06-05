package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/popup"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// Represent model for deletion the resource record from the database.
type DeleteModel struct {
	// Width and height of the screen
	width, height int
	// Row containing the record user want to delete.
	Record table.Row

	// "Yes" and "No" buttons to accept or reject delete.
	buttons []string
	// Selected button.
	cursor int

	// Pointer to the database connection
	transport *transport.Transport
}

// Create new delete model.
func NewDeleteModel(record table.Row, t *transport.Transport, w, h int) DeleteModel {
	return DeleteModel{
		Record:    record,
		buttons:   []string{"Yes", "No"},
		transport: t,
		width:     w,
		height:    h,
	}
}

func (dm DeleteModel) Init() tea.Cmd {
	return nil
}

// Process the user interactions.
func (d DeleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if d.cursor > 0 {
				d.cursor--
			}
		case "right", "l":
			if d.cursor < len(d.buttons)-1 {
				d.cursor++
			}
		case "enter", " ":
			// Button "Yes" is pressed
			if d.cursor == 0 {
				id, err := parseRecordID(d.Record)
				if err != nil {
					return d, func() tea.Msg {
						return popup.PopupMsg{
							Level:    "ERROR",
							Msg:      fmt.Sprintf("Invalid record ID: %v", err),
							Duration: 4 * time.Second,
						}
					}
				}
				return d, func() tea.Msg { return DeleteMsg{id: id} }
			}

			return d, tea.Batch(func() tea.Msg {
				return popup.PopupMsg{
					Level: "INFO", Msg: "Deletion canceled",
					Duration: 4 * time.Second,
				}
			}, func() tea.Msg {
				return DeleteCancelMsg{}
			})

		case "Esc":
			return d, tea.Batch(func() tea.Msg {
				return popup.PopupMsg{Level: "INFO", Msg: "Deletion canceled", Duration: 4 * time.Second}
			}, func() tea.Msg {
				return DeleteCancelMsg{}
			})
		}
	case DeleteMsg:
		err := d.delete(msg.id)
		if err != nil {
			return d, func() tea.Msg {
				return popup.PopupMsg{
					Level: "ERROR", Msg: "Can't delete record",
					Duration: 4 * time.Second,
				}
			}
		}
		return d, tea.Batch(func() tea.Msg {
			return popup.PopupMsg{
				Level:    "SUCCESS",
				Msg:      "Record succesfully deleted",
				Duration: 4 * time.Second,
			}
		},
			func() tea.Msg {
				return DeleteSuccessMsg{
					Id: msg.id,
				}
			})
	}
	return d, nil
}

// Delete record from database.
func (d DeleteModel) delete(id int64) error {
	err := d.transport.DeleteRR(id)
	if err != nil {
		return fmt.Errorf("can't delete from database: %w", err)
	}
	return nil
}

func parseRecordID(r table.Row) (int64, error) {
	id := r[0]
	return strconv.ParseInt(id, 10, 64)
}

func (d DeleteModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Confirm Deletion"))
	s.WriteString("\n\n")
	recordDetails := fmt.Sprintf(
		"ID: %s\nDomain: %s\nData: %s\nType: %s\nClass: %s\nTTL: %s",
		d.Record[0], d.Record[1], d.Record[2], d.Record[3], d.Record[4], d.Record[5],
	)
	s.WriteString(style.BaseStyle.Render(recordDetails))
	s.WriteString("\n\n")
	styledButtons := make([]string, len(d.buttons))
	for i, btn := range d.buttons {
		if d.cursor == i {
			styledButtons[i] = style.SelectedButtonStyle.Render(btn)
		} else {
			styledButtons[i] = style.ButtonStyle.Render(btn)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(d.width - 8).Height(d.height - 36).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, styledButtons...)))
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type DeleteMsg struct {
	id      int64
	confirm bool
}

type DeleteSuccessMsg struct {
	Id int64
}

type DeleteCancelMsg struct{}
