package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var blocks = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

type chart struct {
	width        int
	height       int
	data         []float64
	dates        []time.Time
	minY         float64
	maxY         float64
	title        string
	color        lipgloss.Color
	colorFunc    func(float64) lipgloss.Color
	yFormatter   func(float64) string
	highlightCol int // -1 = none
}

func (c chart) render() string {
	validCount := 0
	for _, v := range c.data {
		if !math.IsNaN(v) {
			validCount++
		}
	}
	if validCount == 0 {
		return ""
	}

	// auto-tighten bounds
	first := true
	dataMin, dataMax := 0.0, 0.0
	for _, v := range c.data {
		if math.IsNaN(v) {
			continue
		}
		if first {
			dataMin, dataMax = v, v
			first = false
		} else {
			if v < dataMin {
				dataMin = v
			}
			if v > dataMax {
				dataMax = v
			}
		}
	}
	padding := (dataMax - dataMin) * 0.1
	if padding < 1 {
		padding = 1
	}
	if dataMin-padding > c.minY {
		c.minY = math.Floor(dataMin - padding)
	}
	if dataMax+padding < c.maxY {
		c.maxY = math.Ceil(dataMax + padding)
	}
	if c.minY < 0 {
		c.minY = 0
	}

	dataRange := c.maxY - c.minY
	if dataRange == 0 {
		dataRange = 1
	}

	// each data point gets a column; figure out column width
	numBars := len(c.data)
	barArea := c.width
	barW := barArea / numBars
	if barW < 1 {
		barW = 1
	}
	gap := 0
	if barW >= 3 {
		gap = 1
		barW--
	}

	// sub-rows: each character row has 8 levels (blocks array)
	subRows := c.height * 8

	// build the bar heights in sub-row units
	barHeights := make([]int, numBars)
	for i, v := range c.data {
		if math.IsNaN(v) {
			barHeights[i] = -1
			continue
		}
		normalized := (v - c.minY) / dataRange
		if normalized < 0 {
			normalized = 0
		}
		if normalized > 1 {
			normalized = 1
		}
		barHeights[i] = int(math.Round(normalized * float64(subRows)))
	}

	// Y axis labels
	yLabelWidth := 6
	numYLabels := 4
	if c.height < 4 {
		numYLabels = c.height
	}

	yLabels := make(map[int]string)
	for i := 0; i < numYLabels; i++ {
		row := int(math.Round(float64(i) / float64(numYLabels-1) * float64(c.height-1)))
		val := c.maxY - float64(row)/float64(c.height-1)*dataRange
		if c.yFormatter != nil {
			yLabels[row] = c.yFormatter(val)
		} else {
			yLabels[row] = fmt.Sprintf("%5.1f", val)
		}
	}

	gridColor := lipgloss.Color("#333333")
	var lines []string

	// title
	lines = append(lines, strings.Repeat(" ", yLabelWidth+1)+
		lipgloss.NewStyle().Bold(true).Foreground(c.color).Render(c.title))

	// render rows top to bottom
	for row := 0; row < c.height; row++ {
		// Y axis
		label := strings.Repeat(" ", yLabelWidth)
		hasLabel := false
		if l, ok := yLabels[row]; ok {
			label = fmt.Sprintf("%*s", yLabelWidth, l)
			hasLabel = true
		}
		axis := "│"
		if hasLabel {
			axis = "┤"
		}

		// the sub-row range for this character row
		// row 0 = top = highest values
		rowBottom := (c.height - 1 - row) * 8
		rowTop := rowBottom + 8

		gridDot := lipgloss.NewStyle().Foreground(gridColor).Render("╌")

		highlightBg := lipgloss.Color("#2a2a3e")

		var barChars []string
		for i := 0; i < numBars; i++ {
			h := barHeights[i]
			isHL := c.highlightCol >= 0 && i == c.highlightCol

			barColor := c.color
			if c.colorFunc != nil && !math.IsNaN(c.data[i]) {
				barColor = c.colorFunc(c.data[i])
			}

			var cell string
			if h < 0 {
				if isHL {
					cell = lipgloss.NewStyle().Background(highlightBg).Render(strings.Repeat(" ", barW))
				} else if hasLabel {
					cell = strings.Repeat(gridDot, barW)
				} else {
					cell = strings.Repeat(" ", barW)
				}
			} else if h >= rowTop {
				s := lipgloss.NewStyle().Foreground(barColor)
				if isHL {
					s = s.Background(highlightBg)
				}
				cell = s.Render(strings.Repeat("█", barW))
			} else if h <= rowBottom {
				if isHL {
					cell = lipgloss.NewStyle().Background(highlightBg).Render(strings.Repeat(" ", barW))
				} else if hasLabel {
					cell = strings.Repeat(gridDot, barW)
				} else {
					cell = strings.Repeat(" ", barW)
				}
			} else {
				fill := h - rowBottom
				blockIdx := fill
				if blockIdx > 8 {
					blockIdx = 8
				}
				s := lipgloss.NewStyle().Foreground(barColor)
				if isHL {
					s = s.Background(highlightBg)
				}
				cell = s.Render(strings.Repeat(blocks[blockIdx], barW))
			}

			barChars = append(barChars, cell)
			if gap > 0 {
				if hasLabel {
					barChars = append(barChars, gridDot)
				} else {
					barChars = append(barChars, " ")
				}
			}
		}

		chartLine := strings.Join(barChars, "")
		lines = append(lines, labelStyle.Render(label+axis)+chartLine)
	}

	// X axis
	pad := strings.Repeat(" ", yLabelWidth)
	totalBarWidth := numBars * (barW + gap)
	axisRunes := []rune(strings.Repeat("─", totalBarWidth))
	if c.highlightCol >= 0 && c.highlightCol < numBars {
		pos := c.highlightCol*(barW+gap) + barW/2
		if pos >= 0 && pos < len(axisRunes) {
			axisRunes[pos] = '┴'
		}
	}
	lines = append(lines, labelStyle.Render(pad+"└"+string(axisRunes)))

	// date labels
	if len(c.dates) > 0 {
		dateLine := make([]byte, totalBarWidth)
		for i := range dateLine {
			dateLine[i] = ' '
		}

		numXLabels := 5
		if numBars <= 7 {
			numXLabels = numBars
		}
		if numXLabels > totalBarWidth/7 {
			numXLabels = totalBarWidth / 7
		}
		if numXLabels < 2 {
			numXLabels = 2
		}

		for i := 0; i < numXLabels; i++ {
			dataIdx := int(math.Round(float64(i) / float64(numXLabels-1) * float64(numBars-1)))
			// center of this bar's position
			colCenter := dataIdx*(barW+gap) + barW/2

			var label string
			if numBars <= 14 {
				label = c.dates[dataIdx].Format("Jan 2")
			} else {
				label = c.dates[dataIdx].Format("1/2")
			}

			start := colCenter - len(label)/2
			if start < 0 {
				start = 0
			}
			if start+len(label) > len(dateLine) {
				start = len(dateLine) - len(label)
			}
			if start < 0 {
				continue
			}

			copy(dateLine[start:], []byte(label))
		}

		// highlight marker
		if c.highlightCol >= 0 && c.highlightCol < numBars {
			markerLine := make([]byte, totalBarWidth)
			for j := range markerLine {
				markerLine[j] = ' '
			}
			pos := c.highlightCol*(barW+gap) + barW/2
			if pos >= 0 && pos < len(markerLine) {
				markerLine[pos] = '^'
			}
			markerStr := string(markerLine)
			markerStr = strings.Replace(markerStr, "^",
				lipgloss.NewStyle().Foreground(white).Bold(true).Render("▲"), 1)
			lines = append(lines, pad+" "+markerStr)
		}

		lines = append(lines, labelStyle.Render(pad+" "+string(dateLine)))
	}

	return strings.Join(lines, "\n")
}
