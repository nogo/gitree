package git

import (
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

	// Sort commits by date (newest first)
	sortCommitsByDate(commits)

	// Apply limit if specified
	if limit > 0 && len(commits) > limit {
		commits = commits[:limit]
	}

	return commits, nil
}

// sortCommitsByDate sorts commits by date descending (newest first)
func sortCommitsByDate(commits []domain.Commit) {
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.After(commits[j].Date)
	})
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
