# ssh.battos

An SSH portfolio you can connect to from any terminal — built with
[Go](https://go.dev), [Wish](https://github.com/charmbracelet/wish) and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).

```sh
ssh ssh.battos.dev
```

![demo](demo.gif)

## Features

- **A real TUI over SSH** — a scrollable list of pages rendered with
  [Glamour](https://github.com/charmbracelet/glamour) Markdown.
- **Responsive layout** — list and content sit side by side on wide terminals
  and stack on narrow ones (works on phone SSH clients).
- **Adaptive rendering** — detects the client's terminal background (without
  blocking the UI) and color profile, and renders the Markdown to match.
  Force it with `SSH_THEME=dark` or `SSH_THEME=light`.
- **Scrollable content** — long pages are paged with `j`/`k`, `pgup`/`pgdn`.
- **GitHub browser** — open the GitHub page to browse your public repositories,
  read READMEs, and explore files without leaving the terminal. Repo and file
  data is fetched from the GitHub API and cached in memory.
- **Hardened server** — ed25519 host key generated on first run, per-IP rate
  limiting, idle/max timeouts, and panic recovery.
- **Configurable** — host, port and host key path via environment variables.

## Quick start

Requires Go 1.21+.

```sh
git clone git@github.com:batt0s/ssh.battos.git
cd ssh.battos

make run     # start the SSH server
make local   # or run the TUI locally in your current terminal
```

Then connect:

```sh
ssh localhost -p 23234
```

On first run the server generates an ed25519 host key at `.ssh/id_ed25519`.

## Keys

| Key               | Action                          |
| ----------------- | ------------------------------- |
| `↑`/`k`, `↓`/`j`  | move selection / scroll content |
| `enter`           | open the selected page          |
| `tab`             | switch between list and content |
| `/`               | filter the page list            |
| `esc`             | back one level / to the list    |
| `pgup`/`pgdn`     | page through content            |
| `q` / `ctrl+c`    | quit                            |

## Configuration

| Variable       | Default              | Description                          |
| -------------- | -------------------- | ------------------------------------ |
| `SSH_HOST`     | `0.0.0.0`            | address to bind                      |
| `SSH_PORT`     | `23234`              | port to listen on                    |
| `SSH_HOST_KEY` | `.ssh/id_ed25519`    | host key path                        |
| `SSH_THEME`    | auto                 | force `dark` or `light` Markdown     |
| `GITHUB_USER`  | `batt0s`             | GitHub user whose public repos show  |
| `GITHUB_TOKEN` | none                 | raises API limit from 60 to 5,000/hr |

## Deployment

### Docker

With [Docker Compose](docker-compose.yml) (recommended):

```sh
# first run
docker compose up -d --build

# update to the latest code later
git pull
docker compose up -d --build
```

Or with plain `docker`:

```sh
docker build -t ssh-battos .
docker rm -f ssh-battos
docker run -d --name ssh-battos --restart unless-stopped \
  -p 22:23234 -v ssh-battos-keys:/app/.ssh ssh-battos
```

The named volume keeps the generated host key across container recreations,
so clients don't see a changed host key on every update.

### GitHub token

Unauthenticated GitHub API access is limited to 60 requests/hour per server
IP, shared by every visitor. A token raises that to 5,000/hour.

Create a **fine-grained personal access token** with read-only access to public
repositories (no write permissions are needed), then keep it in a local `.env`
file on the server — **never commit it to git**.

With Docker Compose, Compose reads `.env` automatically:

```sh
cp .env.example .env
chmod 600 .env
$EDITOR .env        # set GITHUB_TOKEN=github_pat_...

docker compose up -d --build
docker compose exec ssh-battos env | grep GITHUB_USER   # sanity check
```

For plain `docker run`, use `--env-file .env`:

```sh
docker run -d --name ssh-battos --restart unless-stopped \
  --env-file .env -p 22:23234 -v ssh-battos-keys:/app/.ssh ssh-battos
```

For systemd, write the token to `/etc/ssh-battos.env` (root-owned, mode 600):

```sh
sudo install -m 600 /dev/null /etc/ssh-battos.env
sudo tee /etc/ssh-battos.env >/dev/null <<'EOF'
GITHUB_TOKEN=github_pat_...
EOF
sudo systemctl restart ssh-battos
```

The service already loads that file via `EnvironmentFile=`. The token is only
used server-side and is never sent to SSH clients.

### systemd

A hardened unit file is provided at
[`deploy/ssh-battos.service`](deploy/ssh-battos.service). Copy the binary to
`/opt/ssh-battos`, adjust the user if needed, then:

```sh
sudo cp deploy/ssh-battos.service /etc/systemd/system/
sudo systemctl enable --now ssh-battos
```

## Development

```sh
make test    # run the test suite
make vet     # static analysis
make fmt     # format the code
make build   # build a versioned binary into bin/
make demo    # record demo.gif with vhs
```

### Project layout

```
main.go              entrypoint (--local to run without SSH)
tui/server.go        Wish SSH server, middleware and configuration
tui/tui.go           root Bubble Tea model, layout and rendering
tui/pages.go         page content
tui/keymap.go        key bindings
tui/github.go        cached GitHub API client
tui/browse.go        repository/file browser state and rendering
```

### Adding a page

Add a `page` to `Pages()` in [`tui/pages.go`](tui/pages.go):

```go
page{
    id:      "uses",
    title:   "Uses",
    desc:    "Tools I use",
    content: "# Uses\n\n...",
}
```

Pages are Markdown, so links, lists, code blocks and tables all render.
