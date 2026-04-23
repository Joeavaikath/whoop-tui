package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jvaikath/whoop-tui/internal/api"
)

func formatDuration(millis int) string {
	d := time.Duration(millis) * time.Millisecond
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func strainBar(strain float64, width int) string {
	filled := int((strain / 21.0) * float64(width))
	if filled > width {
		filled = width
	}
	color := strainColor(strain)
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(dimmed).Render(strings.Repeat("░", width-filled))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// --- Today view ---

func (m model) cardWidth() int {
	// outer container has Padding(1, 2) = 4 horizontal
	// each card: Width(w) renders as w + 2 (border) + 1 (margin-right) = w + 3
	// 3 cards = 3*(w+3) = 3w + 9
	// available = m.width - 4
	// so w = (m.width - 4 - 9) / 3 = (m.width - 13) / 3
	w := (m.width - 13) / 3
	if w < 18 {
		w = 18
	}
	return w
}

func (m model) viewToday() string {
	var cards []string
	cards = append(cards, m.recoveryCard())
	cards = append(cards, m.strainCard())
	cards = append(cards, m.sleepCard())
	row := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	var sections []string
	sections = append(sections, row)

	if len(m.workouts) > 0 {
		sections = append(sections, m.workoutsCard())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) recoveryCard() string {
	w := m.cardWidth()
	var lines []string
	lines = append(lines, headerStyle.Render("Recovery"))

	var r *api.RecoveryScore
	for i := range m.recoveries {
		if m.recoveries[i].Score != nil {
			r = m.recoveries[i].Score
			break
		}
	}

	if r == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	color := recoveryColor(r.RecoveryScore)
	lines = append(lines, bigNumberStyle.Foreground(color).Render(fmt.Sprintf("%.0f%%", r.RecoveryScore)))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("HRV     ")+valueStyle.Render(fmt.Sprintf("%.0f ms", r.HrvRmssdMilli)))
	lines = append(lines, labelStyle.Render("RHR     ")+valueStyle.Render(fmt.Sprintf("%.0f bpm", r.RestingHeartRate)))
	if r.Spo2Percentage != nil {
		lines = append(lines, labelStyle.Render("SpO2    ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *r.Spo2Percentage)))
	}
	if r.SkinTempCelsius != nil {
		lines = append(lines, labelStyle.Render("Temp    ")+valueStyle.Render(fmt.Sprintf("%.1f°C", *r.SkinTempCelsius)))
	}

	return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) strainCard() string {
	w := m.cardWidth()
	barWidth := w - 8
	if barWidth < 10 {
		barWidth = 10
	}

	var lines []string
	lines = append(lines, headerStyle.Render("Strain"))

	var score *api.CycleScore
	for i := range m.cycles {
		if m.cycles[i].Score != nil {
			score = m.cycles[i].Score
			break
		}
	}

	if score == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	color := strainColor(score.Strain)
	lines = append(lines, bigNumberStyle.Foreground(color).Render(fmt.Sprintf("%.1f", score.Strain)))
	lines = append(lines, "")
	lines = append(lines, strainBar(score.Strain, barWidth))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Avg HR  ")+valueStyle.Render(fmt.Sprintf("%d bpm", score.AverageHeartRate)))
	lines = append(lines, labelStyle.Render("Max HR  ")+valueStyle.Render(fmt.Sprintf("%d bpm", score.MaxHeartRate)))
	lines = append(lines, labelStyle.Render("Cal     ")+valueStyle.Render(fmt.Sprintf("%.0f kcal", score.Kilojoule*0.239006)))

	return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) sleepCard() string {
	w := m.cardWidth()
	var lines []string
	lines = append(lines, headerStyle.Render("Sleep"))

	var score *api.SleepScore
	for i := range m.sleeps {
		if !m.sleeps[i].Nap && m.sleeps[i].Score != nil {
			score = m.sleeps[i].Score
			break
		}
	}

	if score == nil {
		lines = append(lines, labelStyle.Render("No data"))
		return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	stages := score.StageSummary
	totalSleep := stages.TotalLightSleepTimeMilli +
		stages.TotalSlowWaveSleepTimeMilli +
		stages.TotalRemSleepTimeMilli

	lines = append(lines, bigNumberStyle.Foreground(white).Render(formatDuration(totalSleep)))
	lines = append(lines, "")

	if score.SleepPerformancePercentage != nil {
		c := recoveryColor(*score.SleepPerformancePercentage)
		lines = append(lines, labelStyle.Render("Perf    ")+lipgloss.NewStyle().Bold(true).Foreground(c).Render(fmt.Sprintf("%.0f%%", *score.SleepPerformancePercentage)))
	}
	if score.SleepEfficiencyPercentage != nil {
		lines = append(lines, labelStyle.Render("Eff     ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *score.SleepEfficiencyPercentage)))
	}
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("REM     ")+valueStyle.Render(formatDuration(stages.TotalRemSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Deep    ")+valueStyle.Render(formatDuration(stages.TotalSlowWaveSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Light   ")+valueStyle.Render(formatDuration(stages.TotalLightSleepTimeMilli)))
	lines = append(lines, labelStyle.Render("Awake   ")+valueStyle.Render(formatDuration(stages.TotalAwakeTimeMilli)))

	return cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) workoutsCard() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Recent Workouts"))

	for _, w := range m.workouts {
		dur := w.End.Sub(w.Start)
		h := int(dur.Hours())
		mins := int(dur.Minutes()) % 60

		strain := "  —"
		if w.Score != nil {
			strain = fmt.Sprintf("%4.1f", w.Score.Strain)
		}

		timeStr := fmt.Sprintf("%dm", mins)
		if h > 0 {
			timeStr = fmt.Sprintf("%dh%dm", h, mins)
		}

		date := w.Start.Format("Jan 2")
		line := fmt.Sprintf("%-14s %6s  strain %s  %s",
			truncate(w.SportName, 14), timeStr, strain, date)
		lines = append(lines, labelStyle.Render(line))
	}

	return cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// --- Trend view (Week / 30d / 60d) ---

func (m model) viewTrend() string {
	if len(m.dayEntries) == 0 {
		return labelStyle.Render("No data for this period.")
	}

	var sections []string

	// charts
	sections = append(sections, m.trendCharts())
	sections = append(sections, "")

	// day list
	sections = append(sections, m.dayList())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

type chartSeries struct {
	values []float64
	dates  []time.Time
}

func (m model) buildChartData() (sharedDates []time.Time, recovery, strain, sleep chartSeries) {
	// build a shared date axis from all day entries (chronological)
	for i := len(m.dayEntries) - 1; i >= 0; i-- {
		e := m.dayEntries[i]
		sharedDates = append(sharedDates, e.date)

		if e.recovery != nil && e.recovery.Score != nil {
			recovery.values = append(recovery.values, e.recovery.Score.RecoveryScore)
		} else {
			recovery.values = append(recovery.values, math.NaN())
		}

		if e.cycle != nil && e.cycle.Score != nil {
			strain.values = append(strain.values, e.cycle.Score.Strain)
		} else {
			strain.values = append(strain.values, math.NaN())
		}

		if e.sleep != nil && e.sleep.Score != nil {
			stages := e.sleep.Score.StageSummary
			total := stages.TotalLightSleepTimeMilli +
				stages.TotalSlowWaveSleepTimeMilli +
				stages.TotalRemSleepTimeMilli
			sleep.values = append(sleep.values, float64(total)/3600000.0)
		} else {
			sleep.values = append(sleep.values, math.NaN())
		}
	}

	recovery.dates = sharedDates
	strain.dates = sharedDates
	sleep.dates = sharedDates
	return
}

func (m model) trendCharts() string {
	_, recovery, strain, sleep := m.buildChartData()

	chartWidth := m.width - 16
	if chartWidth < 30 {
		chartWidth = 30
	}
	chartHeight := 8

	var charts []string

	if len(recovery.values) > 0 {
		c := chart{
			width:  chartWidth,
			height: chartHeight,
			data:   recovery.values,
			dates:  recovery.dates,
			minY:   0,
			maxY:   100,
			title:  "Recovery %",
			color:  green,
			colorFunc: func(v float64) lipgloss.Color {
				// smooth gradient: red(0) -> yellow(50) -> green(100)
				t := v / 100.0
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				var r, g, b int
				if t < 0.5 {
					// red to yellow
					s := t * 2
					r = 231
					g = int(76 + s*169) // 76 -> 245
					b = int(60 - s*20)  // 60 -> 40
				} else {
					// yellow to green
					s := (t - 0.5) * 2
					r = int(231 - s*209) // 231 -> 22
					g = int(245 - s*46)  // 245 -> 199
					b = int(40 + s*114)  // 40 -> 154
				}
				return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
			},
			yFormatter: func(v float64) string {
				return fmt.Sprintf("%.0f%%", v)
			},
		}
		charts = append(charts, c.render())
	}

	if len(strain.values) > 0 {
		c := chart{
			width:  chartWidth,
			height: chartHeight,
			data:   strain.values,
			dates:  strain.dates,
			minY:   0,
			maxY:   21,
			title:  "Strain",
			color:  blue,
			colorFunc: func(v float64) lipgloss.Color {
				// light blue to dark blue based on strain intensity
				t := v / 21.0
				r := int(100 - t*70)
				g := int(180 - t*100)
				b := int(255 - t*50)
				return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
			},
			yFormatter: func(v float64) string {
				return fmt.Sprintf("%.1f", v)
			},
		}
		charts = append(charts, c.render())
	}

	if len(sleep.values) > 0 {
		maxSleep := 0.0
		for _, v := range sleep.values {
			if v > maxSleep {
				maxSleep = v
			}
		}
		maxSleep = math.Ceil(maxSleep + 1)

		c := chart{
			width:  chartWidth,
			height: chartHeight,
			data:   sleep.values,
			dates:  sleep.dates,
			minY:   0,
			maxY:   maxSleep,
			title:  "Sleep",
			color:  green,
			colorFunc: func(v float64) lipgloss.Color {
				// smooth gradient: red(<5) -> yellow(6.5) -> green(>=8)
				t := (v - 5.0) / 3.0 // 0 at 5h, 1 at 8h
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				var r, g, b int
				if t < 0.5 {
					s := t * 2
					r = 231
					g = int(76 + s*169)
					b = int(60 - s*20)
				} else {
					s := (t - 0.5) * 2
					r = int(231 - s*209)
					g = int(245 - s*46)
					b = int(40 + s*114)
				}
				return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
			},
			yFormatter: func(v float64) string {
				return fmt.Sprintf("%.0fh", v)
			},
		}
		charts = append(charts, c.render())
	}

	return lipgloss.JoinVertical(lipgloss.Left, charts...)
}

func (m model) dayList() string {
	var lines []string
	lines = append(lines, headerStyle.Render(fmt.Sprintf("  %-12s  %8s  %8s  %8s  %s",
		"Date", "Recovery", "Strain", "Sleep", "Workouts")))

	// scrolling window — charts take ~35 lines, header/tabs/help ~6, so day list gets the rest
	visibleRows := m.height - 42
	if visibleRows < 5 {
		visibleRows = 5
	}
	if visibleRows > len(m.dayEntries) {
		visibleRows = len(m.dayEntries)
	}

	// scroll offset to keep cursor visible
	scrollStart := 0
	if m.cursor >= visibleRows {
		scrollStart = m.cursor - visibleRows + 1
	}
	scrollEnd := scrollStart + visibleRows
	if scrollEnd > len(m.dayEntries) {
		scrollEnd = len(m.dayEntries)
		scrollStart = scrollEnd - visibleRows
		if scrollStart < 0 {
			scrollStart = 0
		}
	}

	for i := scrollStart; i < scrollEnd; i++ {
		e := m.dayEntries[i]
		date := e.date.Format("Mon Jan 2")

		rec := "    —   "
		if e.recovery != nil && e.recovery.Score != nil {
			s := e.recovery.Score.RecoveryScore
			c := recoveryColor(s)
			rec = lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("  %5.0f%%", s))
		}

		str := "    —   "
		if e.cycle != nil && e.cycle.Score != nil {
			s := e.cycle.Score.Strain
			c := strainColor(s)
			str = lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("  %5.1f ", s))
		}

		slp := "    —   "
		if e.sleep != nil && e.sleep.Score != nil {
			stages := e.sleep.Score.StageSummary
			total := stages.TotalLightSleepTimeMilli +
				stages.TotalSlowWaveSleepTimeMilli +
				stages.TotalRemSleepTimeMilli
			slp = fmt.Sprintf("  %6s", formatDuration(total))
		}

		workoutStr := ""
		for _, w := range e.workouts {
			if workoutStr != "" {
				workoutStr += ", "
			}
			workoutStr += truncate(w.SportName, 12)
		}
		if workoutStr == "" {
			workoutStr = "—"
		}

		prefix := "  "
		line := fmt.Sprintf("%s%-12s  %s  %s  %s  %s", prefix, date, rec, str, slp, workoutStr)

		if i == m.cursor {
			line = selectedRowStyle.Render("▸ " + line[2:])
		} else {
			line = labelStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// scroll indicator
	if len(m.dayEntries) > visibleRows {
		indicator := fmt.Sprintf("  showing %d-%d of %d", scrollStart+1, scrollEnd, len(m.dayEntries))
		lines = append(lines, labelStyle.Render(indicator))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// --- Detail view (drill-down) ---

func (m model) viewDetail(e dayEntry) string {
	w := m.cardWidth()
	barWidth := w - 8
	if barWidth < 10 {
		barWidth = 10
	}

	var sections []string

	dateStr := e.date.Format("Monday, January 2, 2006")
	sections = append(sections, headerStyle.Render("◀ "+dateStr))
	sections = append(sections, "")

	var cards []string

	// Recovery card
	{
		var lines []string
		lines = append(lines, headerStyle.Render("Recovery"))
		if e.recovery != nil && e.recovery.Score != nil {
			s := e.recovery.Score
			c := recoveryColor(s.RecoveryScore)
			lines = append(lines, bigNumberStyle.Foreground(c).Render(fmt.Sprintf("%.0f%%", s.RecoveryScore)))
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("HRV       ")+valueStyle.Render(fmt.Sprintf("%.0f ms", s.HrvRmssdMilli)))
			lines = append(lines, labelStyle.Render("RHR       ")+valueStyle.Render(fmt.Sprintf("%.0f bpm", s.RestingHeartRate)))
			if s.Spo2Percentage != nil {
				lines = append(lines, labelStyle.Render("SpO2      ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *s.Spo2Percentage)))
			}
			if s.SkinTempCelsius != nil {
				lines = append(lines, labelStyle.Render("Skin Temp ")+valueStyle.Render(fmt.Sprintf("%.1f°C", *s.SkinTempCelsius)))
			}
		} else {
			lines = append(lines, labelStyle.Render("No data"))
		}
		cards = append(cards, cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
	}

	// Strain card
	{
		var lines []string
		lines = append(lines, headerStyle.Render("Strain"))
		if e.cycle != nil && e.cycle.Score != nil {
			s := e.cycle.Score
			c := strainColor(s.Strain)
			lines = append(lines, bigNumberStyle.Foreground(c).Render(fmt.Sprintf("%.1f", s.Strain)))
			lines = append(lines, "")
			lines = append(lines, strainBar(s.Strain, barWidth))
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("Avg HR    ")+valueStyle.Render(fmt.Sprintf("%d bpm", s.AverageHeartRate)))
			lines = append(lines, labelStyle.Render("Max HR    ")+valueStyle.Render(fmt.Sprintf("%d bpm", s.MaxHeartRate)))
			lines = append(lines, labelStyle.Render("Calories  ")+valueStyle.Render(fmt.Sprintf("%.0f kcal", s.Kilojoule*0.239006)))
		} else {
			lines = append(lines, labelStyle.Render("No data"))
		}
		cards = append(cards, cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
	}

	// Sleep card
	{
		var lines []string
		lines = append(lines, headerStyle.Render("Sleep"))
		if e.sleep != nil && e.sleep.Score != nil {
			sc := e.sleep.Score
			stages := sc.StageSummary
			totalSleep := stages.TotalLightSleepTimeMilli +
				stages.TotalSlowWaveSleepTimeMilli +
				stages.TotalRemSleepTimeMilli

			lines = append(lines, bigNumberStyle.Foreground(white).Render(formatDuration(totalSleep)))
			lines = append(lines, "")

			if sc.SleepPerformancePercentage != nil {
				c := recoveryColor(*sc.SleepPerformancePercentage)
				lines = append(lines, labelStyle.Render("Perf      ")+lipgloss.NewStyle().Bold(true).Foreground(c).Render(fmt.Sprintf("%.0f%%", *sc.SleepPerformancePercentage)))
			}
			if sc.SleepEfficiencyPercentage != nil {
				lines = append(lines, labelStyle.Render("Efficiency")+valueStyle.Render(fmt.Sprintf(" %.0f%%", *sc.SleepEfficiencyPercentage)))
			}
			if sc.SleepConsistencyPercentage != nil {
				lines = append(lines, labelStyle.Render("Consist.  ")+valueStyle.Render(fmt.Sprintf("%.0f%%", *sc.SleepConsistencyPercentage)))
			}
			if sc.RespiratoryRate != nil {
				lines = append(lines, labelStyle.Render("Resp Rate ")+valueStyle.Render(fmt.Sprintf("%.1f /min", *sc.RespiratoryRate)))
			}

			lines = append(lines, "")

			// sleep stage bars
			total := float64(stages.TotalInBedTimeMilli)
			if total > 0 {
				remPct := float64(stages.TotalRemSleepTimeMilli) / total
				deepPct := float64(stages.TotalSlowWaveSleepTimeMilli) / total
				lightPct := float64(stages.TotalLightSleepTimeMilli) / total
				awakePct := float64(stages.TotalAwakeTimeMilli) / total

				remBar := int(remPct * float64(barWidth))
				deepBar := int(deepPct * float64(barWidth))
				lightBar := int(lightPct * float64(barWidth))
				awakeBar := barWidth - remBar - deepBar - lightBar

				bar := lipgloss.NewStyle().Foreground(lipgloss.Color("#9b59b6")).Render(strings.Repeat("█", remBar)) +
					lipgloss.NewStyle().Foreground(blue).Render(strings.Repeat("█", deepBar)) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#5dade2")).Render(strings.Repeat("█", lightBar)) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#e74c3c")).Render(strings.Repeat("█", awakeBar))
				lines = append(lines, bar)
				lines = append(lines, "")

				lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#9b59b6")).Render("■")+labelStyle.Render(fmt.Sprintf(" REM   %s (%.0f%%)", formatDuration(stages.TotalRemSleepTimeMilli), remPct*100)))
				lines = append(lines, lipgloss.NewStyle().Foreground(blue).Render("■")+labelStyle.Render(fmt.Sprintf(" Deep  %s (%.0f%%)", formatDuration(stages.TotalSlowWaveSleepTimeMilli), deepPct*100)))
				lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#5dade2")).Render("■")+labelStyle.Render(fmt.Sprintf(" Light %s (%.0f%%)", formatDuration(stages.TotalLightSleepTimeMilli), lightPct*100)))
				lines = append(lines, lipgloss.NewStyle().Foreground(red).Render("■")+labelStyle.Render(fmt.Sprintf(" Awake %s (%.0f%%)", formatDuration(stages.TotalAwakeTimeMilli), awakePct*100)))
			}

			// sleep need
			need := sc.SleepNeeded
			totalNeed := need.BaselineMilli + need.NeedFromSleepDebtMilli + need.NeedFromRecentStrainMilli + need.NeedFromRecentNapMilli
			if totalNeed > 0 {
				lines = append(lines, "")
				lines = append(lines, labelStyle.Render(fmt.Sprintf("Need: %s (base %s + debt %s)",
					formatDuration(int(totalNeed)),
					formatDuration(int(need.BaselineMilli)),
					formatDuration(int(need.NeedFromSleepDebtMilli)))))
			}
		} else {
			lines = append(lines, labelStyle.Render("No data"))
		}
		cards = append(cards, cardStyle.Width(w).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
	}

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top, cards...))

	// Workouts
	if len(e.workouts) > 0 {
		var wLines []string
		wLines = append(wLines, headerStyle.Render("Workouts"))
		for _, w := range e.workouts {
			dur := w.End.Sub(w.Start)
			h := int(dur.Hours())
			mins := int(dur.Minutes()) % 60
			timeStr := fmt.Sprintf("%dm", mins)
			if h > 0 {
				timeStr = fmt.Sprintf("%dh %dm", h, mins)
			}

			wLines = append(wLines, valueStyle.Render(w.SportName)+" "+labelStyle.Render(timeStr))

			if w.Score != nil {
				s := w.Score
				sc := strainColor(s.Strain)
				wLines = append(wLines, fmt.Sprintf("  %s %s  %s %s  %s %s",
					labelStyle.Render("Strain"),
					lipgloss.NewStyle().Foreground(sc).Render(fmt.Sprintf("%.1f", s.Strain)),
					labelStyle.Render("Avg HR"),
					valueStyle.Render(fmt.Sprintf("%d", s.AverageHeartRate)),
					labelStyle.Render("Max HR"),
					valueStyle.Render(fmt.Sprintf("%d", s.MaxHeartRate)),
				))

				// HR zones
				z := s.ZoneDurations
				totalZone := z.ZoneZeroMilli + z.ZoneOneMilli + z.ZoneTwoMilli + z.ZoneThreeMilli + z.ZoneFourMilli + z.ZoneFiveMilli
				if totalZone > 0 {
					zoneBar := func(label string, ms int64, color lipgloss.Color) string {
						pct := float64(ms) / float64(totalZone) * 100
						barLen := int(pct / 100 * 20)
						return fmt.Sprintf("  %s %s %s",
							labelStyle.Render(label),
							lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", barLen)),
							labelStyle.Render(fmt.Sprintf("%.0f%%", pct)))
					}
					wLines = append(wLines, zoneBar("Z1", z.ZoneOneMilli, lipgloss.Color("#3498db")))
					wLines = append(wLines, zoneBar("Z2", z.ZoneTwoMilli, lipgloss.Color("#2ecc71")))
					wLines = append(wLines, zoneBar("Z3", z.ZoneThreeMilli, lipgloss.Color("#f5c542")))
					wLines = append(wLines, zoneBar("Z4", z.ZoneFourMilli, lipgloss.Color("#e67e22")))
					wLines = append(wLines, zoneBar("Z5", z.ZoneFiveMilli, lipgloss.Color("#e74c3c")))
				}

				if s.DistanceMeter != nil && *s.DistanceMeter > 0 {
					miles := *s.DistanceMeter * 0.000621371
					wLines = append(wLines, "  "+labelStyle.Render("Distance  ")+valueStyle.Render(fmt.Sprintf("%.1f mi", miles)))
				}
			}
			wLines = append(wLines, "")
		}
		sections = append(sections, cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, wLines...)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
