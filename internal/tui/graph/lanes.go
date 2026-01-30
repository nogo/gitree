package graph

import (
	"slices"

	"github.com/nogo/gitree/internal/domain"
)

// CommitNode extends commit data with graph layout information
type CommitNode struct {
	Hash      string
	Parents   []string
	Children  []string // computed: hashes of commits that have this as parent
	Lane      int      // assigned column
	Row       int      // index in display order
	MergeFrom []int    // lanes merging INTO this commit (for └ rendering)
	ForkTo    []int    // lanes forking FROM this commit (for ┐ rendering)
}

// GraphLayout holds the complete graph structure
type GraphLayout struct {
	Nodes        []*CommitNode
	HashToNode   map[string]*CommitNode
	MaxLanes     int
	ActiveLanes  []map[int]bool // active lanes at each row (computed during assignLanes)
}

// BuildLayout constructs the graph layout from commits
// Commits are expected in display order (newest first)
func BuildLayout(commits []domain.Commit) *GraphLayout {
	if len(commits) == 0 {
		return &GraphLayout{
			Nodes:      nil,
			HashToNode: make(map[string]*CommitNode),
			MaxLanes:   1,
		}
	}

	layout := &GraphLayout{
		Nodes:      make([]*CommitNode, len(commits)),
		HashToNode: make(map[string]*CommitNode, len(commits)),
		MaxLanes:   0, // Will grow as lanes are allocated
	}

	// Step 1: Create nodes and build hash lookup
	for i, c := range commits {
		node := &CommitNode{
			Hash:    c.Hash,
			Parents: c.Parents,
			Row:     i,
		}
		layout.Nodes[i] = node
		layout.HashToNode[c.Hash] = node
		// Also index by short hash prefix for flexible matching
		if len(c.Hash) >= 7 {
			layout.HashToNode[c.Hash[:7]] = node
		}
	}

	// Step 2: Build children relationships
	for _, node := range layout.Nodes {
		for _, parentHash := range node.Parents {
			if parent := layout.findNode(parentHash); parent != nil {
				parent.Children = append(parent.Children, node.Hash)
			}
		}
	}

	// Step 3: Assign lanes
	layout.assignLanes()

	return layout
}

// findNode looks up a node by hash (supports partial hash matching)
func (l *GraphLayout) findNode(hash string) *CommitNode {
	if node, ok := l.HashToNode[hash]; ok {
		return node
	}
	// Try prefix match for short hashes
	if len(hash) >= 7 {
		if node, ok := l.HashToNode[hash[:7]]; ok {
			return node
		}
	}
	return nil
}

// assignLanes processes commits top-to-bottom assigning lane positions
func (l *GraphLayout) assignLanes() {
	// activeLanes[lane] = hash of commit this lane is "waiting for"
	activeLanes := make(map[int]string)
	// freeLanes tracks lanes that can be reused (sorted for determinism)
	var freeLanes []int

	// Initialize storage for active lanes at each row
	l.ActiveLanes = make([]map[int]bool, len(l.Nodes))

	for i, node := range l.Nodes {
		// Find which lanes are targeting this commit
		var targetingLanes []int
		for lane, targetHash := range activeLanes {
			if hashMatch(targetHash, node.Hash) {
				targetingLanes = append(targetingLanes, lane)
			}
		}

		// Sort targeting lanes for deterministic behavior
		slices.Sort(targetingLanes)

		var assignedLane int
		if len(targetingLanes) > 0 {
			// Use leftmost targeting lane for the node
			assignedLane = targetingLanes[0]

			// Other targeting lanes merge here
			if len(targetingLanes) > 1 {
				for _, lane := range targetingLanes[1:] {
					node.MergeFrom = append(node.MergeFrom, lane)
					delete(activeLanes, lane)
					freeLanes = insertSorted(freeLanes, lane)
				}
			}
		} else {
			// No lane targeting this commit - allocate new lane
			if len(freeLanes) > 0 {
				assignedLane = freeLanes[0]
				freeLanes = freeLanes[1:]
			} else {
				assignedLane = l.MaxLanes
				l.MaxLanes++
			}
		}

		node.Lane = assignedLane

		// Update max lanes
		if assignedLane+1 > l.MaxLanes {
			l.MaxLanes = assignedLane + 1
		}

		// Handle parents
		if len(node.Parents) == 0 {
			// Root commit - free this lane
			delete(activeLanes, assignedLane)
			freeLanes = insertSorted(freeLanes, assignedLane)
		} else {
			// First parent continues in same lane
			activeLanes[assignedLane] = node.Parents[0]

			// Additional parents get new lanes (fork)
			for _, parentHash := range node.Parents[1:] {
				var newLane int
				if len(freeLanes) > 0 {
					newLane = freeLanes[0]
					freeLanes = freeLanes[1:]
				} else {
					newLane = l.MaxLanes
					l.MaxLanes++
				}
				activeLanes[newLane] = parentHash
				node.ForkTo = append(node.ForkTo, newLane)

				// Update max lanes
				if newLane+1 > l.MaxLanes {
					l.MaxLanes = newLane + 1
				}
			}
		}

		// Store active lanes for this row (copy the current state)
		l.ActiveLanes[i] = make(map[int]bool)
		for lane := range activeLanes {
			l.ActiveLanes[i][lane] = true
		}
		// Also include the node's own lane
		l.ActiveLanes[i][assignedLane] = true
	}

	// Ensure at least 1 lane
	if l.MaxLanes < 1 {
		l.MaxLanes = 1
	}
}

// ActiveLanesAt returns which lanes are active at a given row
// Uses pre-computed values from assignLanes
func (l *GraphLayout) ActiveLanesAt(row int) map[int]bool {
	if row < 0 || row >= len(l.ActiveLanes) {
		return make(map[int]bool)
	}
	return l.ActiveLanes[row]
}

// MaxLaneAt returns the highest lane number used at a given row
func (l *GraphLayout) MaxLaneAt(row int) int {
	if row < 0 || row >= len(l.Nodes) {
		return 0
	}

	node := l.Nodes[row]
	maxLane := node.Lane

	// Check active lanes
	for lane := range l.ActiveLanes[row] {
		if lane > maxLane {
			maxLane = lane
		}
	}

	// Check merge/fork connections
	for _, lane := range node.MergeFrom {
		if lane > maxLane {
			maxLane = lane
		}
	}
	for _, lane := range node.ForkTo {
		if lane > maxLane {
			maxLane = lane
		}
	}

	return maxLane
}

// hashMatch checks if two hashes match (handling short hash prefixes)
func hashMatch(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) >= 7 && len(b) >= 7 {
		minLen := min(len(a), len(b))
		if minLen >= 7 {
			return a[:minLen] == b[:minLen]
		}
	}
	return false
}

// insertSorted inserts v into sorted slice s maintaining order
func insertSorted(s []int, v int) []int {
	i, _ := slices.BinarySearch(s, v)
	return slices.Insert(s, i, v)
}
