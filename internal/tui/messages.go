package tui

import "github.com/nogo/gitree/internal/domain"

// RepoChangedMsg signals the repository has changed
type RepoChangedMsg struct{}

// RepoLoadedMsg carries refreshed repository data
type RepoLoadedMsg struct {
	Repo *domain.Repository
	Err  error
}
