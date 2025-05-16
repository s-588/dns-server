package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/ui/tui/style"
)

func (m model) View() string {
	raw := m.rawView() // your existing UI rendering logic
	return style.BaseBorderStyle.Render(raw)
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
	s.WriteString(style.HeaderStyle.Render("DNS Server Dashboard"))
	s.WriteString("\n\n")

	// Tabs row
	tabs := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		if m.selectedTab == i {
			if m.focusLayer == focusTabs {
				tabs[i] = style.SelectedButtonStyle.Render(tab.name)
			} else {
				tabs[i] = style.SecondarySelectedButtonStyle.Render(tab.name)
			}
		} else {
			tabs[i] = style.ButtonStyle.Render(tab.name)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...)))
	s.WriteString("\n\n")

	// Content area
	switch m.focusLayer {
	case focusDeletePage:
		s.WriteString(m.deletePage.View())

	case focusAddPage:
		s.WriteString(m.addPage.View())

	default:
		// Buttons row
		buttons := make([]string, len(m.tabs[m.selectedTab].buttons))
		for i, btn := range m.tabs[m.selectedTab].buttons {
			if m.tabs[m.selectedTab].cursor == i {
				if m.focusLayer == focusButtons {
					buttons[i] = style.SelectedButtonStyle.Render(btn)
				} else {
					buttons[i] = style.SecondarySelectedButtonStyle.Render(btn)
				}
			} else {
				buttons[i] = style.ButtonStyle.Render(btn)
			}
		}
		buttonsAlignCenter = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
		s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, buttons...)))
		s.WriteString("\n\n")

		// Table
		if table.Focused() {
			s.WriteString(style.SelectedBoarderStyle.Render(table.View()))
		} else {
			s.WriteString(style.UnselectedBoarderStyle.Render(table.View()))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(m.popup.View())

	// Footer with border
	s.WriteString("\n")
	s.WriteString(style.FooterStyle.Render("Press q to quit. Use ↑/↓ to navigate."))

	return s.String()
}
