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
	"tkw/pkg/template"
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

func (m model) View() string {
	s := m.timer.View()
	if m.timer.Timedout() {
		s = "Peace!"
	}
	if !m.quitting {
		s = lipgloss.JoinVertical(lipgloss.Top,
			template.Error(m.err.Error()),
			template.SimpleStyle.Render(fmt.Sprintf("exiting in %s", s)),
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
