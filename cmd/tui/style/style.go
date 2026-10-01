// Package style provides styling utilities for the TUI components.
package style

import "github.com/charmbracelet/lipgloss"

var (
	PurpleColor = lipgloss.Color("#8839ef") // Purple color used for borders and highlights
	TextColor   = lipgloss.Color("#ffffff") // Text color
	GreenColor  = lipgloss.Color("#40a02b") // Green color
	BlueColor   = lipgloss.Color("#1e66f5") // Blue color
	PinkColor   = lipgloss.Color("#ff87d7") // Pink color
	RedColor    = lipgloss.Color("#e64553") // Red color

	// BaseBorderWithPaddingsStyle defines the base style for borders with padding.
	BaseBorderWithPaddingsStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(PurpleColor).
					Padding(1, 2)

	// BaseBorderStyle defines the base style for borders with no padding or margin.
	BaseBorderStyle = lipgloss.NewStyle().
			Padding(0, 0, 0, 0).
			Margin(0, 0, 0, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PurpleColor)

	// UnselectedBoarderStyle defines the style for unselected borders.
	UnselectedBoarderStyle = lipgloss.NewStyle().
				Padding(0, 0, 0, 0).
				Margin(0, 0, 0, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PinkColor)

	// SelectedBoarderStyle defines the style for selected borders.
	SelectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleColor)

	// HeaderStyle defines the style for headers.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Background(PurpleColor).
			Padding(0, 2).
			Align(lipgloss.Center)

		// ButtonStyle defines the style for buttons.
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(TextColor)

		// SelectedButtonStyle defines the style for selected buttons.
	SelectedButtonStyle = ButtonStyle.
				Background(PurpleColor).
				Bold(true)

		// SecondarySelectedButtonStyle defines the style for secondary selected buttons.
	SecondarySelectedButtonStyle = ButtonStyle.
					Background(PinkColor).
					Bold(true)

		// FooterStyle defines the style for footers.
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 2)

		// BaseStyle defines the base style for the TUI.
	BaseStyle = lipgloss.NewStyle().Padding(1, 2)

	// RedBoarderStyle defines the style for red borders.
	RedBoarderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#e64553"))

		// GreenBoarderStyle defines the style for green borders.
	GreenBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#40a02b"))

		// BlueBoarderStyle defines the style for blue borders.
	BlueBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1e66f5"))

		// FocusedInputStyle defines the style for focused input fields.
	FocusedInputStyle = lipgloss.NewStyle().Foreground(PurpleColor)

	// BlurredInputStyle defines the style for blurred input fields.
	BlurredInputStyle = lipgloss.NewStyle().Foreground(PinkColor)
)
