package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/martin-dimi/claude-coach/internal/db"
)

// Activity colors - 3 intensity levels (low, medium, high)
// Empty days use a shared dim gray, not a dark version of the color
var activityColors = [][]string{
	{"#2ea043", "#3fb950", "#56d364"}, // green
	{"#1f6feb", "#388bfd", "#58a6ff"}, // blue
	{"#a371f7", "#bc8cff", "#d2a8ff"}, // purple
	{"#d29922", "#e3b341", "#f0c960"}, // amber
	{"#39d2c0", "#56e0cf", "#7aebdf"}, // teal
	{"#f47067", "#ff7b72", "#ffa198"}, // coral
}

var emptyColor = lipgloss.Color("#2d333b")

var (
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	statStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

const days = 30

// RenderGrid renders a 30-day contribution strip for a single activity.
func RenderGrid(activity string, colorIdx int, dayStats map[string]db.DayStat, todayReps int, todaySessions int, allTimeReps int, allTimeSessions int) string {
	colors := activityColors[colorIdx%len(activityColors)]
	now := time.Now()

	// Find max value for intensity scaling
	maxVal := 1
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -(days-1)+i).Format("2006-01-02")
		if s, ok := dayStats[day]; ok {
			val := cellValue(s)
			if val > maxVal {
				maxVal = val
			}
		}
	}

	// Build the strip
	var cells []string
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -(days-1)+i).Format("2006-01-02")
		s := dayStats[day]
		val := cellValue(s)

		var cell string
		if val == 0 {
			cell = lipgloss.NewStyle().Foreground(emptyColor).Render("█")
		} else {
			level := intensityLevel(val, maxVal)
			cell = lipgloss.NewStyle().Foreground(lipgloss.Color(colors[level])).Render("█")
		}
		cells = append(cells, cell)
	}

	strip := strings.Join(cells, " ")

	// Stats line
	var todayStr string
	if todayReps > 0 {
		todayStr = fmt.Sprintf("%d today", todayReps)
	} else if todaySessions > 0 {
		todayStr = fmt.Sprintf("%dx today", todaySessions)
	}

	var allTimeStr string
	if allTimeReps > 0 {
		allTimeStr = fmt.Sprintf("%d all time", allTimeReps)
	} else if allTimeSessions > 0 {
		allTimeStr = fmt.Sprintf("%dx all time", allTimeSessions)
	}

	var b strings.Builder
	b.WriteString("  ")
	b.WriteString(labelStyle.Render(activity))
	if todayStr != "" || allTimeStr != "" {
		parts := []string{}
		if todayStr != "" {
			parts = append(parts, todayStr)
		}
		if allTimeStr != "" {
			parts = append(parts, allTimeStr)
		}
		b.WriteString("  ")
		b.WriteString(statStyle.Render(strings.Join(parts, " · ")))
	}
	b.WriteString("\n  ")
	b.WriteString(strip)

	return b.String()
}

// RenderFooter renders the date range and streak.
func RenderFooter(streak, best int) string {
	now := time.Now()
	startDate := now.AddDate(0, 0, -(days-1)).Format("Jan 2")
	endDate := now.Format("Jan 2")

	// The strip is 30 blocks with spaces = 30 + 29 = 59 chars
	stripWidth := days*2 - 1
	gap := stripWidth - len(startDate) - len(endDate)
	if gap < 1 {
		gap = 1
	}

	var b strings.Builder
	b.WriteString("  ")
	b.WriteString(mutedStyle.Render(startDate))
	b.WriteString(strings.Repeat(" ", gap))
	b.WriteString(mutedStyle.Render(endDate))

	b.WriteString("\n\n  ")
	if streak > 0 {
		streakStr := fmt.Sprintf("%d day streak", streak)
		b.WriteString(labelStyle.Render(streakStr))
		if best > streak {
			b.WriteString(statStyle.Render(fmt.Sprintf(" · best: %d days", best)))
		}
	} else {
		b.WriteString(mutedStyle.Render("No streak yet"))
	}

	return b.String()
}

func cellValue(s db.DayStat) int {
	if s.Reps > 0 {
		return s.Reps
	}
	return s.Done
}

func intensityLevel(val, maxVal int) int {
	ratio := float64(val) / float64(maxVal)
	switch {
	case ratio <= 0.33:
		return 0 // low
	case ratio <= 0.66:
		return 1 // medium
	default:
		return 2 // high
	}
}
