package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case logMsg:
		// Append new log line
		t, err := time.Parse(time.RFC3339, msg["time"].(string))
		if err != nil {
			t = time.Now()
		}
		level := strings.ToUpper(msg["level"].(string))
		message := msg["msg"].(string)

		// Update viewport content and scroll to bottom
		m.logTable.SetRows(append(m.logTable.Rows(), table.Row{t.Format(time.DateTime), level, message}))
		m.logTable.UpdateViewport()
		return m, waitForMsg(m.msg)

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			if m.focusLayer == focusTable {
				if m.getTable().Focused() {
					m.focusLayer--
					m.getTable().Blur()
					return m, nil
				}
			}
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.focusLayer != focusTable {
				if m.focusLayer > 0 {
					m.focusLayer--
				}
			} else {
				return m.updateTable(msg)
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.focusLayer != focusTable {
				if m.focusLayer < 2 {
					m.focusLayer++
				}
				if m.focusLayer == focusTable {
					if m.selectedTab == 0 {
						m.logTable.Focus()
					}
					if m.selectedTab == 1 {
						m.rrTable.Focus()
					}
				}
			} else {
				return m.updateTable(msg)
			}

		case "left", "l":
			switch m.focusLayer {
			case focusTabs:
				if m.selectedTab < len(m.tabs)-1 {
					m.selectedTab++
				}
			case focusButtons:
				if m.tabs[m.selectedTab].cursor < len(m.tabs[m.selectedTab].buttons)-1 {
					m.tabs[m.selectedTab].cursor++
				}
			case focusTable:
				if m.selectedTab == 0 {
					m.logTable, cmd = m.logTable.Update(msg)
					return m, cmd
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					return m, cmd
				}
			}

		case "right", "h":
			switch m.focusLayer {
			case focusTabs:
				if m.selectedTab > 0 {
					m.selectedTab--
				}
			case focusButtons:
				if m.tabs[m.selectedTab].cursor > 0 {
					m.tabs[m.selectedTab].cursor--
				}
			case focusTable:
				if m.selectedTab == 0 {
					m.logTable, cmd = m.logTable.Update(msg)
					return m, cmd
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					return m, cmd
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			switch m.focusLayer {
			case focusTabs:
			case focusButtons:
				if m.selectedTab == 0 && m.tabs[0].cursor == 0 {
					m.logTable.Focus()
				}
				if m.selectedTab == 1 && m.tabs[0].cursor == 0 {
					m.rrTable.Focus()
				}
			case focusTable:
			}

			// Delete key
		case "d":
			// row := m.getTable().SelectedRow()
			// id, err := strconv.ParseUint(row[0], 10, 64)
			// if err != nil {
			// }

			// Add key
		case "a":

			// Search key
		case "s":

			// Filter key
		case "f":

		default:
			fmt.Println("uknown key")
		}
	default:
		if m.selectedTab == 0 {
			m.logTable, cmd = m.logTable.Update(msg)
			return m, cmd
		}
		if m.selectedTab == 1 {
			m.rrTable, cmd = m.rrTable.Update(msg)
			return m, cmd
		}
	}

	select {
	case log := <-m.msg:
		m, cmd := m.Update(logMsg(log))
		return m, cmd
	default:
		return m, nil
	}
}
