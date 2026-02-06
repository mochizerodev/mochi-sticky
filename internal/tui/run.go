package tui

import (
	"mochi-sticky/internal/board"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Bubble Tea program with the provided repository.
func Run(repo *board.Repository) error {
	return RunWithEditor(repo, "")
}

// RunWithEditor starts the Bubble Tea program using the provided editor command.
func RunWithEditor(repo *board.Repository, editor string) error {
	boardRepo, err := board.NewBoardRepositoryWithStorage(repo.BaseDir(), repo.StorageRoot())
	if err != nil {
		return err
	}
	model := NewModel(repo, boardRepo, repo.BaseDir()).SetEditor(editor)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}
