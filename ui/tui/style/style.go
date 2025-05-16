package style

import "github.com/charmbracelet/lipgloss"

var (
	PurpleColor = lipgloss.Color("#8839ef")
	TextColor   = lipgloss.Color("")
	GreenColor  = lipgloss.Color("#40a02b")
	BlueColor   = lipgloss.Color("#1e66f5")
	PinkColor   = lipgloss.Color("#ff87d7")
	RedColor    = lipgloss.Color("#e64553")

	BaseBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PurpleColor).
			Padding(1, 2) // padding inside the border

	UnselectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PinkColor)

	SelectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleColor)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Background(PurpleColor).
			Padding(0, 2).
			Align(lipgloss.Center).
			Width(50)

	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(TextColor)

	SelectedButtonStyle = ButtonStyle.
				Background(PurpleColor).
				Bold(true)

	SecondarySelectedButtonStyle = ButtonStyle.
					Background(PinkColor).
					Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 2)

	BaseStyle = lipgloss.NewStyle().Padding(1, 2)

	RedBoarderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#e64553"))

	GreenBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#40a02b"))

	BlueBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1e66f5"))
)
