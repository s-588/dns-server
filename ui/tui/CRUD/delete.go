package crud

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/database"
	"github.com/prionis/dns-server/sqlite"
	"github.com/prionis/dns-server/ui/tui/popup"
	"github.com/prionis/dns-server/ui/tui/style"
)

type DeleteModel struct {
	width, height int
	// Message with data of Record user wan't to delete.
	Record table.Row

	// "Yes" and "No" buttons to accept or reject delete.
	buttons []string
	cursor  int

	db *sqlite.DB
}

func NewDeleteModel(record table.Row, db *sqlite.DB, w, h int) DeleteModel {
	return DeleteModel{
		Record:  record,
		buttons: []string{"Yes", "No"},
		db:      db,
		width:   w,
		height:  h,
	}
}

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
			if d.cursor == 0 { // Button "Yes" is pressed
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
				return popup.PopupMsg{Level: "INFO", Msg: "Deletion canceled", Duration: 4 * time.Second}
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
		rr, err := d.delete(msg.id)
		if err != nil {
			return d, func() tea.Msg {
				return popup.PopupMsg{Level: "ERROR", Msg: "Can't delete record", Duration: 4 * time.Second}
			}
		}
		return d, tea.Batch(func() tea.Msg {
			return popup.PopupMsg{
				Level:    "SUCCESS",
				Msg:      fmt.Sprintf("Record %s %s was deleted", rr.RR.Domain, net.IP(rr.RR.Data).String()),
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

func (d DeleteModel) delete(id int64) (*database.DBRR, error) {
	rr, err := d.db.DelRR(id)
	if err != nil {
		return nil, fmt.Errorf("Can't delete from database: %w", err)
	}
	return rr, nil
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
