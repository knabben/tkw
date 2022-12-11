package docker

import (
	"context"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"time"
	"tkw/pkg/config"
)

type model struct {
	ready       bool
	err         error
	content     string
	Client      *Docker
	containerID string
	timer       timer.Model
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func NewProgram(cli *Docker, containerID string) (*tea.Program, error) {
	p := tea.NewProgram(
		model{
			containerID: containerID,
			Client:      cli,
			timer:       timer.NewWithInterval(5*time.Hour, time.Second),
			content:     string("Loading..."),
		},
		tea.WithAltScreen(),
	)
	return p, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

	case timer.TickMsg:
		var cmd tea.Cmd
		response, err := m.Client.GetLogs(context.Background(), m.containerID)
		if err != nil {
			m.err = err
		}
		if len(response) >= len(m.content) {
			m.content = response
		}
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		config.ExplodeGraceful(m.err)
	}
	return m.content
}
