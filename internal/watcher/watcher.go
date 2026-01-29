package watcher

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	repoPath  string
	debounce  time.Duration
	changes   chan struct{}
	stop      chan struct{}
}

func New(repoPath string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fsWatcher: fsw,
		repoPath:  repoPath,
		debounce:  100 * time.Millisecond,
		changes:   make(chan struct{}, 1),
		stop:      make(chan struct{}),
	}

	// Watch .git directory for changes
	gitDir := filepath.Join(repoPath, ".git")
	if err := fsw.Add(gitDir); err != nil {
		fsw.Close()
		return nil, err
	}

	// Also watch refs and HEAD specifically
	fsw.Add(filepath.Join(gitDir, "HEAD"))
	fsw.Add(filepath.Join(gitDir, "refs"))
	fsw.Add(filepath.Join(gitDir, "refs", "heads"))
	fsw.Add(filepath.Join(gitDir, "refs", "remotes"))

	return w, nil
}

func (w *Watcher) Start() {
	var timer *time.Timer

	go func() {
		for {
			select {
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					return
				}
				if isRelevantChange(event) {
					// Debounce: reset timer on each event
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(w.debounce, func() {
						select {
						case w.changes <- struct{}{}:
						default: // Don't block if channel full
						}
					})
				}
			case <-w.fsWatcher.Errors:
				// Ignore errors silently
			case <-w.stop:
				if timer != nil {
					timer.Stop()
				}
				return
			}
		}
	}()
}

func (w *Watcher) Changes() <-chan struct{} {
	return w.changes
}

func (w *Watcher) Stop() {
	close(w.stop)
	w.fsWatcher.Close()
}

func isRelevantChange(e fsnotify.Event) bool {
	// Filter for git-relevant files
	name := filepath.Base(e.Name)
	return name == "HEAD" ||
		strings.Contains(e.Name, "refs/") ||
		name == "index" ||
		strings.HasSuffix(name, ".pack")
}
