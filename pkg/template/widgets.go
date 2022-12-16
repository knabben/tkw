package template

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	DescStyle = lipgloss.NewStyle().MarginTop(1).
			Foreground(lipgloss.Color("#fff")).
			Background(lipgloss.Color("9")).
			Bold(true).
			Padding(1, 3, 1, 3).
			MarginBottom(1)

	SimpleStyle = lipgloss.NewStyle().
			Bold(true).
			MarginTop(1).
			MarginBottom(1).
			Foreground(lipgloss.Color("#ADD8E6")).
			Background(lipgloss.Color("#000")).
			Padding(1, 3, 1, 3)

	titleStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderBottom(true).
			BorderForeground(Subtle).
			PaddingTop(2)
)

func Info(msg string) string {
	return SimpleStyle.Copy().Render(msg)
}

func Error(msg string) string {
	return lipgloss.JoinVertical(lipgloss.Top,
		titleStyle.Render("** FATAL ERROR **"),
		DescStyle.Copy().Render(msg),
	)
}

func Warning(msg string) string {
	return lipgloss.JoinVertical(lipgloss.Top,
		titleStyle.Render("** WARNING **"),
		DescStyle.Copy().Background(lipgloss.Color("#FF8C00")).Render(msg),
	)
}
