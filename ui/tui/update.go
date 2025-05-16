package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	crud "github.com/prionis/dns-server/ui/tui/CRUD"
	"github.com/prionis/dns-server/ui/tui/popup"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case popup.PopupMsg:
		updatedModel, cmd := m.popup.Update(msg)
		m.popup = updatedModel.(popup.PopupModel)
		cmds = append(cmds, cmd)

	case popup.ClearPopupMsg:
		updatedModel, cmd := m.popup.Update(msg)
		m.popup = updatedModel.(popup.PopupModel)
		cmds = append(cmds, cmd)

	case crud.DeleteMsg:
		model, cmd := m.deletePage.Update(msg)
		m.deletePage = model.(crud.DeleteModel)
		cmds = append(cmds, cmd)

	case crud.DeleteCancelMsg:
		m.focusLayer = focusTable

	case crud.DeleteSuccessMsg:
		m.focusLayer = focusTable
		rows := m.rrTable.Rows()
		newRows := make([]table.Row, 0, len(rows))
		for _, row := range rows {
			if row[0] != strconv.FormatInt(msg.Id, 10) {
				newRows = append(newRows, row)
			}
		}
		m.rrTable.SetRows(newRows)
		m.rrTable.UpdateViewport()
		if m.rrTable.Cursor() >= len(newRows) && len(newRows) > 0 {
			m.rrTable.SetCursor(len(newRows) - 1)
		}

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
		cmds = append(cmds, waitForLogMsg(m.logMsgChan))

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			switch m.focusLayer {
			case focusDeletePage:
				model, cmd := m.deletePage.Update(msg)
				m.deletePage = model.(crud.DeleteModel)
				cmds = append(cmds, cmd)
				m.focusLayer = focusButtons
				return m, tea.Batch(cmds...)
			case focusTable:
				if m.getTable().Focused() {
					m.focusLayer--
					m.getTable().Blur()
					return m, nil
				}
			}
			m.Close()
			cmds = append(cmds, tea.Quit)

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			switch m.focusLayer {
			case focusTable:
				m, cmd = m.updateTable(msg)
				cmds = append(cmds, cmd)
			case focusButtons, focusTabs:
				if m.focusLayer > 0 {
					m.focusLayer--
				}
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			switch m.focusLayer {
			case focusTable:
				if m.rrTable.Focused() || m.logTable.Focused() {
					m, cmd = m.updateTable(msg)
					cmds = append(cmds, cmd)
				}
			case focusTabs, focusButtons:
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
			}

		case "left", "l":
			switch m.focusLayer {
			case focusDeletePage:
				model, cmd := m.deletePage.Update(msg)
				m.deletePage = model.(crud.DeleteModel)
				cmds = append(cmds, cmd)
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
					cmds = append(cmds, cmd)
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					cmds = append(cmds, cmd)
				}
			}

		case "right", "h":
			switch m.focusLayer {
			case focusDeletePage:
				model, cmd := m.deletePage.Update(msg)
				m.deletePage = model.(crud.DeleteModel)
				cmds = append(cmds, cmd)
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
					cmds = append(cmds, cmd)
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					cmds = append(cmds, cmd)
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			switch m.focusLayer {
			case focusDeletePage:
				model, cmd := m.deletePage.Update(msg)
				m.deletePage = model.(crud.DeleteModel)
				cmds = append(cmds, cmd)
			case focusTabs:
				m.focusLayer = focusButtons
			case focusButtons:
				switch m.selectedTab {
				case 0: // logs page
					switch m.tabs[0].cursor {
					case 0: // view button
						m.logTable.Focus()
					case 1: // filter button
					case 2: // sort button
					}
				case 1: // records page
					switch m.tabs[1].cursor {
					case 0: // view button
						m.rrTable.Focus()
					case 1: // add button
						m.focusLayer = focusAddPage
					case 2: // delete button
						m.focusLayer = focusDeletePage
						m.deletePage.Record = m.rrTable.SelectedRow()
					case 3: // filter button
					case 4: // sort button
					}
				}
			case focusTable:
			}

			// Delete key
		case "d":
			if m.selectedTab == 1 {
				m.focusLayer = focusDeletePage
				m.deletePage.Record = m.rrTable.SelectedRow()
			}

			// Add key
		case "a":
			if m.selectedTab == 1 {
			}

			// Search key
		case "s":

			// Filter key
		case "f":

		}
	default:
		if m.selectedTab == 0 {
			m.logTable, cmd = m.logTable.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.selectedTab == 1 {
			m.rrTable, cmd = m.rrTable.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}
