package template

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

const (
	minTokens  = 1
	halfTokens = 2
	maxTokens  = 4
)

var (
	columns     = []table.Column{{Title: "ID", Width: 50}, {Title: "Value", Width: 40}}
	parseMargin = parsePadding
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var messageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#fff")).
	BorderForeground(lipgloss.Color("#f0f0")).
	Align(lipgloss.Center).
	Height(5).
	Width(100).
	Margin(parseMargin("0 2")).
	Padding(parsePadding("0 4 4 0")).
	Bold(true)

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#fff")).
	BorderForeground(lipgloss.Color("211")).
	Align(lipgloss.Center).
	Border(lipgloss.DoubleBorder()).
	Height(5).
	Width(95).
	Padding(parsePadding("4 4")).
	Bold(true)

// RenderTable plots out the table with the properties of the VM
func RenderTable(propsRows map[string]string, title string) error {
	rows := []table.Row{}
	for k, v := range propsRows {
		rows = append(rows, table.Row{k, v})
	}
	s := table.DefaultStyles()
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
		table.WithWidth(98),
	)
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t, title, len(propsRows)}
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		return err
	}
	return nil
}

// parsePadding parses 1 - 4 integers from a string and returns them in a top,
// right, bottom, left order for use in the lipgloss.Padding() method.
func parsePadding(s string) (int, int, int, int) {
	var ints [maxTokens]int

	tokens := strings.Split(s, " ")

	if len(tokens) > maxTokens {
		return 0, 0, 0, 0
	}

	// All tokens must be an integer
	for i, token := range tokens {
		parsed, err := strconv.Atoi(token)
		if err != nil {
			return 0, 0, 0, 0
		}
		ints[i] = parsed
	}

	if len(tokens) == minTokens {
		return ints[0], ints[0], ints[0], ints[0]
	}

	if len(tokens) == halfTokens {
		return ints[0], ints[1], ints[0], ints[1]
	}

	if len(tokens) == maxTokens {
		return ints[0], ints[1], ints[2], ints[3]
	}

	return 0, 0, 0, 0
}
