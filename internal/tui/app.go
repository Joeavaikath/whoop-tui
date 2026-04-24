package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvaikath/whoop-tui/internal/api"
)

func Run(client *api.Client) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
