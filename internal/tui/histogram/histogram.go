package histogram

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
)

// Bin represents a time bucket with commit count
type Bin struct {
	Start    time.Time
	End      time.Time
	Count    int
	Selected bool
}

// Histogram displays commit density over time
type Histogram struct {
	bins           []Bin
	cursor         int  // current bin position when focused
	selectionStart int  // -1 if no selection
	selectionEnd   int  // -1 if no selection
	focused        bool
	visible        bool
	width          int
	height         int // bar height (not including axis)
	// Viewport for zoom
	viewStart int // first visible bin index
	viewEnd   int // last visible bin index (exclusive)
	zoomLevel int // 0=full, 1=50%, 2=25%, 3=12.5%
}

// New creates a histogram from commits
func New(commits []domain.Commit, width int) Histogram {
	h := Histogram{
		selectionStart: -1,
		selectionEnd:   -1,
		visible:        true,
		width:          width,
		height:         5,
		zoomLevel:      0,
	}
	h.Recalculate(commits, width)
	return h
}

// Recalculate rebuilds bins from commits
func (h *Histogram) Recalculate(commits []domain.Commit, width int) {
	h.width = width
	if len(commits) == 0 {
		h.bins = nil
		return
	}

	// Find date range
	oldest := commits[len(commits)-1].Date
	newest := commits[0].Date

	// Determine bin granularity based on date range
	duration := newest.Sub(oldest)
	var binDuration time.Duration
	var binCount int

	// Available width for bars (minus margins)
	maxBins := (width - 10) / 2 // each bar is ~2 chars wide
	if maxBins < 5 {
		maxBins = 5
	}
	if maxBins > 60 {
		maxBins = 60
	}

	switch {
	case duration < 30*24*time.Hour: // < 30 days: daily
		binDuration = 24 * time.Hour
		binCount = int(duration/(24*time.Hour)) + 1
	case duration < 180*24*time.Hour: // < 6 months: weekly
		binDuration = 7 * 24 * time.Hour
		binCount = int(duration/(7*24*time.Hour)) + 1
	default: // >= 6 months: monthly (approx 30 days)
		binDuration = 30 * 24 * time.Hour
		binCount = int(duration/(30*24*time.Hour)) + 1
	}

	// Clamp bin count to available width
	if binCount > maxBins {
		binCount = maxBins
		binDuration = duration / time.Duration(binCount-1)
	}
	if binCount < 1 {
		binCount = 1
	}

	// Create bins
	h.bins = make([]Bin, binCount)
	for i := range h.bins {
		h.bins[i].Start = oldest.Add(time.Duration(i) * binDuration)
		h.bins[i].End = oldest.Add(time.Duration(i+1) * binDuration)
	}
	// Ensure last bin captures newest commit
	if len(h.bins) > 0 {
		h.bins[len(h.bins)-1].End = newest.Add(time.Hour)
	}

	// Count commits per bin
	for _, c := range commits {
		for i := range h.bins {
			if !c.Date.Before(h.bins[i].Start) && c.Date.Before(h.bins[i].End) {
				h.bins[i].Count++
				break
			}
		}
	}

	// Reset cursor if out of bounds
	if h.cursor >= len(h.bins) {
		h.cursor = len(h.bins) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}

	// Initialize viewport to show all bins
	h.viewStart = 0
	h.viewEnd = len(h.bins)
	h.zoomLevel = 0

	// Reapply selection state to bins
	h.updateSelectionState()
}

// Update handles keyboard input, returns (updated, cmd, selectionChanged)
func (h Histogram) Update(msg tea.Msg) (Histogram, tea.Cmd, bool) {
	if !h.focused || !h.visible {
		return h, nil, false
	}

	selectionChanged := false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if h.cursor > 0 {
				h.cursor--
				h.ensureCursorVisible()
			}
		case "l", "right":
			if h.cursor < len(h.bins)-1 {
				h.cursor++
				h.ensureCursorVisible()
			}
		case "[":
			// Set selection start at cursor
			h.selectionStart = h.cursor
			if h.selectionEnd < h.selectionStart {
				h.selectionEnd = h.selectionStart
			}
			h.updateSelectionState()
			selectionChanged = true
		case "]":
			// Set selection end at cursor
			h.selectionEnd = h.cursor
			if h.selectionStart < 0 || h.selectionStart > h.selectionEnd {
				h.selectionStart = h.selectionEnd
			}
			h.updateSelectionState()
			selectionChanged = true
		case " ", "space":
			// Toggle selection at cursor (quick select)
			selectionChanged = h.toggleSelection()
		case "enter":
			// Apply current selection as filter
			if h.selectionStart >= 0 {
				selectionChanged = true
			}
		case "esc":
			// Clear selection and return focus to list
			if h.selectionStart >= 0 {
				h.clearSelection()
				selectionChanged = true
			}
			h.focused = false
		case "+", "=":
			// Zoom in (center on cursor)
			h.zoomIn()
		case "-", "_":
			// Zoom out
			h.zoomOut()
		case "H":
			// Pan left (jump)
			h.panLeft()
		case "L":
			// Pan right (jump)
			h.panRight()
		}
	}

	return h, nil, selectionChanged
}

// zoomIn halves the visible range, centered on cursor
func (h *Histogram) zoomIn() {
	if len(h.bins) == 0 {
		return
	}
	// Max zoom: show at least 5 bins
	viewSize := h.viewEnd - h.viewStart
	if viewSize <= 5 {
		return
	}

	h.zoomLevel++
	newSize := viewSize / 2
	if newSize < 5 {
		newSize = 5
	}

	// Center on cursor
	center := h.cursor
	h.viewStart = center - newSize/2
	h.viewEnd = h.viewStart + newSize

	// Clamp to bounds
	if h.viewStart < 0 {
		h.viewStart = 0
		h.viewEnd = newSize
	}
	if h.viewEnd > len(h.bins) {
		h.viewEnd = len(h.bins)
		h.viewStart = h.viewEnd - newSize
		if h.viewStart < 0 {
			h.viewStart = 0
		}
	}
}

// zoomOut doubles the visible range
func (h *Histogram) zoomOut() {
	if len(h.bins) == 0 {
		return
	}
	if h.zoomLevel == 0 {
		return // Already at full view
	}

	h.zoomLevel--
	viewSize := h.viewEnd - h.viewStart
	newSize := viewSize * 2
	if newSize > len(h.bins) {
		newSize = len(h.bins)
	}

	// Center on current view
	center := (h.viewStart + h.viewEnd) / 2
	h.viewStart = center - newSize/2
	h.viewEnd = h.viewStart + newSize

	// Clamp to bounds
	if h.viewStart < 0 {
		h.viewStart = 0
		h.viewEnd = newSize
	}
	if h.viewEnd > len(h.bins) {
		h.viewEnd = len(h.bins)
		h.viewStart = h.viewEnd - newSize
		if h.viewStart < 0 {
			h.viewStart = 0
		}
	}

	// Full zoom out
	if h.zoomLevel == 0 {
		h.viewStart = 0
		h.viewEnd = len(h.bins)
	}
}

// panLeft moves viewport left
func (h *Histogram) panLeft() {
	if h.viewStart <= 0 {
		return
	}
	viewSize := h.viewEnd - h.viewStart
	shift := viewSize / 4
	if shift < 1 {
		shift = 1
	}
	h.viewStart -= shift
	h.viewEnd -= shift
	if h.viewStart < 0 {
		h.viewStart = 0
		h.viewEnd = viewSize
	}
}

// panRight moves viewport right
func (h *Histogram) panRight() {
	if h.viewEnd >= len(h.bins) {
		return
	}
	viewSize := h.viewEnd - h.viewStart
	shift := viewSize / 4
	if shift < 1 {
		shift = 1
	}
	h.viewStart += shift
	h.viewEnd += shift
	if h.viewEnd > len(h.bins) {
		h.viewEnd = len(h.bins)
		h.viewStart = h.viewEnd - viewSize
	}
}

// ensureCursorVisible scrolls viewport if cursor is outside
func (h *Histogram) ensureCursorVisible() {
	if h.cursor < h.viewStart {
		shift := h.viewStart - h.cursor
		h.viewStart -= shift
		h.viewEnd -= shift
	} else if h.cursor >= h.viewEnd {
		shift := h.cursor - h.viewEnd + 1
		h.viewStart += shift
		h.viewEnd += shift
	}
}

// toggleSelection handles space key - start/extend/complete selection
func (h *Histogram) toggleSelection() bool {
	if h.selectionStart < 0 {
		// Start new selection
		h.selectionStart = h.cursor
		h.selectionEnd = h.cursor
	} else if h.selectionEnd == h.selectionStart {
		// Extend selection to cursor
		h.selectionEnd = h.cursor
		// Normalize so start <= end
		if h.selectionStart > h.selectionEnd {
			h.selectionStart, h.selectionEnd = h.selectionEnd, h.selectionStart
		}
	} else {
		// Selection complete, start new one
		h.selectionStart = h.cursor
		h.selectionEnd = h.cursor
	}
	h.updateSelectionState()
	return true
}

func (h *Histogram) clearSelection() {
	h.selectionStart = -1
	h.selectionEnd = -1
	h.updateSelectionState()
}

func (h *Histogram) updateSelectionState() {
	for i := range h.bins {
		h.bins[i].Selected = h.selectionStart >= 0 &&
			i >= h.selectionStart && i <= h.selectionEnd
	}
}

// View renders the histogram
func (h Histogram) View() string {
	if !h.visible || len(h.bins) == 0 {
		return ""
	}
	// Get visible bins
	visibleBins := h.bins[h.viewStart:h.viewEnd]
	// Adjust cursor relative to viewport
	relativeCursor := h.cursor - h.viewStart
	// Calculate zoom percentage
	zoomPct := 100
	if len(h.bins) > 0 {
		zoomPct = (h.viewEnd - h.viewStart) * 100 / len(h.bins)
	}
	return renderHistogram(visibleBins, relativeCursor, h.focused, h.width, h.height, zoomPct)
}

// SelectedRange returns the time range of selected bins
func (h Histogram) SelectedRange() (start, end time.Time, hasSelection bool) {
	if h.selectionStart < 0 || len(h.bins) == 0 {
		return time.Time{}, time.Time{}, false
	}
	return h.bins[h.selectionStart].Start, h.bins[h.selectionEnd].End, true
}

// SetFocused sets focus state
func (h *Histogram) SetFocused(focused bool) {
	h.focused = focused
}

// IsFocused returns focus state
func (h Histogram) IsFocused() bool {
	return h.focused
}

// SetVisible sets visibility
func (h *Histogram) SetVisible(visible bool) {
	h.visible = visible
}

// IsVisible returns visibility
func (h Histogram) IsVisible() bool {
	return h.visible
}

// Toggle toggles visibility
func (h *Histogram) Toggle() {
	h.visible = !h.visible
}

// HasSelection returns whether a time range is selected
func (h Histogram) HasSelection() bool {
	return h.selectionStart >= 0
}

// Reset clears selection
func (h *Histogram) Reset() {
	h.clearSelection()
}

// Height returns total rendered height (video player style: 5 lines)
func (h Histogram) Height() int {
	if !h.visible || len(h.bins) == 0 {
		return 0
	}
	// date labels(1) + density bars(1) + track line(1) + playhead(1) + info(1)
	return 5
}
