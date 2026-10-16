package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var p *tea.Program

type focusArea int

const (
	focusList focusArea = iota
	focusContent
)

const (
	// sideBySideMin is the terminal width at or above which the page list and
	// the content are rendered next to each other. Below it they are stacked,
	// keeping the app usable on narrow terminals and phone SSH clients.
	sideBySideMin = 80

	// docStyle frame sizes: border (2) plus margins (4 horizontal, 2 vertical).
	docFrameW = 6
	docFrameH = 4
)

// MainModel is the root Bubble Tea model. It shows a list of pages on one side
// and a scrollable, rendered Markdown page on the other.
type MainModel struct {
	list      list.Model
	viewport  viewport.Model
	page      page
	width     int
	height    int
	focus     focusArea
	renderer  *lipgloss.Renderer
	docStyle  lipgloss.Style
	helpStyle lipgloss.Style
}

// NewMainModel builds the root model. The renderer may be nil, in which case
// lipgloss and glamour fall back to their global, auto-detected defaults.
func NewMainModel(items []list.Item, r *lipgloss.Renderer) MainModel {
	docStyle, helpStyle := stylesFor(r)

	m := MainModel{
		list:      list.New(items, list.NewDefaultDelegate(), 0, 0),
		viewport:  viewport.New(0, 0),
		page:      getSelectedPage("home"),
		renderer:  r,
		docStyle:  docStyle,
		helpStyle: helpStyle,
	}
	m.list.Title = "Pages"
	m.list.SetShowHelp(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{Keymap.Enter, Keymap.Focus, Keymap.Quit}
	}
	m.renderContent()
	return m
}

func stylesFor(r *lipgloss.Renderer) (docStyle, helpStyle lipgloss.Style) {
	newStyle := lipgloss.NewStyle
	if r != nil {
		newStyle = r.NewStyle
	}
	docStyle = newStyle().Margin(1, 2).Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("228")).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(true).
		BorderLeft(true)
	helpStyle = newStyle().Foreground(lipgloss.Color("241"))
	return docStyle, helpStyle
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil
	case tea.KeyMsg:
		// While filtering the list, let it consume all keys.
		if m.focus == focusList && m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keymap.Focus):
			m.toggleFocus()
			return m, nil
		case key.Matches(msg, Keymap.Back):
			m.focus = focusList
			return m, nil
		case key.Matches(msg, Keymap.Enter):
			if m.focus == focusList {
				if selected, ok := m.list.SelectedItem().(page); ok {
					m.page = selected
					m.renderContent()
					m.viewport.GotoTop()
					m.focus = focusContent
				}
			}
			return m, nil
		}
	}

	if m.focus == focusList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *MainModel) toggleFocus() {
	if m.focus == focusList {
		m.focus = focusContent
	} else {
		m.focus = focusList
	}
}

// layout sizes the list and the viewport for the current terminal dimensions,
// choosing a side-by-side or stacked arrangement based on the width.
func (m *MainModel) layout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	bodyH := m.height - lipgloss.Height(m.viewHelp())
	if bodyH < 6 {
		bodyH = 6
	}

	if m.width >= sideBySideMin {
		sidebarW := m.width / 3
		if sidebarW < 28 {
			sidebarW = 28
		}
		if sidebarW > 44 {
			sidebarW = 44
		}
		if sidebarW > m.width/2 {
			sidebarW = m.width / 2
		}
		m.list.SetSize(max(10, sidebarW-docFrameW), max(3, bodyH-docFrameH))
		m.viewport.Width = max(20, m.width-sidebarW)
		m.viewport.Height = bodyH
	} else {
		listH := bodyH / 3
		if listH < 5 {
			listH = 5
		}
		m.list.SetSize(max(10, m.width-docFrameW), max(3, listH-docFrameH))
		m.viewport.Width = max(20, m.width)
		m.viewport.Height = max(3, bodyH-listH)
	}
	m.renderContent()
}

// renderContent re-renders the current page at the viewport's width.
func (m *MainModel) renderContent() {
	width := m.viewport.Width
	if width <= 0 {
		width = 80
	}
	m.viewport.SetContent(renderMarkdown(m.renderer, m.page.content, width))
}

func (m MainModel) View() string {
	sidebar := m.sidebarStyle().Render(m.list.View())
	content := m.viewport.View()

	var body string
	if m.width >= sideBySideMin {
		body = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, sidebar, content)
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.viewHelp())
}

// sidebarStyle highlights the page list when it has focus.
func (m MainModel) sidebarStyle() lipgloss.Style {
	color := lipgloss.Color("240")
	if m.focus == focusList {
		color = lipgloss.Color("228")
	}
	return m.docStyle.BorderForeground(color)
}

// viewHelp returns the contextual key hints for the focused pane.
func (m MainModel) viewHelp() string {
	text := " ↑/k ↓/j: scroll • pgup/pgdn: page • tab: pages • esc: back • q: quit "
	if m.focus == focusList {
		text = " ↑/k ↓/j: select • /: filter • enter: open • tab: content • q: quit "
	}
	return m.helpStyle.Render("\n" + text + "\n")
}

// renderMarkdown renders Markdown for the client's terminal, adapting the
// style to its background and color profile and wrapping to width.
func renderMarkdown(r *lipgloss.Renderer, content string, width int) string {
	if width < 20 {
		width = 80
	}

	style := "dark"
	profile := termenv.TrueColor
	if r != nil {
		profile = r.ColorProfile()
		if !r.HasDarkBackground() {
			style = "light"
		}
	}

	tr, err := glamour.NewTermRenderer(
		glamour.WithColorProfile(profile),
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	out, err := tr.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimSpace(out)
}

// StartTea runs the app locally, outside of an SSH session.
func StartTea() error {
	if f, err := tea.LogToFile("debug.log", ">"); err != nil {
		return err
	} else {
		defer f.Close()
	}

	m := NewMainModel(Pages(), nil)
	p = tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
