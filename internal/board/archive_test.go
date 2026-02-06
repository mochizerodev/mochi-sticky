package board

import "testing"

func TestArchiveRestoreDeleteTask(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)

	task, _ := NewTask("Archive me")
	created, err := repo.CreateTask(task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Act
	_, archiveErr := repo.ArchiveTask(created.ID)
	activeAfterArchive, activeAfterArchiveErr := repo.GetAllTasks()
	archivedAfterArchive, archivedAfterArchiveErr := repo.ListArchivedTasks()
	_, restoreErr := repo.RestoreTask(created.ID)
	activeAfterRestore, activeAfterRestoreErr := repo.GetAllTasks()
	deleteErr := repo.DeleteTask(created.ID)
	activeAfterDelete, activeAfterDeleteErr := repo.GetAllTasks()

	// Assert
	if archiveErr != nil {
		t.Fatalf("archive task: %v", archiveErr)
	}
	if activeAfterArchiveErr != nil {
		t.Fatalf("list tasks: %v", activeAfterArchiveErr)
	}
	if len(activeAfterArchive) != 0 {
		t.Fatalf("expected 0 active tasks, got %d", len(activeAfterArchive))
	}
	if archivedAfterArchiveErr != nil {
		t.Fatalf("list archived: %v", archivedAfterArchiveErr)
	}
	if len(archivedAfterArchive) != 1 {
		t.Fatalf("expected 1 archived task, got %d", len(archivedAfterArchive))
	}
	if restoreErr != nil {
		t.Fatalf("restore task: %v", restoreErr)
	}
	if activeAfterRestoreErr != nil {
		t.Fatalf("list tasks: %v", activeAfterRestoreErr)
	}
	if len(activeAfterRestore) != 1 {
		t.Fatalf("expected 1 active task, got %d", len(activeAfterRestore))
	}
	if deleteErr != nil {
		t.Fatalf("delete task: %v", deleteErr)
	}
	if activeAfterDeleteErr != nil {
		t.Fatalf("list tasks after delete: %v", activeAfterDeleteErr)
	}
	if len(activeAfterDelete) != 0 {
		t.Fatalf("expected 0 active tasks, got %d", len(activeAfterDelete))
	}
}

func TestDeleteArchivedTask(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)

	task, _ := NewTask("Archived delete")
	created, err := repo.CreateTask(task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Act
	_, archiveErr := repo.ArchiveTask(created.ID)
	deleteErr := repo.DeleteArchivedTask(created.ID)
	archived, archivedErr := repo.ListArchivedTasks()

	// Assert
	if archiveErr != nil {
		t.Fatalf("archive task: %v", archiveErr)
	}
	if deleteErr != nil {
		t.Fatalf("delete archived task: %v", deleteErr)
	}
	if archivedErr != nil {
		t.Fatalf("list archived: %v", archivedErr)
	}
	if len(archived) != 0 {
		t.Fatalf("expected 0 archived tasks, got %d", len(archived))
	}
}
