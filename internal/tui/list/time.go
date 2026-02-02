package list

import (
	"fmt"
	"time"
)

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		// Show year if not current year
		if t.Year() != now.Year() {
			return t.Format("Jan 2 '06")
		}
		return t.Format("Jan 2")
	}
}
