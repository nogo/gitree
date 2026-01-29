package domain

import "time"

type Commit struct {
	Hash        string
	ShortHash   string // first 7 chars
	Author      string
	Email       string
	Date        time.Time
	Message     string // first line only
	FullMessage string
	Parents     []string // parent hashes
	BranchRefs  []string // branches pointing here
}

type Branch struct {
	Name     string
	IsRemote bool
	HeadHash string
	Color    string // assigned during rendering
}

type Repository struct {
	Path     string
	Commits  []Commit
	Branches []Branch
	HEAD     string // current HEAD hash or branch name
}
