package histogram

import (
	"strconv"
	"strings"
)

func itoa(n int) string {
	return strconv.Itoa(n)
}

// Braille density characters (5 levels, bottom-up fill)
// Level 0: space (empty)
// Level 1: ⣀ (bottom row)
// Level 2: ⣤ (bottom 2 rows)
// Level 3: ⣶ (bottom 3 rows)
// Level 4: ⣿ (full)
var densityChars = []rune{' ', '⣀', '⣤', '⣶', '⣿'}

// renderHistogram creates the video player timeline view
func renderHistogram(bins []Bin, cursor int, focused bool, width, height int, zoomPct int) string {
	if len(bins) == 0 {
		return ""
	}

	// Find max count for scaling
	maxCount := 0
	for _, b := range bins {
		if b.Count > maxCount {
			maxCount = b.Count
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	// Available width for the timeline (minus margins)
	availableWidth := width - 4
	if availableWidth < len(bins) {
		availableWidth = len(bins)
	}

	// Calculate position for each bin (evenly distributed)
	binPositions := make([]int, len(bins))
	for i := range bins {
		if len(bins) == 1 {
			binPositions[i] = availableWidth / 2
		} else {
			binPositions[i] = (i * (availableWidth - 1)) / (len(bins) - 1)
		}
	}

	// Find selection range positions
	selStart, selEnd := -1, -1
	for i, bin := range bins {
		if bin.Selected {
			if selStart < 0 {
				selStart = binPositions[i]
			}
			selEnd = binPositions[i]
		}
	}

	// Line 1: Date labels
	dateLabels := renderDateLabels(bins, binPositions, availableWidth)

	// Line 2: Density bars (braille)
	var densityLine strings.Builder
	densityLine.WriteString("  ")
	pos := 0
	for i, bin := range bins {
		// Fill spaces to reach bin position
		for pos < binPositions[i] {
			densityLine.WriteString(" ")
			pos++
		}
		// Calculate density level (0-4)
		level := 0
		if bin.Count > 0 {
			level = (bin.Count * 4) / maxCount
			if level < 1 {
				level = 1
			}
			if level > 4 {
				level = 4
			}
		}
		char := string(densityChars[level])
		if bin.Selected {
			char = SelectedBarStyle.Render(char)
		} else if bin.Count > 0 {
			char = BarStyle.Render(char)
		}
		densityLine.WriteString(char)
		pos++
	}

	// Line 3: Track line with selection brackets
	var trackLine strings.Builder
	trackLine.WriteString("  ")
	for i := 0; i < availableWidth; i++ {
		if selStart >= 0 && i == selStart {
			trackLine.WriteString(SelectedBarStyle.Render("["))
		} else if selEnd >= 0 && i == selEnd {
			trackLine.WriteString(SelectedBarStyle.Render("]"))
		} else if selStart >= 0 && i > selStart && i < selEnd {
			trackLine.WriteString(SelectedBarStyle.Render("━"))
		} else {
			trackLine.WriteString(AxisStyle.Render("━"))
		}
	}

	// Line 4: Playhead indicator
	var playheadLine strings.Builder
	playheadLine.WriteString("  ")
	if focused && cursor >= 0 && cursor < len(bins) {
		cursorPos := binPositions[cursor]
		for i := 0; i < availableWidth; i++ {
			if i == cursorPos {
				playheadLine.WriteString(SelectedBarStyle.Render("▲"))
			} else {
				playheadLine.WriteString(" ")
			}
		}
	}

	// Line 5: Cursor info (date and commit count) + zoom indicator
	var infoLine strings.Builder
	infoLine.WriteString("  ")
	if focused && cursor >= 0 && cursor < len(bins) {
		bin := bins[cursor]
		info := bin.Start.Format("Jan 02")
		if bin.Count == 1 {
			info += " (1 commit)"
		} else {
			info += " (" + itoa(bin.Count) + " commits)"
		}
		// Center under playhead
		cursorPos := binPositions[cursor]
		padding := cursorPos - len(info)/2
		if padding < 0 {
			padding = 0
		}
		if padding+len(info) > availableWidth {
			padding = availableWidth - len(info)
		}
		infoLine.WriteString(strings.Repeat(" ", padding))
		infoLine.WriteString(SelectedBarStyle.Render(info))

		// Add zoom indicator on the right if zoomed
		if zoomPct < 100 {
			zoomInfo := " [" + itoa(zoomPct) + "% +/- zoom, H/L pan]"
			remaining := availableWidth - padding - len(info)
			if remaining > len(zoomInfo) {
				infoLine.WriteString(strings.Repeat(" ", remaining-len(zoomInfo)))
				infoLine.WriteString(AxisStyle.Render(zoomInfo))
			}
		}
	} else if zoomPct < 100 {
		// Show zoom indicator even when not focused
		zoomInfo := "[" + itoa(zoomPct) + "% view]"
		infoLine.WriteString(strings.Repeat(" ", availableWidth-len(zoomInfo)))
		infoLine.WriteString(AxisStyle.Render(zoomInfo))
	}

	// Combine all lines
	var result strings.Builder
	result.WriteString("  ")
	result.WriteString(AxisStyle.Render(dateLabels))
	result.WriteString("\n")
	result.WriteString(densityLine.String())
	result.WriteString("\n")
	result.WriteString(trackLine.String())
	result.WriteString("\n")
	result.WriteString(playheadLine.String())
	result.WriteString("\n")
	result.WriteString(infoLine.String())

	return result.String()
}

// renderDateLabels creates evenly spaced date labels
func renderDateLabels(bins []Bin, binPositions []int, width int) string {
	if len(bins) == 0 {
		return ""
	}

	// Determine how many labels we can fit
	labelWidth := 8
	maxLabels := width / labelWidth
	if maxLabels < 2 {
		maxLabels = 2
	}
	if maxLabels > 5 {
		maxLabels = 5
	}
	if maxLabels > len(bins) {
		maxLabels = len(bins)
	}

	// Select which bins to label
	labelIndices := make([]int, maxLabels)
	for i := 0; i < maxLabels; i++ {
		if maxLabels == 1 {
			labelIndices[i] = 0
		} else {
			labelIndices[i] = (i * (len(bins) - 1)) / (maxLabels - 1)
		}
	}

	// Build label line
	result := make([]byte, width)
	for i := range result {
		result[i] = ' '
	}

	for _, idx := range labelIndices {
		label := bins[idx].Start.Format("Jan 02")
		pos := binPositions[idx]

		// Center label on position
		startPos := pos - len(label)/2
		if startPos < 0 {
			startPos = 0
		}
		if startPos+len(label) > width {
			startPos = width - len(label)
		}

		// Check for overlap
		canPlace := true
		for j := startPos; j < startPos+len(label) && j < width; j++ {
			if result[j] != ' ' {
				canPlace = false
				break
			}
		}

		if canPlace {
			copy(result[startPos:], label)
		}
	}

	return string(result)
}
