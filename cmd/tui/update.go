package tui

import (
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	bubbleTable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/export"
	"github.com/prionis/dns-server/cmd/tui/model/account"
	"github.com/prionis/dns-server/cmd/tui/model/crud"
	exportModel "github.com/prionis/dns-server/cmd/tui/model/export"
	"github.com/prionis/dns-server/cmd/tui/model/filter"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/sort"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	logTable "github.com/prionis/dns-server/cmd/tui/model/table/log"
	rrTable "github.com/prionis/dns-server/cmd/tui/model/table/rr"
	userTable "github.com/prionis/dns-server/cmd/tui/model/table/user"
)

// Update handle all messages that comming from user interaction and other models.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case table.ExportToExcelRequestMsg:
		headers := make([]string, len(m.tables[m.selectedTab].Descriptor.Columns))
		for i, col := range m.tables[m.selectedTab].Descriptor.Columns {
			headers[i] = col.Title
		}
		tables := make([]export.TableData, len(m.tables))
		for i, table := range m.tables {
			tables[i] = export.TableData{
				Name: table.Header,
				Rows: table.Table.Rows(),
			}
		}
		file, err := export.ExportToExcel(tables)
		if err != nil {
			return m, popup.Error("can't create excel report: "+err.Error(), "")
		}
		return m, popup.Success("Excel report created, saved with name "+file, "10s")

	case table.ExportToWordRequestMsg:
		m.focusLayer = focusExportModel

	case exportModel.ExportMsg:
		name, err := export.NewWordFile(m.tables[0].UnchangedRows,
			m.tables[1].UnchangedRows, msg.StartTime, msg.EndTime)
		if err != nil {
			return m, popup.Error("can't create Word report: "+err.Error(), "")
		}
		m.focusLayer = focusButtons
		return m, popup.Success("your Word report saved: "+name, "10s")

	case exportModel.ExportCancelMsg:
		m.focusLayer = focusButtons

	case logTable.LogMsg:
		newRows := append(m.tables[0].Table.Rows(), msg.Row)
		sortedRows := m.tables[0].Descriptor.SortFn(m.tables[0].Descriptor.SortedColumn,
			newRows, m.tables[0].Descriptor.SortAscending)
		m.tables[0].Table.SetRows(sortedRows)
		m.tables[0].Table.UpdateViewport()
		cmd = logTable.WaitforLogs(m.logChan)

	case table.AddRequestMsg:
		m.focusLayer = focusAddModel
		m.addModel = crud.NewAddModel(&m.tables[m.selectedTab], m.transport, m.width)

	case table.DeleteRequestMsg:
		m.focusLayer = focusDeleteModel
		m.deletePage = crud.NewDeleteModel(&m.tables[m.selectedTab], m.transport, m.width)

	case table.FilterRequestMsg:
		m.focusLayer = focusFilterModel
		m.filterPage = filter.NewFilterModel(&m.tables[m.selectedTab], m.width)

	case table.SortRequestMsg:
		m.focusLayer = focusSortModel
		m.sortPage = sort.NewSortModel(&m.tables[m.selectedTab], m.width)

	case table.ResetRequestMsg:
		m.tables[m.selectedTab].Table.SetRows(m.tables[m.selectedTab].UnchangedRows)
		m.tables[m.selectedTab].Table.UpdateViewport()

	case table.RefreshRequestMsg:
		rows, err := m.tables[m.selectedTab].Descriptor.RefreshFn(m.transport)
		if err != nil {
			return m, popup.Error("something went wrong: "+err.Error(), "")
		}
		m.tables[m.selectedTab].Table.SetRows(rows)
		m.tables[m.selectedTab].Table.UpdateViewport()

	case table.UpdateRequestMsg:
		m.focusLayer = focusUpdateModel
		id, err := strconv.ParseInt(m.tables[m.selectedTab].Table.SelectedRow()[0], 10, 32)
		if err != nil {
			return m, popup.Error("can't parse id for updating", "")
		}
		m.updatePage = crud.NewUpdateModel(m.transport,
			&m.tables[m.selectedTab],
			int32(id), m.width)

	case table.FocusMsg:
		m.tables[m.selectedTab].Table.Focus()
		m.focusLayer = focusTable

	case table.UnfocusMsg:
		m.tables[m.selectedTab].Table.Blur()
		if m.focusLayer == focusTable {
			m.focusLayer = focusButtons
		} else {
			m.focusLayer = focusTabs
		}

	case account.LoginSuccessMsg, account.LoginCancelMsg:
		m, cmd = m.HandleLoginMsgs(msg)

	case crud.UpdateCancelMsg, crud.UpdateSuccessMsg:
		m = m.handleUpdateMsgs(msg)

	case sort.SortMsg, sort.SortCancelMsg:
		m = m.handleSortMsgs(msg)

	case filter.FilterMsg, filter.FilterCancelMsg:
		m = m.handleFilterMsgs(msg)

	case crud.AddSuccessMsg, crud.AddCancelMsg:
		m = m.handleAddMsgs(msg)

	case tea.WindowSizeMsg:
		m, cmd = m.handleScreenResize(msg)

	case popup.PopupMsg, popup.ClearPopupMsg:
		m, cmd = m.handlePopupMsgs(msg)

	case crud.DeleteMsg, crud.DeleteCancelMsg, crud.DeleteSuccessMsg:
		m, cmd = m.handleDeleteMsgs(msg)

	case tea.KeyMsg:
		m, cmd = m.handleKeys(msg)

	}
	return m, cmd
}

func (m model) handleUpdateMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case crud.UpdateSuccessMsg:
		m.tables[m.selectedTab] = m.tables[m.selectedTab].UpdateRow(msg.Index, msg.Row)
		m.focusLayer = focusButtons
		return m
	case crud.UpdateCancelMsg:
		m.focusLayer = focusButtons
		return m
	}
	return m
}

func (m model) HandleLoginMsgs(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case account.LoginSuccessMsg:
		tables := make([]table.TableModel, 0)
		logt, err := logTable.New(m.width, m.height, m.transport)
		if err != nil {
			return m, popup.Error("can't create table for logs: "+err.Error(), "")
		}
		tables = append(tables, logt)

		rrt, err := rrTable.New(m.transport, m.width, m.height)
		if err != nil {
			return m, popup.Error("can't create table for resource records: "+err.Error(), "")
		}
		tables = append(tables, rrt)

		if msg.User.Role == "admin" {
			usert, err := userTable.New(m.transport, m.width, m.height)
			if err != nil {
				return m, popup.Error("can't create table for users: "+err.Error(), "")
			}
			tables = append(tables, usert)
		}

		err = m.transport.EstablishWebsocketConnection(m.transport.HTTPClient.Jar, m.transport.Addr)
		if err != nil {
			return m, popup.Error("can't connect to the logs websocket: "+err.Error(), "")
		}
		go m.transport.ListenWebSocket(m.logChan)
		cmd = logTable.WaitforLogs(m.logChan)
		tabs := []string{"Logs", "Resource records"}
		if msg.User.Role == "admin" {
			tabs = []string{"Logs", "Resource records", "Users"}
		}
		m.tabs = tabs
		m.tables = tables
		m.focusLayer = focusTabs
		m.user = msg.User
	case account.LoginCancelMsg:
		cmd = m.Close()
	}
	return m, cmd
}

// Handle sort messages of table.
func (m model) handleSortMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case sort.SortMsg:
		m.tables[m.selectedTab].Table.SetRows(msg.Rows)
		m.tables[m.selectedTab].Table.UpdateViewport()
		m.focusLayer = focusTable

	case sort.SortCancelMsg:
		m.focusLayer = focusButtons

	}
	return m
}

// Handle filter messages of table.
func (m model) handleFilterMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case filter.FilterMsg:
		m.tables[m.selectedTab].Table.SetRows(msg.Rows)
		m.tables[m.selectedTab].Table.UpdateViewport()
		m.focusLayer = focusTable

	case filter.FilterCancelMsg:
		m.focusLayer = focusTable
	}
	return m
}

// Handle add messages for resource record table.
func (m model) handleAddMsgs(message tea.Msg) model {
	switch msg := message.(type) {
	case crud.AddSuccessMsg:
		m.tables[m.selectedTab].Table.SetRows(
			slices.Insert(m.tables[m.selectedTab].Table.Rows(), 0, msg.Row))
		m.tables[m.selectedTab].Table.UpdateViewport()
		if m.tables[m.selectedTab].Table.Focused() {
			m.focusLayer = focusTable
		} else {
			m.focusLayer = focusButtons
		}
	case crud.AddCancelMsg:
		if m.tables[m.selectedTab].Table.Focused() {
			m.focusLayer = focusTable
		} else {
			m.focusLayer = focusButtons
		}
	}
	return m
}

// Handle screen resizing messages.
func (m model) handleScreenResize(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	w, h := msg.Width, msg.Height
	m.width = w
	m.height = h
	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	// Update table dimensions
	for i := range m.tables {
		m.tables[i] = m.tables[i].UpdateSize(w, h)
	}
	return m, cmd
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
		model, cmd := m.deletePage.Update(msg)
		m.deletePage = model.(crud.DeleteModel)
		return m, cmd

	case crud.DeleteCancelMsg:
		m.focusLayer = focusTable

	case crud.DeleteSuccessMsg:
		m.focusLayer = focusTable
		rows := m.tables[m.selectedTab].Table.Rows()
		newRows := make([]bubbleTable.Row, 0, len(rows))
		for _, row := range rows {
			if row[0] != strconv.FormatInt(int64(msg.Id), 10) {
				newRows = append(newRows, row)
			}
		}
		m.tables[m.selectedTab].Table.SetRows(newRows)
		m.tables[m.selectedTab].Table.UpdateViewport()
		if m.tables[m.selectedTab].Table.Cursor() >= len(newRows) && len(newRows) > 0 {
			m.tables[m.selectedTab].Table.SetCursor(len(newRows) - 1)
		}

	}
	return m, nil
}

// Handle key messages form user input.
func (m model) handleKeys(message tea.Msg) (model, tea.Cmd) {
	// First of all we check what focus layer is.
	// If user on the different page, not the main model, we just let this model handle msg.
	// If user on the main page, e.g. button or tabs is focused, we handle it by ourself.
	// This aproach of handling message was choosen to reduce boilerplate code in handling
	// each key and then check what the focus layer is
	// (for real we just move this code to selected model)
	var cmd tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch m.focusLayer {
		case focusSearch:
			switch {
			case key.Matches(msg, m.keys.Down):
				if !m.searchInput.Focused() {
					m.focusLayer = focusButtons
				} else {
					m.searchInput, cmd = m.searchInput.Update(msg)
				}
			case key.Matches(msg, m.keys.Unfocus):
				if m.searchInput.Focused() {
					m.searchInput.Blur()
				} else {
					m.focusLayer = focusTabs
				}
			case key.Matches(msg, m.keys.Up):
				if !m.searchInput.Focused() {
					m.focusLayer = focusTabs
				} else {
					m.searchInput, cmd = m.searchInput.Update(msg)
				}
			case key.Matches(msg, m.keys.Enter):
				if !m.searchInput.Focused() {
					m.searchInput.Focus()
				} else {
					m.focusLayer = focusTable
				}
			default:
				if m.searchInput.Focused() {
					model, cmd := m.searchInput.Update(msg)
					m.searchInput = model

					rows := m.tables[m.selectedTab].Descriptor.SearchFn(
						strings.ToLower(m.searchInput.Value()),
						m.tables[m.selectedTab].UnchangedRows,
					)
					m.tables[m.selectedTab].Table.SetRows(rows)
					m.tables[m.selectedTab].Table.UpdateViewport()
					return m, cmd
				}
			}

		case focusExportModel:
			model, cmd := m.exportModel.Update(msg)
			m.exportModel = model.(exportModel.ExportModel)
			return m, cmd

		case focusTable:
			model, cmd := m.tables[m.selectedTab].Update(msg)
			m.tables[m.selectedTab] = model.(table.TableModel)
			return m, cmd

		case focusLoginModel:
			model, cmd := m.loginPage.Update(msg)
			m.loginPage = model.(account.LoginModel)
			return m, cmd

		case focusUpdateModel:
			model, cmd := m.updatePage.Update(msg)
			m.updatePage = model.(crud.UpdateModel)
			return m, cmd

		case focusSortModel:
			model, cmd := m.sortPage.Update(msg)
			m.sortPage = model.(sort.SortModel)
			return m, cmd

		case focusFilterModel:
			model, cmd := m.filterPage.Update(msg)
			m.filterPage = model.(filter.FilterModel)
			return m, cmd

		case focusAddModel:
			model, cmd := m.addModel.Update(msg)
			m.addModel = model.(crud.AddModel)
			return m, cmd

		case focusDeleteModel:
			model, cmd := m.deletePage.Update(msg)
			m.deletePage = model.(crud.DeleteModel)
			m.focusLayer = focusButtons
			return m, cmd

		default:
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, m.Close()

			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll

			case key.Matches(msg, m.keys.Unfocus):
				switch m.focusLayer {
				case focusTabs:
					return m, m.Close()

				case focusSearch:
					m.focusLayer = focusTabs

				case focusButtons:
					m.focusLayer = focusSearch

				}

			case key.Matches(msg, m.keys.Up):
				switch m.focusLayer {
				case focusSearch:
					m.focusLayer = focusTabs

				case focusButtons:
					m.focusLayer = focusSearch

				case focusTable:
					m.tables[m.selectedTab].Table, cmd = m.tables[m.selectedTab].Table.Update(msg)
					return m, cmd

				}

			case key.Matches(msg, m.keys.Down):
				switch m.focusLayer {
				case focusTabs:
					m.focusLayer = focusSearch

				case focusSearch:
					m.focusLayer = focusButtons

				case focusButtons:
					m.focusLayer = focusTable
					m.tables[m.selectedTab].Table.Focus()

				case focusTable:
					if m.tables[m.selectedTab].Table.Focused() || m.tables[m.selectedTab].Table.Focused() {
						m.tables[m.selectedTab].Table, cmd = m.tables[m.selectedTab].Table.Update(msg)
						return m, cmd
					}

				}

			case key.Matches(msg, m.keys.Right):
				switch m.focusLayer {
				case focusTabs:
					if m.selectedTab < len(m.tabs)-1 {
						m.selectedTab++
					}
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}

			case key.Matches(msg, m.keys.Left):
				switch m.focusLayer {
				case focusTabs:
					if m.selectedTab > 0 {
						m.selectedTab--
					}
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}

			case key.Matches(msg, m.keys.Enter):
				switch m.focusLayer {
				case focusTabs:
					m.focusLayer = focusButtons
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}

				// Delete key
			case key.Matches(msg, m.keys.Delete):
				switch m.focusLayer {
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}

				// Add key
			case key.Matches(msg, m.keys.Add):
				switch m.focusLayer {
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}

				// Search key
			case key.Matches(msg, m.keys.Search):
				switch m.focusLayer {
				case focusButtons, focusTable:
					m.focusLayer = focusSearch
				}

				// Filter key
			case key.Matches(msg, m.keys.Filter):
				switch m.focusLayer {
				case focusButtons, focusTable:
					model, cmd := m.tables[m.selectedTab].Update(msg)
					m.tables[m.selectedTab] = model.(table.TableModel)
					return m, cmd
				}
			}
		}
	}
	return m, nil
}
