package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui"
	"github.com/nogo/gitree/internal/watcher"
)

func main() {
	repoPath := "."

	reader := git.NewReader()
	repo, err := reader.LoadRepository(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create watcher (graceful degradation if fails)
	w, err := watcher.New(repoPath)
	if err != nil {
		// Continue without watching
		w = nil
	} else {
		defer w.Stop()
		w.Start()
	}

	model := tui.NewModel(repo, repoPath, w)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
