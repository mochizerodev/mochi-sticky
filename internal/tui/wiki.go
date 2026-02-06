package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/wiki"

	tea "github.com/charmbracelet/bubbletea"
)

type wikiItemKind int

const (
	wikiItemSection wikiItemKind = iota
	wikiItemPage
)

type wikiNavItem struct {
	Kind         wikiItemKind
	Title        string
	Slug         string
	SectionTitle string
	SectionSlug  string
	Depth        int
}

type wikiStateMsg struct {
	nav   []wiki.NavNode
	pages map[string]wiki.Page
}

type wikiExportMsg struct {
	path   string
	status string
}

type wikiFilterMode int

const (
	wikiFilterQuery wikiFilterMode = iota
	wikiFilterTitle
	wikiFilterTags
	wikiFilterSection
)

func loadWikiCmd(root string) tea.Cmd {
	return loadWikiCmdContext(context.Background(), root)
}

func loadWikiCmdContext(ctx context.Context, root string) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		indexPath := filepath.Join(root, "_index.yaml")
		index, indexErr := wiki.LoadIndexContext(ctx, indexPath)
		if indexErr != nil && !errors.Is(indexErr, wiki.ErrIndexNotFound) {
			return errMsg{err: indexErr}
		}

		var pages []wiki.Page
		var err error
		if indexErr == nil {
			pages, err = wiki.ListPagesFromIndexContext(ctx, root, index)
		} else {
			pages, err = wiki.ListPagesContext(ctx, root)
			if err == nil {
				index, err = wiki.GenerateIndexContext(ctx, pages)
			}
		}
		if err != nil {
			return errMsg{err: err}
		}

		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}

		nav, err := wiki.BuildNavTreeContext(ctx, index, pages)
		if err != nil {
			return errMsg{err: err}
		}
		pageMap := make(map[string]wiki.Page, len(pages))
		for _, page := range pages {
			select {
			case <-ctx.Done():
				return errMsg{err: ctx.Err()}
			default:
			}
			slug := strings.TrimSpace(page.Slug)
			if slug == "" && strings.TrimSpace(page.FilePath) != "" {
				slug = wiki.SlugFromPath(root, page.FilePath)
			}
			if slug == "" {
				continue
			}
			page.Slug = slug
			pageMap[slug] = page
		}
		return wikiStateMsg{nav: nav, pages: pageMap}
	}
}

func exportWikiCmdContext(ctx context.Context, baseDir, root, format string, selection wiki.ExportSelection, filters wiki.FilterOptions, includeLinked bool, linkTypes []string) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}

		normalizedFormat := strings.ToLower(strings.TrimSpace(format))
		if normalizedFormat == "" {
			normalizedFormat = "md"
		}
		if normalizedFormat != "md" && normalizedFormat != "pdf" {
			return errMsg{err: fmt.Errorf("unsupported format: %s", format)}
		}

		indexPath := filepath.Join(root, "_index.yaml")
		index, indexErr := wiki.LoadIndex(indexPath)
		if indexErr != nil && !errors.Is(indexErr, wiki.ErrIndexNotFound) {
			return errMsg{err: indexErr}
		}

		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}

		var pages []wiki.Page
		var err error
		if indexErr == nil {
			pages, err = wiki.ListPagesFromIndex(root, index)
		} else {
			pages, err = wiki.ListPages(root)
		}
		if err != nil {
			return errMsg{err: err}
		}

		var manifest []wiki.ManifestEntry
		var baseSection wiki.IndexSection
		var linked []wiki.IndexSection
		if strings.TrimSpace(selection.Page) != "" {
			manifest, err = wiki.BuildPageManifest(root, pages, selection.Page)
		} else if strings.TrimSpace(selection.Section) != "" {
			sectionIndex := index
			if indexErr != nil {
				sectionIndex, err = wiki.GenerateIndex(pages)
				if err != nil {
					return errMsg{err: err}
				}
			}
			if includeLinked {
				baseSection, _, err = wiki.FindSection(sectionIndex, selection.Section)
				if err != nil {
					return errMsg{err: err}
				}
				linked, err = wiki.ResolveLinkedSections(sectionIndex, baseSection, linkTypes)
				if err != nil {
					return errMsg{err: err}
				}
				sections := append([]wiki.IndexSection{baseSection}, linked...)
				manifest, err = wiki.BuildManifestForSections(sectionIndex, pages, sections)
			} else {
				manifest, err = wiki.BuildSectionManifest(sectionIndex, pages, selection.Section)
			}
		} else if indexErr == nil {
			manifest, err = wiki.BuildManifest(index, pages)
		} else {
			manifest = wiki.BuildManifestFromPages(pages)
		}
		if err != nil {
			return errMsg{err: err}
		}
		if len(manifest) == 0 {
			return wikiExportMsg{status: "No pages to export."}
		}

		pageMap := make(map[string]wiki.Page, len(pages))
		for _, page := range pages {
			slug := strings.TrimSpace(page.Slug)
			if slug == "" && strings.TrimSpace(page.FilePath) != "" {
				slug = wiki.SlugFromPath(root, page.FilePath)
			}
			if slug == "" {
				continue
			}
			page.Slug = slug
			pageMap[slug] = page
		}

		manifest = wiki.FilterManifest(manifest, pageMap, filters)
		if len(manifest) == 0 {
			return wikiExportMsg{status: "No pages to export."}
		}

		output := wiki.DefaultExportPath(root, normalizedFormat, selection)
		data, err := wiki.ExportMarkdownContext(ctx, root, manifest)
		if err != nil {
			return errMsg{err: err}
		}
		if normalizedFormat == "md" {
			if err := wiki.WriteExportContext(ctx, output, data); err != nil {
				return errMsg{err: err}
			}
			return wikiExportMsg{path: output}
		}

		if _, err := wiki.WritePDFContext(ctx, data, wiki.PDFOptions{
			Output:  output,
			BaseDir: baseDir,
			TempDir: root,
		}); err != nil {
			return errMsg{err: err}
		}
		return wikiExportMsg{path: output}
	}
}

func buildWikiItems(nav []wiki.NavNode) []wikiNavItem {
	items := make([]wikiNavItem, 0)
	for _, node := range nav {
		items = append(items, wikiNavItem{
			Kind:  wikiItemSection,
			Title: node.Title,
			Slug:  node.Slug,
			Depth: 0,
		})
		for _, page := range node.Pages {
			items = append(items, wikiNavItem{
				Kind:         wikiItemPage,
				Title:        page.Title,
				Slug:         page.Slug,
				SectionTitle: node.Title,
				SectionSlug:  node.Slug,
				Depth:        1,
			})
		}
	}
	return items
}

func (m Model) wikiRoot() string {
	if m.repo != nil {
		return filepath.Join(m.repo.StorageRoot(), "wiki")
	}
	return filepath.Join(m.baseDir, ".sticky", "wiki")
}

func (m Model) currentWikiSelection() (wikiNavItem, bool) {
	if len(m.wikiItems) == 0 || m.wikiIndex < 0 || m.wikiIndex >= len(m.wikiItems) {
		return wikiNavItem{}, false
	}
	return m.wikiItems[m.wikiIndex], true
}

func (m Model) selectedWikiExportSelection() (wiki.ExportSelection, bool) {
	item, ok := m.currentWikiSelection()
	if !ok {
		return wiki.ExportSelection{}, false
	}
	switch item.Kind {
	case wikiItemPage:
		return wiki.ExportSelection{Page: item.Slug}, true
	case wikiItemSection:
		section := strings.TrimSpace(item.Slug)
		if section == "" {
			section = strings.TrimSpace(item.Title)
		}
		return wiki.ExportSelection{Section: section}, true
	default:
		return wiki.ExportSelection{}, false
	}
}

func (m Model) handleWikiKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "E":
		return m.startWikiExport("md", wiki.ExportSelection{})
	case "P":
		return m.startWikiExport("pdf", wiki.ExportSelection{})
	}

	switch normalizedKey(msg) {
	case "esc", "q", "b":
		m = m.cancelInFlight()
		m.screen = screenBoard
		return m, nil
	case "ctrl+r", "f5":
		m = m.cancelInFlight()
		ctx, cancel := context.WithCancel(context.Background())
		m.inFlightCancel = cancel
		m.loading = true
		m.loadingMessage = "Loading wiki..."
		return m, loadWikiCmdContext(ctx, m.wikiRoot())
	case "j":
		m.wikiIndex++
		m.wikiIndex = clampIndex(m.wikiIndex, len(m.wikiItems))
		return m, nil
	case "k":
		m.wikiIndex--
		m.wikiIndex = clampIndex(m.wikiIndex, len(m.wikiItems))
		return m, nil
	case "/", "f":
		m.screen = screenWikiFilterMenu
		m.wikiAction = 0
		return m, nil
	case "c":
		m.wikiQuery = ""
		m.wikiFilterTitle = ""
		m.wikiFilterSection = ""
		m.wikiFilterTags = nil
		m.applyWikiFilters()
		m.wikiStatus = "Filters cleared."
		return m, nil
	case "enter":
		if item, ok := m.currentWikiSelection(); ok && item.Kind == wikiItemPage {
			return m.startWikiPager(item.Slug)
		}
		return m, nil
	case "e":
		if item, ok := m.currentWikiSelection(); ok && item.Kind == wikiItemPage {
			return m.startWikiEdit(item.Slug)
		}
		return m, nil
	case "x":
		m.screen = screenWikiActions
		m.wikiAction = 0
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) startWikiExport(format string, selection wiki.ExportSelection) (tea.Model, tea.Cmd) {
	return m.startWikiExportWithLinks(format, selection, false)
}

func (m Model) startWikiExportWithLinks(format string, selection wiki.ExportSelection, includeLinked bool) (tea.Model, tea.Cmd) {
	m = m.cancelInFlight()
	ctx, cancel := context.WithCancel(context.Background())
	m.inFlightCancel = cancel
	m.loading = true
	m.loadingMessage = "Exporting wiki..."
	m.wikiStatus = ""
	return m, exportWikiCmdContext(ctx, m.baseDir, m.wikiRoot(), format, selection, m.wikiFilterOptions(), includeLinked, nil)
}

func (m Model) startWikiEdit(slug string) (tea.Model, tea.Cmd) {
	page, ok := m.wikiPages[slug]
	if !ok || strings.TrimSpace(page.FilePath) == "" {
		m.wikiStatus = "Unable to locate page on disk."
		return m, nil
	}
	m = m.cancelInFlight()
	m.loading = true
	m.loadingMessage = "Opening editor..."
	return m, openWikiEditorCmd(m.wikiRoot(), page.FilePath, m.editor)
}

func (m Model) startWikiPager(slug string) (tea.Model, tea.Cmd) {
	page, ok := m.wikiPages[slug]
	if !ok || strings.TrimSpace(page.FilePath) == "" {
		m.wikiStatus = "Unable to locate page on disk."
		return m, nil
	}
	m = m.cancelInFlight()
	m.loading = true
	m.loadingMessage = "Opening pager..."
	return m, openWikiPagerCmd(m.wikiRoot(), page.FilePath)
}

func (m Model) handleWikiActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := m.wikiActions()
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenWiki
		return m, nil
	case "j":
		m.wikiAction++
		m.wikiAction = clampIndex(m.wikiAction, len(actions))
		return m, nil
	case "k":
		m.wikiAction--
		m.wikiAction = clampIndex(m.wikiAction, len(actions))
		return m, nil
	case "enter":
		return m.handleWikiActionSelection(actions)
	default:
		return m, nil
	}
}

func (m Model) handleWikiActionSelection(actions []wikiAction) (tea.Model, tea.Cmd) {
	m.screen = screenWiki
	if len(actions) == 0 || m.wikiAction < 0 || m.wikiAction >= len(actions) {
		return m, nil
	}
	switch actions[m.wikiAction].kind {
	case wikiActionExportMD:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("md", selection, false)
		}
		m.wikiStatus = "Select a page or section to export."
		return m, nil
	case wikiActionExportPDF:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("pdf", selection, false)
		}
		m.wikiStatus = "Select a page or section to export."
		return m, nil
	case wikiActionExportSectionMD:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("md", selection, false)
		}
		m.wikiStatus = "Select a section to export."
		return m, nil
	case wikiActionExportSectionPDF:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("pdf", selection, false)
		}
		m.wikiStatus = "Select a section to export."
		return m, nil
	case wikiActionExportSectionLinkedMD:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("md", selection, true)
		}
		m.wikiStatus = "Select a section to export."
		return m, nil
	case wikiActionExportSectionLinkedPDF:
		if selection, ok := m.selectedWikiExportSelection(); ok {
			return m.startWikiExportWithLinks("pdf", selection, true)
		}
		m.wikiStatus = "Select a section to export."
		return m, nil
	case wikiActionEdit:
		if item, ok := m.currentWikiSelection(); ok && item.Kind == wikiItemPage {
			return m.startWikiEdit(item.Slug)
		}
		m.wikiStatus = "Select a page to edit."
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleWikiFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenWiki
		m.wikiFilterInput = ""
		return m, nil
	case tea.KeyEnter:
		input := strings.TrimSpace(m.wikiFilterInput)
		switch m.wikiFilterMode {
		case wikiFilterQuery:
			m.wikiQuery = input
		case wikiFilterTitle:
			m.wikiFilterTitle = input
		case wikiFilterTags:
			m.wikiFilterTags = parseFilterTags(input)
		case wikiFilterSection:
			m.wikiFilterSection = input
		}
		m.wikiFilterInput = ""
		m.applyWikiFilters()
		m.screen = screenWiki
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.wikiFilterInput) > 0 {
			m.wikiFilterInput = m.wikiFilterInput[:len(m.wikiFilterInput)-1]
		}
		return m, nil
	case tea.KeySpace:
		m.wikiFilterInput += " "
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.wikiFilterInput += string(msg.Runes)
		}
		return m, nil
	}
}

func (m Model) handleWikiFilterMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenWiki
		return m, nil
	case "j":
		m.wikiAction++
		m.wikiAction = clampIndex(m.wikiAction, len(wikiFilterMenuItems()))
		return m, nil
	case "k":
		m.wikiAction--
		m.wikiAction = clampIndex(m.wikiAction, len(wikiFilterMenuItems()))
		return m, nil
	case "enter":
		return m.handleWikiFilterMenuSelection(wikiFilterMenuItems())
	default:
		return m, nil
	}
}

func (m Model) handleWikiFilterMenuSelection(items []wikiFilterMenuItem) (tea.Model, tea.Cmd) {
	if len(items) == 0 || m.wikiAction < 0 || m.wikiAction >= len(items) {
		m.screen = screenWiki
		return m, nil
	}
	item := items[m.wikiAction]
	switch item.kind {
	case wikiFilterMenuQuery:
		m.screen = screenWikiFilter
		m.wikiFilterMode = wikiFilterQuery
		m.wikiFilterInput = m.wikiQuery
		return m, nil
	case wikiFilterMenuTitle:
		m.screen = screenWikiFilter
		m.wikiFilterMode = wikiFilterTitle
		m.wikiFilterInput = m.wikiFilterTitle
		return m, nil
	case wikiFilterMenuTags:
		m.screen = screenWikiFilter
		m.wikiFilterMode = wikiFilterTags
		m.wikiFilterInput = strings.Join(m.wikiFilterTags, ", ")
		return m, nil
	case wikiFilterMenuSection:
		m.screen = screenWikiFilter
		m.wikiFilterMode = wikiFilterSection
		m.wikiFilterInput = m.wikiFilterSection
		return m, nil
	case wikiFilterMenuClear:
		m.wikiQuery = ""
		m.wikiFilterTitle = ""
		m.wikiFilterSection = ""
		m.wikiFilterTags = nil
		m.applyWikiFilters()
		m.wikiStatus = "Filters cleared."
		m.screen = screenWiki
		return m, nil
	default:
		m.screen = screenWiki
		return m, nil
	}
}

func (m *Model) applyWikiFilters() {
	items := buildWikiItems(m.wikiNav)
	if len(items) == 0 || len(m.wikiPages) == 0 {
		m.wikiItems = items
		m.wikiIndex = clampIndex(m.wikiIndex, len(m.wikiItems))
		return
	}

	filterOptions := m.wikiFilterOptions()
	pages := make([]wiki.Page, 0, len(m.wikiPages))
	for _, page := range m.wikiPages {
		pages = append(pages, page)
	}
	filtered := wiki.FilterPages(pages, filterOptions)
	allowed := make(map[string]struct{}, len(filtered))
	for _, page := range filtered {
		slug := strings.TrimSpace(page.Slug)
		if slug == "" && strings.TrimSpace(page.FilePath) != "" {
			slug = wiki.SlugFromPath(m.wikiRoot(), page.FilePath)
		}
		if slug == "" {
			continue
		}
		allowed[slug] = struct{}{}
	}

	filteredItems := make([]wikiNavItem, 0)
	for _, node := range m.wikiNav {
		pagesInSection := make([]wiki.PageRef, 0, len(node.Pages))
		for _, page := range node.Pages {
			if _, ok := allowed[page.Slug]; ok {
				pagesInSection = append(pagesInSection, page)
			}
		}
		if len(pagesInSection) == 0 {
			continue
		}
		filteredItems = append(filteredItems, wikiNavItem{
			Kind:  wikiItemSection,
			Title: node.Title,
			Slug:  node.Slug,
			Depth: 0,
		})
		for _, page := range pagesInSection {
			filteredItems = append(filteredItems, wikiNavItem{
				Kind:         wikiItemPage,
				Title:        page.Title,
				Slug:         page.Slug,
				SectionTitle: node.Title,
				SectionSlug:  node.Slug,
				Depth:        1,
			})
		}
	}

	m.wikiItems = filteredItems
	m.wikiIndex = clampIndex(m.wikiIndex, len(m.wikiItems))
}

func (m Model) wikiFilterOptions() wiki.FilterOptions {
	tagMode := m.wikiFilterTagMode
	if strings.TrimSpace(tagMode) == "" {
		tagMode = "any"
	}
	return wiki.FilterOptions{
		Title:           m.wikiFilterTitle,
		Tags:            m.wikiFilterTags,
		TagMode:         tagMode,
		Section:         m.wikiFilterSection,
		Query:           m.wikiQuery,
		CaseInsensitive: true,
	}
}

func parseFilterTags(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

type wikiFilterMenuKind int

const (
	wikiFilterMenuQuery wikiFilterMenuKind = iota
	wikiFilterMenuTitle
	wikiFilterMenuTags
	wikiFilterMenuSection
	wikiFilterMenuClear
	wikiFilterMenuCancel
)

type wikiFilterMenuItem struct {
	label string
	kind  wikiFilterMenuKind
}

func wikiFilterMenuItems() []wikiFilterMenuItem {
	return []wikiFilterMenuItem{
		{label: "search query", kind: wikiFilterMenuQuery},
		{label: "title filter", kind: wikiFilterMenuTitle},
		{label: "tag filter", kind: wikiFilterMenuTags},
		{label: "section filter", kind: wikiFilterMenuSection},
		{label: "clear filters", kind: wikiFilterMenuClear},
		{label: "cancel", kind: wikiFilterMenuCancel},
	}
}

type wikiActionKind int

const (
	wikiActionExportMD wikiActionKind = iota
	wikiActionExportPDF
	wikiActionExportSectionMD
	wikiActionExportSectionPDF
	wikiActionExportSectionLinkedMD
	wikiActionExportSectionLinkedPDF
	wikiActionEdit
	wikiActionCancel
)

type wikiAction struct {
	label string
	kind  wikiActionKind
}

func (m Model) wikiActions() []wikiAction {
	item, ok := m.currentWikiSelection()
	if !ok {
		return []wikiAction{{label: "cancel", kind: wikiActionCancel}}
	}
	if item.Kind == wikiItemSection {
		return []wikiAction{
			{label: "export section (md)", kind: wikiActionExportSectionMD},
			{label: "export section (pdf)", kind: wikiActionExportSectionPDF},
			{label: "export section + linked (md)", kind: wikiActionExportSectionLinkedMD},
			{label: "export section + linked (pdf)", kind: wikiActionExportSectionLinkedPDF},
			{label: "cancel", kind: wikiActionCancel},
		}
	}
	return []wikiAction{
		{label: "export page (md)", kind: wikiActionExportMD},
		{label: "export page (pdf)", kind: wikiActionExportPDF},
		{label: "open in editor", kind: wikiActionEdit},
		{label: "cancel", kind: wikiActionCancel},
	}
}

func openWikiEditorCmd(root, path, editor string) tea.Cmd {
	resolved := resolveEditor(editor)
	parts := strings.Fields(resolved)
	if len(parts) == 0 {
		parts = []string{"nano"}
	}
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err: err}
		}
		return loadWikiCmd(root)()
	})
}

func openWikiPagerCmd(root, path string) tea.Cmd {
	pager := strings.TrimSpace(os.Getenv("PAGER"))
	if pager == "" {
		pager = "less"
	}
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		parts = []string{"less"}
	}
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err: err}
		}
		return loadWikiCmd(root)()
	})
}
