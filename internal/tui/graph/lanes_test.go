package graph

import (
	"testing"

	"github.com/nogo/gitree/internal/domain"
)

func TestBuildLayout_Linear(t *testing.T) {
	// Linear history: A -> B -> C (A is newest)
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb"}},
		{Hash: "bbb", Parents: []string{"ccc"}},
		{Hash: "ccc", Parents: []string{}},
	}

	layout := BuildLayout(commits)

	if layout.MaxLanes != 1 {
		t.Errorf("Expected MaxLanes=1 for linear history, got %d", layout.MaxLanes)
	}

	// All commits should be in lane 0
	for i, node := range layout.Nodes {
		if node.Lane != 0 {
			t.Errorf("Commit %d: expected Lane=0, got %d", i, node.Lane)
		}
	}
}

func TestBuildLayout_SimpleBranch(t *testing.T) {
	// Branch and merge:
	//   A (merge) -> B, C
	//   B -> D
	//   C -> D
	//   D (root)
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{"bbb", "ccc"}}, // merge
		{Hash: "bbb", Parents: []string{"ddd"}},        // main branch
		{Hash: "ccc", Parents: []string{"ddd"}},        // feature branch
		{Hash: "ddd", Parents: []string{}},             // common ancestor
	}

	layout := BuildLayout(commits)

	// Should have at least 2 lanes (main + feature)
	if layout.MaxLanes < 2 {
		t.Errorf("Expected MaxLanes>=2 for branch/merge, got %d", layout.MaxLanes)
	}

	// Merge commit should have fork information
	mergeNode := layout.Nodes[0]
	if len(mergeNode.ForkTo) == 0 {
		t.Error("Merge commit should have ForkTo for second parent")
	}
}

func TestBuildLayout_Empty(t *testing.T) {
	layout := BuildLayout([]domain.Commit{})

	if layout.MaxLanes != 1 {
		t.Errorf("Expected MaxLanes=1 for empty commits, got %d", layout.MaxLanes)
	}

	if len(layout.Nodes) != 0 {
		t.Errorf("Expected 0 nodes for empty commits, got %d", len(layout.Nodes))
	}
}

func TestBuildLayout_SingleCommit(t *testing.T) {
	commits := []domain.Commit{
		{Hash: "aaa", Parents: []string{}},
	}

	layout := BuildLayout(commits)

	if layout.MaxLanes != 1 {
		t.Errorf("Expected MaxLanes=1, got %d", layout.MaxLanes)
	}

	if layout.Nodes[0].Lane != 0 {
		t.Errorf("Single commit should be in lane 0, got %d", layout.Nodes[0].Lane)
	}
}

func TestHashMatch(t *testing.T) {
	tests := []struct {
		a, b   string
		expect bool
	}{
		{"abc123", "abc123", true},
		{"abc1234567890", "abc1234", true},
		{"abc1234", "abc1234567890", true},
		{"abc1234", "def5678", false},
		{"", "", true},
		{"abc", "abc", true},
	}

	for _, tc := range tests {
		got := hashMatch(tc.a, tc.b)
		if got != tc.expect {
			t.Errorf("hashMatch(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.expect)
		}
	}
}

func TestSortInts(t *testing.T) {
	tests := []struct {
		input  []int
		expect []int
	}{
		{[]int{3, 1, 2}, []int{1, 2, 3}},
		{[]int{1, 2, 3}, []int{1, 2, 3}},
		{[]int{}, []int{}},
		{[]int{5}, []int{5}},
	}

	for _, tc := range tests {
		// Make a copy
		s := make([]int, len(tc.input))
		copy(s, tc.input)
		sortInts(s)
		for i := range s {
			if s[i] != tc.expect[i] {
				t.Errorf("sortInts(%v) = %v, want %v", tc.input, s, tc.expect)
				break
			}
		}
	}
}
