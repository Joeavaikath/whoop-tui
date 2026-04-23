package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jvaikath/whoop-tui/internal/api"
)

type tab int

const (
	tabToday tab = iota
	tabWeek
	tab30Day
	tab60Day
)

var tabNames = []string{"Today", "Week", "30 Days", "60 Days"}

type tabCache struct {
	cycles     []api.Cycle
	recoveries []api.Recovery
	sleeps     []api.Sleep
	workouts   []api.Workout
	dayEntries []dayEntry
	fetchedAt  time.Time
}

type model struct {
	client  *api.Client
	tab     tab
	loading bool
	spinner spinner.Model
	err     error
	width   int
	height  int

	profile  *api.UserProfile
	body     *api.BodyMeasurement
	cache    map[tab]*tabCache

	// active view data (points into cache)
	cycles     []api.Cycle
	recoveries []api.Recovery
	sleeps     []api.Sleep
	workouts   []api.Workout

	// drill-down
	detailOpen bool
	cursor     int
	dayEntries []dayEntry
}

type dayEntry struct {
	date     time.Time
	cycle    *api.Cycle
	recovery *api.Recovery
	sleep    *api.Sleep
	workouts []api.Workout
}

type profileLoadedMsg struct {
	profile *api.UserProfile
	body    *api.BodyMeasurement
	err     error
}

type dataLoadedMsg struct {
	tab        tab
	cycles     []api.Cycle
	recoveries []api.Recovery
	sleeps     []api.Sleep
	workouts   []api.Workout
	err        error
}

func newModel(client *api.Client) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(green)

	return model{
		client:  client,
		tab:     tabToday,
		loading: true,
		spinner: s,
		cache:   make(map[tab]*tabCache),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadProfile(), m.loadDataForTab(tabToday))
}

func (m model) loadProfile() tea.Cmd {
	return func() tea.Msg {
		profile, err := m.client.GetProfile()
		if err != nil {
			return profileLoadedMsg{err: err}
		}
		body, _ := m.client.GetBodyMeasurement()
		return profileLoadedMsg{profile: profile, body: body}
	}
}

func (m *model) switchToCache(t tab) {
	c := m.cache[t]
	m.cycles = c.cycles
	m.recoveries = c.recoveries
	m.sleeps = c.sleeps
	m.workouts = c.workouts
	m.dayEntries = c.dayEntries
	m.cursor = 0
}

func (m model) loadDataForTab(t tab) tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		var start time.Time
		switch t {
		case tabToday:
			start = now.AddDate(0, 0, -2)
		case tabWeek:
			start = now.AddDate(0, 0, -7)
		case tab30Day:
			start = now.AddDate(0, 0, -30)
		case tab60Day:
			start = now.AddDate(0, 0, -60)
		}

		startStr := start.Format(time.RFC3339)

		cycles, err := m.client.GetAllCycles(startStr, "")
		if err != nil {
			return dataLoadedMsg{tab: t, err: fmt.Errorf("cycles: %w", err)}
		}

		recoveries, err := m.client.GetAllRecoveries(startStr, "")
		if err != nil {
			return dataLoadedMsg{tab: t, err: fmt.Errorf("recoveries: %w", err)}
		}

		sleeps, err := m.client.GetAllSleeps(startStr, "")
		if err != nil {
			return dataLoadedMsg{tab: t, err: fmt.Errorf("sleeps: %w", err)}
		}

		workouts, err := m.client.GetAllWorkouts(startStr, "")
		if err != nil {
			return dataLoadedMsg{tab: t, err: fmt.Errorf("workouts: %w", err)}
		}

		return dataLoadedMsg{
			tab:        t,
			cycles:     cycles,
			recoveries: recoveries,
			sleeps:     sleeps,
			workouts:   workouts,
		}
	}
}

func (m *model) buildDayEntries() {
	dayMap := make(map[string]*dayEntry)

	for i := range m.cycles {
		c := &m.cycles[i]
		key := c.Start.Format("2006-01-02")
		if _, ok := dayMap[key]; !ok {
			dayMap[key] = &dayEntry{date: c.Start}
		}
		dayMap[key].cycle = c
	}

	for i := range m.recoveries {
		r := &m.recoveries[i]
		key := r.CreatedAt.Format("2006-01-02")
		if _, ok := dayMap[key]; !ok {
			dayMap[key] = &dayEntry{date: r.CreatedAt}
		}
		dayMap[key].recovery = r
	}

	for i := range m.sleeps {
		s := &m.sleeps[i]
		if s.Nap {
			continue
		}
		key := s.End.Format("2006-01-02")
		if _, ok := dayMap[key]; !ok {
			dayMap[key] = &dayEntry{date: s.End}
		}
		dayMap[key].sleep = s
	}

	for i := range m.workouts {
		w := &m.workouts[i]
		key := w.Start.Format("2006-01-02")
		if _, ok := dayMap[key]; !ok {
			dayMap[key] = &dayEntry{date: w.Start}
		}
		dayMap[key].workouts = append(dayMap[key].workouts, *w)
	}

	m.dayEntries = make([]dayEntry, 0, len(dayMap))
	for _, e := range dayMap {
		m.dayEntries = append(m.dayEntries, *e)
	}
	sort.Slice(m.dayEntries, func(i, j int) bool {
		return m.dayEntries[i].date.After(m.dayEntries[j].date)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case profileLoadedMsg:
		if msg.err == nil {
			m.profile = msg.profile
			m.body = msg.body
		}

	case dataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.cycles = msg.cycles
			m.recoveries = msg.recoveries
			m.sleeps = msg.sleeps
			m.workouts = msg.workouts
			m.buildDayEntries()
			m.cursor = 0

			m.cache[msg.tab] = &tabCache{
				cycles:     msg.cycles,
				recoveries: msg.recoveries,
				sleeps:     msg.sleeps,
				workouts:   msg.workouts,
				dayEntries: m.dayEntries,
				fetchedAt:  time.Now(),
			}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "1":
		return m.switchTab(tabToday)
	case "2":
		return m.switchTab(tabWeek)
	case "3":
		return m.switchTab(tab30Day)
	case "4":
		return m.switchTab(tab60Day)

	case "tab", "right", "l":
		if !m.detailOpen {
			return m.switchTab((m.tab + 1) % 4)
		}
	case "shift+tab", "left", "h":
		if !m.detailOpen {
			return m.switchTab((m.tab + 3) % 4)
		}

	case "j", "down":
		if m.tab != tabToday && !m.detailOpen && len(m.dayEntries) > 0 {
			m.cursor++
			if m.cursor >= len(m.dayEntries) {
				m.cursor = len(m.dayEntries) - 1
			}
		}
	case "k", "up":
		if m.tab != tabToday && !m.detailOpen {
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
		}

	case "enter":
		if m.tab != tabToday && !m.detailOpen && len(m.dayEntries) > 0 {
			m.detailOpen = true
		}
	case "esc":
		if m.detailOpen {
			m.detailOpen = false
		}

	case "r":
		delete(m.cache, m.tab)
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadDataForTab(m.tab))
	}

	return m, nil
}

func (m model) switchTab(t tab) (tea.Model, tea.Cmd) {
	m.tab = t
	m.detailOpen = false

	if c, ok := m.cache[t]; ok {
		m.switchToCache(t)
		_ = c
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.spinner.Tick, m.loadDataForTab(t))
}

func (m model) View() string {
	if m.loading {
		s := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center)
		return s.Render(m.spinner.View() + " Loading your WHOOP data...")
	}

	if m.err != nil {
		return m.viewError()
	}

	var sections []string

	// header
	greeting := "WHOOP"
	if m.profile != nil {
		greeting = fmt.Sprintf("WHOOP — %s %s", m.profile.FirstName, m.profile.LastName)
	}
	sections = append(sections, titleStyle.Render(greeting))

	// tabs
	sections = append(sections, m.renderTabs())
	sections = append(sections, "")

	// content
	if m.detailOpen && len(m.dayEntries) > 0 {
		sections = append(sections, m.viewDetail(m.dayEntries[m.cursor]))
	} else {
		switch m.tab {
		case tabToday:
			sections = append(sections, m.viewToday())
		case tabWeek, tab30Day, tab60Day:
			sections = append(sections, m.viewTrend())
		}
	}

	// help
	help := "1-4 tabs • ←→ switch"
	if m.tab != tabToday && !m.detailOpen {
		help += " • ↑↓ select day • enter detail"
	}
	if m.detailOpen {
		help += " • esc back"
	}
	help += " • r refresh • q quit"
	sections = append(sections, "", helpStyle.Render(help))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().Padding(1, 2).Width(m.width).Height(m.height).Render(content)
}

func (m model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if tab(i) == m.tab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, tabs...)
}

func (m model) viewError() string {
	s := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	errBox := cardStyle.
		BorderForeground(red).
		Render(
			lipgloss.NewStyle().Foreground(red).Bold(true).Render("Error") + "\n\n" +
				m.err.Error() + "\n\n" +
				helpStyle.Render("Press r to retry, q to quit"),
		)

	return s.Render(errBox)
}
