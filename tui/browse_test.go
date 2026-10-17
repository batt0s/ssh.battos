package tui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDirectEntries(t *testing.T) {
	tree := []githubTreeEntry{
		{Path: "README.md", Type: "blob", Size: 12},
		{Path: "cmd", Type: "tree"},
		{Path: "cmd/app/main.go", Type: "blob", Size: 100},
		{Path: "cmd/app/util.go", Type: "blob", Size: 50},
		{Path: "docs", Type: "tree"},
		{Path: "docs/guide.md", Type: "blob", Size: 25},
		{Path: "go.mod", Type: "blob", Size: 20},
	}

	root := directEntries(tree, "")
	var rootNames []string
	for _, entry := range root {
		rootNames = append(rootNames, entry.Path)
	}
	wantRoot := []string{"cmd", "docs", "go.mod", "README.md"}
	if strings.Join(rootNames, ",") != strings.Join(wantRoot, ",") {
		t.Errorf("root entries = %v, want %v", rootNames, wantRoot)
	}

	cmd := directEntries(tree, "cmd")
	if len(cmd) != 1 || cmd[0].Path != "cmd/app" || cmd[0].Type != "tree" {
		t.Errorf("cmd entries = %+v, want one cmd/app directory", cmd)
	}

	app := directEntries(tree, "cmd/app")
	if len(app) != 2 || app[0].Path != "cmd/app/main.go" || app[1].Path != "cmd/app/util.go" {
		t.Errorf("cmd/app entries = %+v", app)
	}
}

func TestReadmeItemIndex(t *testing.T) {
	items := []list.Item{
		fileItem{entry: githubTreeEntry{Path: "go.mod", Type: "blob"}},
		fileItem{entry: githubTreeEntry{Path: "README-dir", Type: "tree"}, dir: true},
		fileItem{entry: githubTreeEntry{Path: "README.md", Type: "blob"}},
	}
	if got := readmeItemIndex(items); got != 2 {
		t.Errorf("readmeItemIndex = %d, want 2", got)
	}
	if got := readmeItemIndex(items[:1]); got != -1 {
		t.Errorf("readmeItemIndex = %d, want -1", got)
	}
}

func TestParentPath(t *testing.T) {
	tests := map[string]string{
		"":                "",
		"cmd":             "",
		"cmd/app":         "cmd",
		"cmd/app/main.go": "cmd/app",
	}
	for in, want := range tests {
		if got := parentPath(in); got != want {
			t.Errorf("parentPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGitHubClientCachesResponses(t *testing.T) {
	var repoCalls, fileCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/batt0s/repos":
			repoCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"name":"demo","full_name":"batt0s/demo","default_branch":"main","language":"Go","updated_at":"2026-01-01T00:00:00Z"}]`))
		case "/repos/batt0s/demo/git/trees/main":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tree":[{"path":"README.md","type":"blob","size":7},{"path":"main.go","type":"blob","size":10}]}`))
		case "/repos/batt0s/demo/contents/README.md":
			fileCalls.Add(1)
			_, _ = w.Write([]byte("# Hello"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &githubClient{
		user:       "batt0s",
		baseURL:    server.URL,
		httpClient: server.Client(),
		cache:      make(map[string]githubCacheEntry),
	}

	repos, err := client.repos(context.Background())
	if err != nil {
		t.Fatalf("repos: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "demo" {
		t.Fatalf("repos = %+v", repos)
	}

	tree, err := client.tree(context.Background(), repos[0])
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree.Tree) != 2 {
		t.Fatalf("tree = %+v", tree.Tree)
	}

	data, err := client.file(context.Background(), repos[0], "README.md")
	if err != nil {
		t.Fatalf("file: %v", err)
	}
	if string(data) != "# Hello" {
		t.Errorf("file = %q, want %q", data, "# Hello")
	}

	// Second reads should hit the TTL cache, not the server.
	if _, err := client.repos(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.file(context.Background(), repos[0], "README.md"); err != nil {
		t.Fatal(err)
	}
	if got := repoCalls.Load(); got != 1 {
		t.Errorf("repo endpoint called %d times, want 1", got)
	}
	if got := fileCalls.Load(); got != 1 {
		t.Errorf("file endpoint called %d times, want 1", got)
	}
}

func TestGitHubBrowserFlow(t *testing.T) {
	m := resize(t, newTestModel(t), 120, 40)

	// Open the GitHub page from the portfolio sidebar.
	for i, item := range m.list.Items() {
		if p, ok := item.(page); ok && p.id == "github" {
			m.list.Select(i)
			break
		}
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(MainModel)
	if m.mode != viewRepos {
		t.Fatalf("after opening GitHub, mode = %v, want viewRepos", m.mode)
	}

	repos := []githubRepo{{
		Name:          "demo",
		FullName:      "batt0s/demo",
		Description:   "A demo repository",
		DefaultBranch: "main",
		Language:      "Go",
	}}
	updated, _ = m.Update(reposMsg{repos: repos})
	m = updated.(MainModel)
	if got := len(m.list.Items()); got != 1 {
		t.Fatalf("repo list has %d items, want 1", got)
	}
	if _, ok := m.list.Items()[0].(repoItem); !ok {
		t.Fatalf("repo item has type %T, want repoItem", m.list.Items()[0])
	}

	// Open the repository and feed it a tree with a README.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(MainModel)
	if m.mode != viewRepo {
		t.Fatalf("after opening repo, mode = %v, want viewRepo", m.mode)
	}

	updated, _ = m.Update(treeMsg{repo: "batt0s/demo", entries: []githubTreeEntry{
		{Path: "README.md", Type: "blob", Size: 9},
		{Path: "main.go", Type: "blob", Size: 20},
	}})
	m = updated.(MainModel)
	if m.filePath != "README.md" {
		t.Errorf("default selected file = %q, want README.md", m.filePath)
	}
	if got := len(m.list.Items()); got != 2 {
		t.Fatalf("file list has %d items, want 2", got)
	}

	// README content arrives asynchronously.
	updated, _ = m.Update(fileMsg{repo: "batt0s/demo", path: "README.md", data: []byte("# Hello")})
	m = updated.(MainModel)
	if !strings.Contains(m.viewport.View(), "Hello") {
		t.Errorf("viewport does not show README content")
	}

	// Esc from the repository returns to the repo list, then to the pages.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(MainModel)
	if m.mode != viewRepos {
		t.Errorf("after esc from repo, mode = %v, want viewRepos", m.mode)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(MainModel)
	if m.mode != viewPages || m.list.Title != "Pages" {
		t.Errorf("after esc from repos, mode = %v title = %q, want viewPages/Pages", m.mode, m.list.Title)
	}
}
