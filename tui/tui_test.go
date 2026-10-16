package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) MainModel {
	t.Helper()
	return NewMainModel(Pages(), nil)
}

func resize(t *testing.T, m MainModel, w, ht int) MainModel {
	t.Helper()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: ht})
	got, ok := updated.(MainModel)
	if !ok {
		t.Fatalf("Update returned %T, want MainModel", updated)
	}
	return got
}

func TestLayoutWideIsSideBySide(t *testing.T) {
	m := resize(t, newTestModel(t), 120, 40)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.viewport.Width <= 0 || m.viewport.Height <= 0 {
		t.Fatalf("viewport not sized: %dx%d", m.viewport.Width, m.viewport.Height)
	}
	// Side-by-side leaves room for a sidebar, so the content is narrower
	// than the terminal.
	if m.viewport.Width >= m.width {
		t.Errorf("viewport.Width = %d, want < terminal width %d", m.viewport.Width, m.width)
	}
}

func TestLayoutNarrowStacks(t *testing.T) {
	m := resize(t, newTestModel(t), 60, 40)

	if m.viewport.Width != m.width {
		t.Errorf("stacked viewport.Width = %d, want full width %d", m.viewport.Width, m.width)
	}
	if m.viewport.Height >= m.height {
		t.Errorf("stacked viewport.Height = %d, want < height %d", m.viewport.Height, m.height)
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	for _, w := range []int{40, 80, 140} {
		m := resize(t, newTestModel(t), w, 30)
		if out := m.View(); out == "" {
			t.Errorf("View() at width %d returned empty output", w)
		}
	}
}

func TestEnterOpensSelectedPage(t *testing.T) {
	m := resize(t, newTestModel(t), 120, 40)

	// Move the selection to the second page, then open it.
	m.list.Select(1)
	want := m.list.SelectedItem().(page)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(MainModel)

	if got.page.id != want.id {
		t.Errorf("after enter, page = %q, want %q", got.page.id, want.id)
	}
	if got.focus != focusContent {
		t.Errorf("after enter, focus = %v, want focusContent", got.focus)
	}
}

func TestTabTogglesFocus(t *testing.T) {
	m := resize(t, newTestModel(t), 120, 40)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(MainModel)
	if m.focus != focusContent {
		t.Fatalf("after tab, focus = %v, want focusContent", m.focus)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(MainModel)
	if m.focus != focusList {
		t.Errorf("after esc, focus = %v, want focusList", m.focus)
	}
}

func TestRenderMarkdown(t *testing.T) {
	out := renderMarkdown(nil, "# Hello\n\nSome **bold** text.", 60)
	if out == "" {
		t.Fatal("renderMarkdown returned empty output")
	}
	if !strings.Contains(out, "bold") {
		t.Errorf("renderMarkdown output missing content: %q", out)
	}
}
