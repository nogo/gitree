package graph

import "github.com/nogo/gitree/internal/domain"

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
	Nodes      []*CommitNode
	HashToNode map[string]*CommitNode
	MaxLanes   int
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
		MaxLanes:   1,
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

	for _, node := range l.Nodes {
		// Find which lanes are targeting this commit
		var targetingLanes []int
		for lane, targetHash := range activeLanes {
			if hashMatch(targetHash, node.Hash) {
				targetingLanes = append(targetingLanes, lane)
			}
		}

		// Sort targeting lanes for deterministic behavior
		sortInts(targetingLanes)

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
				assignedLane = len(activeLanes)
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
	}

	// Ensure at least 1 lane
	if l.MaxLanes < 1 {
		l.MaxLanes = 1
	}
}

// ActiveLanesAt returns which lanes are active at a given row
// A lane is active if it has a target hash that appears at or after this row
func (l *GraphLayout) ActiveLanesAt(row int) map[int]bool {
	active := make(map[int]bool)

	// Replay lane assignment up to this row
	activeLanes := make(map[int]string)
	var freeLanes []int

	for i := 0; i <= row && i < len(l.Nodes); i++ {
		node := l.Nodes[i]

		// Find targeting lanes
		var targetingLanes []int
		for lane, targetHash := range activeLanes {
			if hashMatch(targetHash, node.Hash) {
				targetingLanes = append(targetingLanes, lane)
			}
		}
		sortInts(targetingLanes)

		var assignedLane int
		if len(targetingLanes) > 0 {
			assignedLane = targetingLanes[0]
			for _, lane := range targetingLanes[1:] {
				delete(activeLanes, lane)
				freeLanes = insertSorted(freeLanes, lane)
			}
		} else {
			if len(freeLanes) > 0 {
				assignedLane = freeLanes[0]
				freeLanes = freeLanes[1:]
			} else {
				assignedLane = len(activeLanes)
			}
		}

		if len(node.Parents) == 0 {
			delete(activeLanes, assignedLane)
			freeLanes = insertSorted(freeLanes, assignedLane)
		} else {
			activeLanes[assignedLane] = node.Parents[0]
			for _, parentHash := range node.Parents[1:] {
				var newLane int
				if len(freeLanes) > 0 {
					newLane = freeLanes[0]
					freeLanes = freeLanes[1:]
				} else {
					newLane = l.maxLaneUsed(activeLanes) + 1
				}
				activeLanes[newLane] = parentHash
			}
		}
	}

	// Convert to bool map
	for lane := range activeLanes {
		active[lane] = true
	}
	// Also include the current node's lane
	if row < len(l.Nodes) {
		active[l.Nodes[row].Lane] = true
	}

	return active
}

func (l *GraphLayout) maxLaneUsed(lanes map[int]string) int {
	max := -1
	for lane := range lanes {
		if lane > max {
			max = lane
		}
	}
	return max
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

// sortInts sorts a slice of ints in place (simple insertion sort for small slices)
func sortInts(s []int) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// insertSorted inserts v into sorted slice s maintaining order
func insertSorted(s []int, v int) []int {
	i := 0
	for i < len(s) && s[i] < v {
		i++
	}
	s = append(s, 0)
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
