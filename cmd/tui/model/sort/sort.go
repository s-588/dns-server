package sort

import (
	"fmt"
	"strings"
	"time"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/style"
)

type SortModel struct {
	width  int
	table  *table.TableModel
	cursor int
}

func NewSortModel(table *table.TableModel, w int) SortModel {
	m := SortModel{
		width: w,
		table: table,
	}
	return m
}

func (s SortModel) Init() tea.Cmd {
	return nil
}

func (s SortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			s.table.Descriptor.SortAscending = !s.table.Descriptor.SortAscending
		case "left", "h":
			if s.cursor > 0 {
				s.cursor--
			}
		case "right", "l":
			if s.cursor < len(s.table.Buttons)-1 {
				s.cursor++
			}
		case "ctrl+c", "esc", "q":
			return s, tea.Batch(
				func() tea.Msg {
					return popup.PopupMsg{
						Level:    "INFO",
						Msg:      "Sorting canceled",
						Duration: 4 * time.Second,
					}
				},
				func() tea.Msg {
					return SortCancelMsg{}
				},
			)
		case "enter", " ":
			s.table.Descriptor.SortedColumn = s.cursor
			rows := s.table.Descriptor.SortFn(s.table.Descriptor.SortedColumn, s.table.UnchangedRows,
				s.table.Descriptor.SortAscending)
			return s, func() tea.Msg {
				return SortMsg{
					Rows: rows,
				}
			}
		}
	}
	return s, tea.Batch(cmds...)
}

func (m SortModel) View() string {
	s := strings.Builder{}
	header := m.table.Header
	s.WriteString(style.HeaderStyle.Render(header))
	s.WriteString("\n\n")

	buttons := make([]string, len(m.table.Buttons))
	for i, column := range m.table.Table.Columns() {
		if m.cursor == i {
			if m.table.Descriptor.SortAscending {
				buttons[i] = style.SelectedButtonStyle.Render("Ascending " + column.Title + fmt.Sprintf(" %c", '\uea9a'))
			} else {
				buttons[i] = style.SelectedButtonStyle.Render("Descending " + column.Title + fmt.Sprintf(" %c", '\ueaa1'))
			}
		} else {
			buttons[i] = style.ButtonStyle.Render(column.Title)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, buttons...)))

	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type SortMsg struct {
	Rows []bubbleTable.Row
}

type SortCancelMsg struct{}
