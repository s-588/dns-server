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

// Update handle all messages that coming from user interaction and other models.
//
//nolint:gocyclo
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case table.ExportToExcelRequestMsg,
		table.ExportToWordRequestMsg,
		table.AddRequestMsg,
		table.DeleteRequestMsg,
		table.FilterRequestMsg,
		table.SortRequestMsg,
		table.ResetRequestMsg,
		table.RefreshRequestMsg,
		table.UpdateRequestMsg,
		table.FocusMsg,
		table.UnfocusMsg:
		return m.handleTableMsgs(msg)

	case exportModel.Msg, exportModel.CancelMsg:
		return m.handleExportMsgs(msg)

	case logTable.LogMsg:
		return processLogMsg(m, msg)

	case account.LoginSuccessMsg, account.LoginCancelMsg:
		m, cmd = m.HandleLoginMsgs(msg)

	case crud.UpdateCancelMsg, crud.UpdateSuccessMsg:
		m = m.handleUpdateMsgs(msg)

	case sort.Msg, sort.CancelMsg:
		m = m.handleSortMsgs(msg)

	case filter.Msg, filter.CancelMsg:
		m = m.handleFilterMsgs(msg)

	case crud.AddSuccessMsg, crud.AddCancelMsg:
		m = m.handleAddMsgs(msg)

	case tea.WindowSizeMsg:
		m, cmd = m.handleScreenResize(msg)

	case popup.Msg, popup.ClearMsg:
		m, cmd = m.handlePopupMsgs(msg)

	case crud.DeleteMsg, crud.DeleteCancelMsg, crud.DeleteSuccessMsg:
		m, cmd = m.handleDeleteMsgs(msg)

	case tea.KeyMsg:
		m, cmd = m.handleKeys(msg)
	}

	return m, cmd
}

func (m model) handleTableMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case table.ExportToExcelRequestMsg:
		return processExportToExcelRequest(m)

	case table.ExportToWordRequestMsg:
		m.focusLayer = focusExportModel

	case table.AddRequestMsg:
		m.focusLayer = focusAddModel
		m.addModel = crud.NewAddModel(&m.tables[m.selectedTab], m.transport, m.width)

	case table.DeleteRequestMsg:
		m.focusLayer = focusDeleteModel
		m.deletePage = crud.NewDeleteModel(&m.tables[m.selectedTab], m.transport, m.width)

	case table.FilterRequestMsg:
		m.focusLayer = focusFilterModel
		m.filterPage = filter.NewModel(&m.tables[m.selectedTab], m.width)

	case table.SortRequestMsg:
		m.focusLayer = focusSortModel
		m.sortPage = sort.NewModel(&m.tables[m.selectedTab], m.width)

	case table.ResetRequestMsg:
		m.tables[m.selectedTab].Table.SetRows(m.tables[m.selectedTab].UnchangedRows)
		m.tables[m.selectedTab].Table.UpdateViewport()

	case table.RefreshRequestMsg:
		return processTableRefreshReq(m)

	case table.UpdateRequestMsg:
		return processTableUpdateReq(m)

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
	}
	return m, nil
}

func (m model) handleExportMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case exportModel.Msg:
		return processExportMsg(m, msg)

	case exportModel.CancelMsg:
		m.focusLayer = focusButtons
	}
	return m, nil
}

func processTableUpdateReq(m model) (tea.Model, tea.Cmd) {
	m.focusLayer = focusUpdateModel
	id, err := strconv.ParseInt(m.tables[m.selectedTab].Table.SelectedRow()[0], 10, 32)
	if err != nil {
		return m, popup.Error("can't parse id for updating", "")
	}
	m.updatePage = crud.NewUpdateModel(m.transport,
		&m.tables[m.selectedTab],
		id, m.width)
	return m, nil
}

func processTableRefreshReq(m model) (tea.Model, tea.Cmd) {
	rows, err := m.tables[m.selectedTab].Descriptor.RefreshFn(m.transport)
	if err != nil {
		return m, popup.Error("something went wrong: "+err.Error(), "")
	}
	m.tables[m.selectedTab].Table.SetRows(rows)
	m.tables[m.selectedTab].Table.UpdateViewport()
	return m, nil
}

func processLogMsg(m model, msg logTable.LogMsg) (model, tea.Cmd) {
	newRows := append(m.tables[0].Table.Rows(), msg.Row)
	sortedRows := m.tables[0].Descriptor.SortFn(m.tables[0].Descriptor.SortedColumn,
		newRows, m.tables[0].Descriptor.SortAscending)
	m.tables[0].Table.SetRows(sortedRows)
	m.tables[0].Table.UpdateViewport()
	return m, logTable.WaitforLogs(m.logChan)
}

func processExportMsg(m model, msg exportModel.Msg) (tea.Model, tea.Cmd) {
	name, err := export.NewWordFile(m.tables[0].UnchangedRows,
		m.tables[1].UnchangedRows, msg.StartTime, msg.EndTime)
	if err != nil {
		return m, popup.Error("can't create Word report: "+err.Error(), "")
	}
	m.focusLayer = focusButtons
	return m, popup.Success("your Word report saved: "+name, "10s")
}

func processExportToExcelRequest(m model) (tea.Model, tea.Cmd) {
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
	file, err := export.Excel(tables)
	if err != nil {
		return m, popup.Error("can't create excel report: "+err.Error(), "")
	}
	return m, popup.Success("Excel report created, saved with name "+file, "10s")
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
		tables := make([]table.Model, 0)
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
	case sort.Msg:
		m.tables[m.selectedTab].Table.SetRows(msg.Rows)
		m.tables[m.selectedTab].Table.UpdateViewport()
		m.focusLayer = focusTable

	case sort.CancelMsg:
		m.focusLayer = focusButtons

	}
	return m
}

// Handle filter messages of table.
func (m model) handleFilterMsgs(msg tea.Msg) model {
	switch msg := msg.(type) {
	case filter.Msg:
		m.tables[m.selectedTab].Table.SetRows(msg.Rows)
		m.tables[m.selectedTab].Table.UpdateViewport()
		m.focusLayer = focusTable

	case filter.CancelMsg:
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
	case popup.Msg:
		updatedModel, cmd := m.popup.Update(msg)
		p, ok := updatedModel.(popup.Model)
		if !ok {
			return m, popup.Error("failed to update popup model", "")
		}
		m.popup = p
		return m, cmd

	case popup.ClearMsg:
		updatedModel, cmd := m.popup.Update(msg)
		p, ok := updatedModel.(popup.Model)
		if !ok {
			return m, popup.Error("failed to update popup model", "")
		}
		m.popup = p
		return m, cmd
	}
	return m, nil
}

// Handle delete messages for resource record table.
func (m model) handleDeleteMsgs(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case crud.DeleteMsg:
		model, cmd := m.deletePage.Update(msg)
		if v, ok := model.(crud.DeleteModel); ok {
			m.deletePage = v
		}
		return m, cmd

	case crud.DeleteCancelMsg:
		m.focusLayer = focusTable

	case crud.DeleteSuccessMsg:
		m.focusLayer = focusTable
		rows := m.tables[m.selectedTab].Table.Rows()
		newRows := make([]bubbleTable.Row, 0, len(rows))
		for _, row := range rows {
			if row[0] != strconv.FormatInt(msg.ID, 10) {
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
	// This aproach of handling message was chosen to reduce boilerplate code in handling
	// each key and then check what the focus layer is
	// (for real we just move this code to selected model)
	if msg, ok := message.(tea.KeyMsg); ok {
		switch m.focusLayer {
		case focusSearch:
			return processSearchLayer(msg, m)

		case focusExportModel:
			return processExportLayer(m, msg)

		case focusTable:
			return processTableLayer(m, msg)

		case focusLoginModel:
			return processLoginLayer(m, msg)

		case focusUpdateModel:
			return processUpdateLayer(m, msg)

		case focusSortModel:
			return processSortLayer(m, msg)

		case focusFilterModel:
			return processFilterLayer(m, msg)

		case focusAddModel:
			return processAddLayer(m, msg)

		case focusDeleteModel:
			return processDeleteLayer(m, msg)

		default:
			return processDefaultLayer(msg, m)
		}
	}
	return m, nil
}

func processDefaultLayer(msg tea.KeyMsg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, m.Close()

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll

	case key.Matches(msg, m.keys.Unfocus):
		return processUnfocusKeys(m)

	case key.Matches(msg, m.keys.Up):
		return processUpKeys(m, msg)

	case key.Matches(msg, m.keys.Down):
		return processDownKeys(m, msg)

	case key.Matches(msg, m.keys.Right):
		return processRightKeys(m, msg)

	case key.Matches(msg, m.keys.Left):
		return processLeftKeys(m, msg)

	case key.Matches(msg, m.keys.Enter):
		processEnterKeys(m, msg)

		// Delete key
	case key.Matches(msg, m.keys.Delete):
		return processDeleteKeys(m, msg)

		// Add key
	case key.Matches(msg, m.keys.Add):
		return processAddKeys(m, msg)

		// Search key
	case key.Matches(msg, m.keys.Search):
		switch m.focusLayer {
		case focusButtons, focusTable:
			m.focusLayer = focusSearch
		}

		// Filter key
	case key.Matches(msg, m.keys.Filter):
		return processFilterKeys(m, msg)
	}
	return m, cmd
}

func processFilterKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}
		return m, cmd
	}
	return m, nil
}

func processAddKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}
		return m, cmd
	}
	return m, nil
}

func processDeleteKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}
		return m, cmd
	}
	return m, nil
}

func processEnterKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusTabs:
		m.focusLayer = focusButtons
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}
		return m, cmd
	}
	return m, nil
}

func processLeftKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusTabs:
		if m.selectedTab > 0 {
			m.selectedTab--
		}
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}
		return m, cmd
	}
	return m, nil
}

func processRightKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusTabs:
		if m.selectedTab < len(m.tabs)-1 {
			m.selectedTab++
		}
	case focusButtons, focusTable:
		model, cmd := m.tables[m.selectedTab].Update(msg)
		if v, ok := model.(table.Model); ok {
			m.tables[m.selectedTab] = v
		}

		return m, cmd
	}
	return m, nil
}

func processDownKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusLayer {
	case focusTabs:
		m.focusLayer = focusSearch

	case focusSearch:
		m.focusLayer = focusButtons

	case focusButtons:
		m.focusLayer = focusTable
		m.tables[m.selectedTab].Table.Focus()

	case focusTable:
		if m.tables[m.selectedTab].Table.Focused() {
			m.tables[m.selectedTab].Table, cmd = m.tables[m.selectedTab].Table.Update(msg)
			return m, cmd
		}

	}
	return m, cmd
}

func processUpKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusLayer {
	case focusSearch:
		m.focusLayer = focusTabs

	case focusButtons:
		m.focusLayer = focusSearch

	case focusTable:
		m.tables[m.selectedTab].Table, cmd = m.tables[m.selectedTab].Table.Update(msg)
		return m, cmd

	}
	return m, nil
}

func processUnfocusKeys(m model) (model, tea.Cmd) {
	switch m.focusLayer {
	case focusTabs:
		return m, m.Close()

	case focusSearch:
		m.focusLayer = focusTabs

	case focusButtons:
		m.focusLayer = focusSearch

	}
	return m, nil
}

func processAddLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.addModel.Update(msg)
	if v, ok := model.(crud.AddModel); ok {
		m.addModel = v
	}
	return m, cmd
}

func processFilterLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.filterPage.Update(msg)
	if v, ok := model.(filter.Model); ok {
		m.filterPage = v
	}
	return m, cmd
}

func processSortLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.sortPage.Update(msg)
	if v, ok := model.(sort.Model); ok {
		m.sortPage = v
	}
	return m, cmd
}

func processUpdateLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.updatePage.Update(msg)
	if v, ok := model.(crud.UpdateModel); ok {
		m.updatePage = v
	}
	return m, cmd
}

func processLoginLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.loginPage.Update(msg)
	if v, ok := model.(account.LoginModel); ok {
		m.loginPage = v
	}
	return m, cmd
}

func processTableLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.tables[m.selectedTab].Update(msg)
	if v, ok := model.(table.Model); ok {
		m.tables[m.selectedTab] = v
	}
	return m, cmd
}

func processExportLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.exportModel.Update(msg)
	if v, ok := model.(exportModel.Model); ok {
		m.exportModel = v
	}
	return m, cmd
}

func processSearchLayer(msg tea.KeyMsg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Down):
		if !m.searchInput.Focused() {
			m.focusLayer = focusButtons
		} else {
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
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
			return m, cmd
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
	return m, nil
}

func processDeleteLayer(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	model, cmd := m.deletePage.Update(msg)
	if v, ok := model.(crud.DeleteModel); ok {
		m.deletePage = v
	}
	m.focusLayer = focusButtons
	return m, cmd
}
