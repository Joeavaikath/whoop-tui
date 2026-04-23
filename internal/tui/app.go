package tui

import (
	"github.com/jvaikath/whoop-tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(client *api.Client) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
