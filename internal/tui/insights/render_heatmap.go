package insights

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Weekday labels for the calendar (Monday start).
var weekdayLabels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// cellChar returns the block character for a heat level.
func cellChar(level int) string {
	switch level {
	case 0:
		return "░"
	case 1:
		return "▒"
	case 2:
		return "▓"
	case 3:
		return "█"
	case 4:
		return "█"
	default:
		return "░"
	}
}

// renderCalendar renders the calendar heatmap grid.
func renderCalendar(cal CalendarData, width, height int) string {
	if len(cal.Cells) == 0 {
		return "No commit data"
	}

	var lines []string

	// Calculate how many weeks we can display
	weekdayLabelWidth := 4 // "Mon " etc
	availableWidth := width - weekdayLabelWidth
	maxWeeks := availableWidth
	if maxWeeks < 1 {
		maxWeeks = 1
	}

	// Determine start week for viewport (show most recent weeks)
	startWeek := 0
	if len(cal.Cells) > maxWeeks {
		startWeek = len(cal.Cells) - maxWeeks
	}
	visibleWeeks := cal.Cells[startWeek:]

	// Render month labels
	monthLine := renderMonthLabels(visibleWeeks, weekdayLabelWidth)
	lines = append(lines, monthLine)

	// Render each day row (7 rows for days of week)
	for dayIdx := 0; dayIdx < 7; dayIdx++ {
		var row strings.Builder

		// Weekday label (show Mon, Wed, Fri only for compactness)
		if dayIdx%2 == 0 {
			row.WriteString(weekdayLabels[dayIdx][:1] + "  ")
		} else {
			row.WriteString("   ")
		}

		// Cells for this day across all weeks
		for _, week := range visibleWeeks {
			if dayIdx < len(week) {
				cell := week[dayIdx]
				char := cellChar(cell.Level)
				style := HeatStyles[cell.Level]
				row.WriteString(style.Render(char))
			} else {
				row.WriteString(" ")
			}
		}

		lines = append(lines, row.String())
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderMonthLabels creates the month header row for visible weeks.
func renderMonthLabels(weeks [][]DayCell, labelWidth int) string {
	if len(weeks) == 0 {
		return ""
	}

	var b strings.Builder

	// Prefix space for weekday label column
	b.WriteString(strings.Repeat(" ", labelWidth))

	lastMonth := time.Month(0)

	for _, week := range weeks {
		if len(week) == 0 {
			b.WriteString(" ")
			continue
		}

		// Use first day of week to determine month
		firstDay := week[0]
		currentMonth := firstDay.Date.Month()

		if currentMonth != lastMonth {
			// New month - write first letter
			monthName := firstDay.Date.Format("Jan")
			b.WriteString(monthName[:1])
			lastMonth = currentMonth
		} else {
			b.WriteString(" ")
		}
	}

	return b.String()
}
