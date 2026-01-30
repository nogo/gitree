package detail

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/git"
)

type Model struct {
	commit       *domain.Commit
	files        []domain.FileChange
	filesLoading bool
	repoPath     string
	width        int
	height       int
}

func New() Model {
	return Model{}
}

func (m *Model) SetCommit(c *domain.Commit) {
	m.commit = c
	m.files = nil
	m.filesLoading = true
}

func (m *Model) SetRepoPath(path string) {
	m.repoPath = path
}

func (m *Model) SetFiles(files []domain.FileChange) {
	m.files = files
	m.filesLoading = false
}

func (m *Model) SetFilesError() {
	m.files = nil
	m.filesLoading = false
}

func (m Model) LoadFilesCmd() tea.Cmd {
	if m.commit == nil || m.repoPath == "" {
		return nil
	}
	hash := m.commit.Hash
	path := m.repoPath
	return func() tea.Msg {
		reader := git.NewReader()
		files, err := reader.LoadFileChanges(path, hash)
		return FileChangesLoadedMsg{Files: files, Err: err}
	}
}

type FileChangesLoadedMsg struct {
	Files []domain.FileChange
	Err   error
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	if m.commit == nil {
		return ""
	}

	c := m.commit

	// Format detail sections
	header := m.renderHeader(c)
	meta := m.renderMeta(c)
	message := m.renderMessage(c)
	parents := m.renderParents(c)
	refs := m.renderRefs(c)
	files := m.renderFiles()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		meta,
		"",
		message,
		"",
		parents,
	)

	if refs != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			refs,
		)
	}

	content = lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		files,
	)

	// Account for border and padding in dimensions
	innerWidth := m.width - 6
	innerHeight := m.height - 4
	if innerWidth < 20 {
		innerWidth = 20
	}
	if innerHeight < 10 {
		innerHeight = 10
	}

	return DetailStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
}

func (m Model) renderHeader(c *domain.Commit) string {
	return HeaderStyle.Render("Commit " + c.Hash)
}

func (m Model) renderMeta(c *domain.Commit) string {
	author := fmt.Sprintf("%s  %s <%s>",
		LabelStyle.Render("Author:"),
		c.Author,
		c.Email,
	)
	date := fmt.Sprintf("%s  %s",
		LabelStyle.Render("Date:"),
		c.Date.Format("Mon Jan 2 15:04:05 2006 -0700"),
	)
	return author + "\n" + date
}

func (m Model) renderMessage(c *domain.Commit) string {
	return MessageStyle.Render(strings.TrimSpace(c.FullMessage))
}

func (m Model) renderParents(c *domain.Commit) string {
	label := LabelStyle.Render("Parents:")
	if len(c.Parents) == 0 {
		return fmt.Sprintf("%s  (none - initial commit)", label)
	}

	// Show short hashes for parents
	shortParents := make([]string, len(c.Parents))
	for i, p := range c.Parents {
		if len(p) > 7 {
			shortParents[i] = ParentStyle.Render(p[:7])
		} else {
			shortParents[i] = ParentStyle.Render(p)
		}
	}
	return fmt.Sprintf("%s  %s", label, strings.Join(shortParents, ", "))
}

func (m Model) renderRefs(c *domain.Commit) string {
	if len(c.BranchRefs) == 0 {
		return ""
	}
	label := LabelStyle.Render("Refs:")
	return fmt.Sprintf("%s  %s", label, strings.Join(c.BranchRefs, ", "))
}

func (m Model) renderFiles() string {
	if m.filesLoading {
		return LabelStyle.Render("Files: Loading...")
	}

	if m.files == nil || len(m.files) == 0 {
		return LabelStyle.Render("Files: (none)")
	}

	// Calculate totals
	totalAdditions := 0
	totalDeletions := 0
	for _, f := range m.files {
		totalAdditions += f.Additions
		totalDeletions += f.Deletions
	}

	// Header line
	header := fmt.Sprintf("Files Changed (%d)", len(m.files))
	stats := fmt.Sprintf("%s %s",
		AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdditions)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", totalDeletions)),
	)
	headerLine := FileSectionStyle.Render(header) + "  " + stats

	// File list (show first 10, then summary)
	maxFiles := 10
	var lines []string
	lines = append(lines, headerLine)

	displayCount := len(m.files)
	if displayCount > maxFiles {
		displayCount = maxFiles
	}

	for i := 0; i < displayCount; i++ {
		f := m.files[i]
		lines = append(lines, m.renderFileLine(f))
	}

	if len(m.files) > maxFiles {
		remaining := len(m.files) - maxFiles
		lines = append(lines, LabelStyle.Render(fmt.Sprintf("  ... and %d more files", remaining)))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderFileLine(f domain.FileChange) string {
	// Status indicator
	var statusStr string
	switch f.Status {
	case domain.FileAdded:
		statusStr = FileAddedStyle.Render("A")
	case domain.FileModified:
		statusStr = FileModifiedStyle.Render("M")
	case domain.FileDeleted:
		statusStr = FileDeletedStyle.Render("D")
	case domain.FileRenamed:
		statusStr = FileRenamedStyle.Render("R")
	case domain.FileCopied:
		statusStr = FileRenamedStyle.Render("C")
	}

	// Path
	path := f.Path
	if f.Status == domain.FileRenamed && f.OldPath != "" {
		path = f.OldPath + " → " + f.Path
	}

	// Stats
	stats := fmt.Sprintf("%s %s",
		AdditionsStyle.Render(fmt.Sprintf("+%d", f.Additions)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", f.Deletions)),
	)

	return fmt.Sprintf("  %s  %s  %s", statusStr, FilePathStyle.Render(path), stats)
}
