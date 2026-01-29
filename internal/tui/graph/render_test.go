package graph

import (
	"strings"
	"testing"

	"github.com/nogo/gitree/internal/domain"
)

// stripAnsi removes ANSI escape codes for testing
func stripAnsi(s string) string {
	var result strings.Builder
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
		result.WriteRune(r)
	}
	return result.String()
}

func TestRenderRow_Linear(t *testing.T) {
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb"}},
		{Hash: "bbb", Parents: []string{"ccc"}},
		{Hash: "ccc", Parents: []string{}},
	}

	layout := BuildLayout(commits)
	renderer := NewRowRenderer(layout, NewColorPalette())

	for i := range commits {
		row := stripAnsi(renderer.RenderRow(i))
		t.Logf("Row %d: %q", i, row)

		// Should contain the node character
		if !strings.Contains(row, string(CharNode)) {
			t.Errorf("Row %d should contain node character", i)
		}
	}
}

func TestRenderRow_BranchMerge(t *testing.T) {
	// A (merge aaa) has parents bbb and ccc
	// bbb has parent ddd
	// ccc has parent ddd
	// ddd is root
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb", "ccc"}},
		{Hash: "bbb", Parents: []string{"ddd"}},
		{Hash: "ccc", Parents: []string{"ddd"}},
		{Hash: "ddd", Parents: []string{}},
	}

	layout := BuildLayout(commits)
	renderer := NewRowRenderer(layout, NewColorPalette())

	t.Log("Branch/Merge visualization:")
	for i := range commits {
		row := stripAnsi(renderer.RenderRow(i))
		t.Logf("Row %d: %q", i, row)
	}

	// First row (merge) should have a fork connection
	row0 := stripAnsi(renderer.RenderRow(0))
	if !strings.Contains(row0, string(CharCornerTR)) && !strings.Contains(row0, string(CharHorizontal)) {
		t.Logf("Row 0 might need fork indicators: %q", row0)
	}
}

func TestRenderRow_Width(t *testing.T) {
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb", "ccc"}},
		{Hash: "bbb", Parents: []string{"ddd"}},
		{Hash: "ccc", Parents: []string{"ddd"}},
		{Hash: "ddd", Parents: []string{}},
	}

	layout := BuildLayout(commits)
	renderer := NewRowRenderer(layout, NewColorPalette())

	expectedWidth := layout.MaxLanes * 2

	for i := range commits {
		row := stripAnsi(renderer.RenderRow(i))
		runeCount := len([]rune(row))
		if runeCount != expectedWidth {
			t.Errorf("Row %d: expected width %d, got %d (row: %q)", i, expectedWidth, runeCount, row)
		}
	}
}

func TestRenderRow_OctopusMerge(t *testing.T) {
	// A (octopus merge) has parents bbb, ccc, ddd
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb", "ccc", "ddd"}},
		{Hash: "bbb", Parents: []string{"eee"}},
		{Hash: "ccc", Parents: []string{"eee"}},
		{Hash: "ddd", Parents: []string{"eee"}},
		{Hash: "eee", Parents: []string{}},
	}

	layout := BuildLayout(commits)
	renderer := NewRowRenderer(layout, NewColorPalette())

	t.Log("Octopus merge visualization:")
	for i := range commits {
		row := stripAnsi(renderer.RenderRow(i))
		t.Logf("Row %d: %q", i, row)
	}

	// Should have at least 3 lanes
	if layout.MaxLanes < 3 {
		t.Errorf("Octopus merge should have at least 3 lanes, got %d", layout.MaxLanes)
	}
}
