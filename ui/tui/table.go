package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) getTable() *table.Model {
	var table *table.Model
	if m.selectedTab == 0 {
		table = &m.logTable
	}
	if m.selectedTab == 1 {
		table = &m.rrTable
	}
	return table
}

func (m model) updateTable(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	if m.selectedTab == 0 {
		m.logTable, cmd = m.logTable.Update(msg)
	}
	if m.selectedTab == 1 {
		m.rrTable, cmd = m.rrTable.Update(msg)
	}
	return m, cmd
}
