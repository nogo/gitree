package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui"
	"github.com/nogo/gitree/internal/version"
	"github.com/nogo/gitree/internal/watcher"
)

func main() {
	// Handle flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("gitree %s\n", version.String())
			return
		case "--check-update":
			checkUpdate()
			return
		case "--help", "-h":
			printUsage()
			return
		}
	}

	repoPath := "."
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
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

	// Show loading message
	fmt.Printf("Loading repository: %s\n", repoPath)

	reader := git.NewReader()
	repo, err := reader.LoadRepository(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d commits, %d branches\n", len(repo.Commits), len(repo.Branches))

	// Create watcher (graceful degradation if fails)
	w, err := watcher.New(repoPath)
	if err != nil {
		// Continue without watching
		w = nil
	} else {
		defer w.Stop()
		w.Start()
	}

	model := tui.NewModel(repo, repoPath, w, reader)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func checkUpdate() {
	fmt.Printf("gitree %s\n", version.String())
	fmt.Println("Checking for updates...")

	release, err := version.CheckLatestRelease(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
		os.Exit(1)
	}

	if version.IsNewer(version.Version, release.Version) {
		fmt.Printf("\nNew version available: %s\n", release.Version)

		// Find asset for current platform
		asset := findPlatformAsset(release.Assets)
		if asset != nil {
			fmt.Printf("Download (%s): %s\n", formatSize(asset.Size), asset.DownloadURL)
		} else {
			fmt.Printf("Release page: %s\n", release.URL)
		}
	} else {
		fmt.Println("You're running the latest version.")
	}
}

func findPlatformAsset(assets []version.ReleaseAsset) *version.ReleaseAsset {
	os := runtime.GOOS
	arch := runtime.GOARCH

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		// Match OS and arch in filename (e.g., gitree_0.3.0_darwin_arm64.tar.gz)
		if strings.Contains(name, os) && strings.Contains(name, arch) {
			return &asset
		}
	}
	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
	)
	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func printUsage() {
	fmt.Println("gitree - TUI git history visualizer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitree [path]           Open repository at path (default: current directory)")
	fmt.Println("  gitree --version, -v    Show version information")
	fmt.Println("  gitree --check-update   Check for new releases")
	fmt.Println("  gitree --help, -h       Show this help message")
}
