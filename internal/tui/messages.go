package tui

import "github.com/nogo/gitree/internal/domain"

// RepoChangedMsg signals the repository has changed
type RepoChangedMsg struct{}

// RepoLoadedMsg carries refreshed repository data
type RepoLoadedMsg struct {
	Repo *domain.Repository
	Err  error
}

// DiffLoadedMsg carries loaded diff content for a file
type DiffLoadedMsg struct {
	FilePath  string
	Diff      string
	IsBinary  bool
	FileIndex int
	Err       error
}

// ExpandedFilesLoadedMsg carries loaded file changes for expanded commit
type ExpandedFilesLoadedMsg struct {
	Files []domain.FileChange
	Err   error
}
