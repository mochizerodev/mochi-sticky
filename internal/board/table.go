package board

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// FormatTasksTable renders tasks into a styled ASCII table.
func FormatTasksTable(tasks []Task) string {
	headers := []string{"ID", "Title", "Status", "Priority", "Tags", "Created"}
	rows := make([][]string, 0, len(tasks))
	for _, task := range tasks {
		created := ""
		if !task.Created.IsZero() {
			created = task.Created.Format("2006-01-02")
		}
		tags := strings.Join(task.Tags, ", ")
		priority := fmt.Sprintf("%d", effectivePriority(task.Priority))
		rows = append(rows, []string{task.ID, task.Title, task.Status, priority, tags, created})
	}

	widths := columnWidths(headers, rows)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("242"))
	cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	var b strings.Builder
	sep := separatorLine(widths)
	b.WriteString(sep)
	b.WriteString("\n")
	b.WriteString(renderRow(headers, widths, headerStyle))
	b.WriteString("\n")
	b.WriteString(sep)
	for _, row := range rows {
		b.WriteString("\n")
		b.WriteString(renderRow(row, widths, cellStyle))
	}
	b.WriteString("\n")
	b.WriteString(sep)
	return b.String()
}

func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = runeLen(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(widths) <= i {
				continue
			}
			if width := runeLen(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}
	return widths
}

func separatorLine(widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width+2)
	}
	return "+" + strings.Join(parts, "+") + "+"
}

func renderRow(cells []string, widths []int, style lipgloss.Style) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		if i >= len(widths) {
			continue
		}
		padded := padRight(cell, widths[i])
		parts[i] = style.Render(padded)
	}
	return "| " + strings.Join(parts, " | ") + " |"
}

func padRight(value string, width int) string {
	padding := width - runeLen(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func runeLen(value string) int {
	return utf8.RuneCountInString(value)
}
