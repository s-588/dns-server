package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/ui/tui/style"
)

func (m model) View() string {
	raw := m.rawView()
	// Ensure the outer border respects the full terminal height
	return style.BaseBorderStyle.Height(m.height - 2).Width(m.width - 2).Render(raw)
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

	help := fmt.Sprintf(
		"Need help? Press ? for full instructions. Press Esc to exit. Use %c /%c  and %c /%c  to navigate. Press Enter to confirm your choice.",
		'\uea9b', '\uea9c', '\ueaa1', '\uea9a',
	)

	switch m.focusLayer {
	case focusTable:
		if m.selectedTab == 1 {
			help += " Press D to delete record. Press A to add record. Press / to find record. Press F to filter."
		}
	}
	s.WriteString(style.FooterStyle.Render(help))

	return s.String()
}
