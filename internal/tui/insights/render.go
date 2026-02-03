package insights

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout breakpoint for wide vs narrow display.
const wideLayoutMinWidth = 100

// View renders the complete insights view with stats and heatmap.
func (v InsightsView) View() string {
	if v.width == 0 || v.height == 0 {
		return ""
	}

	// Determine layout mode
	isWide := v.width >= wideLayoutMinWidth

	var content string
	if isWide {
		content = v.renderWideLayout()
	} else {
		content = v.renderNarrowLayout()
	}

	result := content

	// Force exact height to keep footer at bottom
	return lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Render(result)
}

// renderWideLayout renders side-by-side stats and heatmap.
func (v InsightsView) renderWideLayout() string {
	// Stats panel takes 40% of width, heatmap takes 60%
	statsPanelWidth := v.width * 40 / 100
	if statsPanelWidth < 30 {
		statsPanelWidth = 30
	}
	heatmapPanelWidth := v.width - statsPanelWidth - 3 // 3 for gap " │ "

	panelHeight := v.height
	if panelHeight < 10 {
		panelHeight = 10
	}

	statsPanel := v.renderStatsPanel(statsPanelWidth, panelHeight)
	heatmapPanel := v.renderHeatmapPanel(heatmapPanelWidth, panelHeight)

	// Apply fixed widths to panels
	statsPanelStyled := lipgloss.NewStyle().
		Width(statsPanelWidth).
		Height(panelHeight).
		Render(statsPanel)

	heatmapPanelStyled := lipgloss.NewStyle().
		Width(heatmapPanelWidth).
		Height(panelHeight).
		Render(heatmapPanel)

	// Build vertical separator spanning full height
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	var sepLines []string
	for i := 0; i < panelHeight; i++ {
		sepLines = append(sepLines, sepStyle.Render("│"))
	}
	separator := strings.Join(sepLines, "\n")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		statsPanelStyled,
		separator,
		heatmapPanelStyled,
	)
}

// renderNarrowLayout renders stacked stats and heatmap.
func (v InsightsView) renderNarrowLayout() string {
	panelWidth := v.width

	// Split height: stats 60%, heatmap 40%
	availableHeight := v.height
	statsPanelHeight := availableHeight * 60 / 100
	heatmapPanelHeight := availableHeight - statsPanelHeight

	if statsPanelHeight < 8 {
		statsPanelHeight = 8
	}
	if heatmapPanelHeight < 8 {
		heatmapPanelHeight = 8
	}

	statsPanel := v.renderStatsPanel(panelWidth, statsPanelHeight)
	heatmapPanel := v.renderHeatmapPanel(panelWidth, heatmapPanelHeight)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statsPanel,
		"",
		heatmapPanel,
	)
}

// renderStatsPanel renders the combined author and file statistics panel.
func (v InsightsView) renderStatsPanel(width, height int) string {
	// Split height between authors and files, with 1 line gap
	availableHeight := height - 1 // Reserve 1 line for gap
	authorsHeight := availableHeight / 2
	filesHeight := availableHeight - authorsHeight

	authorsSection := v.renderAuthorsSection(width, authorsHeight)
	filesSection := v.renderFilesSection(width, filesHeight)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		authorsSection,
		"", // Gap between sections
		filesSection,
	)
}

// renderAuthorsSection renders the top authors list.
func (v InsightsView) renderAuthorsSection(width, height int) string {
	var lines []string

	// Section title
	title := SectionTitleStyle.Render("Top Authors")
	lines = append(lines, title)

	// Calculate total commits for percentage
	totalCommits := 0
	for _, a := range v.authorStats {
		totalCommits += a.Commits
	}

	// Render author rows (leave room for title)
	maxRows := height - 1
	if maxRows < 1 {
		maxRows = 1
	}

	for i, author := range v.authorStats {
		if i >= maxRows {
			break
		}

		pct := 0.0
		if totalCommits > 0 {
			pct = float64(author.Commits) * 100 / float64(totalCommits)
		}

		// Truncate name to fit
		name := author.Name
		maxNameLen := width - 15 // space for count and percentage
		if maxNameLen < 10 {
			maxNameLen = 10
		}
		if len(name) > maxNameLen {
			name = name[:maxNameLen-1] + "..."
		}

		countStr := CountStyle.Render(fmt.Sprintf("%4d", author.Commits))
		pctStr := PercentStyle.Render(fmt.Sprintf("%5.1f%%", pct))
		nameStr := NameStyle.Render(name)

		line := fmt.Sprintf("%s %s %s", countStr, pctStr, nameStr)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderFilesSection renders the most changed files list.
func (v InsightsView) renderFilesSection(width, height int) string {
	var lines []string

	// Section title
	title := SectionTitleStyle.Render("Most Changed Files")
	lines = append(lines, title)

	// Calculate total changes for percentage
	totalChanges := 0
	for _, f := range v.fileStats {
		totalChanges += f.ChangeCount
	}

	// Render file rows (leave room for title)
	maxRows := height - 1
	if maxRows < 1 {
		maxRows = 1
	}

	for i, file := range v.fileStats {
		if i >= maxRows {
			break
		}

		pct := 0.0
		if totalChanges > 0 {
			pct = float64(file.ChangeCount) * 100 / float64(totalChanges)
		}

		// Truncate path to fit (keep end of path)
		path := file.Path
		maxPathLen := width - 15 // space for count and percentage
		if maxPathLen < 10 {
			maxPathLen = 10
		}
		if len(path) > maxPathLen {
			path = "..." + path[len(path)-maxPathLen+3:]
		}

		countStr := CountStyle.Render(fmt.Sprintf("%4d", file.ChangeCount))
		pctStr := PercentStyle.Render(fmt.Sprintf("%5.1f%%", pct))
		pathStr := NameStyle.Render(path)

		line := fmt.Sprintf("%s %s %s", countStr, pctStr, pathStr)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderHeatmapPanel renders the calendar heatmap with title.
func (v InsightsView) renderHeatmapPanel(width, height int) string {
	var lines []string

	// Section title
	title := SectionTitleStyle.Render("Activity")
	lines = append(lines, title)

	// Render calendar (leave room for title)
	calendarHeight := height - 1
	if calendarHeight < 8 {
		calendarHeight = 8
	}

	calendar := renderCalendar(v.calendar, width, calendarHeight)
	lines = append(lines, calendar)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderSummaryLine renders the bottom summary statistics.
