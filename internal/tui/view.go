package tui

import (
	"fmt"
	"sort"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"
	"mochi-sticky/internal/wiki"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
)

var (
	bg                = lipgloss.Color("#141821")
	panelBg           = lipgloss.Color("#1C1F26")
	accent            = lipgloss.Color("#5FB0FF")
	accentSoft        = lipgloss.Color("#5FB0FF")
	textBright        = lipgloss.Color("#F5F7FA")
	textMuted         = lipgloss.Color("#9CA3AF")
	successColor      = lipgloss.Color("#42C47A")
	dangerColor       = lipgloss.Color("#E35D5D")
	infoColor         = accentSoft
	borderColor       = lipgloss.Color("#334155")
	columnBorder      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	sidebarStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	infoBoxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	kanbanStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	headerStyle       = lipgloss.NewStyle().Bold(true).Foreground(accent).Background(panelBg)
	taskStyle         = lipgloss.NewStyle().Foreground(textBright).Background(panelBg)
	selectedTask      = lipgloss.NewStyle().Foreground(lipgloss.Color("#0B1016")).Background(accent).Bold(true)
	activeBoard       = lipgloss.NewStyle().Foreground(accentSoft).Background(panelBg).Bold(true)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
	barStyle          = lipgloss.NewStyle().Background(bg).Foreground(textBright).Bold(true).Padding(0, 1)
	footerStyle       = lipgloss.NewStyle().Background(bg).Foreground(textMuted).Padding(0, 1)
	toastStyle        = lipgloss.NewStyle().Background(panelBg).Foreground(textBright).Padding(0, 1)
	toastInfoStyle    = lipgloss.NewStyle().Foreground(infoColor).Bold(true)
	toastSuccessStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	toastErrorStyle   = lipgloss.NewStyle().Foreground(dangerColor).Bold(true)
	tabActive         = lipgloss.NewStyle().Foreground(lipgloss.Color("#0B1016")).Background(accent).Bold(true).Padding(0, 1)
	tabInactive       = lipgloss.NewStyle().Foreground(textMuted).Padding(0, 1)
)

// View renders the TUI.
func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.loading {
		message := "Loading tasks..."
		if strings.TrimSpace(m.loadingMessage) != "" {
			message = m.loadingMessage
		}
		return m.frame("mochi-sticky", message, "")
	}
	switch m.screen {
	case screenBoardActions:
		return m.viewBoardActions()
	case screenBoardEdit:
		return m.viewBoardEdit()
	case screenBoardDetail:
		return m.viewBoardDetail()
	case screenBoardFilter:
		return m.viewBoardFilter()
	case screenBoardFilterMenu:
		return m.viewBoardFilterMenu()
	case screenBoardSortMenu:
		return m.viewBoardSortMenu()
	case screenConfirm:
		return m.viewConfirm()
	case screenTaskActions:
		return m.viewTaskActions()
	case screenStatusPicker:
		return m.viewStatusPicker()
	case screenTaskCreate:
		return m.viewTaskCreate()
	case screenTaskDetail:
		return m.viewTaskDetail()
	case screenTaskEdit:
		return m.viewTaskEdit()
	case screenArchive:
		return m.viewArchive()
	case screenWiki:
		return m.viewWiki()
	case screenWikiActions:
		return m.viewWikiActions()
	case screenWikiFilter:
		return m.viewWikiFilter()
	case screenWikiFilterMenu:
		return m.viewWikiFilterMenu()
	case screenADR:
		return m.viewADR()
	case screenADRActions:
		return m.viewADRActions()
	case screenADRStatusPicker:
		return m.viewADRStatusPicker()
	case screenADRCreate:
		return m.viewADRCreate()
	case screenADRDetail:
		return m.viewADRDetail()
	default:
	}
	if len(m.columns) == 0 {
		return m.frame("mochi-sticky", "No columns configured.", "")
	}
	return m.renderBoardScreen("")
}

func (m Model) renderBoardScreen(helpOverride string) string {
	header := fmt.Sprintf("Boards ▸ %s ▸ Tasks", m.activeBoardName())
	if m.boardFocus == focusBoards {
		header = "Boards ▸ Boards"
	}
	help := helpOverride
	if strings.TrimSpace(help) == "" {
		help = m.boardHelpText()
	}
	availableHeight := m.bodyHeight(header, help)
	availableWidth := m.width
	contentHeight := availableHeight
	if availableWidth <= 0 {
		return m.frame(header, m.renderBoardMainPanel(0, contentHeight), help)
	}

	switch {
	case availableWidth >= 120:
		const (
			leftWidth = 26
			gap       = 1
			minCenter = 36
		)
		if availableWidth < leftWidth+minCenter+gap {
			return m.frame(header, m.renderBoardTwoPane(availableWidth, contentHeight), help)
		}
		centerWidth := availableWidth - leftWidth - gap
		left := m.renderBoardSidebar(leftWidth, contentHeight)
		center := m.renderBoardCenterPanel(centerWidth, contentHeight)
		body := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), center)
		return m.frame(header, body, help)
	case availableWidth >= 90:
		return m.frame(header, m.renderBoardTwoPane(availableWidth, contentHeight), help)
	default:
		if m.boardFocus == focusBoards {
			return m.frame(header, m.renderBoardSidebar(availableWidth, contentHeight), help)
		}
		return m.frame(header, m.renderBoardMainPanel(availableWidth, contentHeight), help)
	}
}

func (m Model) renderBoardTwoPane(availableWidth, availableHeight int) string {
	const (
		leftWidth = 26
		gap       = 1
	)
	if availableWidth <= leftWidth+gap+24 {
		return m.renderBoardMainPanel(availableWidth, availableHeight)
	}
	centerWidth := availableWidth - leftWidth - gap
	left := m.renderBoardSidebar(leftWidth, availableHeight)
	center := m.renderBoardCenterPanel(centerWidth, availableHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), center)
}

func (m Model) renderBoardCenterPanel(panelWidth, panelHeight int) string {
	if m.boardFocus == focusBoards {
		return m.renderBoardPreviewPanel(panelWidth, panelHeight)
	}
	return m.renderBoardMainPanel(panelWidth, panelHeight)
}

func (m Model) renderBoardMainPanel(panelWidth, panelHeight int) string {
	availableWidth := panelWidth
	if availableWidth < 0 {
		availableWidth = 0
	}
	if m.boardUsesListViewForWidth(availableWidth) {
		return m.renderBoardListPanel(availableWidth, panelHeight)
	}
	boxWidth := boxedContentWidth(availableWidth)

	viewportStart, viewportCount, columnWidth := m.boardKanbanViewport(availableWidth)
	taskIndex := buildTaskIndex(m.columns)
	infoBox := m.renderBoardInfoBox(availableWidth)
	infoBoxHeight := 0
	if infoBox != "" {
		infoBoxHeight = lipgloss.Height(infoBox)
	}
	viewportFooter := m.renderKanbanViewportFooter(viewportStart, viewportCount)
	viewportFooterHeight := 0
	if strings.TrimSpace(viewportFooter) != "" {
		viewportFooterHeight = 1
	}

	spacing := 0
	if infoBox != "" {
		spacing = 1
	}
	contentHeightCap := 0
	if panelHeight > 0 {
		contentHeightCap = panelHeight - infoBoxHeight - spacing - viewportFooterHeight
		if contentHeightCap < 0 {
			contentHeightCap = 0
		}
	}

	kanbanHeight := 0
	if contentHeightCap > 0 {
		kanbanHeight = contentHeightCap
		if kanbanHeight < 3 {
			kanbanHeight = 3 // minimum to show border + one line
		}
	}
	if panelHeight > 0 && kanbanHeight > 0 {
		totalBodyHeight := infoBoxHeight + spacing + kanbanHeight
		if totalBodyHeight > panelHeight {
			overflow := totalBodyHeight - panelHeight
			if overflow > 0 {
				kanbanHeight -= overflow
				if kanbanHeight < 0 {
					kanbanHeight = 0
				}
			}
		}
	}

	columnHeight := 0
	if kanbanHeight >= 3 {
		columnHeight = kanbanHeight - 2
	}

	rendered := make([]string, 0, viewportCount)
	for i := 0; i < viewportCount; i++ {
		columnIndex := viewportStart + i
		column := m.columns[columnIndex]
		rendered = append(rendered, m.renderColumn(column, columnIndex == m.active, columnWidth, i == viewportCount-1, taskIndex, columnHeight))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	kanbanStyleSized := kanbanStyle
	if boxWidth > 0 {
		kanbanStyleSized = kanbanStyleSized.Width(boxWidth)
	}
	if kanbanHeight > 0 {
		kanbanContentHeight := max(0, kanbanHeight-2)
		if kanbanContentHeight > 0 {
			kanbanStyleSized = kanbanStyleSized.Height(kanbanContentHeight)
		}
	}
	kanbanBox := kanbanStyleSized.Render(board)
	sections := make([]string, 0, 3)
	if infoBox != "" {
		sections = append(sections, infoBox)
	}
	if viewportCount > 0 && viewportCount < len(m.columns) {
		sections = append(sections, taskStyle.Render(fmt.Sprintf("Kanban viewport: %d/%d columns visible (h/l to switch)", viewportCount, len(m.columns))))
	}
	sections = append(sections, kanbanBox)
	if strings.TrimSpace(viewportFooter) != "" {
		sections = append(sections, taskStyle.Render(viewportFooter))
	}
	body := strings.Join(sections, "\n")
	if panelHeight > 0 {
		body = clampToHeight(padToHeight(body, panelHeight), panelHeight)
	}
	return body
}

func (m Model) renderKanbanViewportFooter(start, visible int) string {
	total := len(m.columns)
	if total == 0 || visible == 0 || visible >= total {
		return ""
	}
	end := start + visible
	if end > total {
		end = total
	}
	left := " "
	if start > 0 {
		left = "←"
	}
	right := " "
	if end < total {
		right = "→"
	}

	activeLabel := "n/a"
	if m.active >= 0 && m.active < total {
		label := strings.TrimSpace(m.columns[m.active].Title)
		if label == "" {
			label = strings.TrimSpace(m.columns[m.active].Key)
		}
		if label == "" {
			label = "column"
		}
		activeLabel = fmt.Sprintf("%s (%d/%d)", label, m.active+1, total)
	}

	return fmt.Sprintf("%s %d-%d/%d %s  •  Active: %s", left, start+1, end, total, right, activeLabel)
}

func (m Model) boardUsesListViewForWidth(panelWidth int) bool {
	return m.boardListView
}

func (m Model) boardKanbanViewport(panelWidth int) (int, int, int) {
	totalColumns := len(m.columns)
	if totalColumns == 0 {
		return 0, 0, 24
	}
	if panelWidth <= 0 {
		return 0, totalColumns, 30
	}

	const (
		kanbanFrameWidth   = 4
		perColumnAllowance = 4
		minColumnWidth     = 24
		columnGap          = 2
	)
	maxVisible := totalColumns
	for visible := totalColumns; visible >= 1; visible-- {
		columnAreaWidth := panelWidth - kanbanFrameWidth - (perColumnAllowance * visible)
		if columnAreaWidth <= 0 {
			continue
		}
		required := (visible * minColumnWidth) + ((visible - 1) * columnGap)
		if columnAreaWidth >= required {
			maxVisible = visible
			break
		}
		if visible == 1 {
			maxVisible = 1
		}
	}

	columnAreaWidth := panelWidth - kanbanFrameWidth - (perColumnAllowance * maxVisible)
	if columnAreaWidth < 0 {
		columnAreaWidth = 0
	}
	columnWidth := m.columnWidthForCount(columnAreaWidth, maxVisible)
	start := 0
	if maxVisible < totalColumns {
		start = m.active - (maxVisible - 1)
		if start < 0 {
			start = 0
		}
		if start > totalColumns-maxVisible {
			start = totalColumns - maxVisible
		}
	}
	return start, maxVisible, columnWidth
}

func (m Model) renderBoardListPanel(panelWidth, panelHeight int) string {
	entries := m.boardListEntries()
	filterSummary := m.boardFilterSummary()
	sortSummary := m.boardSortSummary()

	lines := []string{
		headerStyle.Render("List view"),
		taskStyle.Render(fmt.Sprintf("Filters [%s]   Sort [%s]   View [List]", filterSummary, sortSummary)),
		taskStyle.Render("F: filters  O: sort  L: kanban"),
		"",
	}

	boxWidth := boxedContentWidth(panelWidth)
	if boxWidth <= 0 {
		boxWidth = 96
	}
	idW := 8
	statusW := 8
	priW := 3
	tagsW := 14
	createdW := 10
	depsW := 12
	fixed := idW + statusW + priW + tagsW + createdW + depsW + (6 * 3)
	titleW := boxWidth - fixed
	if titleW < 18 {
		titleW = 18
	}
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s", idW, "ID", statusW, "Status", priW, "Pri", titleW, "Title", tagsW, "Tags", createdW, "Created", depsW, "Deps")
	lines = append(lines, headerStyle.Render(fitText(header, boxWidth)))

	if len(entries) == 0 {
		lines = append(lines, taskStyle.Render("No tasks"))
	} else {
		taskIndex := buildTaskIndex(m.columns)
		for _, entry := range entries {
			task := entry.Task
			status := strings.TrimSpace(task.Status)
			if status == "" {
				status = "-"
			}
			priority := fmt.Sprintf("P%d", effectivePriority(task.Priority))
			title := fitText(strings.TrimSpace(task.Title), titleW)
			tags := "-"
			if len(task.Tags) > 0 {
				tags = strings.Join(task.Tags, ",")
			}
			tags = fitText(tags, tagsW)
			created := "-"
			if !task.Created.IsZero() {
				created = task.Created.Format("2006-01-02")
			}
			deps := "-"
			if len(task.DependsOn) > 0 {
				deps = strings.Join(task.DependsOn, ",")
			}
			ready, _ := board.IsReady(task, taskIndex)
			if !ready && deps == "-" {
				deps = "blocked"
			}
			deps = fitText(deps, depsW)

			row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s", idW, fitText(task.ID, idW), statusW, fitText(status, statusW), priW, fitText(priority, priW), titleW, title, tagsW, tags, createdW, fitText(created, createdW), depsW, deps)
			row = fitText(row, boxWidth)
			if entry.Ref.ColumnIndex == m.active && entry.Ref.TaskIndex == m.columns[entry.Ref.ColumnIndex].Selected {
				lines = append(lines, selectedTask.Render(row))
			} else {
				lines = append(lines, taskStyle.Render(row))
			}
		}
	}

	body := strings.Join(lines, "\n")
	style := kanbanStyle
	if boxWidth > 0 {
		style = style.Width(boxWidth)
	}
	if panelHeight > 0 {
		contentHeight := max(0, panelHeight-2)
		if contentHeight > 0 {
			style = style.Height(contentHeight)
			body = clampToHeight(padToHeight(body, contentHeight), contentHeight)
		}
	}
	return style.Render(body)
}

func (m Model) boardFilterSummary() string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(m.boardFilterStatus) != "" {
		parts = append(parts, "status:"+m.boardFilterStatus)
	}
	if strings.TrimSpace(m.boardFilterTitle) != "" {
		parts = append(parts, "title:"+m.boardFilterTitle)
	}
	if len(m.boardFilterTags) > 0 {
		parts = append(parts, "tags:"+strings.Join(m.boardFilterTags, ","))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func (m Model) boardSortSummary() string {
	sortBy := strings.TrimSpace(m.boardSortBy)
	if sortBy == "" {
		sortBy = "readiness"
	}
	dir := "asc"
	if m.boardSortDesc {
		dir = "desc"
	}
	return sortBy + " " + dir
}

func (m Model) columnWidthForCount(totalWidth, count int) int {
	if count == 0 {
		return 24
	}
	if totalWidth == 0 {
		return 30
	}
	gap := 2
	available := totalWidth - (gap * (count - 1))
	width := available / count
	if width < 24 {
		return 24
	}
	return width
}

func (m Model) sidebarWidth() int {
	if m.width == 0 {
		return 0
	}
	if m.width < 90 {
		return 0
	}
	return 26
}

func (m Model) renderColumn(column columnModel, active bool, width int, isLast bool, index map[string]board.Task, height int) string {
	title := column.Title
	if strings.TrimSpace(title) == "" {
		title = column.Key
	}
	if column.Key != "" && !strings.EqualFold(title, column.Key) {
		title = fmt.Sprintf("%s (%s)", title, column.Key)
	}

	lines := []string{headerStyle.Render(title)}
	if len(column.Tasks) == 0 {
		lines = append(lines, taskStyle.Render("No tasks"))
	} else {
		for i, task := range column.Tasks {
			ready, unmet := board.IsReady(task, index)
			line := fmt.Sprintf("P%d %s %s", effectivePriority(task.Priority), task.ID, task.Title)
			if !ready {
				line = fmt.Sprintf("%s ⏳ blocked by %s", line, strings.Join(unmet, ","))
			}
			if active && i == column.Selected {
				lines = append(lines, selectedTask.Render(line))
				continue
			}
			lines = append(lines, taskStyle.Render(line))
		}
	}

	contentHeight := 0
	if height > 0 {
		contentHeight = max(0, height-2)
	}
	if contentHeight > 0 && len(lines) > contentHeight {
		overflow := len(lines) - contentHeight
		if contentHeight >= 2 {
			lines = append(lines[:contentHeight-1], taskStyle.Render(fmt.Sprintf("… %d more", overflow)))
		} else {
			lines = lines[:contentHeight]
		}
	}

	style := columnBorder.Width(width)
	if height > 0 {
		contentHeight := max(0, height-2)
		if contentHeight > 0 {
			style = style.Height(contentHeight)
		}
	}
	if active {
		style = style.BorderForeground(accentSoft)
	} else {
		style = style.BorderForeground(borderColor)
	}
	if !isLast {
		style = style.MarginRight(2)
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderContextBlock() string {
	lines := m.contextLines()
	if len(lines) == 0 {
		return ""
	}
	return fmt.Sprintf("%s\n%s", headerStyle.Render("Context"), strings.Join(lines, "\n"))
}

func (m Model) renderBoardInfoBox(width int) string {
	lines := []string{taskStyle.Render(m.boardStatsSummary())}
	lines = append(lines, m.contextLines()...)
	if len(lines) == 0 {
		lines = []string{taskStyle.Render("(empty)")}
	}
	body := fmt.Sprintf("%s\n%s", headerStyle.Render("Board Info"), strings.Join(lines, "\n"))
	style := infoBoxStyle
	if width > 0 {
		style = style.Width(boxedContentWidth(width))
	}
	return style.Render(body)
}

func (m Model) renderBoardPreviewPanel(width, height int) string {
	lines := []string{headerStyle.Render("Board info")}
	selected, ok := m.selectedBoard()
	if !ok {
		lines = append(lines, taskStyle.Render("No board selected"))
	} else {
		preview := m.boardPreview
		if preview.BoardID != selected.ID {
			preview = boardPreviewData{
				BoardID:   selected.ID,
				BoardName: selected.Name,
				Archived:  selected.Archived,
			}
			if !selected.Created.IsZero() {
				preview.Created = selected.Created.Format("2006-01-02")
			}
		}

		name := strings.TrimSpace(preview.BoardName)
		if name == "" {
			name = selected.ID
		}
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Name: %s", name)))
		if preview.Archived {
			lines = append(lines, taskStyle.Render("Status: Archived"))
		} else {
			lines = append(lines, taskStyle.Render("Status: Active"))
		}
		if preview.Loaded {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Tasks: %d (Open %d / Doing %d / Done %d)", preview.Total, preview.Open, preview.Doing, preview.Done)))
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Archived tasks: %d", preview.ArchivedTasks)))
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Last activity: %s", preview.LastActivity)))
		}
		if strings.TrimSpace(preview.Created) != "" {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Created: %s", preview.Created)))
		}
		lines = append(lines, "")
		lines = append(lines, taskStyle.Render(fmt.Sprintf("ID: %s", selected.ID)))
		contextLines := m.contextLinesFor(preview.Context)
		if len(contextLines) > 0 {
			lines = append(lines, contextLines...)
		}
		if m.boardPreviewLoading {
			lines = append(lines, "")
			lines = append(lines, taskStyle.Render("Loading board info..."))
		} else if strings.TrimSpace(preview.Error) != "" {
			lines = append(lines, "")
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Info unavailable: %s", preview.Error)))
		}
	}

	body := strings.Join(lines, "\n")
	style := infoBoxStyle
	if width > 0 {
		style = style.Width(boxedContentWidth(width))
	}
	if height > 0 {
		contentHeight := max(0, height-2)
		if contentHeight > 0 {
			style = style.Height(contentHeight)
		}
	}
	return style.Render(body)
}

func (m Model) boardStatsSummary() string {
	total := 0
	doing := 0
	done := 0
	for _, column := range m.columns {
		count := len(column.Tasks)
		total += count
		key := strings.ToLower(strings.TrimSpace(column.Key))
		switch key {
		case "doing", "in-progress", "in_progress":
			doing += count
		case "done":
			done += count
		}
	}
	open := total - done
	archived := len(m.archived)
	return fmt.Sprintf("Stats  Open:%d  Doing:%d  Done:%d  Archived:%d", open, doing, done, archived)
}

func (m Model) contextLines() []string {
	return m.contextLinesFor(m.boardContext)
}

func (m Model) contextLinesFor(ctx board.BoardContext) []string {
	lines := make([]string, 0)
	addLine := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		lines = append(lines, taskStyle.Render(fmt.Sprintf("%s: %s", label, value)))
	}
	addLine("Scope", ctx.Scope)
	addLine("Release", ctx.Release)
	addLine("Target", ctx.Target)
	if len(ctx.Owners) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Owners: %s", strings.Join(ctx.Owners, ", "))))
	}
	addLine("Notes", ctx.Notes)
	return lines
}

func (m Model) renderBoardSidebar(width, height int) string {
	lines := []string{headerStyle.Render("Boards")}
	if len(m.boards) == 0 {
		lines = append(lines, taskStyle.Render("No boards"))
	} else {
		for i, board := range m.boards {
			label := board.Name
			if strings.TrimSpace(label) == "" {
				label = board.ID
			}
			if board.Archived {
				label = fmt.Sprintf("%s (archived)", label)
			}
			if m.boardFocus == focusBoards && i == m.boardIndex {
				lines = append(lines, selectedTask.Render(label))
				continue
			}
			if board.ID == m.activeBoard {
				lines = append(lines, activeBoard.Render(label))
				continue
			}
			lines = append(lines, taskStyle.Render(label))
		}
	}
	body := strings.Join(lines, "\n")
	style := sidebarStyle
	if width > 0 {
		style = style.Width(boxedContentWidth(width))
	}
	if height > 0 {
		contentHeight := max(0, height-2)
		if contentHeight > 0 {
			style = style.Height(contentHeight)
		}
	}
	return style.Render(body)
}

func (m Model) viewBoardActions() string {
	board, ok := m.selectedBoard()
	title := "Board Actions"
	if ok {
		title = fmt.Sprintf("Board Actions: %s", board.Name)
	}
	lines := []string{headerStyle.Render(title)}
	items := boardActionItems()
	for i, item := range items {
		label := titleCase(item)
		if item == "cancel" {
			label = "Cancel"
		}
		if i == m.boardAction {
			lines = append(lines, selectedTask.Render(label))
			continue
		}
		lines = append(lines, taskStyle.Render(label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderModalOverlay("Board Actions", body, help)
}

func titleCase(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func (m Model) viewBoardEdit() string {
	title := "Create Board"
	if m.editMode == editRename {
		title = fmt.Sprintf("Rename Board: %s", m.editBoardID)
	}
	lines := []string{headerStyle.Render(title), ""}
	lines = append(lines, taskStyle.Render("Name: "+m.editInput))
	body := strings.Join(lines, "\n")
	help := "enter save • esc cancel"
	return m.frame(title, body, help)
}

func (m Model) viewBoardDetail() string {
	if strings.TrimSpace(m.activeBoard) == "" {
		return m.frame("Board Detail", "No board selected.", "esc back")
	}

	boardName := m.activeBoardName()
	title := fmt.Sprintf("Board Detail: %s", boardName)

	lines := []string{
		headerStyle.Render(title),
		"",
		taskStyle.Render(fmt.Sprintf("ID: %s", m.activeBoard)),
		taskStyle.Render(fmt.Sprintf("Name: %s", boardName)),
	}

	for _, board := range m.boards {
		if board.ID != m.activeBoard {
			continue
		}
		if strings.TrimSpace(board.Path) != "" {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Path: %s", board.Path)))
		}
		archived := "false"
		if board.Archived {
			archived = "true"
		}
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Archived: %s", archived)))
		if !board.Created.IsZero() {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Created: %s", board.Created.Format("2006-01-02"))))
		}
		break
	}

	lines = append(lines, "")
	lines = append(lines, taskStyle.Render("Description:"))
	if strings.TrimSpace(m.boardDesc) != "" {
		lines = append(lines, taskStyle.Render(strings.TrimRight(m.boardDesc, "\n")))
	} else {
		lines = append(lines, taskStyle.Render("(empty)"))
	}

	if contextBlock := m.renderContextBlock(); contextBlock != "" {
		lines = append(lines, "")
		lines = append(lines, contextBlock)
	}

	body := strings.Join(lines, "\n")
	helpParts := []string{"e edit", "ctrl+r/F5 refresh", "x actions"}
	if m.sidebarWidth() > 0 {
		helpParts = append(helpParts, "b boards")
	}
	helpParts = append(helpParts, "esc back")
	help := strings.Join(helpParts, " • ")
	return m.frame(title, body, help)
}

func (m Model) viewBoardFilter() string {
	title := "Board Filter"
	prompt := "Status filter"
	switch m.boardFilterMode {
	case boardListFilterTitle:
		prompt = "Title filter"
	case boardListFilterTags:
		prompt = "Tag filter (comma-separated)"
	}
	lines := []string{
		headerStyle.Render(prompt),
		"",
		taskStyle.Render(m.boardFilterInput),
	}
	body := strings.Join(lines, "\n")
	help := "enter apply • esc cancel"
	return m.renderModalOverlay(title, body, help)
}

func (m Model) viewBoardFilterMenu() string {
	items := boardFilterMenuItems()
	lines := make([]string, 0, len(items)+2)
	lines = append(lines, headerStyle.Render("Board Filters"), "")
	for i, item := range items {
		if i == m.boardListAction {
			lines = append(lines, selectedTask.Render(item.label))
			continue
		}
		lines = append(lines, taskStyle.Render(item.label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderModalOverlay("Board Filters", body, help)
}

func (m Model) viewBoardSortMenu() string {
	items := boardSortMenuItems()
	lines := make([]string, 0, len(items)+3)
	lines = append(lines, headerStyle.Render("Board Sort"), "")
	lines = append(lines, taskStyle.Render("Current: "+m.boardSortSummary()))
	lines = append(lines, "")
	for i, item := range items {
		if i == m.boardListAction {
			lines = append(lines, selectedTask.Render(item.label))
			continue
		}
		lines = append(lines, taskStyle.Render(item.label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderModalOverlay("Board Sort", body, help)
}

func (m Model) viewConfirm() string {
	message := "Confirm action?"
	actionHint := "[y] Confirm   [n] Cancel"
	messageStyle := taskStyle
	switch m.confirmAction {
	case confirmArchiveBoard:
		message = fmt.Sprintf("Archive board %q?", m.confirmBoard)
	case confirmDeleteBoard:
		message = fmt.Sprintf("Delete board %q? This cannot be undone.", m.confirmBoard)
		messageStyle = errorStyle
		actionHint = "[y] Delete (danger)   [n] Cancel"
	case confirmArchiveTask:
		message = fmt.Sprintf("Archive task %q?", m.confirmTask)
	case confirmDeleteTask:
		message = fmt.Sprintf("Delete task %q? This cannot be undone.", m.confirmTask)
		messageStyle = errorStyle
		actionHint = "[y] Delete (danger)   [n] Cancel"
	case confirmDeleteADR:
		message = fmt.Sprintf("Delete ADR %s? This cannot be undone.", adr.FormatID(m.confirmADR))
		messageStyle = errorStyle
		actionHint = "[y] Delete (danger)   [n] Cancel"
	}
	lines := []string{
		headerStyle.Render("Confirm"),
		"",
		messageStyle.Render(message),
		"",
		taskStyle.Render(actionHint),
	}
	body := strings.Join(lines, "\n")
	help := "y confirm • n cancel"
	if m.confirmAction == confirmDeleteADR {
		return m.renderADRModalOverlay("Confirm", body, help)
	}
	return m.renderModalOverlay("Confirm", body, help)
}

func (m Model) viewTaskActions() string {
	task, ok := m.currentTask()
	title := "Task Actions"
	if ok {
		title = fmt.Sprintf("Task Actions: %s", task.ID)
	}
	lines := []string{headerStyle.Render(title)}
	items := taskActionItems()
	for i, item := range items {
		label := titleCase(item)
		if i == m.taskAction {
			lines = append(lines, selectedTask.Render(label))
			continue
		}
		lines = append(lines, taskStyle.Render(label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderModalOverlay("Task Actions", body, help)
}

func (m Model) viewStatusPicker() string {
	lines := []string{headerStyle.Render("Pick Status")}
	for i, column := range m.columns {
		label := column.Title
		if strings.TrimSpace(label) == "" {
			label = column.Key
		}
		if i == m.statusIndex {
			lines = append(lines, selectedTask.Render(label))
			continue
		}
		lines = append(lines, taskStyle.Render(label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.frame("Pick Status", body, help)
}

func (m Model) viewTaskCreate() string {
	title := headerStyle.Render("New Task")
	status := m.taskStatus
	if strings.TrimSpace(status) == "" {
		status = "todo"
	}
	lineTitle := "Title: " + m.taskTitle
	linePriority := fmt.Sprintf("Priority: %d", effectivePriority(m.taskPriority))
	lineTags := "Tags: " + m.taskTags
	switch m.taskField {
	case 0:
		lineTitle = selectedTask.Render(lineTitle)
	case 1:
		linePriority = selectedTask.Render(linePriority)
	default:
		lineTags = selectedTask.Render(lineTags)
	}
	lines := []string{
		title,
		"",
		taskStyle.Render("Status: " + status),
		lineTitle,
		linePriority,
		lineTags,
	}
	body := strings.Join(lines, "\n")
	help := "tab switch field • 1-3 set priority • enter save • esc cancel"
	return m.frame("New Task", body, help)
}

func (m Model) viewTaskDetail() string {
	task, ok := m.currentTask()
	if !ok {
		return m.frame("Task Detail", "No task selected.", "esc back")
	}
	boardLabel := board.TaskBoardLabel(task)
	lines := []string{
		headerStyle.Render("Task Detail"),
		"",
		taskStyle.Render(fmt.Sprintf("Board: %s", boardLabel)),
		taskStyle.Render(fmt.Sprintf("ID: %s", task.ID)),
		m.fieldLine("Title", task.Title, fieldTitle),
		m.fieldLine("Status", task.Status, fieldStatus),
		m.fieldLine("Priority", fmt.Sprintf("%d", effectivePriority(task.Priority)), fieldPriority),
		m.fieldLine("Tags", strings.Join(task.Tags, ", "), fieldTags),
	}
	if !task.Created.IsZero() {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Created: %s", task.Created.Format("2006-01-02"))))
	}
	lines = append(lines, "")
	lines = append(lines, m.fieldLine("Description", "", fieldDescription))
	if strings.TrimSpace(task.Content) != "" {
		lines = append(lines, taskStyle.Render(task.Content))
	} else {
		lines = append(lines, taskStyle.Render("(empty)"))
	}
	body := strings.Join(lines, "\n")
	help := "tab next • enter edit • a archive • d delete • e editor • x actions • esc back"
	return m.frame("Task Detail", body, help)
}

func (m Model) viewTaskEdit() string {
	title := "Edit Task"
	switch m.taskEditMode {
	case editTitle:
		title = "Edit Title"
	case editTags:
		title = "Edit Tags"
	case editDescription:
		title = "Edit Description"
	case editPriority:
		title = "Edit Priority"
	}
	lines := []string{
		headerStyle.Render(title),
		"",
		taskStyle.Render(m.taskEditInput),
	}
	body := strings.Join(lines, "\n")
	help := "enter save • esc cancel"
	return m.frame(title, body, help)
}

func (m Model) viewArchive() string {
	lines := []string{headerStyle.Render("Archived Tasks")}
	if len(m.archived) == 0 {
		lines = append(lines, taskStyle.Render("No archived tasks"))
	} else {
		for i, task := range m.archived {
			line := fmt.Sprintf("%s %s", task.ID, task.Title)
			if i == m.archiveIndex {
				lines = append(lines, selectedTask.Render(line))
				continue
			}
			lines = append(lines, taskStyle.Render(line))
		}
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter restore • esc back"
	return m.frame("Archived Tasks", body, help)
}

func (m Model) viewWikiActions() string {
	items := m.wikiActions()
	lines := make([]string, 0, len(items)+2)
	lines = append(lines, headerStyle.Render("Wiki Actions"), "")
	for i, item := range items {
		if i == m.wikiAction {
			lines = append(lines, selectedTask.Render(item.label))
			continue
		}
		lines = append(lines, taskStyle.Render(item.label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderWikiModalOverlay("Wiki Actions", body, help)
}

func (m Model) viewWikiFilter() string {
	title := "Wiki Filter"
	prompt := m.wikiFilterPrompt()
	lines := []string{
		headerStyle.Render(prompt),
		"",
		taskStyle.Render(m.wikiFilterInput),
	}
	body := strings.Join(lines, "\n")
	help := "enter apply • esc cancel"
	return m.renderWikiModalOverlay(title, body, help)
}

func (m Model) viewWikiFilterMenu() string {
	items := wikiFilterMenuItems()
	lines := make([]string, 0, len(items)+2)
	lines = append(lines, headerStyle.Render("Wiki Filters"), "")
	for i, item := range items {
		if i == m.wikiAction {
			lines = append(lines, selectedTask.Render(item.label))
			continue
		}
		lines = append(lines, taskStyle.Render(item.label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderWikiModalOverlay("Wiki Filters", body, help)
}

func (m Model) viewWiki() string {
	header := "Wiki ▸ Navigation"
	help := m.wikiHelpText()
	body := m.renderWikiLayout(header, help)
	return m.frame(header, body, help)
}

func (m Model) renderWikiList() string {
	lines := make([]string, 0, len(m.wikiItems))
	terms := m.wikiQueryTerms()
	for i, item := range m.wikiItems {
		prefix := "  - "
		if item.Kind == wikiItemSection {
			prefix = "> "
		}
		title := item.Title
		if item.Kind == wikiItemPage && len(terms) > 0 {
			title = highlightTerms(title, terms, accentSoft)
		}
		line := prefix + title
		if item.Kind == wikiItemPage {
			if page, ok := m.wikiPages[item.Slug]; ok {
				status := strings.TrimSpace(page.Status)
				if status != "" {
					line = fmt.Sprintf("%s (%s)", line, status)
				}
			}
		}
		if i == m.wikiIndex {
			lines = append(lines, selectedTask.Render(line))
			continue
		}
		if item.Kind == wikiItemSection {
			lines = append(lines, headerStyle.Render(line))
		} else {
			lines = append(lines, taskStyle.Render(line))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderWikiLayout(header, help string) string {
	navContent := m.renderWikiNavContent()
	infoContent := m.renderWikiInfoContent()
	pageContent := m.renderWikiPageContent()

	if m.width <= 0 {
		parts := []string{
			navContent,
			"",
			infoBoxStyle.Render(infoContent),
			"",
			kanbanStyle.Render(pageContent),
		}
		return strings.Join(parts, "\n")
	}

	gap := 1
	navWidth := m.wikiNavWidth(gap)
	rightWidth := m.width - navWidth - gap
	if rightWidth < 0 {
		rightWidth = 0
	}

	infoBoxWidth := boxedContentWidth(rightWidth)
	infoBoxStyleSized := infoBoxStyle
	if infoBoxWidth > 0 {
		infoBoxStyleSized = infoBoxStyleSized.Width(infoBoxWidth)
	}
	infoBox := infoBoxStyleSized.Render(infoContent)
	contentBoxStyle := kanbanStyle
	if infoBoxWidth > 0 {
		contentBoxStyle = contentBoxStyle.Width(infoBoxWidth)
	}
	contentBox := contentBoxStyle.Render(pageContent)
	rightPanel := strings.Join([]string{infoBox, "", contentBox}, "\n")

	availableHeight := m.bodyHeight(header, help)
	if availableHeight > 0 {
		rightPanel = clampToHeight(padToHeight(rightPanel, availableHeight), availableHeight)
	}

	navText := navContent
	if availableHeight > 0 {
		navContentHeight := availableHeight - 2
		if navContentHeight < 1 {
			navContentHeight = 1
		}
		navText = padToHeight(clampToHeight(navText, navContentHeight), navContentHeight)
	}
	navStyleSized := sidebarStyle
	if navInner := boxedContentWidth(navWidth); navInner > 0 {
		navStyleSized = navStyleSized.Width(navInner)
	}
	navPanel := navStyleSized.Render(navText)

	separator := strings.Repeat(" ", gap)
	return lipgloss.JoinHorizontal(lipgloss.Top, navPanel, separator, rightPanel)
}

func (m Model) renderWikiNavContent() string {
	if len(m.wikiItems) == 0 {
		return taskStyle.Render("No wiki pages found.")
	}
	return m.renderWikiList()
}

func (m Model) renderWikiInfoContent() string {
	lines := make([]string, 0, 6)
	lines = append(lines, headerStyle.Render("Page Info"))
	if status := strings.TrimSpace(m.wikiStatus); status != "" {
		lines = append(lines, taskStyle.Render(status))
	}
	filterLines := m.wikiFilterSummaryLines()
	if len(filterLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, taskStyle.Render("Filters:"))
		lines = append(lines, filterLines...)
	}

	item, ok := m.currentWikiSelection()
	if !ok {
		lines = append(lines, taskStyle.Render("Select a page to view details."))
		return strings.Join(lines, "\n")
	}

	if item.Kind == wikiItemSection {
		sectionTitle := strings.TrimSpace(item.Title)
		sectionSlug := strings.TrimSpace(item.Slug)
		if sectionTitle != "" {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Section: %s", sectionTitle)))
		}
		if sectionSlug != "" {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Slug: %s", sectionSlug)))
		}
		if meta, ok := m.sectionMeta(sectionSlug); ok {
			if len(meta.Tags) > 0 {
				lines = append(lines, taskStyle.Render(fmt.Sprintf("Tags: %s", strings.Join(meta.Tags, ", "))))
			}
			lines = append(lines, m.sectionLinkLines(meta.Links)...)
		}
		return strings.Join(lines, "\n")
	}

	page, ok := m.wikiPages[item.Slug]
	if !ok {
		lines = append(lines, taskStyle.Render("Unable to load page metadata."))
		return strings.Join(lines, "\n")
	}

	title := strings.TrimSpace(page.Title)
	if title == "" {
		title = item.Title
	}
	if title != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Title: %s", title)))
	}
	lines = append(lines, taskStyle.Render(fmt.Sprintf("Slug: %s", item.Slug)))

	sectionTitle := strings.TrimSpace(item.SectionTitle)
	if sectionTitle == "" {
		sectionTitle = strings.TrimSpace(page.Section)
	}
	if sectionTitle != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Section: %s", sectionTitle)))
	}
	if status := strings.TrimSpace(page.Status); status != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Status: %s", status)))
	}
	if len(page.Tags) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Tags: %s", strings.Join(page.Tags, ", "))))
	}
	if meta, ok := m.sectionMeta(item.SectionSlug); ok {
		if len(meta.Tags) > 0 {
			lines = append(lines, taskStyle.Render(fmt.Sprintf("Section Tags: %s", strings.Join(meta.Tags, ", "))))
		}
		lines = append(lines, m.sectionLinkLines(meta.Links)...)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderWikiPageContent() string {
	item, ok := m.currentWikiSelection()
	if !ok || item.Kind != wikiItemPage {
		return taskStyle.Render("Select a page to read its content.")
	}
	page, ok := m.wikiPages[item.Slug]
	if !ok {
		return taskStyle.Render("Unable to load wiki page.")
	}
	content := strings.TrimSpace(page.Content)
	if content == "" {
		return taskStyle.Render("No content.")
	}
	terms := m.wikiQueryTerms()
	if len(terms) > 0 {
		content = highlightTerms(content, terms, accentSoft)
	}
	return taskStyle.Render(content)
}

func (m Model) wikiNavWidth(gap int) int {
	if m.width <= 0 {
		return 30
	}
	navWidth := m.width / 4
	if navWidth < 22 {
		navWidth = 22
	}
	if navWidth > 40 {
		navWidth = 40
	}
	minRight := 24
	if m.width-navWidth-gap < minRight {
		navWidth = m.width - minRight - gap
		if navWidth < 10 {
			navWidth = 10
		}
	}
	return navWidth
}

func (m Model) wikiHelpText() string {
	parts := make([]string, 0, 12)
	if len(m.wikiItems) > 1 {
		parts = append(parts, "j/k move")
	}
	if item, ok := m.currentWikiSelection(); ok && item.Kind == wikiItemPage {
		parts = append(parts, "enter pager", "e edit")
	}
	parts = append(parts,
		"x actions",
		"f filters",
		"E/P export all",
		"alt+1/2/3 or F1/F2/F3 tabs",
		"ctrl+r/F5 refresh",
		"b/esc/q back",
	)
	if summary := m.wikiFilterSummaryShort(); summary != "" {
		parts = append(parts, summary)
	}
	return strings.Join(parts, " • ")
}

func (m Model) bodyHeight(header, footer string) int {
	if m.height <= 0 {
		return 0
	}
	headerHeight := lipgloss.Height(m.renderHeader(header))
	footerText := m.footerText(footer)
	footerHeight := lipgloss.Height(m.renderBar(footerText, footerStyle))
	footerGap := footerGapFor(footerText)
	available := m.height - headerHeight - footerHeight - footerGap
	if available < 0 {
		return 0
	}
	return available
}

func (m Model) renderWikiModalOverlay(title, body, help string) string {
	header := "Wiki ▸ Navigation"
	background := m.frame(header, m.renderWikiLayout(header, m.wikiHelpText()), m.wikiHelpText())
	modal := m.renderModalBox(body)

	if m.width <= 0 || m.height <= 0 {
		return m.renderModal(title, body, help)
	}

	modalWidth := lipgloss.Width(modal)
	modalHeight := lipgloss.Height(modal)
	if modalWidth == 0 || modalHeight == 0 {
		return background
	}

	x := max(0, (m.width-modalWidth)/2)
	y := max(0, (m.height-modalHeight)/2)
	return overlayAt(background, modal, x, y, m.width)
}

func (m Model) wikiFilterPrompt() string {
	switch m.wikiFilterMode {
	case wikiFilterQuery:
		return "Search query"
	case wikiFilterTitle:
		return "Title filter"
	case wikiFilterTags:
		return "Tag filter (comma-separated)"
	case wikiFilterSection:
		return "Section filter"
	default:
		return "Filter"
	}
}

func (m Model) wikiFilterSummaryLines() []string {
	lines := make([]string, 0, 4)
	if strings.TrimSpace(m.wikiQuery) != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Query: %s", m.wikiQuery)))
	}
	if strings.TrimSpace(m.wikiFilterTitle) != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Title: %s", m.wikiFilterTitle)))
	}
	if strings.TrimSpace(m.wikiFilterSection) != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Section: %s", m.wikiFilterSection)))
	}
	if len(m.wikiFilterTags) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Tags: %s", strings.Join(m.wikiFilterTags, ", "))))
	}
	return lines
}

func (m Model) wikiFilterSummaryShort() string {
	parts := make([]string, 0, 4)
	if strings.TrimSpace(m.wikiQuery) != "" {
		parts = append(parts, "query="+m.wikiQuery)
	}
	if strings.TrimSpace(m.wikiFilterTitle) != "" {
		parts = append(parts, "title="+m.wikiFilterTitle)
	}
	if strings.TrimSpace(m.wikiFilterSection) != "" {
		parts = append(parts, "section="+m.wikiFilterSection)
	}
	if len(m.wikiFilterTags) > 0 {
		parts = append(parts, "tags="+strings.Join(m.wikiFilterTags, ","))
	}
	if len(parts) == 0 {
		return ""
	}
	return "filters: " + strings.Join(parts, " ") + " • c clear"
}

func (m Model) wikiQueryTerms() []string {
	return wiki.SplitQueryTerms(m.wikiQuery)
}

func (m Model) sectionMeta(slug string) (wiki.NavNode, bool) {
	needle := strings.TrimSpace(slug)
	for _, node := range m.wikiNav {
		if strings.EqualFold(strings.TrimSpace(node.Slug), needle) {
			return node, true
		}
	}
	return wiki.NavNode{}, false
}

func (m Model) sectionLinkLines(links wiki.SectionLinks) []string {
	lines := make([]string, 0, 2)
	if len(links.DependsOn) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Depends on: %s", strings.Join(links.DependsOn, ", "))))
	}
	if len(links.RelatedTo) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Related to: %s", strings.Join(links.RelatedTo, ", "))))
	}
	return lines
}

func highlightTerms(text string, terms []string, color lipgloss.Color) string {
	if len(terms) == 0 || text == "" {
		return text
	}
	lower := strings.ToLower(text)
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0)
	for _, term := range terms {
		needle := strings.ToLower(strings.TrimSpace(term))
		if needle == "" {
			continue
		}
		idx := 0
		for {
			pos := strings.Index(lower[idx:], needle)
			if pos < 0 {
				break
			}
			start := idx + pos
			end := start + len(needle)
			spans = append(spans, span{start: start, end: end})
			idx = end
		}
	}
	if len(spans) == 0 {
		return text
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].start == spans[j].start {
			return spans[i].end < spans[j].end
		}
		return spans[i].start < spans[j].start
	})
	merged := make([]span, 0, len(spans))
	current := spans[0]
	for i := 1; i < len(spans); i++ {
		next := spans[i]
		if next.start <= current.end {
			if next.end > current.end {
				current.end = next.end
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)

	highlightStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	var buf strings.Builder
	last := 0
	for _, sp := range merged {
		if sp.start > last {
			buf.WriteString(text[last:sp.start])
		}
		if sp.start < 0 || sp.end > len(text) || sp.start >= sp.end {
			continue
		}
		buf.WriteString(highlightStyle.Render(text[sp.start:sp.end]))
		last = sp.end
	}
	if last < len(text) {
		buf.WriteString(text[last:])
	}
	return buf.String()
}

func (m Model) activeBoardName() string {
	if strings.TrimSpace(m.activeBoard) == "" {
		return "Unknown"
	}
	for _, board := range m.boards {
		if board.ID == m.activeBoard {
			if strings.TrimSpace(board.Name) != "" {
				return board.Name
			}
			return board.ID
		}
	}
	return m.activeBoard
}

func (m Model) fieldLine(label, value string, field detailField) string {
	line := fmt.Sprintf("%s: %s", label, value)
	if m.detailField == field {
		return selectedTask.Render(line)
	}
	return taskStyle.Render(line)
}

func (m Model) frame(title, body, footer string) string {
	head := m.renderHeader(title)
	footerText := m.footerText(footer)
	foot := m.renderBar(footerText, footerStyle)
	toast := m.renderToastBar()
	footerGap := footerGapFor(footerText)
	if m.height > 0 {
		headerHeight := lipgloss.Height(head)
		footerHeight := lipgloss.Height(foot)
		toastHeight := lipgloss.Height(toast)
		available := m.height - headerHeight - footerHeight - footerGap - toastHeight
		if available > 0 {
			body = clampToHeight(body, available)
			body = padToHeight(body, available)
		}
	}
	parts := []string{head, body}
	if strings.TrimSpace(toast) != "" {
		parts = append(parts, toast)
	}
	for i := 0; i < footerGap; i++ {
		parts = append(parts, "")
	}
	parts = append(parts, foot)
	return strings.Join(parts, "\n")
}

func (m Model) renderToastBar() string {
	if len(m.toastQueue) == 0 {
		return ""
	}
	item := m.toastQueue[len(m.toastQueue)-1]
	prefix := "INFO"
	style := toastInfoStyle
	switch item.Level {
	case toastSuccess:
		prefix = "SUCCESS"
		style = toastSuccessStyle
	case toastError:
		prefix = "ERROR"
		style = toastErrorStyle
	}
	content := fmt.Sprintf("%s: %s", style.Render(prefix), item.Message)
	return m.renderBar(content, toastStyle)
}

func (m Model) renderBar(content string, style lipgloss.Style) string {
	if m.width <= 0 {
		return style.Render(content)
	}
	return style.Width(m.width).Render(content)
}

func footerGapFor(footer string) int {
	if strings.TrimSpace(footer) == "" {
		return 0
	}
	return 2
}

func boxedContentWidth(total int) int {
	if total <= 0 {
		return 0
	}
	if total <= 4 {
		return 1
	}
	return total - 4
}

func (m Model) renderHeader(title string) string {
	lines := []string{
		m.renderBar(m.renderTabBar(), barStyle),
		m.renderBar(title, barStyle),
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderTabBar() string {
	active := inferTabFromScreen(m.screen, m.confirmAction)
	tabs := []struct {
		tab   appTab
		label string
	}{
		{tab: tabBoards, label: "Boards"},
		{tab: tabWiki, label: "Wiki"},
		{tab: tabADR, label: "ADRs"},
	}
	parts := make([]string, 0, len(tabs)+1)
	parts = append(parts, "Tabs:")
	for _, item := range tabs {
		if item.tab == active {
			parts = append(parts, tabActive.Render(item.label))
			continue
		}
		parts = append(parts, tabInactive.Render(item.label))
	}
	return strings.Join(parts, " ")
}

func (m Model) footerText(help string) string {
	parts := make([]string, 0, 2)
	if trimmed := strings.TrimSpace(help); trimmed != "" {
		parts = append(parts, trimmed)
	}
	parts = append(parts, m.statusLine())
	return strings.Join(parts, " • ")
}

func (m Model) statusLine() string {
	tabLabel := "Boards"
	switch inferTabFromScreen(m.screen, m.confirmAction) {
	case tabWiki:
		tabLabel = "Wiki"
	case tabADR:
		tabLabel = "ADRs"
	}

	state := "Ready"
	if m.loading {
		state = "Loading"
		if strings.TrimSpace(m.loadingMessage) != "" {
			state = m.loadingMessage
		}
	}
	parts := []string{fmt.Sprintf("Tab: %s", tabLabel), state}
	if m.width > 0 && m.height > 0 {
		parts = append(parts, fmt.Sprintf("Size: %dx%d", m.width, m.height))
	}
	return strings.Join(parts, " • ")
}

func (m Model) boardHelpText() string {
	parts := make([]string, 0, 12)
	if m.boardFocus == focusBoards {
		_, hasBoard := m.selectedBoard()
		if len(m.boards) > 1 {
			parts = append(parts, "j/k boards")
		}
		if hasBoard {
			parts = append(parts, "enter open board")
		}
		parts = append(parts, "a add board")
		if hasBoard {
			parts = append(parts, "e edit board", "i board detail", "x board actions")
		}
		parts = append(parts,
			"alt+1/2/3 or F1/F2/F3 tabs",
			"tab/b tasks",
			"ctrl+r/F5 refresh",
			"q quit",
		)
		return strings.Join(parts, " • ")
	}

	listView := m.boardUsesListView()
	if !listView && len(m.columns) > 1 {
		parts = append(parts, "h/l columns")
	}
	if listView {
		if len(m.boardListEntries()) > 1 {
			parts = append(parts, "j/k tasks")
		}
	} else if m.active >= 0 && m.active < len(m.columns) && len(m.columns[m.active].Tasks) > 1 {
		parts = append(parts, "j/k tasks")
	}
	parts = append(parts, "F filters")
	if listView {
		parts = append(parts, "O sort")
	}
	if listView {
		parts = append(parts, "L kanban")
	} else {
		parts = append(parts, "L list")
	}
	parts = append(parts, "a add task")
	if m.currentTaskExists() {
		parts = append(parts, "x task actions", "i task info")
	}
	if m.canMoveSelectedTaskForward() {
		parts = append(parts, "m move")
	}
	if m.canMoveSelectedTaskBack() {
		parts = append(parts, "M move back")
	}
	parts = append(parts, "z archive")
	boardsFocusAvailable := m.sidebarWidth() > 0 || (m.width > 0 && m.width < 90)
	if boardsFocusAvailable {
		parts = append(parts, "tab/b boards")
	}
	parts = append(parts,
		"alt+1/2/3 or F1/F2/F3 tabs",
		"ctrl+r/F5 refresh",
		"q quit",
	)
	return strings.Join(parts, " • ")
}

func (m Model) renderModal(title, body, help string) string {
	modal := m.renderModalBox(body)

	if m.height > 0 && m.width > 0 {
		headerHeight := lipgloss.Height(m.renderHeader(title))
		footerText := m.footerText(help)
		footerHeight := lipgloss.Height(m.renderBar(footerText, footerStyle))
		footerGap := footerGapFor(footerText)
		availableHeight := m.height - headerHeight - footerHeight - footerGap
		if availableHeight < 0 {
			availableHeight = 0
		}
		if availableHeight > 0 {
			modal = lipgloss.Place(m.width, availableHeight, lipgloss.Center, lipgloss.Center, modal)
		}
	}
	return m.frame(title, modal, help)
}

func (m Model) renderModalOverlay(title, body, help string) string {
	background := m.renderBoardScreen(help)
	modal := m.renderModalBox(body)

	if m.width <= 0 || m.height <= 0 {
		return m.renderModal(title, body, help)
	}

	modalWidth := lipgloss.Width(modal)
	modalHeight := lipgloss.Height(modal)
	if modalWidth == 0 || modalHeight == 0 {
		return background
	}

	x := max(0, (m.width-modalWidth)/2)
	y := max(0, (m.height-modalHeight)/2)
	return overlayAt(background, modal, x, y, m.width)
}

func (m Model) renderModalBox(body string) string {
	contentLines := strings.Split(body, "\n")
	maxWidth := 0
	for _, line := range contentLines {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}
	if maxWidth == 0 {
		maxWidth = 1
	}
	const modalMinWidth = 26
	if maxWidth < modalMinWidth {
		maxWidth = modalMinWidth
	}
	if m.width > 0 {
		capWidth := m.width - 6
		if capWidth > 0 && maxWidth > capWidth {
			maxWidth = capWidth
		}
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentSoft).
		Padding(0, 1).
		Background(panelBg).
		Width(maxWidth)
	return modalStyle.Render(body)
}

func overlayAt(background, overlay string, x, y, width int) string {
	bgLines := strings.Split(background, "\n")
	ovLines := strings.Split(overlay, "\n")
	if width <= 0 {
		for _, line := range bgLines {
			if w := lipgloss.Width(line); w > width {
				width = w
			}
		}
	}

	for i, ovLine := range ovLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]
		ovWidth := ansi.PrintableRuneWidth(ovLine)
		if ovWidth == 0 {
			continue
		}

		localX := x
		if localX < 0 {
			ovLine = sliceANSI(ovLine, -localX, ovWidth)
			ovWidth = ansi.PrintableRuneWidth(ovLine)
			localX = 0
		}
		if width > 0 && localX >= width {
			continue
		}
		if width > 0 && localX+ovWidth > width {
			ovLine = sliceANSI(ovLine, 0, width-localX)
			ovWidth = ansi.PrintableRuneWidth(ovLine)
		}

		prefix := sliceANSI(bgLine, 0, localX)
		suffix := sliceANSI(bgLine, localX+ovWidth, width)
		bgLines[row] = prefix + ovLine + suffix
	}

	return strings.Join(bgLines, "\n")
}

func sliceANSI(line string, start, end int) string {
	if end <= start {
		return ""
	}
	if start < 0 {
		start = 0
	}

	var buf strings.Builder
	printable := 0
	inANSI := false
	var seq strings.Builder
	currentStyle := ""
	wroteStyle := false

	writeStyle := func() {
		if !wroteStyle && currentStyle != "" {
			buf.WriteString(currentStyle)
			wroteStyle = true
		}
	}

	for _, r := range line {
		if r == ansi.Marker {
			inANSI = true
			seq.Reset()
			seq.WriteRune(r)
			continue
		}
		if inANSI {
			seq.WriteRune(r)
			if ansi.IsTerminator(r) {
				inANSI = false
				seqStr := seq.String()
				if strings.HasSuffix(seqStr, "[0m") {
					currentStyle = ""
				} else if r == 'm' {
					currentStyle += seqStr
				}
				if printable >= start && printable < end {
					buf.WriteString(seqStr)
				}
			}
			continue
		}

		width := runewidth.RuneWidth(r)
		next := printable + width
		if next <= start {
			printable = next
			continue
		}
		if printable >= end {
			break
		}
		writeStyle()
		buf.WriteRune(r)
		printable = next
	}

	if wroteStyle && currentStyle != "" {
		buf.WriteString("\x1b[0m")
	}

	return buf.String()
}

func fitText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "-"
	}
	if runewidth.StringWidth(trimmed) <= width {
		return trimmed
	}
	if width <= 1 {
		return "…"
	}
	var buf strings.Builder
	current := 0
	for _, r := range trimmed {
		rw := runewidth.RuneWidth(r)
		if current+rw >= width {
			break
		}
		buf.WriteRune(r)
		current += rw
	}
	return strings.TrimSpace(buf.String()) + "…"
}

func padToHeight(body string, height int) string {
	current := lipgloss.Height(body)
	if current >= height {
		return body
	}
	return body + strings.Repeat("\n", height-current)
}

func clampToHeight(body string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(body, "\n")
	if len(lines) <= height {
		return body
	}
	return strings.Join(lines[:height], "\n")
}
