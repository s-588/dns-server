package models

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/popup"
	"github.com/prionis/dns-server/cmd/tui/style"
)

type SortModel struct {
	width, height int
	buttons       []string
	cursor        int
	ascending     bool
	IsLogTable    bool
	logTable      *table.Model
	rrTable       *table.Model
}

func NewSortModel(rrTable, logTable *table.Model, w, h int, isLogTable bool) SortModel {
	m := SortModel{
		width:      w,
		height:     h,
		rrTable:    rrTable,
		logTable:   logTable,
		IsLogTable: isLogTable,
	}
	if isLogTable {
		m.buttons = []string{"Time", "Level", "Message"}
	} else {
		m.buttons = []string{"ID", "Domain", "Data", "Type", "Class", "TTL"}
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
			s.ascending = !s.ascending
		case "left", "h":
			if s.cursor > 0 {
				s.cursor--
			}
		case "right", "l":
			if s.cursor < len(s.buttons)-1 {
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
			rows, err := s.sortRows()
			if err != nil {
				return s, func() tea.Msg {
					return popup.PopupMsg{
						Level:    "ERROR",
						Msg:      fmt.Sprintf("Failed to sort: %v", err),
						Duration: 4 * time.Second,
					}
				}
			}
			return s, tea.Batch(
				func() tea.Msg {
					return popup.PopupMsg{
						Level:    "SUCCESS",
						Msg:      fmt.Sprintf("Sorted by %s %s", s.buttons[s.cursor], s.getDirection()),
						Duration: 4 * time.Second,
					}
				},
				func() tea.Msg {
					return SortMsg{
						Rows:       rows,
						IsLogTable: s.IsLogTable,
					}
				},
			)
		}
	}
	return s, tea.Batch(cmds...)
}

func (s SortModel) getTable() *table.Model {
	if s.IsLogTable {
		return s.logTable
	}
	return s.rrTable
}

func (s SortModel) getDirection() string {
	if s.ascending {
		return "ascending"
	}
	return "descending"
}

func (s SortModel) sortRows() ([]table.Row, error) {
	rows := s.getTable().Rows()
	if len(rows) == 0 {
		return rows, nil
	}

	// Create a copy to avoid modifying the original
	sortedRows := make([]table.Row, len(rows))
	copy(sortedRows, rows)

	column := s.cursor
	ascending := s.ascending

	if s.IsLogTable {
		switch column {
		case 0: // Time
			sort.Slice(sortedRows, func(i, j int) bool {
				t1, err1 := time.Parse(time.DateTime, sortedRows[i][0])
				t2, err2 := time.Parse(time.DateTime, sortedRows[j][0])
				if err1 != nil || err2 != nil {
					return false
				}
				if ascending {
					return t1.Before(t2)
				}
				return t2.Before(t1)
			})
		case 1, 2: // Level, Message
			sort.Slice(sortedRows, func(i, j int) bool {
				if ascending {
					return strings.ToLower(sortedRows[i][column]) < strings.ToLower(sortedRows[j][column])
				}
				return strings.ToLower(sortedRows[i][column]) > strings.ToLower(sortedRows[j][column])
			})
		}
	} else {
		switch column {
		case 0: // ID
			sort.Slice(sortedRows, func(i, j int) bool {
				id1, err1 := strconv.ParseInt(sortedRows[i][0], 10, 64)
				id2, err2 := strconv.ParseInt(sortedRows[j][0], 10, 64)
				if err1 != nil || err2 != nil {
					return false
				}
				if ascending {
					return id1 < id2
				}
				return id1 > id2
			})
		case 5: // TTL
			sort.Slice(sortedRows, func(i, j int) bool {
				ttl1, err1 := strconv.ParseInt(sortedRows[i][5], 10, 64)
				ttl2, err2 := strconv.ParseInt(sortedRows[j][5], 10, 64)
				if err1 != nil || err2 != nil {
					return false
				}
				if ascending {
					return ttl1 < ttl2
				}
				return ttl1 > ttl2
			})
		case 2: // Data (IP)
			sort.Slice(sortedRows, func(i, j int) bool {
				ip1 := net.ParseIP(sortedRows[i][2])
				ip2 := net.ParseIP(sortedRows[j][2])
				if ip1 == nil || ip2 == nil {
					return false
				}
				if ascending {
					return bytesCompare(ip1, ip2) < 0
				}
				return bytesCompare(ip1, ip2) > 0
			})
		case 1, 3, 4: // Domain, Type, Class
			sort.Slice(sortedRows, func(i, j int) bool {
				if ascending {
					return strings.ToLower(sortedRows[i][column]) < strings.ToLower(sortedRows[j][column])
				}
				return strings.ToLower(sortedRows[i][column]) > strings.ToLower(sortedRows[j][column])
			})
		}
	}
	return sortedRows, nil
}

// bytesCompare compares two IP addresses byte-by-byte for sorting
func bytesCompare(a, b net.IP) int {
	a = a.To16()
	b = b.To16()
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return int(a[i]) - int(b[i])
		}
	}
	return 0
}

func (sm SortModel) View() string {
	s := strings.Builder{}
	header := "Sort Logs"
	if !sm.IsLogTable {
		header = "Sort Records"
	}
	s.WriteString(style.HeaderStyle.Render(header))
	s.WriteString("\n\n")

	buttons := make([]string, len(sm.buttons))
	for i, btn := range sm.buttons {
		if sm.cursor == i {
			if sm.ascending {
				buttons[i] = style.SelectedButtonStyle.Render("Ascending " + btn + fmt.Sprintf(" %c", '\uea9a'))
			} else {
				buttons[i] = style.SelectedButtonStyle.Render("Descending " + btn + fmt.Sprintf(" %c", '\ueaa1'))
			}
		} else {
			buttons[i] = style.ButtonStyle.Render(btn)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(sm.width - 4).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, buttons...)))

	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

type SortMsg struct {
	Rows       []table.Row
	IsLogTable bool
}

type SortCancelMsg struct{}
