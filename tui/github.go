package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	defaultGitHubUser = "batt0s"
	githubAPIBase     = "https://api.github.com"

	maxAPIResponse = 16 << 20  // 16 MiB
	maxFileBytes   = 512 << 10 // 512 KiB rendered, larger files are truncated
	maxFileLines   = 4000

	reposTTL = 5 * time.Minute
	treeTTL  = 10 * time.Minute
	fileTTL  = 30 * time.Minute
)

type githubRepo struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	HTMLURL       string    `json:"html_url"`
	Language      string    `json:"language"`
	UpdatedAt     time.Time `json:"updated_at"`
	Fork          bool      `json:"fork"`
	Archived      bool      `json:"archived"`
}

type githubTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob", "tree" or "commit"
	Size int    `json:"size"`
}

type githubTree struct {
	Tree      []githubTreeEntry `json:"tree"`
	Truncated bool              `json:"truncated"`
}

type githubCacheEntry struct {
	data    []byte
	expires time.Time
}

// githubClient is a small, read-only GitHub API client with an in-memory TTL
// cache. A single client is shared by all SSH sessions so the cache is shared
// too and the API is not hammered once per visitor.
type githubClient struct {
	user       string
	token      string
	baseURL    string
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]githubCacheEntry
}

var (
	sharedGitHubOnce   sync.Once
	sharedGitHubClient *githubClient
)

func defaultGitHubClient() *githubClient {
	sharedGitHubOnce.Do(func() {
		sharedGitHubClient = newGitHubClient()
	})
	return sharedGitHubClient
}

func newGitHubClient() *githubClient {
	return &githubClient{
		user:    envOr("GITHUB_USER", defaultGitHubUser),
		token:   os.Getenv("GITHUB_TOKEN"),
		baseURL: strings.TrimRight(envOr("GITHUB_API_URL", githubAPIBase), "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		cache: make(map[string]githubCacheEntry),
	}
}

// repos returns the user's public repositories, newest activity first.
func (c *githubClient) repos(ctx context.Context) ([]githubRepo, error) {
	var all []githubRepo
	for page := 1; page <= 5; page++ {
		path := fmt.Sprintf("/users/%s/repos?type=public&sort=updated&per_page=100&page=%d",
			url.PathEscape(c.user), page)
		data, err := c.get(ctx, path, "application/vnd.github+json", reposTTL)
		if err != nil {
			return nil, err
		}
		var batch []githubRepo
		if err := json.Unmarshal(data, &batch); err != nil {
			return nil, fmt.Errorf("decode repositories: %w", err)
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}

	// Pin non-forks above forks while keeping GitHub's recency order otherwise.
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Fork != all[j].Fork {
			return !all[i].Fork
		}
		return all[i].UpdatedAt.After(all[j].UpdatedAt)
	})
	return all, nil
}

// tree returns the recursive file tree for the repository's default branch.
func (c *githubClient) tree(ctx context.Context, repo githubRepo) (githubTree, error) {
	branch := repo.DefaultBranch
	if branch == "" {
		branch = "HEAD"
	}
	path := fmt.Sprintf("/repos/%s/git/trees/%s?recursive=1", repoSegments(repo), url.PathEscape(branch))
	data, err := c.get(ctx, path, "application/vnd.github+json", treeTTL)
	if err != nil {
		return githubTree{}, err
	}
	var tree githubTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return githubTree{}, fmt.Errorf("decode repository tree: %w", err)
	}
	return tree, nil
}

// file returns the raw content of a file at path in repo.
func (c *githubClient) file(ctx context.Context, repo githubRepo, path string) ([]byte, error) {
	apiPath := fmt.Sprintf("/repos/%s/contents/%s", repoSegments(repo), escapePath(path))
	data, err := c.get(ctx, apiPath, "application/vnd.github.raw+json", fileTTL)
	if err != nil {
		return nil, err
	}
	if len(data) > maxFileBytes {
		data = data[:maxFileBytes]
		// Do not leave an incomplete UTF-8 rune at the truncation point.
		for i := 0; i < 3 && len(data) > 0 && !utf8.Valid(data); i++ {
			data = data[:len(data)-1]
		}
	}
	return data, nil
}

func repoSegments(repo githubRepo) string {
	full := repo.FullName
	if full == "" {
		full = repo.Name
	}
	parts := strings.SplitN(full, "/", 2)
	if len(parts) == 1 {
		return url.PathEscape(parts[0])
	}
	return url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1])
}

func escapePath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func (c *githubClient) get(ctx context.Context, path, accept string, ttl time.Duration) ([]byte, error) {
	key := path + "\x00" + accept

	c.mu.Lock()
	entry, ok := c.cache[key]
	c.mu.Unlock()
	if ok && time.Now().Before(entry.expires) {
		return entry.data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "ssh.battos")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect to GitHub: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, maxAPIResponse+1))
	if err != nil {
		return nil, fmt.Errorf("read GitHub response: %w", err)
	}
	if len(body) > maxAPIResponse {
		return nil, fmt.Errorf("GitHub response is too large")
	}
	if res.StatusCode != http.StatusOK {
		return nil, githubAPIError(res.StatusCode, body)
	}

	c.mu.Lock()
	c.cache[key] = githubCacheEntry{
		data:    body,
		expires: time.Now().Add(ttl),
	}
	c.mu.Unlock()
	return body, nil
}

func githubAPIError(status int, body []byte) error {
	var payload struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)
	msg := payload.Message
	if msg == "" {
		msg = http.StatusText(status)
	}
	lower := strings.ToLower(msg)
	if status == http.StatusForbidden || status == http.StatusTooManyRequests {
		if strings.Contains(lower, "rate limit") || status == http.StatusTooManyRequests {
			return fmt.Errorf("GitHub API rate limit reached; please try again later")
		}
	}
	if status == http.StatusNotFound {
		return fmt.Errorf("not found on GitHub")
	}
	return fmt.Errorf("GitHub API error: %s (%d)", msg, status)
}

// directEntries reduces a recursive tree to the entries directly under dir,
// with nested directories synthesised when the tree does not include them.
func directEntries(entries []githubTreeEntry, dir string) []githubTreeEntry {
	prefix := ""
	if dir != "" {
		prefix = strings.TrimSuffix(dir, "/") + "/"
	}

	seen := make(map[string]bool)
	var out []githubTreeEntry
	for _, entry := range entries {
		if entry.Type != "blob" && entry.Type != "tree" {
			continue
		}
		if !strings.HasPrefix(entry.Path, prefix) || entry.Path == prefix {
			continue
		}
		rest := strings.TrimPrefix(entry.Path, prefix)
		if rest == "" {
			continue
		}

		if i := strings.IndexByte(rest, '/'); i >= 0 {
			childDir := prefix + rest[:i]
			if !seen[childDir] {
				seen[childDir] = true
				out = append(out, githubTreeEntry{Path: childDir, Type: "tree"})
			}
			continue
		}
		if !seen[entry.Path] {
			seen[entry.Path] = true
			out = append(out, entry)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		iDir := out[i].Type == "tree"
		jDir := out[j].Type == "tree"
		if iDir != jDir {
			return iDir
		}
		return strings.ToLower(out[i].Path) < strings.ToLower(out[j].Path)
	})
	return out
}

func isMarkdownPath(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".md") ||
		strings.HasSuffix(lower, ".markdown") ||
		strings.HasSuffix(lower, ".mdown") ||
		strings.HasSuffix(lower, ".mkd")
}

func isBinary(data []byte) bool {
	if !utf8.Valid(data) {
		return true
	}
	head := data
	if len(head) > 8000 {
		head = head[:8000]
	}
	return strings.IndexByte(string(head), 0) >= 0
}

func languageFor(path string) string {
	lower := strings.ToLower(path)
	ext := ""
	if i := strings.LastIndexByte(lower, '.'); i >= 0 {
		ext = lower[i+1:]
	}
	langs := map[string]string{
		"go": "go", "py": "python", "rs": "rust", "js": "javascript",
		"jsx": "jsx", "ts": "typescript", "tsx": "tsx", "c": "c",
		"h": "c", "cc": "cpp", "cpp": "cpp", "hpp": "cpp", "java": "java",
		"rb": "ruby", "php": "php", "sh": "bash", "bash": "bash",
		"zsh": "zsh", "fish": "fish", "json": "json", "yaml": "yaml",
		"yml": "yaml", "toml": "toml", "xml": "xml", "html": "html",
		"css": "css", "scss": "scss", "sql": "sql", "tex": "latex",
		"lua": "lua", "vim": "vim", "dockerfile": "dockerfile",
		"makefile": "makefile", "cmake": "cmake", "nix": "nix",
	}
	if lang, ok := langs[ext]; ok {
		return lang
	}
	base := lower
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	switch base {
	case "dockerfile":
		return "dockerfile"
	case "makefile":
		return "makefile"
	}
	return ""
}

func fenceFor(text string) string {
	longest, run := 0, 0
	for _, r := range text {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}
	if longest < 3 {
		longest = 3
	}
	return strings.Repeat("`", longest+1)
}
