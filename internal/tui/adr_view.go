package tui

import (
	"fmt"
	"strings"

	"mochi-sticky/internal/adr"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) adrHelpText() string {
	return "h/l columns • j/k adrs • a add adr • x actions • i detail • m/M move • e editor • ctrl+r/F5 refresh • b/esc back"
}

func (m Model) viewADR() string {
	return m.renderADRScreen("")
}

func (m Model) renderADRScreen(helpOverride string) string {
	header := "mochi-sticky • ADRs"
	help := helpOverride
	if strings.TrimSpace(help) == "" {
		help = m.adrHelpText()
	}
	if len(m.adrColumns) == 0 {
		return m.frame(header, "No ADR statuses configured.", help)
	}

	availableWidth := m.width
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
		columnAreaWidth -= perColumnAllowance * len(m.adrColumns)
		if columnAreaWidth < 0 {
			columnAreaWidth = 0
		}
	}

	columnWidth := m.adrColumnWidthFor(columnAreaWidth)

	maxContentLines := 0
	for _, column := range m.adrColumns {
		lines := 1
		if len(column.ADRs) == 0 {
			lines++
		} else {
			lines += len(column.ADRs)
		}
		if lines > maxContentLines {
			maxContentLines = lines
		}
	}

	availableHeight := m.bodyHeight(header, help)
	kanbanHeight := 0
	if availableHeight > 0 {
		kanbanHeight = availableHeight
		if kanbanHeight < 3 {
			kanbanHeight = 3
		}
	}
	columnHeight := 0
	if kanbanHeight >= 3 {
		columnHeight = kanbanHeight - 2
	}

	rendered := make([]string, 0, len(m.adrColumns))
	for i, column := range m.adrColumns {
		rendered = append(rendered, m.renderADRColumn(column, i == m.adrActive, columnWidth, i == len(m.adrColumns)-1, columnHeight))
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
	return m.frame(header, kanbanBox, help)
}

func (m Model) adrColumnWidthFor(totalWidth int) int {
	count := len(m.adrColumns)
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

func (m Model) renderADRColumn(column adrColumnModel, active bool, width int, isLast bool, height int) string {
	title := strings.TrimSpace(column.Title)
	if title == "" {
		title = strings.TrimSpace(column.Key)
	}
	if strings.TrimSpace(column.Key) != "" && !strings.EqualFold(title, column.Key) {
		title = fmt.Sprintf("%s (%s)", title, column.Key)
	}

	lines := []string{headerStyle.Render(title)}
	if len(column.ADRs) == 0 {
		lines = append(lines, taskStyle.Render("No ADRs"))
	} else {
		for i, record := range column.ADRs {
			dateLabel := "-"
			if !record.Date.IsZero() {
				dateLabel = record.Date.Format("2006-01-02")
			}
			line := fmt.Sprintf("%s %s %s", adr.FormatID(record.ID), dateLabel, record.Title)
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

func (m Model) viewADRActions() string {
	record, ok := m.currentADR()
	title := "ADR Actions"
	if ok {
		title = fmt.Sprintf("ADR Actions: %s", adr.FormatID(record.ID))
	}
	lines := []string{headerStyle.Render(title)}
	items := adrActionItems()
	for i, item := range items {
		label := titleCase(item)
		if item == "create adr" {
			label = "Create ADR"
		}
		if item == "delete adr" {
			label = "Delete ADR"
		}
		if i == m.adrAction {
			lines = append(lines, selectedTask.Render(label))
			continue
		}
		lines = append(lines, taskStyle.Render(label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.renderADRModalOverlay("ADR Actions", body, help)
}

func (m Model) viewADRStatusPicker() string {
	lines := []string{headerStyle.Render("Pick Status")}
	for i, column := range m.adrStatusColumns {
		label := strings.TrimSpace(column.Title)
		if label == "" {
			label = strings.TrimSpace(column.Key)
		}
		if i == m.adrStatusIndex {
			lines = append(lines, selectedTask.Render(label))
			continue
		}
		lines = append(lines, taskStyle.Render(label))
	}
	body := strings.Join(lines, "\n")
	help := "j/k move • enter select • esc back"
	return m.frame("Pick Status", body, help)
}

func (m Model) viewADRCreate() string {
	title := headerStyle.Render("New ADR")
	status := strings.TrimSpace(m.adrStatus)
	if status == "" {
		status = "proposed"
	}
	lineTitle := "Title: " + m.adrTitle
	lineTags := "Tags: " + m.adrTags
	switch m.adrField {
	case 0:
		lineTitle = selectedTask.Render(lineTitle)
	default:
		lineTags = selectedTask.Render(lineTags)
	}
	lines := []string{
		title,
		"",
		taskStyle.Render("Status: " + status),
		lineTitle,
		lineTags,
	}
	body := strings.Join(lines, "\n")
	help := "tab switch field • enter create • esc cancel"
	return m.frame("New ADR", body, help)
}

func (m Model) viewADRDetail() string {
	record, ok := m.currentADR()
	if !ok {
		return m.frame("ADR Detail", "No ADR selected.", "esc back")
	}
	lines := []string{
		headerStyle.Render("ADR Detail"),
		"",
		taskStyle.Render(fmt.Sprintf("ID: %s", adr.FormatID(record.ID))),
		taskStyle.Render(fmt.Sprintf("Title: %s", record.Title)),
		taskStyle.Render(fmt.Sprintf("Status: %s", record.Status)),
	}
	if !record.Date.IsZero() {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Date: %s", record.Date.Format("2006-01-02"))))
	}
	if len(record.Tags) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Tags: %s", strings.Join(record.Tags, ", "))))
	}
	if len(record.Links) > 0 {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Links: %s", strings.Join(record.Links, ", "))))
	}
	if strings.TrimSpace(record.FilePath) != "" {
		lines = append(lines, taskStyle.Render(fmt.Sprintf("Path: %s", record.FilePath)))
	}
	lines = append(lines, "")
	lines = append(lines, taskStyle.Render("Content:"))
	if strings.TrimSpace(record.Content) != "" {
		lines = append(lines, taskStyle.Render(strings.TrimRight(record.Content, "\n")))
	} else {
		lines = append(lines, taskStyle.Render("(empty)"))
	}
	body := strings.Join(lines, "\n")
	help := "s status • e editor • x actions • esc back"
	return m.frame("ADR Detail", body, help)
}

func (m Model) renderADRModalOverlay(title, body, help string) string {
	background := m.renderADRScreen(help)
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
