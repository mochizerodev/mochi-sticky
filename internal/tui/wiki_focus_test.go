package tui

import (
	"strings"
	"testing"

	"mochi-sticky/internal/wiki"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleWikiKeyTabTogglesFocus(t *testing.T) {
	m := Model{screen: screenWiki, wikiFocus: focusWikiNav}

	next, _ := m.handleWikiKey(tea.KeyMsg{Type: tea.KeyTab})
	updated := next.(Model)
	if updated.wikiFocus != focusWikiContent {
		t.Fatalf("expected content focus after tab, got %v", updated.wikiFocus)
	}

	next, _ = updated.handleWikiKey(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(Model)
	if updated.wikiFocus != focusWikiNav {
		t.Fatalf("expected nav focus after second tab, got %v", updated.wikiFocus)
	}
}

func TestHandleWikiKeyScrollsPreviewWhenContentFocused(t *testing.T) {
	content := strings.Repeat("line\n", 30)
	m := Model{
		screen:         screenWiki,
		wikiFocus:      focusWikiContent,
		wikiItems:      []wikiNavItem{{Kind: wikiItemPage, Slug: "page-1", Title: "Page 1"}},
		wikiPages:      map[string]wiki.Page{"page-1": {Slug: "page-1", Content: content}},
		wikiIndex:      0,
		wikiStatus:     "",
		wikiAction:     0,
		wikiFilterMode: wikiFilterQuery,
	}

	next, _ := m.handleWikiKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := next.(Model)
	if updated.wikiContentOffset != 1 {
		t.Fatalf("expected content offset 1, got %d", updated.wikiContentOffset)
	}
	if updated.wikiIndex != 0 {
		t.Fatalf("expected wiki index unchanged while scrolling content, got %d", updated.wikiIndex)
	}
}

func TestHandleWikiKeyMovesSelectionWhenNavFocused(t *testing.T) {
	m := Model{
		screen:            screenWiki,
		wikiFocus:         focusWikiNav,
		wikiIndex:         0,
		wikiContentOffset: 4,
		wikiItems: []wikiNavItem{
			{Kind: wikiItemPage, Slug: "page-1", Title: "Page 1"},
			{Kind: wikiItemPage, Slug: "page-2", Title: "Page 2"},
		},
	}

	next, _ := m.handleWikiKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := next.(Model)
	if updated.wikiIndex != 1 {
		t.Fatalf("expected wiki index to move to 1, got %d", updated.wikiIndex)
	}
	if updated.wikiContentOffset != 0 {
		t.Fatalf("expected content offset reset on selection change, got %d", updated.wikiContentOffset)
	}
}
