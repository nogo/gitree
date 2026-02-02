package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// SetDisplayWidth sets the maximum display width for the graph.
// This prevents drawing connections to lanes that would be truncated.
func (r *Renderer) SetDisplayWidth(width int) {
	lanes := width / 2
	r.rowRender.SetDisplayLanes(lanes)
}

// MaxLanesInRange returns the maximum lane used in a range of commits
func (r *Renderer) MaxLanesInRange(start, end int) int {
	if start < 0 {
		start = 0
	}
	if end > len(r.commits) {
		end = len(r.commits)
	}

	maxLane := 0
	for i := start; i < end; i++ {
		if lane := r.layout.MaxLaneAt(i); lane > maxLane {
			maxLane = lane
		}
	}
	return maxLane
}

// WidthForRange returns the graph width needed for a range of commits
func (r *Renderer) WidthForRange(start, end int) int {
	maxLane := r.MaxLanesInRange(start, end)
	return (maxLane + 1) * 2 // +1 because lanes are 0-indexed
}

// RenderGraphCell returns the graph portion for commit at index
func (r *Renderer) RenderGraphCell(i int) string {
	if i < 0 || i >= len(r.commits) {
		return strings.Repeat(" ", r.Width())
	}
	return r.rowRender.RenderRow(i)
}

// RenderGraphCellDimmed returns a dimmed graph portion for commit at index
func (r *Renderer) RenderGraphCellDimmed(i int) string {
	if i < 0 || i >= len(r.commits) {
		return strings.Repeat(" ", r.Width())
	}
	return r.rowRender.RenderRowDimmed(i)
}

// RenderBranchBadges returns styled branch labels for commit
// Merges local and remote branches with same name (e.g., "main | origin")
func (r *Renderer) RenderBranchBadges(c domain.Commit) string {
	if len(c.BranchRefs) == 0 {
		return ""
	}

	// Group branches by base name
	type branchInfo struct {
		hasLocal  bool
		hasRemote bool
		remotes   []string // remote names like "origin", "upstream"
	}
	groups := make(map[string]*branchInfo)

	for _, ref := range c.BranchRefs {
		if strings.HasPrefix(ref, "origin/") {
			baseName := strings.TrimPrefix(ref, "origin/")
			if groups[baseName] == nil {
				groups[baseName] = &branchInfo{}
			}
			groups[baseName].hasRemote = true
			groups[baseName].remotes = append(groups[baseName].remotes, "origin")
		} else if strings.Contains(ref, "/") {
			// Other remotes like "upstream/main"
			parts := strings.SplitN(ref, "/", 2)
			if len(parts) == 2 {
				remoteName, baseName := parts[0], parts[1]
				if groups[baseName] == nil {
					groups[baseName] = &branchInfo{}
				}
				groups[baseName].hasRemote = true
				groups[baseName].remotes = append(groups[baseName].remotes, remoteName)
			}
		} else {
			// Local branch
			if groups[ref] == nil {
				groups[ref] = &branchInfo{}
			}
			groups[ref].hasLocal = true
		}
	}

	// Build badges
	var badges []string
	for baseName, info := range groups {
		var label string
		if info.hasLocal && info.hasRemote {
			// Merge: "main | origin" or "main | origin, upstream"
			label = baseName + " | " + strings.Join(info.remotes, ", ")
		} else if info.hasLocal {
			label = baseName
		} else {
			// Remote only
			label = info.remotes[0] + "/" + baseName
		}

		style := r.badgeStyleForGroup(baseName, info.hasLocal, info.hasRemote)
		badges = append(badges, style.Render(label))
	}

	if len(badges) == 0 {
		return ""
	}
	return strings.Join(badges, " ") + " "
}

// badgeStyleForGroup returns style based on branch type
func (r *Renderer) badgeStyleForGroup(baseName string, hasLocal, hasRemote bool) lipgloss.Style {
	if hasLocal {
		// Use local branch color
		color := r.colors.ForBranch(baseName)
		return BadgeStyle.Background(color.GetForeground()).Foreground(lipgloss.Color("0"))
	}
	// Remote only
	return OriginBadgeStyle
}

// RenderTagBadges returns styled tag labels for commit
// Format: <tagname> with yellow/gold background
func (r *Renderer) RenderTagBadges(c domain.Commit) string {
	if len(c.Tags) == 0 {
		return ""
	}

	var badges []string
	for _, tag := range c.Tags {
		label := "<" + tag + ">"
		badges = append(badges, TagBadgeStyle.Render(label))
	}

	return strings.Join(badges, " ") + " "
}

// RenderContinuation returns continuation lines for expanded row areas
func (r *Renderer) RenderContinuation(i int) string {
	if i < 0 || i >= len(r.commits) {
		return strings.Repeat(" ", r.Width())
	}
	return r.rowRender.RenderContinuation(i)
}

func (r *Renderer) isHead(hash string) bool {
	return hash == r.head || strings.HasPrefix(hash, r.head) || strings.HasPrefix(r.head, hash)
}

func (r *Renderer) isBranchTip(hash string) bool {
	return r.branchTips[hash]
}
