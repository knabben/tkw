package template

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	table table.Model
	title string
	rows  int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) View() string {
	var style string
	if m.rows == 0 {
		style = messageStyle.Render("No properties exists on this template.")
	} else {
		style = baseStyle.Render(m.table.View())
	}
	return fmt.Sprintf("%s\n%s\n", titleStyle.Render(m.title), style)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "enter":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
