package domain

type GitReader interface {
	LoadRepository(path string) (*Repository, error)
	LoadCommits(path string, limit int) ([]Commit, error)
	LoadBranches(path string) ([]Branch, error)
}

type RepositoryWatcher interface {
	Watch(path string) (<-chan struct{}, error)
	Stop()
}
