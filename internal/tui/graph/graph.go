package graph

import (
	"strings"

	"github.com/nogo/gitree/internal/domain"
)

// Renderer handles graph visualization for commit lists
type Renderer struct {
	commits    []domain.Commit
	branches   []domain.Branch
	head       string
	colors     *ColorPalette
	branchTips map[string]bool
	layout     *GraphLayout
	rowRender  *RowRenderer
}

// NewRenderer creates a graph renderer for the given commits
func NewRenderer(commits []domain.Commit, branches []domain.Branch, head string) *Renderer {
	// Build branch tips lookup
	tips := make(map[string]bool)
	for _, b := range branches {
		tips[b.HeadHash] = true
	}

	// Build the DAG layout
	layout := BuildLayout(commits)
	colors := NewColorPalette()

	return &Renderer{
		commits:    commits,
		branches:   branches,
		head:       head,
		colors:     colors,
		branchTips: tips,
		layout:     layout,
		rowRender:  NewRowRenderer(layout, colors),
	}
}

// Width returns the display width of the graph column
func (r *Renderer) Width() int {
	// Each lane takes 2 chars (symbol + connector/space)
	return r.layout.MaxLanes * 2
}

// RenderGraphCell returns the graph portion for commit at index
func (r *Renderer) RenderGraphCell(i int) string {
	if i < 0 || i >= len(r.commits) {
		return strings.Repeat(" ", r.Width())
	}
	return r.rowRender.RenderRow(i)
}

// RenderBranchBadges returns styled branch labels for commit
func (r *Renderer) RenderBranchBadges(c domain.Commit) string {
	if len(c.BranchRefs) == 0 {
		return ""
	}

	var badges []string
	for _, ref := range c.BranchRefs {
		style := r.badgeStyle(ref)
		// Truncate long branch names (rune-aware)
		name := ref
		runes := []rune(name)
		if len(runes) > 15 {
			name = string(runes[:12]) + "…"
		}
		badges = append(badges, style.Render(name))
	}

	return strings.Join(badges, " ") + " "
}

func (r *Renderer) isHead(hash string) bool {
	return hash == r.head || strings.HasPrefix(hash, r.head) || strings.HasPrefix(r.head, hash)
}

func (r *Renderer) isBranchTip(hash string) bool {
	return r.branchTips[hash]
}
