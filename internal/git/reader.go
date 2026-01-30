package git

import (
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/nogo/gitree/internal/domain"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) LoadRepository(path string) (*domain.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	commits, err := r.loadCommitsFromRepo(repo, 0)
	if err != nil {
		return nil, err
	}

	branches, err := r.loadBranchesFromRepo(repo)
	if err != nil {
		return nil, err
	}

	head := ""
	headRef, err := repo.Head()
	if err == nil {
		if headRef.Name().IsBranch() {
			head = headRef.Name().Short()
		} else {
			head = headRef.Hash().String()[:7]
		}
	}

	return &domain.Repository{
		Path:     path,
		Commits:  commits,
		Branches: branches,
		HEAD:     head,
	}, nil
}

func (r *Reader) LoadCommits(path string, limit int) ([]domain.Commit, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return r.loadCommitsFromRepo(repo, limit)
}

func (r *Reader) loadCommitsFromRepo(repo *git.Repository, limit int) ([]domain.Commit, error) {
	// Build map of branch refs pointing to each commit
	branchRefs := make(map[string][]string)
	branches, _ := repo.Branches()
	branches.ForEach(func(ref *plumbing.Reference) error {
		branchRefs[ref.Hash().String()] = append(branchRefs[ref.Hash().String()], ref.Name().Short())
		return nil
	})

	// Also include remote branches
	refs, _ := repo.References()
	refs.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(ref.Name().String(), "refs/remotes/") {
			branchRefs[ref.Hash().String()] = append(branchRefs[ref.Hash().String()], ref.Name().Short())
		}
		return nil
	})

	head, err := repo.Head()
	if err != nil {
		// Empty repository
		return []domain.Commit{}, nil
	}

	// Collect all branch head hashes to include commits from all branches
	var allBranchHashes []plumbing.Hash
	allBranchHashes = append(allBranchHashes, head.Hash())

	localBranches, _ := repo.Branches()
	localBranches.ForEach(func(ref *plumbing.Reference) error {
		allBranchHashes = append(allBranchHashes, ref.Hash())
		return nil
	})

	allRefs, _ := repo.References()
	allRefs.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(ref.Name().String(), "refs/remotes/") {
			allBranchHashes = append(allBranchHashes, ref.Hash())
		}
		return nil
	})

	// Use a map to deduplicate commits from multiple branches
	seen := make(map[string]bool)
	var commits []domain.Commit

	for _, branchHash := range allBranchHashes {
		logOpts := &git.LogOptions{
			From:  branchHash,
			Order: git.LogOrderCommitterTime,
		}

		iter, err := repo.Log(logOpts)
		if err != nil {
			continue
		}

		iter.ForEach(func(c *object.Commit) error {
			hash := c.Hash.String()
			if seen[hash] {
				return nil
			}
			seen[hash] = true

			parents := make([]string, len(c.ParentHashes))
			for i, p := range c.ParentHashes {
				parents[i] = p.String()
			}

			commits = append(commits, domain.Commit{
				Hash:        hash,
				ShortHash:   hash[:7],
				Author:      c.Author.Name,
				Email:       c.Author.Email,
				Date:        c.Committer.When, // Use committer date for proper ordering
				Message:     firstLine(c.Message),
				FullMessage: c.Message,
				Parents:     parents,
				BranchRefs:  branchRefs[hash],
			})
			return nil
		})
		iter.Close()
	}

	// Topological sort: children before parents, with date as tiebreaker
	commits = topoSortCommits(commits)

	// Apply limit if specified
	if limit > 0 && len(commits) > limit {
		commits = commits[:limit]
	}

	return commits, nil
}

// topoSortCommits sorts commits topologically (children before parents)
// with date as secondary sort key for commits at the same level
func topoSortCommits(commits []domain.Commit) []domain.Commit {
	if len(commits) == 0 {
		return commits
	}

	// Build hash -> commit map and hash -> index map
	hashToCommit := make(map[string]*domain.Commit)
	for i := range commits {
		hashToCommit[commits[i].Hash] = &commits[i]
		if len(commits[i].Hash) >= 7 {
			hashToCommit[commits[i].Hash[:7]] = &commits[i]
		}
	}

	// Build child count (in-degree for reverse topo sort)
	childCount := make(map[string]int)
	for i := range commits {
		childCount[commits[i].Hash] = 0
	}
	for i := range commits {
		for _, parentHash := range commits[i].Parents {
			// Find the parent in our commit set
			if _, exists := hashToCommit[parentHash]; exists {
				childCount[parentHash]++
			} else if len(parentHash) >= 7 {
				if _, exists := hashToCommit[parentHash[:7]]; exists {
					childCount[parentHash[:7]]++
				}
			}
		}
	}

	// Find all commits with no children (roots of our view)
	var ready []domain.Commit
	for i := range commits {
		if childCount[commits[i].Hash] == 0 {
			ready = append(ready, commits[i])
		}
	}

	// Sort ready list by date (newest first)
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].Date.After(ready[j].Date)
	})

	// Process commits in topological order
	var result []domain.Commit
	seen := make(map[string]bool)

	for len(ready) > 0 {
		// Take the newest commit from ready list
		commit := ready[0]
		ready = ready[1:]

		if seen[commit.Hash] {
			continue
		}
		seen[commit.Hash] = true
		result = append(result, commit)

		// Decrement child count for parents
		for _, parentHash := range commit.Parents {
			var parentCommit *domain.Commit
			if c, exists := hashToCommit[parentHash]; exists {
				parentCommit = c
			} else if len(parentHash) >= 7 {
				if c, exists := hashToCommit[parentHash[:7]]; exists {
					parentCommit = c
				}
			}

			if parentCommit != nil && !seen[parentCommit.Hash] {
				childCount[parentCommit.Hash]--
				if childCount[parentCommit.Hash] == 0 {
					// Insert into ready list maintaining date order
					inserted := false
					for i := range ready {
						if parentCommit.Date.After(ready[i].Date) {
							ready = append(ready[:i], append([]domain.Commit{*parentCommit}, ready[i:]...)...)
							inserted = true
							break
						}
					}
					if !inserted {
						ready = append(ready, *parentCommit)
					}
				}
			}
		}
	}

	return result
}

func (r *Reader) LoadBranches(path string) ([]domain.Branch, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return r.loadBranchesFromRepo(repo)
}

func (r *Reader) loadBranchesFromRepo(repo *git.Repository) ([]domain.Branch, error) {
	var branches []domain.Branch

	// Local branches
	branchIter, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, domain.Branch{
			Name:     ref.Name().Short(),
			IsRemote: false,
			HeadHash: ref.Hash().String(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Remote branches
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(ref.Name().String(), "refs/remotes/") {
			branches = append(branches, domain.Branch{
				Name:     ref.Name().Short(),
				IsRemote: true,
				HeadHash: ref.Hash().String(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return branches, nil
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		return s[:idx]
	}
	return s
}

// LoadFileDiff returns the diff for a specific file in a commit
func (r *Reader) LoadFileDiff(path string, commitHash string, filePath string) (string, bool, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", false, err
	}

	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return "", false, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return "", false, err
	}

	var changes object.Changes
	if commit.NumParents() == 0 {
		// Initial commit: compare with empty tree
		changes, err = object.DiffTree(nil, commitTree)
		if err != nil {
			return "", false, err
		}
	} else {
		parent, err := commit.Parent(0)
		if err != nil {
			return "", false, err
		}
		parentTree, err := parent.Tree()
		if err != nil {
			return "", false, err
		}
		changes, err = object.DiffTree(parentTree, commitTree)
		if err != nil {
			return "", false, err
		}
	}

	// Find the change for the requested file
	for _, change := range changes {
		var changePath string
		if change.To.Name != "" {
			changePath = change.To.Name
		} else {
			changePath = change.From.Name
		}

		if changePath == filePath {
			patch, err := change.Patch()
			if err != nil {
				return "", false, err
			}
			if patch == nil {
				return "", true, nil // Binary file
			}
			patchStr := patch.String()
			// Check if it's a binary diff
			if strings.Contains(patchStr, "Binary files") {
				return "", true, nil
			}
			return patchStr, false, nil
		}
	}

	return "", false, nil // File not found
}

func (r *Reader) LoadFileChanges(path string, commitHash string) ([]domain.FileChange, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var changes object.Changes
	if commit.NumParents() == 0 {
		// Initial commit: compare with empty tree
		changes, err = object.DiffTree(nil, commitTree)
		if err != nil {
			return nil, err
		}
	} else {
		parent, err := commit.Parent(0)
		if err != nil {
			return nil, err
		}
		parentTree, err := parent.Tree()
		if err != nil {
			return nil, err
		}
		changes, err = object.DiffTree(parentTree, commitTree)
		if err != nil {
			return nil, err
		}
	}

	var result []domain.FileChange
	for _, change := range changes {
		fc := domain.FileChange{}

		// Determine status and path
		action, err := change.Action()
		if err != nil {
			continue
		}

		switch action {
		case merkletrie.Insert:
			fc.Status = domain.FileAdded
			fc.Path = change.To.Name
		case merkletrie.Delete:
			fc.Status = domain.FileDeleted
			fc.Path = change.From.Name
		case merkletrie.Modify:
			fc.Status = domain.FileModified
			fc.Path = change.To.Name
			// Check for rename
			if change.From.Name != change.To.Name {
				fc.Status = domain.FileRenamed
				fc.OldPath = change.From.Name
			}
		}

		// Get line stats
		patch, err := change.Patch()
		if err == nil && patch != nil {
			for _, fileStat := range patch.Stats() {
				fc.Additions += fileStat.Addition
				fc.Deletions += fileStat.Deletion
			}
		}

		result = append(result, fc)
	}

	return result, nil
}
