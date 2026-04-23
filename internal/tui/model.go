package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jvaikath/whoop-tui/internal/api"
)

type view int

const (
	viewLoading view = iota
	viewDashboard
)

type dashboardData struct {
	profile  *api.UserProfile
	body     *api.BodyMeasurement
	cycle    *api.Cycle
	recovery *api.Recovery
	sleep    *api.Sleep
	workouts []api.Workout
}

type dataLoadedMsg struct {
	data dashboardData
	err  error
}

type model struct {
	client   *api.Client
	view     view
	data     dashboardData
	spinner  spinner.Model
	err      error
	width    int
	height   int
}

func newModel(client *api.Client) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#16c79a"))

	return model{
		client:  client,
		view:    viewLoading,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadData())
}

func (m model) loadData() tea.Cmd {
	return func() tea.Msg {
		var d dashboardData

		profile, err := m.client.GetProfile()
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading profile: %w", err)}
		}
		d.profile = profile

		body, err := m.client.GetBodyMeasurement()
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading body: %w", err)}
		}
		d.body = body

		now := time.Now()
		start := now.AddDate(0, 0, -1).Format(time.RFC3339)

		cycles, err := m.client.GetCycles(start, "", 1)
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading cycles: %w", err)}
		}
		if len(cycles.Records) > 0 {
			d.cycle = &cycles.Records[0]
		}

		recoveries, err := m.client.GetRecoveries(start, "", 1)
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading recovery: %w", err)}
		}
		if len(recoveries.Records) > 0 {
			d.recovery = &recoveries.Records[0]
		}

		sleeps, err := m.client.GetSleeps(start, "", 1)
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading sleep: %w", err)}
		}
		if len(sleeps.Records) > 0 {
			d.sleep = &sleeps.Records[0]
		}

		weekStart := now.AddDate(0, 0, -7).Format(time.RFC3339)
		workouts, err := m.client.GetWorkouts(weekStart, "", 10)
		if err != nil {
			return dataLoadedMsg{err: fmt.Errorf("loading workouts: %w", err)}
		}
		d.workouts = workouts.Records

		return dataLoadedMsg{data: d}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.view = viewLoading
			return m, tea.Batch(m.spinner.Tick, m.loadData())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.view = viewDashboard
		} else {
			m.data = msg.data
			m.view = viewDashboard
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.view {
	case viewLoading:
		return m.viewLoading()
	case viewDashboard:
		if m.err != nil {
			return m.viewError()
		}
		return m.viewDashboard()
	}
	return ""
}
