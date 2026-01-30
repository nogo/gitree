package list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

const (
	expandedHeight     = 10 // Fixed height for expanded section
	minTwoColumnWidth  = 100
	minSingleColWidth  = 60
	maxVisibleFiles    = 6
)

// Styles for expanded view
var (
	ExpandedBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62"))

	ExpandedLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	ExpandedValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	ExpandedHashStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	ExpandedMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	FileAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	FileModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	FileDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	FileRenamedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("51"))

	FileSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("237"))

	AdditionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	DeletionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	ExpandedHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// renderExpanded renders the expanded detail section for a commit
// Returns multiple lines that should be inserted after the commit row
func (m Model) renderExpanded(commit *domain.Commit, files []domain.FileChange, fileCursor int, fileScrollOffset int) []string {
	graphWidth := m.graph.Width()
	contentWidth := m.width - graphWidth - 3 // -3 for cursor and spacing

	if contentWidth < 30 {
		contentWidth = 30
	}

	// Determine layout mode
	useTwoColumns := m.width >= minTwoColumnWidth

	var lines []string

	if useTwoColumns {
		lines = m.renderTwoColumnExpanded(commit, files, fileCursor, fileScrollOffset, contentWidth)
	} else {
		lines = m.renderSingleColumnExpanded(commit, files, fileCursor, fileScrollOffset, contentWidth)
	}

	// Prepend graph continuation to each line
	graphCont := m.graph.RenderContinuation(m.cursor)
	var result []string
	for _, line := range lines {
		result = append(result, "  "+graphCont+line)
	}

	return result
}

func (m Model) renderTwoColumnExpanded(commit *domain.Commit, files []domain.FileChange, fileCursor int, fileScrollOffset int, totalWidth int) []string {
	// Split width: 50% metadata, 50% files
	leftWidth := totalWidth / 2
	rightWidth := totalWidth - leftWidth - 1 // -1 for separator

	var lines []string

	// Top border
	lines = append(lines, ExpandedBorderStyle.Render("╔"+strings.Repeat("═", leftWidth)+"╤"+strings.Repeat("═", rightWidth)+"╗"))

	// Content rows
	leftLines := m.renderMetadataColumn(commit, leftWidth-2)
	rightLines := m.renderFilesColumn(files, fileCursor, fileScrollOffset, rightWidth-2)

	// Pad to same height
	maxLines := expandedHeight - 2 // -2 for borders
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	for i := 0; i < maxLines; i++ {
		left := padRight(leftLines[i], leftWidth-2)
		right := padRight(rightLines[i], rightWidth-2)
		lines = append(lines, ExpandedBorderStyle.Render("║")+" "+left+" "+ExpandedBorderStyle.Render("│")+" "+right+" "+ExpandedBorderStyle.Render("║"))
	}

	// Bottom border with help
	help := "[↑/↓] file  [Enter] diff  [Esc] close"
	helpPadded := padCenter(help, totalWidth)
	lines = append(lines, ExpandedBorderStyle.Render("╚")+ExpandedHelpStyle.Render(helpPadded)+ExpandedBorderStyle.Render("╝"))

	return lines
}

func (m Model) renderSingleColumnExpanded(commit *domain.Commit, files []domain.FileChange, fileCursor int, fileScrollOffset int, totalWidth int) []string {
	var lines []string
	innerWidth := totalWidth - 4 // borders + padding

	// Top border
	lines = append(lines, ExpandedBorderStyle.Render("╔"+strings.Repeat("═", totalWidth-2)+"╗"))

	// Metadata (abbreviated)
	hashLine := fmt.Sprintf("%s %s", ExpandedLabelStyle.Render("Commit:"), ExpandedHashStyle.Render(truncateStr(commit.Hash, 12)))
	authorLine := fmt.Sprintf("%s %s", ExpandedLabelStyle.Render("Author:"), ExpandedValueStyle.Render(truncateStr(commit.Author, innerWidth-10)))
	lines = append(lines, m.wrapInBorder(hashLine, totalWidth))
	lines = append(lines, m.wrapInBorder(authorLine, totalWidth))

	// Separator
	lines = append(lines, ExpandedBorderStyle.Render("╟"+strings.Repeat("─", totalWidth-2)+"╢"))

	// Files
	fileLines := m.renderFilesColumn(files, fileCursor, fileScrollOffset, innerWidth)
	for _, fl := range fileLines {
		lines = append(lines, m.wrapInBorder(fl, totalWidth))
	}

	// Pad to height
	for len(lines) < expandedHeight-1 {
		lines = append(lines, m.wrapInBorder("", totalWidth))
	}

	// Bottom with help
	help := "[↑/↓] file  [Enter] diff  [Esc] close"
	helpPadded := padCenter(help, totalWidth-2)
	lines = append(lines, ExpandedBorderStyle.Render("╚")+ExpandedHelpStyle.Render(helpPadded)+ExpandedBorderStyle.Render("╝"))

	return lines
}

func (m Model) renderMetadataColumn(commit *domain.Commit, width int) []string {
	var lines []string

	// Hash
	hashLabel := ExpandedLabelStyle.Render("Commit:")
	hashValue := ExpandedHashStyle.Render(truncateStr(commit.Hash, width-10))
	lines = append(lines, hashLabel+" "+hashValue)

	// Author
	authorLabel := ExpandedLabelStyle.Render("Author:")
	authorValue := ExpandedValueStyle.Render(truncateStr(fmt.Sprintf("%s <%s>", commit.Author, commit.Email), width-10))
	lines = append(lines, authorLabel+" "+authorValue)

	// Date
	dateLabel := ExpandedLabelStyle.Render("Date:")
	dateValue := ExpandedValueStyle.Render(commit.Date.Format("Jan 2, 2006 15:04"))
	lines = append(lines, dateLabel+"   "+dateValue)

	// Parents
	if len(commit.Parents) > 0 {
		parentLabel := ExpandedLabelStyle.Render("Parents:")
		parentHashes := make([]string, len(commit.Parents))
		for i, p := range commit.Parents {
			if len(p) > 7 {
				parentHashes[i] = p[:7]
			} else {
				parentHashes[i] = p
			}
		}
		parentValue := ExpandedValueStyle.Render(strings.Join(parentHashes, ", "))
		lines = append(lines, parentLabel+" "+parentValue)
	}

	// Empty line
	lines = append(lines, "")

	// Message (may span multiple lines)
	msgLines := wrapText(commit.FullMessage, width)
	for i, ml := range msgLines {
		if i >= 3 { // Limit to 3 lines of message
			lines = append(lines, ExpandedMessageStyle.Render("..."))
			break
		}
		lines = append(lines, ExpandedMessageStyle.Render(ml))
	}

	return lines
}

func (m Model) renderFilesColumn(files []domain.FileChange, cursor int, scrollOffset int, width int) []string {
	var lines []string

	if len(files) == 0 {
		lines = append(lines, ExpandedLabelStyle.Render("No files changed"))
		return lines
	}

	// Calculate totals
	totalAdd, totalDel := 0, 0
	for _, f := range files {
		totalAdd += f.Additions
		totalDel += f.Deletions
	}

	// Header
	header := fmt.Sprintf("Files (%d)  %s %s",
		len(files),
		AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdd)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", totalDel)),
	)
	lines = append(lines, header)

	// File list with scrolling
	visibleFiles := maxVisibleFiles
	if len(files) < visibleFiles {
		visibleFiles = len(files)
	}

	// Adjust scroll offset to keep cursor visible
	if cursor < scrollOffset {
		scrollOffset = cursor
	}
	if cursor >= scrollOffset+visibleFiles {
		scrollOffset = cursor - visibleFiles + 1
	}

	endIdx := scrollOffset + visibleFiles
	if endIdx > len(files) {
		endIdx = len(files)
	}

	for i := scrollOffset; i < endIdx; i++ {
		f := files[i]
		selected := i == cursor

		// Status indicator
		var statusStr string
		switch f.Status {
		case domain.FileAdded:
			statusStr = FileAddedStyle.Render("A")
		case domain.FileModified:
			statusStr = FileModifiedStyle.Render("M")
		case domain.FileDeleted:
			statusStr = FileDeletedStyle.Render("D")
		case domain.FileRenamed:
			statusStr = FileRenamedStyle.Render("R")
		default:
			statusStr = " "
		}

		// Cursor
		cursorStr := "  "
		if selected {
			cursorStr = "> "
		}

		// Path (truncated)
		pathWidth := width - 20 // space for cursor, status, stats
		if pathWidth < 10 {
			pathWidth = 10
		}
		path := truncateStr(f.Path, pathWidth)

		// Stats
		stats := fmt.Sprintf("%s %s",
			AdditionsStyle.Render(fmt.Sprintf("+%d", f.Additions)),
			DeletionsStyle.Render(fmt.Sprintf("-%d", f.Deletions)),
		)

		line := fmt.Sprintf("%s%s %s %s", cursorStr, statusStr, padRight(path, pathWidth), stats)
		if selected {
			line = FileSelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	// Scroll indicator
	if len(files) > maxVisibleFiles {
		indicator := fmt.Sprintf("  (%d/%d)", cursor+1, len(files))
		lines = append(lines, ExpandedLabelStyle.Render(indicator))
	}

	return lines
}

func (m Model) wrapInBorder(content string, totalWidth int) string {
	innerWidth := totalWidth - 4
	padded := padRight(content, innerWidth)
	return ExpandedBorderStyle.Render("║") + " " + padded + " " + ExpandedBorderStyle.Render("║")
}

// Helper functions

func padRight(s string, width int) string {
	sLen := displayLen(s)
	if sLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sLen)
}

func padCenter(s string, width int) string {
	sLen := len(s)
	if sLen >= width {
		return s
	}
	leftPad := (width - sLen) / 2
	rightPad := width - sLen - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

func wrapText(s string, width int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			lines = append(lines, "")
			continue
		}

		// Simple word wrap
		for len(line) > width {
			// Find break point
			breakAt := width
			for breakAt > 0 && line[breakAt] != ' ' {
				breakAt--
			}
			if breakAt == 0 {
				breakAt = width
			}
			lines = append(lines, line[:breakAt])
			line = strings.TrimSpace(line[breakAt:])
		}
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

// displayLen returns display width excluding ANSI codes
func displayLen(s string) int {
	count := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}
