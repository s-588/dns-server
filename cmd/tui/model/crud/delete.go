package crud

import (
	"fmt"
	"strconv"
	"strings"

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
			if d.cursor == 0 { // Yes
				row := d.table.Table.SelectedRow()
				if len(row) == 0 {
					return d, popup.Error("no row selected", "")
				}
				id, err := strconv.ParseInt(row[0], 10, 32)
				if err != nil {
					return d, popup.Error(fmt.Sprintf("invalid record id: %v", err), "")
				}
				if d.table.Descriptor.DeleteFn == nil {
					return d, popup.Error("delete is not supported for this table", "")
				}
				return d, func() tea.Msg { return DeleteMsg{id: id} }
			}
			return d, tea.Batch(
				popup.Info("Deletion canceled", ""),
				func() tea.Msg { return DeleteCancelMsg{} },
			)
		case "esc", "ctrl+c": // was "Esc" — never matched
			return d, tea.Batch(
				popup.Info("Deletion canceled", ""),
				func() tea.Msg { return DeleteCancelMsg{} },
			)
		}

	case DeleteMsg:
		if err := d.delete(msg.id); err != nil {
			return d, popup.Error(transport.UserMessage(err), "")
		}
		return d, tea.Batch(
			popup.Success("Record successfully deleted", ""),
			func() tea.Msg { return DeleteSuccessMsg{ID: msg.id} },
		)
	}
	return d, nil
}

// delete removes a record via the table's descriptor, so the request is
// routed to the correct endpoint (RRs vs users).
func (d DeleteModel) delete(id int64) error {
	return d.table.Descriptor.DeleteFn(d.transport, id)
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
