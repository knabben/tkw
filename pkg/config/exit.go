package config

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8s.io/klog/v2"
	"os"
	"time"
)

var timeout = time.Second * 5

type model struct {
	timer    timer.Model
	help     help.Model
	quitting bool
	err      error
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	descStyle = lipgloss.NewStyle().MarginTop(1).
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("9")).
		Bold(true).Padding(3)
	infoStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(subtle)
)

func (m model) View() string {
	s := m.timer.View()
	if m.timer.Timedout() {
		s = "Peace!"
	}
	if !m.quitting {
		s = lipgloss.JoinVertical(lipgloss.Top,
			descStyle.Render(m.err.Error()),
			infoStyle.Render(fmt.Sprintf("exiting in %s", s)),
		)
	}
	return s
}

func ExplodeGraceful(err error) {
	if err == nil {
		return
	}

	m := model{
		timer: timer.NewWithInterval(timeout, time.Millisecond),
		help:  help.New(),
		err:   err,
	}
	if err := tea.NewProgram(m).Start(); err != nil {
		klog.Error("Uh oh, we encountered an error:", err)
	}

	os.Exit(1)
}
