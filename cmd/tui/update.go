package tui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	crud "github.com/prionis/dns-server/cmd/tui/CRUD"
	"github.com/prionis/dns-server/cmd/tui/auth"
	"github.com/prionis/dns-server/cmd/tui/popup"
)

// Update handle all messages that comming from user interaction and other models.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case auth.LoginSuccessMsg, auth.LoginCancelMsg:
		m, cmd = m.HandleLoginMsgs(msg)
		cmds = append(cmds, cmd)

	case crud.SortMsg, crud.SortCancelMsg:
		m = m.handleSortMsgs(msg)

	case crud.FilterMsg, crud.FilterCancelMsg:
		m = m.handleFilterMsgs(msg)

	case crud.AddSuccessMsg, crud.AddCancelMsg:
		m = m.handleAddMsgs(msg)

	case tea.WindowSizeMsg:
		m = m.handleScreenResize(msg)

	case popup.PopupMsg, popup.ClearPopupMsg:
		m, cmd = m.handlePopupMsgs(msg)
		cmds = append(cmds, cmd)

	case crud.DeleteMsg, crud.DeleteCancelMsg, crud.DeleteSuccessMsg:
		m, cmd = m.handleDeleteMsgs(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		m, cmd = m.handleKeys(msg)
		cmds = append(cmds, cmd)

	}
	return m, tea.Batch(cmds...)
}

func (m model) HandleLoginMsgs(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case auth.LoginSuccessMsg:
		m.focusLayer = focusTabs
		m.user = msg.User
	case auth.LoginCancelMsg:
		cmd = m.Close()
	}
	return m, cmd
}

// Handle sort messages of table.
func (m model) handleSortMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case crud.SortMsg:
		if m.selectedTab == 0 {
			m.logTable.SetRows(msg.Rows)
			m.logTable.UpdateViewport()
			m.focusLayer = focusTable
			m.logTable.Focus()
		} else {
			m.rrTable.SetRows(msg.Rows)
			m.rrTable.UpdateViewport()
			m.focusLayer = focusTable
			m.rrTable.Focus()
		}

	case crud.SortCancelMsg:
		m.focusLayer = focusButtons

	}
	return m
}

// Handle filter messages of table.
func (m model) handleFilterMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case crud.FilterMsg:
		if m.selectedTab == 0 {
			m.logTable.SetRows(msg.Rows)
			m.logTable.UpdateViewport()
			m.focusLayer = focusTable
			m.logTable.Focus()
		} else {
			m.rrTable.SetRows(msg.Rows)
			m.rrTable.UpdateViewport()
			m.focusLayer = focusTable
			m.rrTable.Focus()
		}
	case crud.FilterCancelMsg:
		m.focusLayer = focusTable
	}
	return m
}

// Handle add messages for resource record table.
func (m model) handleAddMsgs(msg tea.Msg) model {
	switch msg.(type) {
	case crud.AddSuccessMsg:
		if m.rrTable.Focused() {
			m.focusLayer = focusTable
		} else {
			m.focusLayer = focusButtons
		}
	case crud.AddCancelMsg:
		if m.rrTable.Focused() {
			m.focusLayer = focusTable
		} else {
			m.focusLayer = focusButtons
		}
	}
	return m
}

// Handle screen resizing messages.
func (m model) handleScreenResize(msg tea.WindowSizeMsg) model {
	m.width = msg.Width
	m.height = msg.Height
	// Update table dimensions
	m.rrTable = rrTable(m.transport, msg.Width, msg.Height)
	m.logTable = logTable(msg.Width, msg.Height)
	// Ensure the cursor stays within bounds
	if m.rrTable.Cursor() >= len(m.rrTable.Rows()) && len(m.rrTable.Rows()) > 0 {
		m.rrTable.SetCursor(len(m.rrTable.Rows()) - 1)
	}
	if m.logTable.Cursor() >= len(m.logTable.Rows()) && len(m.logTable.Rows()) > 0 {
		m.logTable.SetCursor(len(m.logTable.Rows()) - 1)
	}
	return m
}

// Handle popup messages.
func (m model) handlePopupMsgs(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case popup.PopupMsg:
		updatedModel, cmd := m.popup.Update(msg)
		m.popup = updatedModel.(popup.PopupModel)
		return m, cmd

	case popup.ClearPopupMsg:
		updatedModel, cmd := m.popup.Update(msg)
		m.popup = updatedModel.(popup.PopupModel)
		return m, cmd
	}
	return m, nil
}

// Handle delete messages for resource record table.
func (m model) handleDeleteMsgs(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case crud.DeleteMsg:
		model, cmd := m.deleteModel.Update(msg)
		m.deleteModel = model.(crud.DeleteModel)
		return m, cmd

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

	}
	return m, nil
}

// Handle key messages form user input.
func (m model) handleKeys(message tea.Msg) (model, tea.Cmd) {
	// First of all we check what focus layer is.
	// If user on the different page, not the main model we just let this model handle msg.
	// If user on the main page, e.g. button, table, tabs is focused, we handle it by ourself.
	// This aproach of handling message was choosen to reduce boilerplate code in handling
	// each key and then check what the focus layer is.
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch m.focusLayer {

		case focusLoginPage:
			model, cmd := m.loginPage.Update(msg)
			m.loginPage = model.(auth.LoginModel)
			return m, cmd

		case focusUpdatePage:
			model, cmd := m.updatePage.Update(msg)
			m.updatePage = model.(crud.UpdateModel)
			return m, cmd

		case focusSortPage:
			model, cmd := m.sortPage.Update(msg)
			m.sortPage = model.(crud.SortModel)
			return m, cmd

		case focusFilterPage:
			model, cmd := m.filterPage.Update(msg)
			m.filterPage = model.(crud.FilterModel)
			return m, cmd

		case focusAddPage:
			model, cmd := m.addModel.Update(msg)
			m.addModel = model.(crud.AddModel)
			return m, cmd

		case focusDeletePage:
			model, cmd := m.deleteModel.Update(msg)
			m.deleteModel = model.(crud.DeleteModel)
			m.focusLayer = focusButtons
			return m, cmd

		default:
			var cmd tea.Cmd
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				switch m.focusLayer {
				case focusTable:
					if m.getTable().Focused() {
						m.focusLayer--
						m.getTable().Blur()
					}
				default:
					m.Close()
					return m, nil
				}

			case "up", "k":
				switch m.focusLayer {
				case focusTable:
					m, cmd = m.updateTable(msg)
					return m, cmd
				case focusButtons, focusTabs:
					if m.focusLayer > 0 {
						m.focusLayer--
					}
				}

			case "down", "j":
				switch m.focusLayer {
				case focusTable:
					if m.rrTable.Focused() || m.logTable.Focused() {
						m, cmd = m.updateTable(msg)
						return m, cmd
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

			case "enter", " ":
				switch m.focusLayer {
				case focusTabs:
					m.focusLayer = focusButtons
				case focusButtons:
					switch m.selectedTab {
					case 0: // logs page
						switch m.tabs[0].cursor {
						case 0: // view button
							m.logTable.Focus()
						case 1: // filter button
							m.filterPage = crud.NewFilterModel(nil, &m.logTable, m.width, m.height)
							m.focusLayer = focusFilterPage
						case 2: // sort button
							m.sortPage = crud.NewSortModel(nil, &m.logTable, m.width, m.height, true)
							m.focusLayer = focusSortPage
						}
					case 1: // records page
						switch m.tabs[1].cursor {
						case 0: // view button
							m.rrTable.Focus()
						case 1: // add button
							m.focusLayer = focusAddPage
						case 2: // delete button
							m.focusLayer = focusDeletePage
							m.deleteModel.Record = m.rrTable.SelectedRow()
						case 3: // update button
							m.updatePage = crud.NewUpdateModel(m.transport, m.rrTable.SelectedRow(), m.width, m.height)
							m.focusLayer = focusUpdatePage
						case 4: // filter button
							m.filterPage = crud.NewFilterModel(&m.rrTable, nil, m.width, m.height)
							m.focusLayer = focusFilterPage
						case 5: // sort button
							m.sortPage = crud.NewSortModel(&m.rrTable, nil, m.width, m.height, false)
							m.focusLayer = focusSortPage
						}
					}
				case focusTable:
				}

				// Delete key
			case "d":
				if m.selectedTab == 1 {
					m.focusLayer = focusDeletePage
					m.deleteModel.Record = m.rrTable.SelectedRow()
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
		}
	}
	return m, nil
}
