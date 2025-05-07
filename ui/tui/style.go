package tui

import "github.com/charmbracelet/lipgloss"

var (
	purpleColor = lipgloss.Color("#8839ef")
	textColor   = lipgloss.Color("")
	greenColor  = lipgloss.Color("#40a02b")
	blueColor   = lipgloss.Color("#1e66f5")
	pinkColor   = lipgloss.Color("#ff87d7")
	redColor    = lipgloss.Color("#d20f39")

	baseBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleColor).
			Padding(1, 2) // padding inside the border

	unselectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(pinkColor)

	selectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purpleColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Background(purpleColor).
			Padding(0, 2).
			Align(lipgloss.Center).
			Width(50)

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(textColor)

	selectedButtonStyle = buttonStyle.
				Background(purpleColor).
				Bold(true)

	secondarySelectedButtonStyle = buttonStyle.
					Background(pinkColor).
					Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 2)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	baseStyle = lipgloss.NewStyle().Padding(1, 2)

	errorBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(redColor)

	successBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(greenColor)

	infoBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(blueColor)
)
