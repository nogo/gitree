package list

import (
	"strings"

	"github.com/nogo/gitree/internal/tui/text"
)

// Row represents a single commit row with all column data.
// Separates data collection from rendering for cleaner code.
type Row struct {
	Cursor  string // "> ", " *", ">*", or "  "
	Graph   string // graph visualization (may contain ANSI)
	Message string // commit message with optional badges (may contain ANSI)
	Author  string // author name
	Date    string // relative date
	Hash    string // short commit hash
}

// RowStyle defines how a row should be styled.
type RowStyle struct {
	Selected bool
	Dimmed   bool
	Width    int // total row width for selected row highlighting
}

// Render formats the row using the given layout and style.
func (r Row) Render(layout RowLayout, style RowStyle) string {
	// Fit each column to its layout width
	cursor := text.Fit(r.Cursor, layout.Cursor)
	graph := text.FitAnsi(r.Graph, layout.Graph)
	message := text.FitAnsi(r.Message, layout.Message)
	author := text.FitLeft(r.Author, layout.Author)
	date := text.FitLeft(r.Date, layout.Date)
	hash := text.Fit(r.Hash, layout.Hash)

	// Apply styles to non-selected rows (styles defined in styles.go)
	if !style.Selected {
		if style.Dimmed {
			hash = DimmedHashStyle.Render(hash)
			author = DimmedAuthorStyle.Render(author)
			date = DimmedDateStyle.Render(date)
			message = DimmedMessageStyle.Render(message)
		} else {
			hash = HashStyle.Render(hash)
			author = AuthorStyle.Render(author)
			date = DateStyle.Render(date)
		}
	}

	// Build row: cursor | graph | message | author | date | hash
	var b strings.Builder
	b.WriteString(cursor)
	b.WriteString(graph)
	b.WriteString(" ")
	b.WriteString(message)
	b.WriteString("  ")
	b.WriteString(author)
	b.WriteString("  ")
	b.WriteString(date)
	b.WriteString("  ")
	b.WriteString(hash)

	row := b.String()

	if style.Selected {
		return SelectedRowStyle.Width(style.Width).Render(row)
	}
	return row
}
