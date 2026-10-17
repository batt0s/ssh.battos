package tui

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type reposMsg struct {
	repos []githubRepo
	err   error
}

type treeMsg struct {
	repo    string
	entries []githubTreeEntry
	err     error
}

type fileMsg struct {
	repo string
	path string
	data []byte
	err  error
}

type repoItem struct {
	repo githubRepo
}

func (r repoItem) Title() string { return r.repo.Name }

func (r repoItem) Description() string {
	desc := strings.TrimSpace(r.repo.Description)
	if desc == "" {
		desc = "No description"
	}
	if r.repo.Fork {
		desc += " · fork"
	}
	if r.repo.Archived {
		desc += " · archived"
	}
	return desc
}

func (r repoItem) FilterValue() string {
	return r.repo.Name + " " + r.repo.Description + " " + r.repo.Language
}

type fileItem struct {
	entry githubTreeEntry
	dir   bool
}

func (f fileItem) Title() string {
	name := path.Base(f.entry.Path)
	if f.dir {
		return name + "/"
	}
	return name
}

func (f fileItem) Description() string {
	if f.dir {
		return "directory"
	}
	if f.entry.Size > 0 {
		return humanSize(f.entry.Size)
	}
	return "file"
}

func (f fileItem) FilterValue() string { return f.entry.Path }

// --- Commands -------------------------------------------------------------

func (m MainModel) loadReposCmd() tea.Cmd {
	client := m.gh
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		repos, err := client.repos(ctx)
		return reposMsg{repos: repos, err: err}
	}
}

func (m MainModel) loadTreeCmd(repo githubRepo) tea.Cmd {
	client := m.gh
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		tree, err := client.tree(ctx, repo)
		return treeMsg{repo: repoName(repo), entries: tree.Tree, err: err}
	}
}

func (m MainModel) loadFileCmd(filePath string) tea.Cmd {
	client := m.gh
	repo := m.repo
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		data, err := client.file(ctx, repo, filePath)
		return fileMsg{repo: repoName(repo), path: filePath, data: data, err: err}
	}
}

// --- Navigation -----------------------------------------------------------

// openSelected handles enter for the active sidebar. Pages open like before,
// repositories load their file tree, files open in the viewport and directories
// become the current tree level.
func (m MainModel) openSelected() (tea.Model, tea.Cmd) {
	if m.focus != focusList {
		return m, nil
	}

	switch m.mode {
	case viewPages:
		selected, ok := m.list.SelectedItem().(page)
		if !ok {
			return m, nil
		}
		m.page = selected
		if selected.id == "github" {
			m.enterRepos()
			return m, m.loadReposCmd()
		}
		m.renderContent()
		m.viewport.GotoTop()
		m.focus = focusContent

	case viewRepos:
		selected, ok := m.list.SelectedItem().(repoItem)
		if !ok {
			return m, nil
		}
		m.enterRepo(selected.repo)
		return m, m.loadTreeCmd(selected.repo)

	case viewRepo:
		selected, ok := m.list.SelectedItem().(fileItem)
		if !ok {
			return m, nil
		}
		if selected.dir {
			m.repoPath = selected.entry.Path
			m.filePath = ""
			m.fileData = nil
			m.fileErr = ""
			m.treeErr = ""
			m.loading = false
			return m, m.rebuildFileList(false)
		}
		return m, m.loadSelectedFile(true)
	}
	return m, nil
}

// goBack walks one level up: content -> list, file directory -> parent,
// repository -> repository list, repository list -> portfolio pages.
func (m MainModel) goBack() (tea.Model, tea.Cmd) {
	// On the static pages, esc first returns focus to the sidebar. Inside the
	// GitHub browser it always walks one level up, even from the content pane.
	if m.focus == focusContent && m.mode == viewPages {
		m.focus = focusList
		return m, nil
	}

	switch m.mode {
	case viewRepo:
		if m.repoPath != "" {
			m.repoPath = parentPath(m.repoPath)
			m.filePath = ""
			m.fileData = nil
			m.fileErr = ""
			m.loading = false
			m.treeErr = ""
			return m, m.rebuildFileList(false)
		}
		m.enterReposView()
		return m, nil
	case viewRepos:
		m.enterPages()
		return m, nil
	}
	return m, nil
}

func (m *MainModel) enterPages() {
	m.mode = viewPages
	m.focus = focusList
	m.loading = false
	m.list.Title = "Pages"

	items := Pages()
	m.list.ResetFilter()
	m.list.SetItems(items)

	selected := 0
	for i, item := range items {
		if p, ok := item.(page); ok && p.id == m.page.id {
			selected = i
			break
		}
	}
	m.list.Select(selected)
	m.renderContent()
	m.viewport.GotoTop()
}

func (m *MainModel) enterRepos() {
	m.mode = viewRepos
	m.focus = focusList
	m.loading = true
	m.reposErr = ""
	m.repo = githubRepo{}
	m.tree = nil
	m.repoPath = ""
	m.filePath = ""
	m.fileData = nil
	m.fileErr = ""
	m.treeErr = ""

	m.list.Title = "Repositories"
	m.list.ResetFilter()
	m.list.SetItems(nil)
	m.renderContent()
	m.viewport.GotoTop()
}

func (m *MainModel) enterReposView() {
	previous := repoName(m.repo)
	m.mode = viewRepos
	m.focus = focusList
	m.loading = false
	m.repo = githubRepo{}
	m.tree = nil
	m.repoPath = ""
	m.filePath = ""
	m.fileData = nil
	m.fileErr = ""
	m.treeErr = ""

	m.setRepoItems()
	if previous != "" {
		m.selectRepo(previous)
	}
	m.renderContent()
	m.viewport.GotoTop()
}

func (m *MainModel) enterRepo(repo githubRepo) {
	m.mode = viewRepo
	m.focus = focusList
	m.repo = repo
	m.tree = nil
	m.repoPath = ""
	m.filePath = ""
	m.fileData = nil
	m.fileErr = ""
	m.treeErr = ""
	m.loading = true

	m.list.Title = repo.Name
	m.list.ResetFilter()
	m.list.SetItems(nil)
	m.renderContent()
	m.viewport.GotoTop()
}

func (m *MainModel) setRepoItems() {
	items := make([]list.Item, 0, len(m.repos))
	for _, repo := range m.repos {
		items = append(items, repoItem{repo: repo})
	}
	m.list.Title = "Repositories"
	m.list.ResetFilter()
	m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
	}
}

func (m *MainModel) selectRepo(name string) {
	for i, item := range m.list.Items() {
		if repo, ok := item.(repoItem); ok && repoName(repo.repo) == name {
			m.list.Select(i)
			return
		}
	}
}

// rebuildFileList rebuilds the sidebar from the current tree level. At the
// repository root it selects README (if present) and starts loading it.
func (m *MainModel) rebuildFileList(selectReadme bool) tea.Cmd {
	entries := directEntries(m.tree, m.repoPath)
	items := make([]list.Item, 0, len(entries))
	for _, entry := range entries {
		dir := entry.Type == "tree"
		items = append(items, fileItem{entry: entry, dir: dir})
	}

	m.list.Title = m.repo.Name
	if m.repoPath != "" {
		m.list.Title = path.Base(m.repoPath)
	}
	m.list.ResetFilter()
	m.list.SetItems(items)

	m.fileData = nil
	m.fileErr = ""
	m.treeErr = ""
	m.loading = false
	m.filePath = ""

	if selectReadme && m.repoPath == "" {
		if idx := readmeItemIndex(items); idx >= 0 {
			m.list.Select(idx)
			return m.loadSelectedFile(false)
		}
	}

	if len(items) > 0 {
		m.list.Select(0)
	}
	m.renderContent()
	m.viewport.GotoTop()
	return nil
}

// loadSelectedFile starts an async fetch for the highlighted file. When focus
// is true (an explicit enter) the content pane is focused as well.
func (m *MainModel) loadSelectedFile(focus bool) tea.Cmd {
	selected, ok := m.list.SelectedItem().(fileItem)
	if !ok || selected.dir {
		m.loading = false
		m.renderContent()
		return nil
	}

	m.filePath = selected.entry.Path
	m.fileData = nil
	m.fileErr = ""
	m.loading = true
	if focus {
		m.focus = focusContent
	}
	m.renderContent()
	m.viewport.GotoTop()
	return m.loadFileCmd(selected.entry.Path)
}

// --- Rendering ------------------------------------------------------------

func (m MainModel) paneHeader() string {
	title, subtitle := m.headerText()

	width := m.viewport.Width
	if width <= 0 {
		width = 80
	}

	newStyle := lipgloss.NewStyle
	if m.renderer != nil {
		newStyle = m.renderer.NewStyle
	}
	titleStyle := newStyle().Bold(true).Foreground(lipgloss.Color("228"))
	subtitleStyle := newStyle().Foreground(lipgloss.Color("241"))

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(truncateString(title, width)),
		subtitleStyle.Render(truncateString(subtitle, width)),
	)
}

func (m MainModel) headerText() (title, subtitle string) {
	switch m.mode {
	case viewRepos:
		title = "GitHub"
		subtitle = "@" + m.gh.user
		switch {
		case m.loading:
			subtitle += " · loading repositories…"
		case m.reposErr != "":
			subtitle += " · could not load repositories"
		default:
			subtitle += fmt.Sprintf(" · %d public repositories", len(m.repos))
		}

	case viewRepo:
		title = repoName(m.repo)
		switch {
		case m.loading:
			subtitle = "loading…"
		case m.filePath != "":
			subtitle = m.filePath
			if m.repo.DefaultBranch != "" {
				subtitle += " · " + m.repo.DefaultBranch
			}
		case m.repoPath != "":
			subtitle = m.repoPath + "/"
			if m.repo.DefaultBranch != "" {
				subtitle += " · " + m.repo.DefaultBranch
			}
		case m.repo.Description != "":
			subtitle = m.repo.Description
		default:
			subtitle = "public repository"
			if m.repo.DefaultBranch != "" {
				subtitle += " · " + m.repo.DefaultBranch
			}
		}

	default:
		title = m.page.title
		subtitle = m.page.desc
	}
	return title, subtitle
}

func (m MainModel) reposMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# GitHub\n\nPublic repositories of **@%s**.\n\n", m.gh.user)

	switch {
	case m.loading:
		b.WriteString("_Loading repositories…_\n")
	case m.reposErr != "":
		fmt.Fprintf(&b, "> **Could not load repositories.**\n>\n> %s\n", m.reposErr)
	case len(m.repos) == 0:
		b.WriteString("_No public repositories found._\n")
	default:
		b.WriteString("Select a repository from the sidebar to browse its README and files.\n")
	}

	b.WriteString("\nPress `esc` to return to the portfolio pages.\n")
	return b.String()
}

func (m MainModel) repoContent(width int) string {
	switch {
	case m.treeErr != "":
		return renderMarkdown(m.renderer, m.style, errorMarkdown("Could not load repository", m.treeErr), width)
	case m.fileErr != "":
		return renderMarkdown(m.renderer, m.style, errorMarkdown("Could not load file", m.fileErr), width)
	case m.loading:
		return renderMarkdown(m.renderer, m.style, "_Loading…_", width)
	case m.fileData != nil:
		return renderFile(m.renderer, m.style, m.filePath, m.fileData, width)
	default:
		return renderMarkdown(m.renderer, m.style, m.directoryMarkdown(), width)
	}
}

func renderFile(r *lipgloss.Renderer, style, filePath string, data []byte, width int) string {
	if isBinary(data) {
		md := fmt.Sprintf("# %s\n\n_Binary file — preview not available._\n\n`%s`",
			path.Base(filePath), humanSize(len(data)))
		return renderMarkdown(r, style, md, width)
	}

	text := string(data)
	lines := strings.Split(text, "\n")
	truncated := false
	if len(lines) > maxFileLines {
		lines = lines[:maxFileLines]
		truncated = true
	}
	text = strings.Join(lines, "\n")

	var md string
	if isMarkdownPath(filePath) {
		md = text
	} else {
		fence := fenceFor(text)
		md = fence + languageFor(filePath) + "\n" + text + "\n" + fence
	}
	if truncated {
		md += "\n\n_… file truncated (showing first " + fmt.Sprint(maxFileLines) + " lines) …_"
	}
	return renderMarkdown(r, style, md, width)
}

func (m MainModel) directoryMarkdown() string {
	var b strings.Builder
	if m.repoPath == "" {
		fmt.Fprintf(&b, "# %s\n\n", m.repo.Name)
		if m.repo.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", m.repo.Description)
		}
		if m.repo.HTMLURL != "" {
			fmt.Fprintf(&b, "[View on GitHub](%s)\n\n", m.repo.HTMLURL)
		}
		if m.repo.Language != "" {
			fmt.Fprintf(&b, "**Language:** %s\n\n", m.repo.Language)
		}
		b.WriteString("## Files\n\n")
	} else {
		fmt.Fprintf(&b, "# %s/\n\n", m.repoPath)
	}

	entries := directEntries(m.tree, m.repoPath)
	if len(entries) == 0 {
		b.WriteString("_This directory is empty._\n")
		return b.String()
	}
	for _, entry := range entries {
		name := path.Base(entry.Path)
		if entry.Type == "tree" {
			fmt.Fprintf(&b, "- **%s/**\n", name)
		} else {
			fmt.Fprintf(&b, "- %s · %s\n", name, humanSize(entry.Size))
		}
	}
	b.WriteString("\n_Select a file from the sidebar to view it._\n")
	return b.String()
}

func errorMarkdown(title, detail string) string {
	return fmt.Sprintf("# %s\n\n> %s\n\nPress `esc` to go back.", title, detail)
}

func repoName(repo githubRepo) string {
	if repo.FullName != "" {
		return repo.FullName
	}
	return repo.Name
}

// readmeItemIndex returns the index of the first root-level README file, or -1.
func readmeItemIndex(items []list.Item) int {
	for i, item := range items {
		file, ok := item.(fileItem)
		if !ok || file.dir {
			continue
		}
		if strings.HasPrefix(strings.ToLower(path.Base(file.entry.Path)), "readme") {
			return i
		}
	}
	return -1
}

func parentPath(p string) string {
	p = strings.TrimSuffix(p, "/")
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return ""
}

func humanSize(n int) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := int64(n) / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func truncateString(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	if width <= 1 {
		return "…"
	}
	if len(runes) > width-1 {
		runes = runes[:width-1]
	}
	return string(runes) + "…"
}
