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

func NewDescriptor(width int) table.TableDescriptor {
	return table.TableDescriptor{
		Columns: GetColumns(width),

		FilterFn:     FilterFn,
		FilterFields: GetFilteredFields(width),

		SortFn:   SortFn,
		SearchFn: SearchFn,

		RefreshFn: RefreshFn,
	}
}

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

func SearchFn(query string, rows []bubbleTable.Row) []bubbleTable.Row {
	result := make([]bubbleTable.Row, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.Join(row, ""), query) {
			result = append(result, row)
		}
	}
	return result
}

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
						return fmt.Errorf("Uknow level: %s. Possible levels: %s", s, levels)
					}
				}
				return nil
			}
		}
		inputs[i] = t
	}
	return inputs
}

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

func FilterFn(inputs []textinput.Model, rows []bubbleTable.Row) ([]bubbleTable.Row, error) {
	result := make([]bubbleTable.Row, 0)
	if inputs[0].Err != nil {
		return nil, inputs[0].Err
	}
	if inputs[1].Err != nil {
		return nil, inputs[1].Err
	}

	// Skip errors because inputs use ValidateTimeFunc which use ParseTime.
	startTime, _ := util.ParseTime(inputs[0].Value())
	endTime, _ := util.ParseTime(inputs[1].Value())

	if !endTime.IsZero() && !startTime.IsZero() {
		if endTime.Before(startTime) {
			return []bubbleTable.Row{}, fmt.Errorf("Start date is after end date: %s > %s",
				startTime.Format(time.DateTime),
				endTime.Format(time.DateTime))
		}
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

func LogButtonsHandler(index int, m table.TableModel) (table.TableModel, tea.Cmd) {
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

// Create new table for server logs.
func New(w, h int, t *transport.Transport) (table.TableModel, error) {
	buttons := make([]string, 0)
	buttons = []string{
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

func WaitforLogs(msgs <-chan transport.LogMsg) tea.Cmd {
	return func() tea.Msg {
		msg := <-msgs
		return LogMsg{[]string{msg.Time.Format(time.DateTime), msg.Level, msg.Msg}}
	}
}

type LogMsg struct {
	Row bubbleTable.Row
}
