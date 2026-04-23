package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	green  = lipgloss.Color("#16c79a")
	yellow = lipgloss.Color("#f5c542")
	red    = lipgloss.Color("#e74c3c")
	white  = lipgloss.Color("#ffffff")
	gray   = lipgloss.Color("#666666")
	dim    = lipgloss.Color("#444444")
	bg     = lipgloss.Color("#1a1a2e")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dim).
			Padding(1, 2).
			MarginRight(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(gray)

	valueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(gray).
			MarginTop(1)
)

func recoveryColor(score float64) lipgloss.Color {
	switch {
	case score >= 67:
		return green
	case score >= 34:
		return yellow
	default:
		return red
	}
}

func strainColor(strain float64) lipgloss.Color {
	switch {
	case strain >= 18:
		return red
	case strain >= 14:
		return yellow
	default:
		return green
	}
}

func formatDuration(millis int) string {
	d := time.Duration(millis) * time.Millisecond
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func (m model) viewLoading() string {
	s := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return s.Render(
		m.spinner.View() + " Loading your WHOOP data...",
	)
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
				lipgloss.NewStyle().Foreground(gray).Render("Press r to retry, q to quit"),
		)

	return s.Render(errBox)
}

func (m model) viewDashboard() string {
	var sections []string

	greeting := "WHOOP"
	if m.data.profile != nil {
		greeting = fmt.Sprintf("WHOOP — %s %s", m.data.profile.FirstName, m.data.profile.LastName)
	}
	sections = append(sections, titleStyle.Render(greeting))

	var cards []string

	cards = append(cards, m.recoveryCard())
	cards = append(cards, m.strainCard())
	cards = append(cards, m.sleepCard())

	row := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	sections = append(sections, row)

	if len(m.data.workouts) > 0 {
		sections = append(sections, m.workoutsCard())
	}

	if m.data.body != nil {
		sections = append(sections, m.bodyCard())
	}

	sections = append(sections, helpStyle.Render("r refresh • q quit"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(content)
}

func (m model) recoveryCard() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Recovery"))

	if m.data.recovery == nil || m.data.recovery.Score == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	score := m.data.recovery.Score
	color := recoveryColor(score.RecoveryScore)

	scoreStr := fmt.Sprintf("%.0f%%", score.RecoveryScore)
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(color).Render(scoreStr))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("HRV     ")+valueStyle.Render(fmt.Sprintf("%.0f ms", score.HrvRmssdMilli)))
	lines = append(lines, labelStyle.Render("RHR     ")+valueStyle.Render(fmt.Sprintf("%.0f bpm", score.RestingHeartRate)))

	if score.Spo2Percentage != nil {
		lines = append(lines, labelStyle.Render("SpO2    ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *score.Spo2Percentage)))
	}
	if score.SkinTempCelsius != nil {
		lines = append(lines, labelStyle.Render("Temp    ")+valueStyle.Render(fmt.Sprintf("%.1f°C", *score.SkinTempCelsius)))
	}

	return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) strainCard() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Strain"))

	if m.data.cycle == nil || m.data.cycle.Score == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	score := m.data.cycle.Score
	color := strainColor(score.Strain)

	strainStr := fmt.Sprintf("%.1f", score.Strain)
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(color).Render(strainStr))
	lines = append(lines, "")

	barWidth := 20
	filled := int((score.Strain / 21.0) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(dim).Render(strings.Repeat("░", barWidth-filled))
	lines = append(lines, bar)
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Avg HR  ")+valueStyle.Render(fmt.Sprintf("%d bpm", score.AverageHeartRate)))
	lines = append(lines, labelStyle.Render("Max HR  ")+valueStyle.Render(fmt.Sprintf("%d bpm", score.MaxHeartRate)))
	lines = append(lines, labelStyle.Render("Cal     ")+valueStyle.Render(fmt.Sprintf("%.0f kcal", score.Kilojoule*0.239006)))

	return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) sleepCard() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Sleep"))

	if m.data.sleep == nil || m.data.sleep.Score == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	score := m.data.sleep.Score
	stages := score.StageSummary

	totalSleep := stages.TotalLightSleepTimeMilli +
		stages.TotalSlowWaveSleepTimeMilli +
		stages.TotalRemSleepTimeMilli

	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render(formatDuration(totalSleep)))
	lines = append(lines, "")

	if score.SleepPerformancePercentage != nil {
		perfColor := recoveryColor(*score.SleepPerformancePercentage)
		lines = append(lines, labelStyle.Render("Perf    ")+lipgloss.NewStyle().Bold(true).Foreground(perfColor).Render(fmt.Sprintf("%.0f%%", *score.SleepPerformancePercentage)))
	}
	if score.SleepEfficiencyPercentage != nil {
		lines = append(lines, labelStyle.Render("Eff     ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *score.SleepEfficiencyPercentage)))
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("REM     ")+valueStyle.Render(formatDuration(stages.TotalRemSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Deep    ")+valueStyle.Render(formatDuration(stages.TotalSlowWaveSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Light   ")+valueStyle.Render(formatDuration(stages.TotalLightSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Awake   ")+valueStyle.Render(formatDuration(stages.TotalAwakeTimeMilli)))

	return cardStyle.Width(26).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) workoutsCard() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Recent Workouts"))

	for _, w := range m.data.workouts {
		dur := w.End.Sub(w.Start)
		h := int(dur.Hours())
		mins := int(dur.Minutes()) % 60

		strain := "—"
		if w.Score != nil {
			strain = fmt.Sprintf("%.1f", w.Score.Strain)
		}

		timeStr := ""
		if h > 0 {
			timeStr = fmt.Sprintf("%dh%dm", h, mins)
		} else {
			timeStr = fmt.Sprintf("%dm", mins)
		}

		date := w.Start.Format("Jan 2")
		line := fmt.Sprintf("%-12s  %5s  strain %s  %s",
			truncate(w.SportName, 12),
			timeStr,
			strain,
			date,
		)
		lines = append(lines, labelStyle.Render(line))
	}

	return cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) bodyCard() string {
	b := m.data.body
	heightFt := b.HeightMeter * 3.28084
	weightLb := b.WeightKilogram * 2.20462

	content := fmt.Sprintf("%s %s  %s %s  %s %s",
		labelStyle.Render("Height"),
		valueStyle.Render(fmt.Sprintf("%.0f″", heightFt*12)),
		labelStyle.Render("Weight"),
		valueStyle.Render(fmt.Sprintf("%.0f lb", weightLb)),
		labelStyle.Render("Max HR"),
		valueStyle.Render(fmt.Sprintf("%d bpm", b.MaxHeartRate)),
	)
	return lipgloss.NewStyle().Foreground(gray).MarginTop(1).Render(content)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
