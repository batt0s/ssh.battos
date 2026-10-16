package tui

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/charmbracelet/wish/recover"
	"golang.org/x/time/rate"
)

const (
	defaultHost    = "0.0.0.0"
	defaultPort    = "23234"
	defaultHostKey = ".ssh/id_ed25519"
)

// Version is the server version, overridable at build time with:
//
//	go build -ldflags "-X github.com/batt0s/ssh.battos/tui.Version=v1.0.0"
var Version = "dev"

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	m := NewMainModel(Pages(), bubbletea.MakeRenderer(s))
	if pty.Window.Width > 0 && pty.Window.Height > 0 {
		m.width, m.height = pty.Window.Width, pty.Window.Height
		m.layout()
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ensureHostKey generates an ed25519 host key at path if none exists yet.
func ensureHostKey(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	_, err := keygen.New(path, keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite())
	return err
}

func SSSHMain() {
	host := envOr("SSH_HOST", defaultHost)
	port := envOr("SSH_PORT", defaultPort)
	hostKey := envOr("SSH_HOST_KEY", defaultHostKey)

	if err := ensureHostKey(hostKey); err != nil {
		log.Fatal("Could not create host key", "path", hostKey, "error", err)
	}

	limiter := ratelimiter.NewRateLimiter(rate.Limit(5), 10, 128)

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKey),
		wish.WithVersion(Version),
		wish.WithIdleTimeout(30*time.Second),
		wish.WithMaxTimeout(10*time.Minute),
		wish.WithMiddleware(
			recover.Middleware(
				ratelimiter.Middleware(limiter),
				activeterm.Middleware(),
				logging.Middleware(),
				bubbletea.Middleware(teaHandler),
			),
		),
	)
	if err != nil {
		log.Fatal("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Info("Starting SSH server", "host", host, "port", port, "version", Version)
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
