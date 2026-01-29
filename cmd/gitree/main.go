package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui"
	"github.com/nogo/gitree/internal/watcher"
)

func main() {
	repoPath := "."
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	// Expand ~ to home directory
	if strings.HasPrefix(repoPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot resolve home directory: %v\n", err)
			os.Exit(1)
		}
		repoPath = filepath.Join(home, repoPath[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid path: %s\n", repoPath)
		os.Exit(1)
	}
	repoPath = absPath

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
