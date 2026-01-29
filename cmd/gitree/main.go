package main

import (
	"fmt"
	"os"

	"github.com/nogo/gitree/internal/git"
)

func main() {
	reader := git.NewReader()
	repo, err := reader.LoadRepository(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d commits, %d branches\n",
		len(repo.Commits), len(repo.Branches))
}
