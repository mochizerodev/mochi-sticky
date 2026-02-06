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
	bg           = lipgloss.Color("#0F111A")
	panelBg      = lipgloss.Color("#111827")
	accent       = lipgloss.Color("#2DD4BF")
	accentSoft   = lipgloss.Color("#14B8A6")
	textBright   = lipgloss.Color("#E5E7EB")
	textMuted    = lipgloss.Color("#9CA3AF")
	borderColor  = lipgloss.Color("#334155")
	columnBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	sidebarStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	infoBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	kanbanStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Background(panelBg)
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(accent).Background(panelBg)
	taskStyle    = lipgloss.NewStyle().Foreground(textBright).Background(panelBg)
	selectedTask = lipgloss.NewStyle().Foreground(lipgloss.Color("#0B1016")).Background(accent).Bold(true)
	activeBoard  = lipgloss.NewStyle().Foreground(accentSoft).Background(panelBg).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
	barStyle     = lipgloss.NewStyle().Background(bg).Foreground(textBright).Bold(true).Padding(0, 1)
	footerStyle  = lipgloss.NewStyle().Background(bg).Foreground(textMuted).Padding(0, 1)
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
	header := fmt.Sprintf("mochi-sticky • Board: %s", m.activeBoardName())
	help := helpOverride
	if strings.TrimSpace(help) == "" {
		help = m.boardHelpText()
	}
	sidebarWidth := m.sidebarWidth()
	availableWidth := m.width
	if sidebarWidth > 0 {
		// Sidebar width does not include padding/border.
		availableWidth -= sidebarWidth + 4
	}
	if availableWidth < 0 {
		availableWidth = 0
	}

	// Account for the kanban frame (rounded border + 1 space of horizontal padding on each side).
	const kanbanFrameWidth = 4
	columnAreaWidth := availableWidth
	if columnAreaWidth > 0 {
		columnAreaWidth -= kanbanFrameWidth
		if columnAreaWidth < 0 {
			columnAreaWidth = 0
		}
	}

	if columnAreaWidth > 0 {
		// Column width excludes padding and borders.
		perColumnAllowance := 4 // 2 for border + 2 for padding
		columnAreaWidth -= perColumnAllowance * len(m.columns)
		if columnAreaWidth < 0 {
			columnAreaWidth = 0
		}
	}

	columnWidth := m.columnWidthFor(columnAreaWidth)
	taskIndex := buildTaskIndex(m.columns)
	infoBox := m.renderBoardInfoBox(availableWidth)
	infoBoxHeight := 0
	if infoBox != "" {
		infoBoxHeight = lipgloss.Height(infoBox)
	}

	maxContentLines := 0
	for _, column := range m.columns {
		lines := 1
		if len(column.Tasks) == 0 {
			lines++
		} else {
			lines += len(column.Tasks)
		}
		if lines > maxContentLines {
			maxContentLines = lines
		}
	}

	availableHeight := 0
	if m.height > 0 {
		headerHeight := lipgloss.Height(m.renderBar(header, barStyle))
		footerHeight := lipgloss.Height(m.renderBar(help, footerStyle))
		footerGap := footerGapFor(help)
		availableHeight = m.height - headerHeight - footerHeight - footerGap
		if availableHeight < 0 {
			availableHeight = 0
		}
	}
	spacing := 0
	if infoBox != "" {
		spacing = 1
	}
	contentHeightCap := 0
	if availableHeight > 0 {
		contentHeightCap = availableHeight - infoBoxHeight - spacing
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
	if availableHeight > 0 && kanbanHeight > 0 {
		totalBodyHeight := infoBoxHeight + spacing + kanbanHeight
		if totalBodyHeight > availableHeight {
			overflow := totalBodyHeight - availableHeight
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

	rendered := make([]string, 0, len(m.columns))
	for i, column := range m.columns {
		rendered = append(rendered, m.renderColumn(column, i == m.active, columnWidth, i == len(m.columns)-1, taskIndex, columnHeight))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	kanbanStyleSized := kanbanStyle
	if availableWidth > 0 {
		kanbanStyleSized = kanbanStyleSized.Width(availableWidth)
	}
	if kanbanHeight > 0 {
		kanbanContentHeight := max(0, kanbanHeight-2)
		if kanbanContentHeight > 0 {
			kanbanStyleSized = kanbanStyleSized.Height(kanbanContentHeight)
		}
	}
	kanbanBox := kanbanStyleSized.Render(board)
	sections := make([]string, 0, 2)
	if infoBox != "" {
		sections = append(sections, infoBox)
	}
	sections = append(sections, kanbanBox)
	body := strings.Join(sections, "\n")
	if sidebarWidth > 0 {
		bodyHeight := lipgloss.Height(body)
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderBoardSidebar(sidebarWidth, bodyHeight), body)
	}
	return m.frame(header, body, help)
}

func (m Model) columnWidthFor(totalWidth int) int {
	count := len(m.columns)
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
	if m.width < 100 {
		return 0
	}
	if m.width >= 140 {
		return 28
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
	lines := m.contextLines()
	if len(lines) == 0 {
		lines = []string{taskStyle.Render("(empty)")}
	}
	body := fmt.Sprintf("%s\n%s", headerStyle.Render("Board Info"), strings.Join(lines, "\n"))
	style := infoBoxStyle
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(body)
}

func (m Model) contextLines() []string {
	ctx := m.boardContext
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
	style := sidebarStyle.Width(width)
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
	help := "e edit • ctrl+r/F5 refresh • x actions • b boards • esc back"
	return m.frame(title, body, help)
}

func (m Model) viewConfirm() string {
	message := "Confirm action?"
	switch m.confirmAction {
	case confirmArchiveBoard:
		message = fmt.Sprintf("Archive board %q?", m.confirmBoard)
	case confirmDeleteBoard:
		message = fmt.Sprintf("Delete board %q? This cannot be undone.", m.confirmBoard)
	case confirmArchiveTask:
		message = fmt.Sprintf("Archive task %q?", m.confirmTask)
	case confirmDeleteTask:
		message = fmt.Sprintf("Delete task %q? This cannot be undone.", m.confirmTask)
	case confirmDeleteADR:
		message = fmt.Sprintf("Delete ADR %s? This cannot be undone.", adr.FormatID(m.confirmADR))
	}
	lines := []string{
		headerStyle.Render("Confirm"),
		"",
		taskStyle.Render(message),
		"",
		taskStyle.Render("[y] confirm  [n] cancel"),
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
	header := "mochi-sticky • Wiki"
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

	infoBox := infoBoxStyle.Width(rightWidth).Render(infoContent)
	contentBox := kanbanStyle.Width(rightWidth).Render(pageContent)
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
	navPanel := sidebarStyle.Width(navWidth).Render(navText)

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
	base := "j/k move • enter pager • e edit • x actions • f filters • E/P export all • ctrl+r/F5 refresh • b/esc back • q quit"
	if summary := m.wikiFilterSummaryShort(); summary != "" {
		return base + " • " + summary
	}
	return base
}

func (m Model) bodyHeight(header, footer string) int {
	if m.height <= 0 {
		return 0
	}
	headerHeight := lipgloss.Height(m.renderBar(header, barStyle))
	footerHeight := 0
	if strings.TrimSpace(footer) != "" {
		footerHeight = lipgloss.Height(m.renderBar(footer, footerStyle))
	}
	footerGap := footerGapFor(footer)
	available := m.height - headerHeight - footerHeight - footerGap
	if available < 0 {
		return 0
	}
	return available
}

func (m Model) renderWikiModalOverlay(title, body, help string) string {
	header := "mochi-sticky • Wiki"
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
	head := m.renderBar(title, barStyle)
	foot := ""
	if strings.TrimSpace(footer) != "" {
		foot = m.renderBar(footer, footerStyle)
	}
	footerGap := footerGapFor(footer)
	if m.height > 0 {
		headerHeight := lipgloss.Height(head)
		footerHeight := 0
		if foot != "" {
			footerHeight = lipgloss.Height(foot)
		}
		available := m.height - headerHeight - footerHeight - footerGap
		if available > 0 {
			body = clampToHeight(body, available)
			body = padToHeight(body, available)
		}
	}
	parts := []string{head, body}
	if foot != "" {
		for i := 0; i < footerGap; i++ {
			parts = append(parts, "")
		}
		parts = append(parts, foot)
	}
	return strings.Join(parts, "\n")
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

func (m Model) boardHelpText() string {
	if m.boardFocus == focusBoards && m.sidebarWidth() > 0 {
		return "j/k boards • enter use • a add board • e edit board • i board detail • x board actions • w wiki • d adrs • tab/b kanban • ctrl+r/F5 refresh • q quit"
	}
	return "h/l columns • j/k tasks • a add task • x task actions • i task info • m/M move • z archive • w wiki • d adrs • tab/b boards • ctrl+r/F5 refresh • q quit"
}

func (m Model) renderModal(title, body, help string) string {
	modal := m.renderModalBox(body)

	if m.height > 0 && m.width > 0 {
		headerHeight := lipgloss.Height(m.renderBar(title, barStyle))
		footerHeight := 0
		if strings.TrimSpace(help) != "" {
			footerHeight = lipgloss.Height(m.renderBar(help, footerStyle))
		}
		footerGap := footerGapFor(help)
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
