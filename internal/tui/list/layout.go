package list

// RowLayout defines fixed column widths for consistent rendering.
// Calculated once per render pass, reused for all rows.
//
// Layout: | Cursor | Graph | Message (with badges) | Author | Date | Hash |
type RowLayout struct {
	Cursor  int // cursor indicator width (2: "> " or "  ")
	Graph   int // graph column width (capped)
	Message int // commit message width (includes badges, flexible)
	Author  int // author name width (fixed)
	Date    int // relative date width (fixed)
	Hash    int // short hash width (fixed)
}

// Layout constants - single source of truth
const (
	ColCursor     = 2
	ColAuthor     = 12
	ColDate       = 10
	ColHash       = 7
	MaxGraphWidth = 40 // cap graph to prevent overflow with many branches

	// Separators: space after graph(1), after message(2), after author(2), after date(2)
	colSeparators = 7

	// Minimum message width
	minMessageWidth = 20
)

// NewRowLayout calculates column widths based on terminal width and graph lanes.
// This should be called once per render pass, not per row.
func NewRowLayout(termWidth, graphLanes int) RowLayout {
	// Graph width: 2 chars per lane, capped
	graphWidth := graphLanes * 2
	if graphWidth > MaxGraphWidth {
		graphWidth = MaxGraphWidth
	}
	if graphWidth < 2 {
		graphWidth = 2
	}

	// Fixed columns total (excluding message)
	fixedWidth := ColCursor + graphWidth + ColAuthor + ColDate + ColHash + colSeparators

	// Message gets remaining space
	msgWidth := termWidth - fixedWidth
	if msgWidth < minMessageWidth {
		msgWidth = minMessageWidth
	}

	return RowLayout{
		Cursor:  ColCursor,
		Graph:   graphWidth,
		Message: msgWidth,
		Author:  ColAuthor,
		Date:    ColDate,
		Hash:    ColHash,
	}
}

// NewRowLayoutWithGraph creates a layout with a specific graph width.
// Used for dynamic viewport-based graph sizing.
func NewRowLayoutWithGraph(termWidth, graphWidth int) RowLayout {
	if graphWidth > MaxGraphWidth {
		graphWidth = MaxGraphWidth
	}
	if graphWidth < 2 {
		graphWidth = 2
	}

	fixedWidth := ColCursor + graphWidth + ColAuthor + ColDate + ColHash + colSeparators
	msgWidth := termWidth - fixedWidth
	if msgWidth < minMessageWidth {
		msgWidth = minMessageWidth
	}

	return RowLayout{
		Cursor:  ColCursor,
		Graph:   graphWidth,
		Message: msgWidth,
		Author:  ColAuthor,
		Date:    ColDate,
		Hash:    ColHash,
	}
}

// TotalWidth returns the total row width
func (l RowLayout) TotalWidth() int {
	return l.Cursor + l.Graph + 1 + l.Message + 2 + l.Author + 2 + l.Date + 2 + l.Hash
}

// GraphMaxLanes returns the max lanes that fit in the graph column
func (l RowLayout) GraphMaxLanes() int {
	return l.Graph / 2
}
