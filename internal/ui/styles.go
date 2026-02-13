package ui

import "github.com/charmbracelet/lipgloss"

var (
	baseStyle = lipgloss.NewStyle().Padding(1)

	heroStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229"))

	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("121")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("67")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	activeColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("229")).
				Padding(0, 1)

	drawerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("111")).
			Padding(0, 1)

	activePromptStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("229")).
				Padding(0, 1)

	helpPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("229")).
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("252")).
			Padding(1, 2)

	focusedBoardStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("60")).
				Bold(true)

	contextMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("186")).
				Bold(true)

	promptBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("67")).
			Padding(0, 1)

	promptBarFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("229")).
				Padding(0, 1)

	promptModeBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("16")).
				Background(lipgloss.Color("117")).
				Bold(true).
				Padding(0, 1)

	promptMoveSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Bold(true)

	promptMoveNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246"))
)

func badge(text string, fg, bg lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		Bold(true).
		Render(text)
}

func colorForName(s string) lipgloss.Color {
	palette := []string{"149", "117", "186", "179", "213", "81", "221", "151"}
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += int(s[i])
	}
	return lipgloss.Color(palette[sum%len(palette)])
}

func listBadge(name string) string {
	return badge(name, lipgloss.Color("16"), colorForName(name))
}
