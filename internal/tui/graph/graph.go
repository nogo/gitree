package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

// Unicode characters for graph rendering
const (
	NodeFilled  = "●"
	NodeHollow  = "○"
	LineVert    = "│"
	LineHoriz   = "─"
	MergeRight  = "╯"
	BranchRight = "╮"
)

type Renderer struct {
	commits    []domain.Commit
	branches   []domain.Branch
	head       string
	colors     *ColorPalette
	branchTips map[string]bool
}

func NewRenderer(commits []domain.Commit, branches []domain.Branch, head string) *Renderer {
	// Build branch tips lookup
	tips := make(map[string]bool)
	for _, b := range branches {
		tips[b.HeadHash] = true
	}

	return &Renderer{
		commits:    commits,
		branches:   branches,
		head:       head,
		colors:     NewColorPalette(),
		branchTips: tips,
	}
}

// RenderGraphCell returns the graph portion for commit at index
func (r *Renderer) RenderGraphCell(i int) string {
	c := r.commits[i]

	// Determine node character
	node := NodeFilled
	if r.isHead(c.Hash) || r.isBranchTip(c.Hash) {
		node = NodeHollow
	}

	// Determine if this is a merge commit
	hasMultipleParents := len(c.Parents) > 1

	// Get color for this commit
	color := r.colorForCommit(c)
	nodeStr := color.Render(node)

	// Check if there's continuation line below (not last commit, has parent)
	hasLineBelow := i < len(r.commits)-1 && len(c.Parents) > 0

	// Simple rendering based on commit type
	if hasMultipleParents {
		// Merge commit
		return color.Render(MergeRight) + nodeStr + "  "
	}

	if hasLineBelow && r.hasSideBranch(i) {
		// Commit with parallel branch
		return nodeStr + " " + color.Render(LineVert) + " "
	}

	return "  " + nodeStr + "  "
}

// RenderBranchBadges returns styled branch labels for commit
func (r *Renderer) RenderBranchBadges(c domain.Commit) string {
	if len(c.BranchRefs) == 0 {
		return ""
	}

	var badges []string
	for _, ref := range c.BranchRefs {
		style := r.badgeStyle(ref)
		// Truncate long branch names
		name := ref
		if len(name) > 15 {
			name = name[:12] + "..."
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

func (r *Renderer) childCount(hash string) int {
	count := 0
	for _, c := range r.commits {
		for _, parent := range c.Parents {
			if parent == hash || strings.HasPrefix(parent, hash) || strings.HasPrefix(hash, parent) {
				count++
			}
		}
	}
	return count
}

func (r *Renderer) hasSideBranch(i int) bool {
	// Check if this commit has multiple children (branch point)
	if i >= len(r.commits) {
		return false
	}
	return r.childCount(r.commits[i].Hash) > 1
}

func (r *Renderer) colorForCommit(c domain.Commit) lipgloss.Style {
	if len(c.BranchRefs) > 0 {
		return r.colors.ForBranch(c.BranchRefs[0])
	}
	return r.colors.Default()
}
