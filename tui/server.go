package tui

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "23234"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	renderer := bubbletea.MakeRenderer(s)
	docStyle := renderer.NewStyle().Margin(1, 2).Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("228")).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(true).
		BorderLeft(true)
	helpStyle := renderer.NewStyle().Foreground(lipgloss.Color("241"))

	items := Pages()

	m := MainModel{
		list:      list.New(items, list.NewDefaultDelegate(), 0, 0),
		docStyle:  docStyle,
		helpStyle: helpStyle,
		help:      helpStyle.Render("\n • up/k: previous • down/j: next • /: filter • enter: select • q/ctrl+c: quit\n"),
		width:     pty.Window.Width,
		height:    pty.Window.Height,
	}
	m.list.Title = "Pages"
	m.list.SetShowHelp(false)

	m.page, _ = items[0].(page)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func SSSHMain() {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_rsa"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Info("Starting SSH Server", "host", host, "port", port)
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH Server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
