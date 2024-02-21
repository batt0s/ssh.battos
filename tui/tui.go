package tui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	p *tea.Program
)

type MainModel struct {
	list      list.Model
	page      page
	width     int
	height    int
	docStyle  lipgloss.Style
	helpStyle lipgloss.Style
	help      string
}

func NewMainModel(items []list.Item) tea.Model {
	m := MainModel{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = "Pages"
	m.list.SetShowHelp(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keymap.Enter,
			Keymap.Quit,
		}
	}
	m.page = getSelectedPage("home")
	docStyle := lipgloss.NewStyle().Margin(1, 2).Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("228")).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(true).
		BorderLeft(true)
	m.docStyle = docStyle
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	m.helpStyle = helpStyle
	help := m.helpStyle.Render("\n • up/k: previous • down/j: next • enter: select • q/ctrl+c: quit\n")
	m.help = help
	return m
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := m.docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-lipgloss.Height(m.help))
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keymap.Enter):
			selected := m.list.SelectedItem()
			page, ok := selected.(page)
			if !ok {
				panic("selected.(page)")
			}
			m.page = page
		}
	}
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	sidebar := m.docStyle.Render(m.list.View())
	content, _ := glamour.Render(m.page.content, "dark")
	s := lipgloss.JoinHorizontal(lipgloss.Center, sidebar, content)
	s = lipgloss.JoinVertical(lipgloss.Left, s, m.help)
	return s
}

func StartTea() error {
	if f, err := tea.LogToFile("debug.log", ">"); err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	items := Pages()
	menu := NewMainModel(items)
	p = tea.NewProgram(menu, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
