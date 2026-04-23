package tui

import "github.com/charmbracelet/lipgloss"

var (
	green  = lipgloss.Color("#16c79a")
	yellow = lipgloss.Color("#f5c542")
	red    = lipgloss.Color("#e74c3c")
	blue   = lipgloss.Color("#3498db")
	white  = lipgloss.Color("#ffffff")
	gray   = lipgloss.Color("#888888")
	dimmed = lipgloss.Color("#444444")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmed).
			Padding(1, 2).
			MarginRight(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(gray)

	valueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(gray)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			Background(lipgloss.Color("#2a2a3e")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(gray).
				Padding(0, 2)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(green)

	bigNumberStyle = lipgloss.NewStyle().
			Bold(true)
)

func recoveryColor(score float64) lipgloss.Color {
	switch {
	case score >= 67:
		return green
	case score >= 34:
		return yellow
	default:
		return red
	}
}

func strainColor(strain float64) lipgloss.Color {
	switch {
	case strain >= 18:
		return red
	case strain >= 14:
		return yellow
	default:
		return blue
	}
}
