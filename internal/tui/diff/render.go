package diff

import (
	"strings"
)

// renderDiff applies syntax coloring to diff output
func renderDiff(diff string) string {
	if diff == "" {
		return InfoStyle.Render("No changes")
	}

	lines := strings.Split(diff, "\n")
	var rendered []string

	for _, line := range lines {
		if len(line) == 0 {
			rendered = append(rendered, "")
			continue
		}

		switch {
		case strings.HasPrefix(line, "@@"):
			// Hunk header
			rendered = append(rendered, DiffHunkStyle.Render(line))
		case strings.HasPrefix(line, "+"):
			// Added line
			rendered = append(rendered, DiffAddedStyle.Render(line))
		case strings.HasPrefix(line, "-"):
			// Deleted line
			rendered = append(rendered, DiffDeletedStyle.Render(line))
		default:
			// Context line
			rendered = append(rendered, DiffContextStyle.Render(line))
		}
	}

	return strings.Join(rendered, "\n")
}
