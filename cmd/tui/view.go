package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/style"
)

// Render view on the screen.
func (m model) View() string {
	raw := m.rawView()
	// Ensure the outer border respects the full terminal height
	return style.BaseBorderWithPaddingsStyle.Height(m.height - 2).Width(m.width - 2).Render(raw)
}

// Render fill of the interface
// This logic was separated from the View function for rendering boarder of the interface.
func (m model) rawView() string {
	s := strings.Builder{}

	s.WriteString(style.HeaderStyle.Width(m.width - 6).Render("DNS Server Dashboard"))
	s.WriteString("\n\n")

	switch m.focusLayer {
	case focusTabs, focusSearch, focusTable, focusButtons:
		styledButtons := make([]string, len(m.tabs))
		for i, tab := range m.tabs {
			if m.selectedTab == i {
				if m.focusLayer == focusTabs {
					styledButtons[i] = style.SelectedButtonStyle.Underline(true).Render(tab)
				} else {
					styledButtons[i] = style.SecondarySelectedButtonStyle.Render(tab)
				}
			} else {
				styledButtons[i] = style.ButtonStyle.Render(tab)
			}
		}

		buttonsAlignCenter := lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center)
		s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, styledButtons...)))
		s.WriteString("\n")
	}

	switch m.focusLayer {
	case focusExportModel:
		s.WriteString(m.exportModel.View())

	case focusLoginModel:
		s.WriteString(m.loginPage.View())

	case focusUpdateModel:
		s.WriteString(m.updatePage.View())

	case focusSortModel:
		s.WriteString(m.sortPage.View())

	case focusDeleteModel:
		s.WriteString(m.deletePage.View())

	case focusAddModel:
		s.WriteString(m.addModel.View())

	case focusFilterModel:
		s.WriteString(m.filterPage.View())

	case focusButtons, focusSearch, focusTabs, focusTable:
		if m.selectedTab <= 2 {
			if m.focusLayer == focusSearch {
				s.WriteString(lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center).
					Render(
						style.BaseBorderStyle.Width((m.width - 4) / 3).Render(
							style.FocusedInputStyle.Render(m.searchInput.View()),
						),
					),
				)
			} else {
				s.WriteString(lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center).
					Render(
						style.UnselectedBoarderStyle.Width((m.width - 4) / 3).Render(
							style.BlurredInputStyle.Render(m.searchInput.View()),
						),
					),
				)
			}
			s.WriteString("\n")

			s.WriteString(m.tables[m.selectedTab].View())
		}
	}

	s.WriteString("\n")
	s.WriteString(m.popup.View())

	s.WriteString("\n")
	s.WriteString(style.FooterStyle.Render("Curently selected: " + focusNames[m.focusLayer]))
	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	return s.String()
}
