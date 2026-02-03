package insights

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatCount formats a number with thousands separators or "k" suffix for large numbers.
func formatCount(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 10000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	if n < 100000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%dk", n/1000)
}

// formatPercent calculates and formats a percentage.
func formatPercent(n, total int) string {
	if total == 0 {
		return "0%"
	}
	pct := (n * 100) / total
	return fmt.Sprintf("%d%%", pct)
}

// renderAuthorStats renders the top authors table.
// Format: Name        123  45%
func renderAuthorStats(stats []AuthorStats, width, height int) string {
	if len(stats) == 0 {
		return ""
	}

	var lines []string

	// Title
	title := SectionTitleStyle.Render("TOP AUTHORS")
	lines = append(lines, title)

	// Calculate total commits for percentage
	totalCommits := 0
	for _, s := range stats {
		totalCommits += s.Commits
	}

	// Calculate column widths
	// Format: Name        123  45%
	// Reserve space for: count (6) + space (1) + percent (4) + margins (2)
	countWidth := 6
	pctWidth := 4
	spacing := 3 // spaces between columns
	nameWidth := width - countWidth - pctWidth - spacing
	if nameWidth < 5 {
		nameWidth = 5
	}

	// Render rows (up to height-2 to account for title and padding)
	maxRows := height - 2
	if maxRows < 1 {
		maxRows = 1
	}

	for i, s := range stats {
		if i >= maxRows {
			break
		}

		// Truncate name to fit
		name := truncateWithEllipsis(s.Name, nameWidth)
		name = padRight(name, nameWidth)

		// Format count and percent
		count := padLeft(formatCount(s.Commits), countWidth)
		pct := padLeft(formatPercent(s.Commits, totalCommits), pctWidth)

		// Apply styles
		styledName := NameStyle.Render(name)
		styledCount := CountStyle.Render(count)
		styledPct := PercentStyle.Render(pct)

		line := styledName + " " + styledCount + " " + styledPct
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderFileStats renders the top files table.
// Format: path/to/file.go   89
func renderFileStats(stats []FileStats, width, height int) string {
	if len(stats) == 0 {
		return ""
	}

	var lines []string

	// Title
	title := SectionTitleStyle.Render("TOP FILES")
	lines = append(lines, title)

	// Calculate column widths
	// Format: path/to/file.go   89
	countWidth := 6
	spacing := 2
	pathWidth := width - countWidth - spacing
	if pathWidth < 10 {
		pathWidth = 10
	}

	// Render rows (up to height-2 to account for title and padding)
	maxRows := height - 2
	if maxRows < 1 {
		maxRows = 1
	}

	for i, s := range stats {
		if i >= maxRows {
			break
		}

		// Truncate path to fit (with leading "..." if needed)
		path := truncatePathWithEllipsis(s.Path, pathWidth)
		path = padRight(path, pathWidth)

		// Format count
		count := padLeft(formatCount(s.ChangeCount), countWidth)

		// Apply styles
		styledPath := NameStyle.Render(path)
		styledCount := CountStyle.Render(count)

		line := styledPath + "  " + styledCount
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderSummary renders the summary line.
// Full format: 847 commits · 12 authors · +12,431/-8,291 lines
// Narrow format: 847 commits · +12k/-8k
func renderSummary(summary Summary, width int) string {
	// Build full format
	commits := fmt.Sprintf("%s commits", formatCount(summary.TotalCommits))
	authors := fmt.Sprintf("%d authors", summary.TotalAuthors)
	additions := "+" + formatCount(summary.TotalAdditions)
	deletions := "-" + formatCount(summary.TotalDeletions)
	lines := additions + "/" + deletions + " lines"

	fullLine := commits + " · " + authors + " · " + lines
	narrowLine := commits + " · " + additions + "/" + deletions

	// Choose format based on width
	var content string
	if len(fullLine) <= width {
		content = fullLine
	} else if len(narrowLine) <= width {
		content = narrowLine
	} else {
		content = truncateWithEllipsis(narrowLine, width)
	}

	// Center-align
	style := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	return style.Render(content)
}

// truncateWithEllipsis truncates a string and adds "..." if it exceeds maxLen.
func truncateWithEllipsis(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// truncatePathWithEllipsis truncates a path from the beginning with "..." prefix.
func truncatePathWithEllipsis(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[len(runes)-maxLen:])
	}
	return "..." + string(runes[len(runes)-maxLen+3:])
}

// padRight pads a string with spaces on the right to reach width.
func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// padLeft pads a string with spaces on the left to reach width.
func padLeft(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(runes)) + s
}
