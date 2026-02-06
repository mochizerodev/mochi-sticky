package board

import (
	"fmt"
	"strings"
)

// FormatTaskDetail renders a task's metadata and content.
func FormatTaskDetail(task Task) string {
	var b strings.Builder
	writeLine := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		fmt.Fprintf(&b, "%s: %s\n", label, value)
	}

	writeLine("Board", TaskBoardLabel(task))
	writeLine("ID", task.ID)
	writeLine("UID", task.UID)
	writeLine("Title", task.Title)
	writeLine("Status", task.Status)
	writeLine("Priority", fmt.Sprintf("%d", effectivePriority(task.Priority)))
	if len(task.Tags) > 0 {
		writeLine("Tags", strings.Join(task.Tags, ", "))
	}
	if len(task.DependsOn) > 0 {
		writeLine("Depends On", strings.Join(task.DependsOn, ", "))
	}
	if !task.Created.IsZero() {
		writeLine("Created", task.Created.Format("2006-01-02"))
	}
	writeLine("Path", task.FilePath)

	if strings.TrimSpace(task.Content) != "" {
		b.WriteString("\n")
		b.WriteString(task.Content)
		if !strings.HasSuffix(task.Content, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// TaskBoardLabel returns the most descriptive board name for the task.
func TaskBoardLabel(task Task) string {
	if trimmed := strings.TrimSpace(task.BoardName); trimmed != "" {
		return trimmed
	}
	if trimmed := strings.TrimSpace(task.BoardID); trimmed != "" {
		return trimmed
	}
	return "unknown"
}
