package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	raw := m.rawView() // your existing UI rendering logic
	return baseBorderStyle.Render(raw)
}

func (m model) rawView() string {
	s := strings.Builder{}

	var table *table.Model
	switch m.selectedTab {
	case 0:
		table = &m.logTable
	case 1:
		table = &m.rrTable
	}

	// Header with border
	s.WriteString(headerStyle.Render("DNS Server Dashboard"))
	s.WriteString("\n\n")

	// Tabs row
	tabs := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		if m.selectedTab == i {
			if m.focusLayer == focusTabs {
				tabs[i] = selectedButtonStyle.Render(tab.name)
			} else {
				tabs[i] = secondarySelectedButtonStyle.Render(tab.name)
			}
		} else {
			tabs[i] = buttonStyle.Render(tab.name)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...)))
	s.WriteString("\n\n")

	// Content area with border
	buttons := make([]string, len(m.tabs[m.selectedTab].buttons))
	for i, btn := range m.tabs[m.selectedTab].buttons {
		if m.tabs[m.selectedTab].cursor == i {
			if m.focusLayer == focusButtons {
				buttons[i] = selectedButtonStyle.Render(btn)
			} else {
				buttons[i] = secondarySelectedButtonStyle.Render(btn)
			}
		} else {
			buttons[i] = buttonStyle.Render(btn)
		}
	}

	// Second button row
	buttonsAlignCenter = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, buttons...)))
	s.WriteString("\n\n")

	if table.Focused() {
		s.WriteString(selectedBoarderStyle.Render(table.View()))
	} else {
		s.WriteString(unselectedBoarderStyle.Render(table.View()))
	}

	s.WriteString("\n\n")

	// Footer with border
	s.WriteString(footerStyle.Render("Press q to quit. Use ↑/↓ to navigate."))

	return s.String()
}
