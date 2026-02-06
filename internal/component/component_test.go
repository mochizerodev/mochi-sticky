package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/wiki"
)

func TestIntegrationBoard(t *testing.T) {
	t.Run("BoardRegistry", func(t *testing.T) {
		// Arrange
		_, storageRoot, _, boardRepo := setupBoardStore(t)

		// Act
		boards, active, err := boardRepo.ListBoards()
		requireNoError(t, err, "list boards")
		workBoard, err := boardRepo.CreateBoard("Work Board")
		requireNoError(t, err, "create work board")
		opsBoard, err := boardRepo.CreateBoard("Ops Board")
		requireNoError(t, err, "create ops board")
		renamedWork, err := boardRepo.RenameBoard(workBoard.ID, "Work Board Renamed")
		requireNoError(t, err, "rename work board")
		requireNoError(t, boardRepo.SetActiveBoard(workBoard.ID), "set active board")
		archivedWork, err := boardRepo.ArchiveBoard(workBoard.ID)
		requireNoError(t, err, "archive work board")
		boardsAfterArchive, activeAfterArchive, err := boardRepo.ListBoards()
		requireNoError(t, err, "list boards after archive")
		requireNoError(t, boardRepo.DeleteBoard(workBoard.ID), "delete work board")
		boardsAfterDelete, activeAfterDelete, err := boardRepo.ListBoards()
		requireNoError(t, err, "list boards after delete")
		resolvedBoard, boardDir, tasksDir, configPath, err := boardRepo.ResolveBoardPaths(opsBoard.ID)
		requireNoError(t, err, "resolve board paths")

		// Assert
		if active != "default" {
			t.Fatalf("expected active board to be default, got %q", active)
		}
		if len(boards) != 1 {
			t.Fatalf("expected 1 board, got %d", len(boards))
		}
		if renamedWork.Name != "Work Board Renamed" {
			t.Fatalf("expected board name to be updated, got %q", renamedWork.Name)
		}
		if !archivedWork.Archived {
			t.Fatalf("expected work board to be archived")
		}
		if activeAfterArchive != "default" {
			t.Fatalf("expected active board to fall back to default, got %q", activeAfterArchive)
		}
		if len(boardsAfterArchive) != 3 {
			t.Fatalf("expected 3 boards after archive, got %d", len(boardsAfterArchive))
		}
		if len(boardsAfterDelete) != 2 {
			t.Fatalf("expected 2 boards after delete, got %d", len(boardsAfterDelete))
		}
		if containsBoardID(boardsAfterDelete, workBoard.ID) {
			t.Fatalf("expected work board to be removed from registry")
		}
		if activeAfterDelete != "default" {
			t.Fatalf("expected active board to remain default, got %q", activeAfterDelete)
		}
		if resolvedBoard.ID != opsBoard.ID {
			t.Fatalf("expected resolved board %q, got %q", opsBoard.ID, resolvedBoard.ID)
		}
		if !strings.HasPrefix(boardDir, storageRoot) {
			t.Fatalf("expected board dir %q to live under storage root %q", boardDir, storageRoot)
		}
		if !strings.HasPrefix(tasksDir, boardDir) {
			t.Fatalf("expected tasks dir %q to live under board dir %q", tasksDir, boardDir)
		}
		if !strings.HasPrefix(configPath, boardDir) {
			t.Fatalf("expected config path %q to live under board dir %q", configPath, boardDir)
		}
	})

	t.Run("TaskLifecycle", func(t *testing.T) {
		// Arrange
		baseDir, storageRoot, _, boardRepo := setupBoardStore(t)
		workBoard, err := boardRepo.CreateBoard("Work Board")
		requireNoError(t, err, "create work board")
		requireNoError(t, boardRepo.SetActiveBoard(workBoard.ID), "set active board")
		repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
		requireNoError(t, err, "create repo for active board")
		taskA := createTask(t, repo, "First task", []string{"alpha"})
		taskB := createTask(t, repo, "Second task", []string{"beta"})

		// Act
		requireNoError(t, repo.UpdateTaskTitle(taskB.ID, "Second task updated"), "update task title")
		requireNoError(t, repo.UpdateTaskTags(taskB.ID, []string{"beta", "gamma"}), "update task tags")
		requireNoError(t, repo.UpdateTaskContent(taskB.ID, "Details for second task"), "update task content")
		requireNoError(t, repo.UpdateTaskPriority(taskB.ID, 3), "update task priority")
		requireNoError(t, repo.UpdateTaskDependencies(taskB.ID, []string{taskA.ID}), "update task dependencies")
		requireNoError(t, repo.UpdateTaskStatus(taskA.ID, "done"), "update task status")
		ready, err := repo.ListReadyTasks()
		requireNoError(t, err, "list ready tasks")
		updatedB, err := repo.GetTaskByID(taskB.ID)
		requireNoError(t, err, "get updated task")
		tasks, err := repo.GetAllTasks()
		requireNoError(t, err, "list tasks")
		requireNoError(t, repo.UpdateBoardDescription("Board notes"), "update board description")
		description, err := repo.LoadBoardDescription()
		requireNoError(t, err, "load board description")
		context := board.BoardContext{
			Scope:   " Sprint 1 ",
			Owners:  []string{"alice", " bob "},
			Release: " v1 ",
			Target:  " Q1 ",
			Notes:   " notes ",
		}
		requireNoError(t, repo.UpdateBoardContext(context), "update board context")
		cfg, err := repo.LoadConfig()
		requireNoError(t, err, "load config")
		archivedTask, err := repo.ArchiveTask(taskA.ID)
		requireNoError(t, err, "archive task")
		archivedTasks, err := repo.ListArchivedTasks()
		requireNoError(t, err, "list archived tasks")
		restoredTask, err := repo.RestoreTask(taskA.ID)
		requireNoError(t, err, "restore task")
		moved, err := repo.ArchiveBefore(time.Now().Add(time.Hour))
		requireNoError(t, err, "archive before")
		_, err = repo.RestoreTask(taskB.ID)
		requireNoError(t, err, "restore task for delete")
		requireNoError(t, repo.DeleteTask(taskB.ID), "delete task")
		archivedTasksAfterDelete, err := repo.ListArchivedTasks()
		requireNoError(t, err, "list archived tasks after delete")
		requireNoError(t, repo.DeleteArchivedTask(taskA.ID), "delete archived task")
		remaining, err := repo.GetAllTasks()
		requireNoError(t, err, "list tasks after deletions")

		// Assert
		if !containsTaskID(ready, taskB.ID) {
			t.Fatalf("expected task %s to be ready", taskB.ID)
		}
		if updatedB.Title != "Second task updated" {
			t.Fatalf("expected updated title, got %q", updatedB.Title)
		}
		if updatedB.Content != "Details for second task" {
			t.Fatalf("expected updated content, got %q", updatedB.Content)
		}
		if updatedB.Priority != 3 {
			t.Fatalf("expected priority 3, got %d", updatedB.Priority)
		}
		if len(updatedB.DependsOn) != 1 || updatedB.DependsOn[0] != taskA.ID {
			t.Fatalf("expected dependency on %s, got %+v", taskA.ID, updatedB.DependsOn)
		}
		if len(updatedB.Tags) != 2 || updatedB.Tags[0] != "beta" || updatedB.Tags[1] != "gamma" {
			t.Fatalf("expected updated tags, got %+v", updatedB.Tags)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}
		if strings.TrimSpace(description) != "Board notes" {
			t.Fatalf("expected board description to match, got %q", description)
		}
		if cfg.Context.Scope != "Sprint 1" || cfg.Context.Release != "v1" || cfg.Context.Target != "Q1" || cfg.Context.Notes != "notes" {
			t.Fatalf("expected trimmed board context, got %+v", cfg.Context)
		}
		if len(cfg.Context.Owners) != 2 || cfg.Context.Owners[0] != "alice" || cfg.Context.Owners[1] != "bob" {
			t.Fatalf("expected owners to be normalized, got %+v", cfg.Context.Owners)
		}
		if archivedTask.ID != taskA.ID {
			t.Fatalf("expected archived task %s, got %s", taskA.ID, archivedTask.ID)
		}
		if len(archivedTasks) != 1 || !containsTaskID(archivedTasks, taskA.ID) {
			t.Fatalf("expected archived list to include %s, got %+v", taskA.ID, archivedTasks)
		}
		if restoredTask.ID != taskA.ID {
			t.Fatalf("expected restored task %s, got %s", taskA.ID, restoredTask.ID)
		}
		if len(moved) != 2 || !containsTaskID(moved, taskA.ID) || !containsTaskID(moved, taskB.ID) {
			t.Fatalf("expected archive before to move both tasks, got %+v", moved)
		}
		if len(archivedTasksAfterDelete) != 1 || !containsTaskID(archivedTasksAfterDelete, taskA.ID) {
			t.Fatalf("expected only task %s to remain archived, got %+v", taskA.ID, archivedTasksAfterDelete)
		}
		if len(remaining) != 0 {
			t.Fatalf("expected no remaining tasks, got %d", len(remaining))
		}
	})
}

func TestIntegrationWiki(t *testing.T) {
	t.Run("IndexAndExport", func(t *testing.T) {
		// Arrange
		_, storageRoot, _, _ := setupBoardStore(t)
		wikiRoot := filepath.Join(storageRoot, "wiki")
		pages := []wiki.Page{
			{
				Title:   "Intro",
				Slug:    "guide/intro",
				Section: "guide",
				Order:   1,
				Tags:    []string{"howto"},
				Status:  "published",
				Content: "Hello wiki",
			},
			{
				Title:   "Setup",
				Slug:    "guide/setup",
				Section: "guide",
				Order:   2,
				Tags:    []string{"howto", "setup"},
				Status:  "published",
				Content: "Setup steps",
			},
			{
				Title:   "FAQ",
				Slug:    "reference/faq",
				Section: "reference",
				Order:   1,
				Tags:    []string{"faq"},
				Status:  "published",
				Content: "Common questions",
			},
			{
				Title:   "Draft",
				Slug:    "guide/draft",
				Section: "guide",
				Order:   3,
				Tags:    []string{"draft"},
				Status:  "draft",
				Content: "Draft content",
			},
		}
		for _, page := range pages {
			writeWikiPage(t, wikiRoot, page)
		}
		template := wiki.Page{
			Title:   "Template",
			Slug:    "templates/guide/template",
			Section: "templates",
			Order:   1,
			Tags:    []string{"template"},
			Status:  "published",
			Content: "Template content",
		}
		writeWikiPage(t, wikiRoot, template)

		// Act
		listedPages, err := wiki.ListPages(wikiRoot)
		requireNoError(t, err, "list wiki pages")
		listedWithTemplates, err := wiki.ListPagesWithTemplates(wikiRoot, true)
		requireNoError(t, err, "list wiki pages with templates")
		searchResults, err := wiki.SearchPages(wikiRoot, wiki.SearchOptions{
			Query:           "hello",
			CaseInsensitive: true,
		})
		requireNoError(t, err, "search wiki pages")
		index, err := wiki.GenerateIndex(listedPages)
		requireNoError(t, err, "generate wiki index")
		indexPath := filepath.Join(wikiRoot, "_index.yaml")
		requireNoError(t, wiki.SaveIndex(indexPath, index), "save wiki index")
		loadedIndex, err := wiki.LoadIndex(indexPath)
		requireNoError(t, err, "load wiki index")
		orderedPages, err := wiki.ListPagesFromIndex(wikiRoot, loadedIndex)
		requireNoError(t, err, "list pages from index")
		manifest, err := wiki.BuildManifest(loadedIndex, listedPages)
		requireNoError(t, err, "build manifest")
		exportData, err := wiki.ExportMarkdown(wikiRoot, manifest)
		requireNoError(t, err, "export markdown")
		exportPath := filepath.Join(storageRoot, "exports", "wiki.md")
		requireNoError(t, wiki.WriteExport(exportPath, exportData), "write export")
		exportOnDisk, err := os.ReadFile(exportPath)
		requireNoError(t, err, "read exported markdown")

		// Assert
		if len(listedPages) != 4 {
			t.Fatalf("expected 4 wiki pages, got %d", len(listedPages))
		}
		if len(listedWithTemplates) != 5 {
			t.Fatalf("expected 5 wiki pages including templates, got %d", len(listedWithTemplates))
		}
		if !containsPageSlug(listedPages, "guide/intro") {
			t.Fatalf("expected guide/intro to be listed")
		}
		if !containsPageSlug(listedWithTemplates, "templates/guide/template") {
			t.Fatalf("expected template page to be listed")
		}
		if len(searchResults) == 0 || !containsSearchSlug(searchResults, "guide/intro") {
			t.Fatalf("expected search results to include guide/intro, got %+v", searchResults)
		}
		if len(loadedIndex.Sections) != 2 {
			t.Fatalf("expected 2 index sections, got %d", len(loadedIndex.Sections))
		}
		if len(orderedPages) != len(listedPages) {
			t.Fatalf("expected %d ordered pages, got %d", len(listedPages), len(orderedPages))
		}
		if len(manifest) != 3 {
			t.Fatalf("expected 3 manifest entries (drafts excluded), got %d", len(manifest))
		}
		if !strings.Contains(string(exportData), "# Intro") {
			t.Fatalf("expected export to include intro page")
		}
		if !strings.Contains(string(exportOnDisk), "Hello wiki") {
			t.Fatalf("expected export on disk to include wiki content")
		}
	})
}

func setupBoardStore(t *testing.T) (string, string, *board.Repository, *board.BoardRepository) {
	t.Helper()
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	requireNoError(t, err, "create repository")
	requireNoError(t, repo.InitStore(), "init store")

	boardRepo, err := board.NewBoardRepositoryWithStorage(baseDir, storageRoot)
	requireNoError(t, err, "create board repository")

	return baseDir, storageRoot, repo, boardRepo
}

func createTask(t *testing.T, repo *board.Repository, title string, tags []string) board.Task {
	t.Helper()
	task, err := board.NewTask(title)
	requireNoError(t, err, "create task model")
	task.Tags = tags
	created, err := repo.CreateTask(task)
	requireNoError(t, err, "create task")
	return created
}

func writeWikiPage(t *testing.T, root string, page wiki.Page) {
	t.Helper()
	pagePath := filepath.Join(root, filepath.FromSlash(page.Slug)+".md")
	requireNoError(t, wiki.SavePage(pagePath, page), "save wiki page")
}

func containsTaskID(tasks []board.Task, id string) bool {
	for _, task := range tasks {
		if task.ID == id {
			return true
		}
	}
	return false
}

func containsBoardID(boards []board.Board, id string) bool {
	for _, board := range boards {
		if board.ID == id {
			return true
		}
	}
	return false
}

func containsPageSlug(pages []wiki.Page, slug string) bool {
	for _, page := range pages {
		if page.Slug == slug {
			return true
		}
	}
	return false
}

func containsSearchSlug(results []wiki.SearchResult, slug string) bool {
	for _, result := range results {
		if result.Slug == slug {
			return true
		}
	}
	return false
}

func requireNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}
