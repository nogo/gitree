package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

// DiffView displays the diff for a file
type DiffView struct {
	visible    bool
	loading    bool
	filePath   string
	diff       string
	additions  int
	deletions  int
	fileIndex  int
	totalFiles int
	files      []domain.FileChange
	viewport   viewport.Model
	width      int
	height     int
	isBinary   bool
}

// New creates a new DiffView
func New() DiffView {
	return DiffView{}
}

// Show displays the diff view for a file
func (d *DiffView) Show(files []domain.FileChange, fileIndex int) {
	d.visible = true
	d.loading = true
	d.files = files
	d.fileIndex = fileIndex
	d.totalFiles = len(files)
	if fileIndex >= 0 && fileIndex < len(files) {
		d.filePath = files[fileIndex].Path
		d.additions = files[fileIndex].Additions
		d.deletions = files[fileIndex].Deletions
	}
	d.diff = ""
	d.isBinary = false
}

// SetDiff sets the loaded diff content
func (d *DiffView) SetDiff(diff string, isBinary bool) {
	d.loading = false
	d.diff = diff
	d.isBinary = isBinary

	// Initialize viewport with rendered diff
	d.viewport = viewport.New(d.contentWidth(), d.contentHeight())
	d.viewport.SetContent(renderDiff(diff))
}

// Hide hides the diff view
func (d *DiffView) Hide() {
	d.visible = false
	d.loading = false
	d.diff = ""
	d.files = nil
}

// IsVisible returns whether the diff view is visible
func (d DiffView) IsVisible() bool {
	return d.visible
}

// IsLoading returns whether a diff is being loaded
func (d DiffView) IsLoading() bool {
	return d.loading
}

// SetSize sets the diff view dimensions
func (d *DiffView) SetSize(w, h int) {
	d.width = w
	d.height = h
	d.viewport.Width = d.contentWidth()
	d.viewport.Height = d.contentHeight()
}

func (d DiffView) contentWidth() int {
	// Account for border and padding
	w := d.width - 6
	if w < 20 {
		w = 20
	}
	return w
}

func (d DiffView) contentHeight() int {
	// Account for border, header, footer
	h := d.height - 8
	if h < 5 {
		h = 5
	}
	return h
}

// CurrentFile returns the current file path
func (d DiffView) CurrentFile() string {
	return d.filePath
}

// FileIndex returns the current file index
func (d DiffView) FileIndex() int {
	return d.fileIndex
}

// NextFile moves to the next file
func (d *DiffView) NextFile() bool {
	if d.fileIndex < d.totalFiles-1 {
		d.fileIndex++
		d.updateCurrentFile()
		return true
	}
	return false
}

// PrevFile moves to the previous file
func (d *DiffView) PrevFile() bool {
	if d.fileIndex > 0 {
		d.fileIndex--
		d.updateCurrentFile()
		return true
	}
	return false
}

func (d *DiffView) updateCurrentFile() {
	if d.fileIndex >= 0 && d.fileIndex < len(d.files) {
		d.filePath = d.files[d.fileIndex].Path
		d.additions = d.files[d.fileIndex].Additions
		d.deletions = d.files[d.fileIndex].Deletions
		d.loading = true
		d.diff = ""
		d.isBinary = false
	}
}

// Update handles input for the diff view
func (d DiffView) Update(msg tea.Msg) (DiffView, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.viewport.LineDown(1)
		case "k", "up":
			d.viewport.LineUp(1)
		case "ctrl+d":
			d.viewport.HalfViewDown()
		case "ctrl+u":
			d.viewport.HalfViewUp()
		case "g":
			d.viewport.GotoTop()
		case "G":
			d.viewport.GotoBottom()
		}
	}

	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// View renders the diff view
func (d DiffView) View() string {
	if !d.visible {
		return ""
	}

	// Header with file path and stats
	header := d.renderHeader()

	// Content
	var content string
	if d.loading {
		content = InfoStyle.Render("Loading diff...")
	} else if d.isBinary {
		content = InfoStyle.Render("Binary file - diff not available")
	} else if d.diff == "" {
		content = InfoStyle.Render("No changes")
	} else {
		content = d.viewport.View()
	}

	// Footer with keybindings
	footer := d.renderFooter()

	// Combine
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		strings.Repeat("─", d.contentWidth()),
		content,
		strings.Repeat("─", d.contentWidth()),
		footer,
	)

	return ContainerStyle.
		Width(d.contentWidth()).
		Height(d.height - 4).
		Render(body)
}

func (d DiffView) renderHeader() string {
	// File path
	path := FilePathStyle.Render(d.filePath)

	// Stats
	stats := fmt.Sprintf("%s %s",
		AdditionsStyle.Render(fmt.Sprintf("+%d", d.additions)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", d.deletions)),
	)

	// File indicator
	indicator := FileIndicatorStyle.Render(fmt.Sprintf("File %d/%d", d.fileIndex+1, d.totalFiles))

	// Combine: path on left, stats and indicator on right
	left := path
	right := fmt.Sprintf("%s  %s", stats, indicator)

	// Calculate spacing
	spacing := d.contentWidth() - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 1 {
		spacing = 1
	}

	return left + strings.Repeat(" ", spacing) + right
}

func (d DiffView) renderFooter() string {
	return FooterStyle.Render("[↑/↓] scroll  [h/l] prev/next file  [Ctrl+d/u] page  [g/G] top/bottom  [Esc] back")
}
