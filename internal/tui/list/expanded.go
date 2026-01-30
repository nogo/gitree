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
	boxWidth := m.width - graphWidth - 2

	if boxWidth < 30 {
		boxWidth = 30
	}

	graphCont := m.graph.RenderContinuation(m.cursor)
	innerWidth := boxWidth - 4 // space inside borders (║ + space + content + space + ║)

	// Build content lines
	var contentLines []string

	// Commit info
	if commit != nil {
		contentLines = append(contentLines, fmt.Sprintf("Commit: %s", ExpandedHashStyle.Render(commit.Hash[:12])))
		contentLines = append(contentLines, fmt.Sprintf("Author: %s <%s>", commit.Author, commit.Email))
		contentLines = append(contentLines, fmt.Sprintf("Date:   %s", commit.Date.Format("Jan 2, 2006 15:04")))
		if len(commit.Parents) > 0 {
			parentShort := commit.Parents[0]
			if len(parentShort) > 7 {
				parentShort = parentShort[:7]
			}
			contentLines = append(contentLines, fmt.Sprintf("Parent: %s", parentShort))
		}
		contentLines = append(contentLines, "") // blank line
	}

	// Files section
	if len(files) > 0 {
		// Calculate totals
		totalAdd, totalDel := 0, 0
		for _, f := range files {
			totalAdd += f.Additions
			totalDel += f.Deletions
		}
		contentLines = append(contentLines, fmt.Sprintf("Files (%d)  %s %s",
			len(files),
			AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdd)),
			DeletionsStyle.Render(fmt.Sprintf("-%d", totalDel))))

		// Show files with cursor
		visibleFiles := maxVisibleFiles
		if len(files) < visibleFiles {
			visibleFiles = len(files)
		}
		// Adjust scroll
		if fileCursor < fileScrollOffset {
			fileScrollOffset = fileCursor
		}
		if fileCursor >= fileScrollOffset+visibleFiles {
			fileScrollOffset = fileCursor - visibleFiles + 1
		}
		endIdx := fileScrollOffset + visibleFiles
		if endIdx > len(files) {
			endIdx = len(files)
		}

		for i := fileScrollOffset; i < endIdx; i++ {
			f := files[i]
			cursor := "  "
			if i == fileCursor {
				cursor = "> "
			}
			status := "M"
			statusStyle := FileModifiedStyle
			switch f.Status {
			case domain.FileAdded:
				status = "A"
				statusStyle = FileAddedStyle
			case domain.FileDeleted:
				status = "D"
				statusStyle = FileDeletedStyle
			case domain.FileRenamed:
				status = "R"
				statusStyle = FileRenamedStyle
			}
			// Truncate path to fit
			maxPath := innerWidth - 20
			if maxPath < 10 {
				maxPath = 10
			}
			path := f.Path
			if len(path) > maxPath {
				path = path[:maxPath-1] + "…"
			}
			fileLine := fmt.Sprintf("%s%s %s %s %s",
				cursor,
				statusStyle.Render(status),
				path,
				AdditionsStyle.Render(fmt.Sprintf("+%d", f.Additions)),
				DeletionsStyle.Render(fmt.Sprintf("-%d", f.Deletions)))
			contentLines = append(contentLines, fileLine)
		}
	}

	// Pad to fixed height
	for len(contentLines) < expandedHeight-3 { // -3 for top, bottom, help
		contentLines = append(contentLines, "")
	}

	// Build result with borders
	var result []string

	// Top border
	topBorder := ExpandedBorderStyle.Render("╔" + strings.Repeat("═", boxWidth-2) + "╗")
	result = append(result, "  "+graphCont+topBorder)

	// Content rows
	for _, content := range contentLines {
		// Truncate and pad content
		contentDisplay := truncateToWidth(content, innerWidth)
		padded := contentDisplay + strings.Repeat(" ", innerWidth-displayLen(contentDisplay))
		line := ExpandedBorderStyle.Render("║") + " " + padded + " " + ExpandedBorderStyle.Render("║")
		result = append(result, "  "+graphCont+line)
	}

	// Help line
	help := "[↑/↓] file  [Enter] diff  [Esc] close"
	helpPad := innerWidth - len(help)
	if helpPad < 0 {
		helpPad = 0
	}
	helpLine := ExpandedHelpStyle.Render(help) + strings.Repeat(" ", helpPad)
	result = append(result, "  "+graphCont+ExpandedBorderStyle.Render("║")+" "+helpLine+" "+ExpandedBorderStyle.Render("║"))

	// Bottom border
	bottomBorder := ExpandedBorderStyle.Render("╚" + strings.Repeat("═", boxWidth-2) + "╝")
	result = append(result, "  "+graphCont+bottomBorder)

	return result
}

// truncateToWidth truncates a string (with ANSI codes) to a display width
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var result strings.Builder
	displayCount := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if displayCount >= width {
			break
		}
		result.WriteRune(r)
		displayCount++
	}
	return result.String()
}

func (m Model) renderTwoColumnExpanded(commit *domain.Commit, files []domain.FileChange, fileCursor int, fileScrollOffset int, totalWidth int) []string {
	// Box structure for two columns:
	// Top: ╔ + ═×left + ╤ + ═×right + ╗
	// Row: ║ + content + │ + content + ║
	// Bot: ╚ + ═×inner + ╝
	//
	// Box width = totalWidth
	// Inner width (between ╔ and ╗) = totalWidth - 2
	// Split inner: leftInterior + 1 (╤) + rightInterior = totalWidth - 2
	// So: leftInterior + rightInterior = totalWidth - 3
	innerWidth := totalWidth - 2
	leftInterior := (innerWidth - 1) / 2 // -1 for center separator ╤
	rightInterior := innerWidth - 1 - leftInterior

	var lines []string

	// Top border: ╔ + ═×leftInterior + ╤ + ═×rightInterior + ╗ = totalWidth
	lines = append(lines, ExpandedBorderStyle.Render("╔"+strings.Repeat("═", leftInterior)+"╤"+strings.Repeat("═", rightInterior)+"╗"))

	// Content rows - leave 1 char padding each side
	leftContentWidth := leftInterior - 2
	rightContentWidth := rightInterior - 2
	if leftContentWidth < 10 {
		leftContentWidth = 10
	}
	if rightContentWidth < 10 {
		rightContentWidth = 10
	}

	leftLines := m.renderMetadataColumn(commit, leftContentWidth)
	rightLines := m.renderFilesColumn(files, fileCursor, fileScrollOffset, rightContentWidth)

	// Pad to same height
	maxLines := expandedHeight - 2 // -2 for borders
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	for i := 0; i < maxLines; i++ {
		left := padRight(leftLines[i], leftContentWidth)
		right := padRight(rightLines[i], rightContentWidth)
		// Row: ║ + space + leftContent + space + │ + space + rightContent + space + ║
		// Width = 1 + 1 + leftContentWidth + 1 + 1 + 1 + rightContentWidth + 1 + 1
		//       = leftContentWidth + rightContentWidth + 7
		//       = (leftInterior-2) + (rightInterior-2) + 7
		//       = leftInterior + rightInterior + 3
		//       = (totalWidth - 3) + 3 = totalWidth ✓
		lines = append(lines, ExpandedBorderStyle.Render("║")+" "+left+" "+ExpandedBorderStyle.Render("│")+" "+right+" "+ExpandedBorderStyle.Render("║"))
	}

	// Bottom border with help text centered
	// Bottom: ╚ + inner + ╝ = totalWidth, so inner = totalWidth - 2
	help := " [↑/↓] file  [Enter] diff  [Esc] close "
	bottomInner := totalWidth - 2
	helpLen := len(help)
	if helpLen > bottomInner {
		help = help[:bottomInner]
		helpLen = bottomInner
	}
	leftBorder := (bottomInner - helpLen) / 2
	rightBorder := bottomInner - helpLen - leftBorder
	bottomBorder := "╚" + strings.Repeat("═", leftBorder) + help + strings.Repeat("═", rightBorder) + "╝"
	lines = append(lines, ExpandedBorderStyle.Render(bottomBorder))

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

	// Bottom border with help text centered
	help := " [↑/↓] file  [Enter] diff  [Esc] close "
	borderWidth := totalWidth - 2 // -2 for ╚ and ╝
	helpLen := len(help)
	if helpLen > borderWidth {
		helpLen = borderWidth
	}
	leftBorder := (borderWidth - helpLen) / 2
	rightBorder := borderWidth - helpLen - leftBorder
	bottomBorder := "╚" + strings.Repeat("═", leftBorder) + help + strings.Repeat("═", rightBorder) + "╝"
	lines = append(lines, ExpandedBorderStyle.Render(bottomBorder))

	return lines
}

func (m Model) renderMetadataColumn(commit *domain.Commit, width int) []string {
	var lines []string

	// Hash
	hashLabel := ExpandedLabelStyle.Render("Commit:")
	hashValue := ExpandedHashStyle.Render(truncateStr(commit.Hash, width-10))
	lines = append(lines, truncateWithAnsi(hashLabel+" "+hashValue, width))

	// Author
	authorLabel := ExpandedLabelStyle.Render("Author:")
	authorValue := ExpandedValueStyle.Render(truncateStr(fmt.Sprintf("%s <%s>", commit.Author, commit.Email), width-10))
	lines = append(lines, truncateWithAnsi(authorLabel+" "+authorValue, width))

	// Date
	dateLabel := ExpandedLabelStyle.Render("Date:")
	dateValue := ExpandedValueStyle.Render(commit.Date.Format("Jan 2, 2006 15:04"))
	lines = append(lines, truncateWithAnsi(dateLabel+"   "+dateValue, width))

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
		lines = append(lines, truncateWithAnsi(parentLabel+" "+parentValue, width))
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

	// Header - truncate to fit width
	header := fmt.Sprintf("Files (%d)  %s %s",
		len(files),
		AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdd)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", totalDel)),
	)
	lines = append(lines, truncateWithAnsi(header, width))

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
		// Ensure line fits within column width
		lines = append(lines, truncateWithAnsi(line, width))
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
		// Truncate if too long (preserving ANSI codes is complex, so just cut)
		return truncateWithAnsi(s, width)
	}
	return s + strings.Repeat(" ", width-sLen)
}

// truncateWithAnsi truncates a string to width display characters, handling ANSI codes
func truncateWithAnsi(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var result strings.Builder
	displayCount := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if displayCount >= width {
			break
		}
		result.WriteRune(r)
		displayCount++
	}
	// Reset any open ANSI sequences
	if inEscape || displayCount > 0 {
		result.WriteString("\x1b[0m")
	}
	return result.String()
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
