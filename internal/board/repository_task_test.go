package board

import (
	"errors"
	"testing"
)

func TestCreateTaskSequentialIDs(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)

	task1, err := NewTask("First")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	task2, err := NewTask("Second")
	if err != nil {
		t.Fatalf("new task2: %v", err)
	}

	// Act
	created1, createErr1 := repo.CreateTask(task1)
	created2, createErr2 := repo.CreateTask(task2)

	// Assert
	if createErr1 != nil {
		t.Fatalf("create task1: %v", createErr1)
	}
	if createErr2 != nil {
		t.Fatalf("create task2: %v", createErr2)
	}
	if created1.ID != "T-000001" {
		t.Fatalf("expected first task id T-000001, got %s", created1.ID)
	}
	if created2.ID != "T-000002" {
		t.Fatalf("expected second task id T-000002, got %s", created2.ID)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)
	task, _ := NewTask("Status update")
	created, err := repo.CreateTask(task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Act
	updateErr := repo.UpdateTaskStatus(created.ID, "done")
	loaded, loadErr := repo.GetTaskByID(created.ID)

	// Assert
	if updateErr != nil {
		t.Fatalf("update status: %v", updateErr)
	}
	if loadErr != nil {
		t.Fatalf("reload task: %v", loadErr)
	}
	if loaded.Status != "done" {
		t.Fatalf("expected status done, got %q", loaded.Status)
	}
}

func TestUpdateTaskDependenciesCycle(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)
	task1, _ := NewTask("One")
	created1, _ := repo.CreateTask(task1)
	task2, _ := NewTask("Two")
	created2, _ := repo.CreateTask(task2)

	// Act
	depsErr := repo.UpdateTaskDependencies(created2.ID, []string{created1.ID})
	err := repo.UpdateTaskDependencies(created1.ID, []string{created2.ID})

	// Assert
	if depsErr != nil {
		t.Fatalf("set deps: %v", depsErr)
	}
	if err == nil {
		t.Fatalf("expected dependency cycle error")
	}
	if !errors.Is(err, ErrInvalidDependency) {
		t.Fatalf("expected ErrInvalidDependency, got %v", err)
	}
}

func TestUpdateTaskPriorityInvalid(t *testing.T) {
	// Arrange
	repo, _, _ := setupRepo(t)
	task, _ := NewTask("Priority")
	created, _ := repo.CreateTask(task)

	// Act
	err := repo.UpdateTaskPriority(created.ID, 99)

	// Assert
	if err == nil {
		t.Fatalf("expected invalid priority error")
	}
	if !errors.Is(err, ErrInvalidPriority) {
		t.Fatalf("expected ErrInvalidPriority, got %v", err)
	}
}
