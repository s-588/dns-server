package crud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// DeleteModel represent model for deletion the resource record from the database.
type DeleteModel struct {
	// Width and height of the screen
	width, height int
	// Row containing the record user want to delete.
	table *table.Model

	// "Yes" and "No" buttons to accept or reject delete.
	buttons []string
	// Selected button.
	cursor int

	// Pointer to the database connection
	transport *transport.Transport
}

// NewDeleteModel creates a new delete model.
func NewDeleteModel(table *table.Model, t *transport.Transport, w int) DeleteModel {
	return DeleteModel{
		table:     table,
		buttons:   []string{"Yes", "No"},
		transport: t,
		width:     w,
	}
}

// Init initializes the delete model.
func (dm DeleteModel) Init() tea.Cmd {
	return nil
}

// Update process the user interactions.
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
				id, err := strconv.ParseInt(d.table.Table.SelectedRow()[0], 10, 32)
				if err != nil {
					return d, func() tea.Msg {
						return popup.Msg{
							Level:    "ERROR",
							Msg:      fmt.Sprintf("Invalid record ID: %v", err),
							Duration: 4 * time.Second,
						}
					}
				}
				return d, func() tea.Msg { return DeleteMsg{id: id} }
			}

			return d, tea.Batch(func() tea.Msg {
				return popup.Msg{
					Level: "INFO", Msg: "Deletion canceled",
					Duration: 4 * time.Second,
				}
			}, func() tea.Msg {
				return DeleteCancelMsg{}
			})

		case "Esc":
			return d, tea.Batch(func() tea.Msg {
				return popup.Msg{Level: "INFO", Msg: "Deletion canceled", Duration: 4 * time.Second}
			}, func() tea.Msg {
				return DeleteCancelMsg{}
			})
		}
	case DeleteMsg:
		err := d.delete(msg.id)
		if err != nil {
			return d, func() tea.Msg {
				return popup.Msg{
					Level: "ERROR", Msg: "Can't delete record",
					Duration: 4 * time.Second,
				}
			}
		}
		return d, tea.Batch(func() tea.Msg {
			return popup.Msg{
				Level:    "SUCCESS",
				Msg:      "Record successfully deleted",
				Duration: 4 * time.Second,
			}
		},
			func() tea.Msg {
				return DeleteSuccessMsg{
					ID: msg.id,
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

// View renders the delete model.
func (d DeleteModel) View() string {
	s := strings.Builder{}
	s.WriteString(style.HeaderStyle.Render("Confirm Deletion"))
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

// DeleteMsg is a message that indicates a request to delete a record with a specific ID.
type DeleteMsg struct {
	id int64
}

// DeleteSuccessMsg is a message that indicates a successful deletion of a record with a specific ID.
type DeleteSuccessMsg struct {
	ID int64
}

// DeleteCancelMsg is a message that indicates the deletion operation was canceled by the user.
type DeleteCancelMsg struct{}
