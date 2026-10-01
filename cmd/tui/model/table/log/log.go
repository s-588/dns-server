// Package logTable provides a model for displaying and interacting with log data in a table format within a TUI application.
package logTable

import (
	"fmt"
	"sort"
	"strings"
	"time"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/transport"
	"github.com/prionis/dns-server/cmd/tui/util"
)

// NewDescriptor creates a new table descriptor for the log table.
func NewDescriptor(width int) table.Descriptor {
	return table.Descriptor{
		Columns: GetColumns(width),

		FilterFn:     FilterFn,
		FilterFields: GetFilteredFields(width),

		SortFn:   SortFn,
		SearchFn: SearchFn,

		RefreshFn: RefreshFn,
	}
}

// RefreshFn retrieves all logs from the transport and returns them as table rows.
func RefreshFn(t *transport.Transport) ([]bubbleTable.Row, error) {
	logs, err := t.GetAllLogs()
	if err != nil {
		return []bubbleTable.Row{}, err
	}
	rows := make([]bubbleTable.Row, len(logs))

	for i, log := range logs {
		rows[i] = []string{log.Time.Format(time.DateTime), log.Level, log.Msg}
	}

	return rows, nil
}

// SearchFn searches for a query string in the log rows and returns the matching rows.
func SearchFn(query string, rows []bubbleTable.Row) []bubbleTable.Row {
	result := make([]bubbleTable.Row, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.Join(row, ""), query) {
			result = append(result, row)
		}
	}
	return result
}

// GetFilteredFields returns the input fields for filtering the log table based on the provided width.
func GetFilteredFields(width int) []textinput.Model {
	inputs := make([]textinput.Model, 3)
	for i := range len(inputs) {
		t := textinput.New()
		t.CharLimit = 32
		t.Prompt = "> "
		t.Width = width / 3

		switch i {
		case 0:
			t.Focus()
			t.Placeholder = "Enter date(e.g., 2025-05-21 or 12:47:58)"
			t.Validate = util.ValidateTimeFunc
		case 1:
			t.Placeholder = "Enter date(e.g., 2025-05-21 or 12:47:58)"
			t.Validate = util.ValidateTimeFunc
		case 2:
			t.Placeholder = "Enter level(e.g., ERROR)"
			t.Validate = func(s string) error {
				if s != "" {
					levels := "ERROR,INFO,WARNING,DEBUG"
					if !strings.Contains(s, levels) {
						return fmt.Errorf("unknown level: %s. Possible levels: %s", s, levels)
					}
				}
				return nil
			}
		}
		inputs[i] = t
	}
	return inputs
}

// GetColumns returns the column definitions for the log table based on the provided width.
func GetColumns(width int) []bubbleTable.Column {
	return []bubbleTable.Column{
		{
			Title: "Time",
			Width: max(8, width/5-5),
		},
		{
			Title: "Level",
			Width: max(5, width/5-5),
		},
		{
			Title: "Message",
			Width: max(8, (width/5)*3-5),
		},
	}
}

// FilterFn filters the log rows based on the provided input fields (start date, end date, and level).
func FilterFn(inputs []textinput.Model, rows []bubbleTable.Row) ([]bubbleTable.Row, error) {
	result := make([]bubbleTable.Row, 0)
	if inputs[0].Err != nil {
		return nil, inputs[0].Err
	}
	if inputs[1].Err != nil {
		return nil, inputs[1].Err
	}

	startTime, endTime, err := parseTimeFilter(inputs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		logDate, err := time.Parse(time.DateTime, row[0])
		if err != nil {
			return []bubbleTable.Row{}, fmt.Errorf("can't parse row time: %w", err)
		}
		if !startTime.IsZero() {
			if logDate.Before(startTime) {
				continue
			}
		}
		if !endTime.IsZero() {
			if logDate.After(endTime) {
				continue
			}
		}

		if inputs[2].Value() != "" {
			if inputs[2].Value() != row[1] {
				continue
			}
		}
		result = append(result, row)
	}
	return result, nil
}

func parseTimeFilter(inputs []textinput.Model) (time.Time, time.Time, error) {
	startTime, err := util.ParseTime(inputs[0].Value())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endTime, err := util.ParseTime(inputs[1].Value())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if !endTime.IsZero() && !startTime.IsZero() {
		if endTime.Before(startTime) {
			return time.Time{}, time.Time{}, fmt.Errorf("start date is after end date: %s > %s",
				startTime.Format(time.DateTime),
				endTime.Format(time.DateTime))
		}
	}
	return startTime, endTime, nil
}

// SortFn sorts the log rows based on the specified column index and order.
func SortFn(index int, r []bubbleTable.Row, asc bool) []bubbleTable.Row {
	switch index {
	case 0: // Time
		sort.Slice(r, func(i, j int) bool {
			t1, err1 := time.Parse(time.DateTime, r[i][0])
			t2, err2 := time.Parse(time.DateTime, r[j][0])
			if err1 != nil || err2 != nil {
				return false
			}
			if asc {
				return t2.Before(t1)
			}
			return t1.Before(t2)
		})
	case 1, 2: // Level, Message
		sort.Slice(r, func(i, j int) bool {
			if asc {
				return strings.ToLower(r[i][index]) < strings.ToLower(r[j][index])
			}
			return strings.ToLower(r[i][index]) > strings.ToLower(r[j][index])
		})
	}
	return r
}

// LogButtonsHandler handles button actions for the log table.
func LogButtonsHandler(index int, m table.Model) (table.Model, tea.Cmd) {
	switch index {
	case 0: // View
		return m, func() tea.Msg {
			return table.FocusMsg{}
		}
	case 1: // Filter
		return m, func() tea.Msg {
			return table.FilterRequestMsg{}
		}
	case 2: // Sort
		return m, func() tea.Msg {
			return table.SortRequestMsg{}
		}
	case 3: // Reset
		m.Table.SetRows(m.UnchangedRows)
		m.Table.UpdateViewport()
		return m, nil

	case 4: // Refresh
		return m, func() tea.Msg {
			return table.RefreshRequestMsg{}
		}

	case 5: // Export to Word
		return m, func() tea.Msg {
			return table.ExportToWordRequestMsg{}
		}
	case 6: // Export to Excel
		return m, func() tea.Msg {
			return table.ExportToExcelRequestMsg{}
		}
	}
	return m, nil
}

// New creates a new table for server logs.
func New(w, h int, t *transport.Transport) (table.Model, error) {
	buttons := []string{
		fmt.Sprintf("View %c ", '\uebb7'),
		fmt.Sprintf("Filter %c ", '\ueaf1'),
		fmt.Sprintf("Sort %c ", '\ueaf1'),
		fmt.Sprintf("Reset %c ", '\ueaf1'),
		fmt.Sprintf("Refresh %c ", '\ueaf1'),
		fmt.Sprintf("Export to Word %c ", '\ue6a5'),
		fmt.Sprintf("Export to Excel %c ", '\uf1c3'),
	}
	return table.NewModel(NewDescriptor(w), w, h, buttons, ParseLogs(t), LogButtonsHandler, "Logs"), nil
}

// ParseLogs retrieves logs from the transport and converts them into table rows.
func ParseLogs(t *transport.Transport) []bubbleTable.Row {
	rows := make([]bubbleTable.Row, 0)
	logs, err := t.GetAllLogs()
	if err != nil {
		return rows
	}
	for _, log := range logs {
		rows = append(rows, bubbleTable.Row{
			log.Time.Format(time.DateTime),
			log.Level,
			log.Msg,
		})
	}
	return rows
}

// WaitforLogs listens for incoming log messages from the transport and returns a tea.Cmd.
func WaitforLogs(msgs <-chan transport.LogMsg) tea.Cmd {
	return func() tea.Msg {
		msg := <-msgs
		return LogMsg{[]string{msg.Time.Format(time.DateTime), msg.Level, msg.Msg}}
	}
}

// LogMsg represents a log message with a row of data for the table.
type LogMsg struct {
	Row bubbleTable.Row
}
