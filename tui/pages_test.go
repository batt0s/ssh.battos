package tui

import "testing"

func TestPagesUniqueIDsAndContent(t *testing.T) {
	pages := Pages()
	if len(pages) == 0 {
		t.Fatal("Pages() returned no items")
	}

	seen := make(map[string]bool, len(pages))
	for _, item := range pages {
		p, ok := item.(page)
		if !ok {
			t.Fatalf("Pages() returned item of type %T, want page", item)
		}
		if p.id == "" {
			t.Errorf("page %q has an empty id", p.title)
		}
		if p.title == "" {
			t.Errorf("page %q has an empty title", p.id)
		}
		if p.content == "" {
			t.Errorf("page %q has empty content", p.id)
		}
		if seen[p.id] {
			t.Errorf("duplicate page id %q", p.id)
		}
		seen[p.id] = true
	}

	if !seen["home"] {
		t.Error(`Pages() must contain a "home" page`)
	}
}

func TestGetSelectedPage(t *testing.T) {
	got := getSelectedPage("about")
	if got.id != "about" {
		t.Errorf("getSelectedPage(about).id = %q, want %q", got.id, "about")
	}

	if missing := getSelectedPage("does-not-exist"); missing.id != "" {
		t.Errorf("getSelectedPage(unknown).id = %q, want empty", missing.id)
	}
}
