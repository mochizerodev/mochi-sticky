package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"
	"mochi-sticky/internal/wiki"

	tea "github.com/charmbracelet/bubbletea"
)

type columnModel struct {
	Key      string
	Title    string
	Tasks    []board.Task
	Selected int
}

type adrColumnModel struct {
	Key      string
	Title    string
	ADRs     []adr.ADR
	Selected int
}

type screen int

const (
	screenBoard screen = iota
	screenBoardActions
	screenBoardEdit
	screenBoardDetail
	screenConfirm
	screenTaskActions
	screenStatusPicker
	screenTaskCreate
	screenTaskDetail
	screenTaskEdit
	screenArchive
	screenWiki
	screenWikiActions
	screenWikiFilter
	screenWikiFilterMenu
	screenADR
	screenADRActions
	screenADRStatusPicker
	screenADRCreate
	screenADRDetail
)

type boardFocus int

const (
	focusKanban boardFocus = iota
	focusBoards
)

type boardEditMode int

const (
	editCreate boardEditMode = iota
	editRename
)

type taskEditMode int

const (
	editTitle taskEditMode = iota
	editTags
	editDescription
	editPriority
)

type confirmAction int

const (
	confirmNone confirmAction = iota
	confirmArchiveBoard
	confirmDeleteBoard
	confirmArchiveTask
	confirmDeleteTask
	confirmDeleteADR
)

type detailField int

const (
	fieldTitle detailField = iota
	fieldStatus
	fieldPriority
	fieldTags
	fieldDescription
)

// Model holds the TUI state.
type Model struct {
	repo                 *board.Repository
	boardRepo            *board.BoardRepository
	baseDir              string
	editor               string
	inFlightCancel       context.CancelFunc
	columns              []columnModel
	active               int
	boards               []board.Board
	activeBoard          string
	boardDesc            string
	boardContext         board.BoardContext
	selectedTaskID       string
	boardIndex           int
	boardAction          int
	taskAction           int
	wikiAction           int
	statusIndex          int
	screen               screen
	confirmAction        confirmAction
	confirmBoard         string
	confirmTask          string
	confirmADR           int
	editMode             boardEditMode
	editInput            string
	editBoardID          string
	taskTitle            string
	taskTags             string
	taskStatus           string
	taskPriority         int
	taskField            int
	taskEditMode         taskEditMode
	taskEditInput        string
	detailField          detailField
	archived             []board.Task
	archiveIndex         int
	boardFocus           boardFocus
	boardActionFromBoard bool
	boardEditFromBoard   bool
	confirmFromBoard     bool
	boardCommandReturn   bool
	loadingMessage       string
	width                int
	height               int
	loading              bool
	pendingRefresh       bool
	pendingBoardDescEdit bool
	pendingBoardDetail   bool
	err                  error
	wikiItems            []wikiNavItem
	wikiIndex            int
	wikiPages            map[string]wiki.Page
	wikiStatus           string
	wikiNav              []wiki.NavNode
	wikiQuery            string
	wikiFilterTitle      string
	wikiFilterSection    string
	wikiFilterTags       []string
	wikiFilterTagMode    string
	wikiFilterInput      string
	wikiFilterMode       wikiFilterMode
	adrColumns           []adrColumnModel
	adrStatusColumns     []adr.Column
	adrActive            int
	adrAction            int
	adrStatusIndex       int
	selectedADRID        int
	adrTitle             string
	adrTags              string
	adrStatus            string
	adrField             int
}

// NewModel creates a TUI model backed by a repository.
func NewModel(repo *board.Repository, boardRepo *board.BoardRepository, baseDir string) Model {
	return Model{
		repo:      repo,
		boardRepo: boardRepo,
		baseDir:   baseDir,
		loading:   true,
	}
}

// SetEditor overrides the editor command for this model.
func (m Model) SetEditor(editor string) Model {
	m.editor = editor
	return m
}

// Init loads tasks from the repository.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadStateCmd(m.repo), loadBoardsCmd(m.boardRepo))
}

func (m Model) withInFlight(ctxFn func(context.Context) tea.Cmd) (Model, tea.Cmd) {
	m = m.cancelInFlight()
	ctx, cancel := context.WithCancel(context.Background())
	m.inFlightCancel = cancel
	return m, ctxFn(ctx)
}

func (m Model) startRefresh() (Model, tea.Cmd) {
	m.captureSelection()
	m.loading = true
	m.pendingRefresh = true
	return m.withInFlight(func(ctx context.Context) tea.Cmd {
		return loadBoardsCmdContext(ctx, m.boardRepo)
	})
}

func (m Model) cancelInFlight() Model {
	if m.inFlightCancel != nil {
		m.inFlightCancel()
		m.inFlightCancel = nil
	}
	return m
}

// Update handles keyboard input and state updates.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case stateMsg:
		if m.repo != nil && msg.boardID != "" && msg.boardID != m.repo.BoardID() {
			return m, nil
		}
		m.columns = buildColumns(msg.columns, msg.tasks)
		m.boardDesc = msg.desc
		m.boardContext = msg.context
		m.loading = false
		m.loadingMessage = ""
		if m.active >= len(m.columns) {
			m.active = max(0, len(m.columns)-1)
		}
		for i := range m.columns {
			clampSelection(&m.columns[i])
		}
		if strings.TrimSpace(m.selectedTaskID) != "" {
			m.restoreSelection()
		}
		if m.pendingBoardDescEdit {
			m.pendingBoardDescEdit = false
			if m.repo == nil {
				m.err = fmt.Errorf("tui: repository unavailable")
				return m, nil
			}
			path, err := m.repo.BoardDescriptionPath()
			if err != nil {
				return m, func() tea.Msg {
					return errMsg{err: err}
				}
			}
			m.screen = screenBoard
			return m, openEditorCmd(m.repo, path, m.editor)
		}
		if m.pendingBoardDetail {
			m.pendingBoardDetail = false
			m.screen = screenBoardDetail
		}
		return m, nil
	case boardStateMsg:
		m.boards = msg.boards
		m.activeBoard = msg.active
		m.boardIndex = clampBoardIndex(m.boards, m.boardIndex, m.activeBoard)
		if m.pendingRefresh {
			m.pendingRefresh = false
			if m.repo == nil {
				m.err = fmt.Errorf("tui: repository unavailable")
				return m, nil
			}
			if strings.TrimSpace(msg.active) != "" && msg.active != m.repo.BoardID() {
				newRepo, err := board.NewRepositoryForBoardWithStorage(m.baseDir, msg.active, m.repo.StorageRoot())
				if err != nil {
					m.err = err
					return m, nil
				}
				m.repo = newRepo
			}
			m.loading = true
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadStateCmdContext(ctx, m.repo)
			})
		}
		if m.repo != nil && strings.TrimSpace(msg.active) != "" && msg.active != m.repo.BoardID() {
			newRepo, err := board.NewRepositoryForBoardWithStorage(m.baseDir, msg.active, m.repo.StorageRoot())
			if err != nil {
				m.err = err
				return m, nil
			}
			m.repo = newRepo
			m.loading = true
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadStateCmdContext(ctx, m.repo)
			})
		}
		return m, nil
	case boardActionMsg:
		m = m.cancelInFlight()
		m.boards = msg.boards
		m.activeBoard = msg.active
		m.boardIndex = clampBoardIndex(m.boards, m.boardIndex, m.activeBoard)
		returnToBoard := msg.returnToBoard || m.boardCommandReturn
		if returnToBoard {
			m.screen = screenBoard
			if m.sidebarWidth() > 0 && m.boardCommandReturn {
				m.boardFocus = focusBoards
			}
		}
		m.boardCommandReturn = false
		if msg.reloadTasks {
			if m.repo == nil {
				m.err = fmt.Errorf("tui: repository unavailable")
				return m, nil
			}
			newRepo, err := board.NewRepositoryForBoardWithStorage(m.baseDir, m.activeBoard, m.repo.StorageRoot())
			if err != nil {
				m.err = err
				return m, nil
			}
			m.repo = newRepo
			m.loading = true
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadStateCmdContext(ctx, m.repo)
			})
		}
		return m, nil
	case statusUpdatedMsg:
		m = m.cancelInFlight()
		m = m.applyStatusUpdate(msg.id, msg.status)
		return m, nil
	case wikiStateMsg:
		m = m.cancelInFlight()
		m.wikiNav = msg.nav
		m.wikiPages = msg.pages
		m.applyWikiFilters()
		m.loading = false
		m.loadingMessage = ""
		if len(m.wikiItems) == 0 {
			m.wikiStatus = "No wiki pages found."
		}
		return m, nil
	case wikiExportMsg:
		m = m.cancelInFlight()
		m.loading = false
		m.loadingMessage = ""
		if msg.status != "" {
			m.wikiStatus = msg.status
		} else if msg.path != "" {
			m.wikiStatus = fmt.Sprintf("Exported to %s", msg.path)
		}
		return m, nil
	case adrStateMsg:
		m = m.cancelInFlight()
		m.adrStatusColumns = msg.columns
		m.adrColumns = buildADRColumns(msg.columns, msg.adrs)
		m.loading = false
		m.loadingMessage = ""
		if m.adrActive >= len(m.adrColumns) {
			m.adrActive = max(0, len(m.adrColumns)-1)
		}
		for i := range m.adrColumns {
			clampADRSelection(&m.adrColumns[i])
		}
		if m.selectedADRID > 0 {
			m.restoreADRSelection()
		}
		return m, nil
	case adrStatusUpdatedMsg:
		m = m.cancelInFlight()
		m = m.applyADRStatusUpdate(msg.id, msg.status)
		return m, nil
	case adrCreatedMsg:
		m = m.cancelInFlight()
		m.selectedADRID = msg.record.ID
		m.loading = true
		m.loadingMessage = "Opening editor..."
		m.screen = screenADR
		return m, openADREditorCmd(m.adrRoot(), msg.record.FilePath, m.editor)
	case errMsg:
		m = m.cancelInFlight()
		m.pendingRefresh = false
		// Ignore context.Canceled errors - these are expected when cancelling
		// in-flight operations (e.g., when refreshing with Ctrl+R/F5)
		if msg.err != nil && msg.err != context.Canceled {
			m.err = msg.err
		}
		m.loading = false
		m.loadingMessage = ""
		return m, nil
	case archiveStateMsg:
		m = m.cancelInFlight()
		m.archived = msg.tasks
		m.archiveIndex = clampIndex(m.archiveIndex, len(m.archived))
		if msg.reloadTasks {
			m.loading = true
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadStateCmdContext(ctx, m.repo)
			})
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenBoardActions:
		return m.handleBoardActionsKey(msg)
	case screenBoardEdit:
		return m.handleBoardEditKey(msg)
	case screenBoardDetail:
		return m.handleBoardDetailKey(msg)
	case screenConfirm:
		return m.handleConfirmKey(msg)
	case screenTaskActions:
		return m.handleTaskActionsKey(msg)
	case screenStatusPicker:
		return m.handleStatusPickerKey(msg)
	case screenTaskCreate:
		return m.handleTaskCreateKey(msg)
	case screenTaskDetail:
		return m.handleTaskDetailKey(msg)
	case screenTaskEdit:
		return m.handleTaskEditKey(msg)
	case screenArchive:
		return m.handleArchiveKey(msg)
	case screenWiki:
		return m.handleWikiKey(msg)
	case screenWikiActions:
		return m.handleWikiActionsKey(msg)
	case screenWikiFilter:
		return m.handleWikiFilterKey(msg)
	case screenWikiFilterMenu:
		return m.handleWikiFilterMenuKey(msg)
	case screenADR:
		return m.handleADRKey(msg)
	case screenADRActions:
		return m.handleADRActionsKey(msg)
	case screenADRStatusPicker:
		return m.handleADRStatusPickerKey(msg)
	case screenADRCreate:
		return m.handleADRCreateKey(msg)
	case screenADRDetail:
		return m.handleADRDetailKey(msg)
	default:
	}

	if m.boardFocus == focusBoards && m.sidebarWidth() > 0 {
		switch normalizedKey(msg) {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "b":
			m.boardFocus = focusKanban
			return m, nil
		case "ctrl+r", "f5":
			return m.startRefresh()
		case "w":
			m.screen = screenWiki
			m.loading = true
			m.loadingMessage = "Loading wiki..."
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadWikiCmdContext(ctx, m.wikiRoot())
			})
		case "d":
			m.screen = screenADR
			m.loading = true
			m.loadingMessage = "Loading ADRs..."
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return loadADRCmdContext(ctx, m.adrRoot())
			})
		case "j":
			m.boardIndex++
			m.boardIndex = clampIndex(m.boardIndex, len(m.boards))
			return m, nil
		case "k":
			m.boardIndex--
			m.boardIndex = clampIndex(m.boardIndex, len(m.boards))
			return m, nil
		case "enter", "u":
			if board, ok := m.selectedBoard(); ok {
				return m.withInFlight(func(ctx context.Context) tea.Cmd {
					return boardUseCmdContext(ctx, m.boardRepo, board.ID)
				})
			}
			return m, nil
		case "i":
			if board, ok := m.selectedBoard(); ok {
				if board.ID != m.activeBoard {
					m.pendingBoardDetail = true
					return m.withInFlight(func(ctx context.Context) tea.Cmd {
						return boardUseCmdContext(ctx, m.boardRepo, board.ID)
					})
				}
				m.screen = screenBoardDetail
			}
			return m, nil
		case "a":
			m.screen = screenBoardEdit
			m.boardEditFromBoard = true
			m.editMode = editCreate
			m.editInput = ""
			m.editBoardID = ""
			return m, nil
		case "e", "r":
			if board, ok := m.selectedBoard(); ok {
				m.screen = screenBoardEdit
				m.boardEditFromBoard = true
				m.editMode = editRename
				m.editInput = board.Name
				m.editBoardID = board.ID
			}
			return m, nil
		case "x":
			m.screen = screenBoardActions
			m.boardAction = 0
			m.boardActionFromBoard = true
			return m, nil
		default:
			return m, nil
		}
	}

	if msg.String() == "M" {
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return m.moveSelectedTaskBackCmdContext(ctx)
		})
	}
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		if m.sidebarWidth() > 0 {
			m.boardFocus = focusBoards
			m.boardIndex = clampBoardIndex(m.boards, m.boardIndex, m.activeBoard)
			return m, nil
		}
		return m, nil
	case "b":
		if m.sidebarWidth() > 0 {
			m.boardFocus = focusBoards
			m.boardIndex = clampBoardIndex(m.boards, m.boardIndex, m.activeBoard)
		}
		return m, nil
	case "ctrl+r", "f5":
		return m.startRefresh()
	case "w":
		m.screen = screenWiki
		m.loading = true
		m.loadingMessage = "Loading wiki..."
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return loadWikiCmdContext(ctx, m.wikiRoot())
		})
	case "d":
		m.screen = screenADR
		m.loading = true
		m.loadingMessage = "Loading ADRs..."
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return loadADRCmdContext(ctx, m.adrRoot())
		})
	case "h":
		if m.active > 0 {
			m.active--
		}
		return m, nil
	case "l":
		if m.active < len(m.columns)-1 {
			m.active++
		}
		return m, nil
	case "j":
		m = m.moveSelection(1)
		return m, nil
	case "k":
		m = m.moveSelection(-1)
		return m, nil
	case "enter", "i":
		if m.currentTaskExists() {
			m.screen = screenTaskDetail
			m.detailField = fieldTitle
		}
		return m, nil
	case "m":
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return m.moveSelectedTaskCmdContext(ctx)
		})
	case "x":
		if m.currentTaskExists() {
			m.screen = screenTaskActions
			m.taskAction = 0
		}
		return m, nil
	case "z":
		m.screen = screenArchive
		m.archiveIndex = 0
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return loadArchiveCmdContext(ctx, m.repo)
		})
	case "a":
		status := ""
		if len(m.columns) > 0 {
			status = m.columns[m.active].Key
		}
		m.taskTitle = ""
		m.taskTags = ""
		m.taskStatus = status
		m.taskPriority = board.DefaultPriority
		m.taskField = 0
		m.screen = screenTaskCreate
		return m, nil
	default:
		return m, nil
	}
}

type stateMsg struct {
	boardID string
	columns []board.Column
	tasks   []board.Task
	desc    string
	context board.BoardContext
}

type boardStateMsg struct {
	boards []board.Board
	active string
}

type boardActionMsg struct {
	boards        []board.Board
	active        string
	reloadTasks   bool
	returnToBoard bool
}

type statusUpdatedMsg struct {
	id     string
	status string
}

type archiveStateMsg struct {
	tasks       []board.Task
	reloadTasks bool
}

type errMsg struct {
	err error
}

func loadStateCmd(repo *board.Repository) tea.Cmd {
	return loadStateCmdContext(context.Background(), repo)
}

func loadStateCmdContext(ctx context.Context, repo *board.Repository) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		config, err := repo.LoadConfigContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		description, err := repo.LoadBoardDescriptionContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		tasks, err := repo.GetAllTasksContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return stateMsg{
			boardID: repo.BoardID(),
			columns: config.Columns,
			tasks:   tasks,
			desc:    description,
			context: config.Context,
		}
	}
}

func loadBoardsCmd(repo *board.BoardRepository) tea.Cmd {
	return loadBoardsCmdContext(context.Background(), repo)
}

func loadBoardsCmdContext(ctx context.Context, repo *board.BoardRepository) tea.Cmd {
	return func() tea.Msg {
		if repo == nil {
			return errMsg{err: fmt.Errorf("tui: board repository unavailable")}
		}
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardStateMsg{boards: boards, active: active}
	}
}

func loadArchiveCmdContext(ctx context.Context, repo *board.Repository) tea.Cmd {
	return func() tea.Msg {
		tasks, err := repo.ListArchivedTasksContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return archiveStateMsg{tasks: tasks, reloadTasks: false}
	}
}

func restoreArchiveCmdContext(ctx context.Context, repo *board.Repository, id string) tea.Cmd {
	return func() tea.Msg {
		if _, err := repo.RestoreTaskContext(ctx, id); err != nil {
			return errMsg{err: err}
		}
		tasks, err := repo.ListArchivedTasksContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return archiveStateMsg{tasks: tasks, reloadTasks: true}
	}
}

func boardUseCmdContext(ctx context.Context, repo *board.BoardRepository, boardID string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.SetActiveBoardContext(ctx, boardID); err != nil {
			return errMsg{err: err}
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardActionMsg{boards: boards, active: active, reloadTasks: true, returnToBoard: true}
	}
}

func boardArchiveCmdContext(ctx context.Context, repo *board.BoardRepository, boardID string) tea.Cmd {
	return func() tea.Msg {
		if _, err := repo.ArchiveBoardContext(ctx, boardID); err != nil {
			return errMsg{err: err}
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardActionMsg{boards: boards, active: active, reloadTasks: true}
	}
}

func boardDeleteCmdContext(ctx context.Context, repo *board.BoardRepository, boardID string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.DeleteBoardContext(ctx, boardID); err != nil {
			return errMsg{err: err}
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardActionMsg{boards: boards, active: active, reloadTasks: true}
	}
}

func boardCreateCmdContext(ctx context.Context, repo *board.BoardRepository, name string) tea.Cmd {
	return func() tea.Msg {
		if _, err := repo.CreateBoardContext(ctx, name); err != nil {
			return errMsg{err: err}
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardActionMsg{boards: boards, active: active, reloadTasks: false}
	}
}

func boardRenameCmdContext(ctx context.Context, repo *board.BoardRepository, boardID, name string) tea.Cmd {
	return func() tea.Msg {
		if _, err := repo.RenameBoardContext(ctx, boardID, name); err != nil {
			return errMsg{err: err}
		}
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return boardActionMsg{boards: boards, active: active, reloadTasks: false}
	}
}

func taskArchiveCmdContext(ctx context.Context, repo *board.Repository, id string) tea.Cmd {
	return func() tea.Msg {
		if _, err := repo.ArchiveTaskContext(ctx, id); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func taskDeleteCmdContext(ctx context.Context, repo *board.Repository, id string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.DeleteTaskContext(ctx, id); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func taskUpdateTitleCmdContext(ctx context.Context, repo *board.Repository, id, title string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.UpdateTaskTitleContext(ctx, id, title); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func taskUpdateTagsCmdContext(ctx context.Context, repo *board.Repository, id string, tags []string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.UpdateTaskTagsContext(ctx, id, tags); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func taskUpdateContentCmdContext(ctx context.Context, repo *board.Repository, id, content string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.UpdateTaskContentContext(ctx, id, content); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func taskUpdatePriorityCmdContext(ctx context.Context, repo *board.Repository, id string, priority int) tea.Cmd {
	return func() tea.Msg {
		if err := repo.UpdateTaskPriorityContext(ctx, id, priority); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func openEditorCmd(repo *board.Repository, path string, editor string) tea.Cmd {
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
		return loadStateCmd(repo)()
	})
}

func resolveEditor(editor string) string {
	if strings.TrimSpace(editor) != "" {
		return editor
	}
	if env := strings.TrimSpace(os.Getenv("MOCHI_EDITOR")); env != "" {
		return env
	}
	if env := strings.TrimSpace(os.Getenv("EDITOR")); env != "" {
		return env
	}
	return "nano"
}

func taskCreateCmdContext(ctx context.Context, repo *board.Repository, title, status string, tags []string, priority int) tea.Cmd {
	return func() tea.Msg {
		task, err := board.NewTask(title)
		if err != nil {
			return errMsg{err: err}
		}
		if status != "" {
			task.Status = status
		}
		task.Tags = tags
		task.Priority = priority
		if _, err := repo.CreateTaskContext(ctx, task); err != nil {
			return errMsg{err: err}
		}
		return loadStateCmdContext(ctx, repo)()
	}
}

func updateStatusCmdContext(ctx context.Context, repo *board.Repository, id, status string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.UpdateTaskStatusContext(ctx, id, status); err != nil {
			return errMsg{err: err}
		}
		return statusUpdatedMsg{id: id, status: status}
	}
}

func buildColumns(columns []board.Column, tasks []board.Task) []columnModel {
	if len(columns) == 0 {
		return nil
	}
	taskIndex := make(map[string]board.Task, len(tasks))
	for _, t := range tasks {
		taskIndex[t.ID] = t
	}
	result := make([]columnModel, len(columns))
	index := make(map[string]int, len(columns))
	for i, column := range columns {
		result[i] = columnModel{
			Key:   column.Key,
			Title: column.Title,
		}
		index[strings.ToLower(column.Key)] = i
	}

	unknownIndex := -1
	for _, task := range tasks {
		key := strings.ToLower(task.Status)
		idx, ok := index[key]
		if !ok {
			if unknownIndex == -1 {
				result = append(result, columnModel{
					Key:   "unknown",
					Title: "Unknown",
				})
				unknownIndex = len(result) - 1
			}
			idx = unknownIndex
		}
		result[idx].Tasks = append(result[idx].Tasks, task)
	}
	for i := range result {
		sortTasksByReadiness(result[i].Tasks, taskIndex)
	}
	return result
}

func (m Model) moveSelection(delta int) Model {
	if len(m.columns) == 0 {
		return m
	}
	column := &m.columns[m.active]
	if len(column.Tasks) == 0 {
		return m
	}
	column.Selected += delta
	clampSelection(column)
	return m
}

func (m Model) moveSelectedTaskCmdContext(ctx context.Context) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	column := m.columns[m.active]
	if len(column.Tasks) == 0 {
		return nil
	}
	if column.Selected < 0 || column.Selected >= len(column.Tasks) {
		return nil
	}
	if m.active >= len(m.columns)-1 {
		return nil
	}

	task := column.Tasks[column.Selected]
	nextStatus := m.columns[m.active+1].Key
	return updateStatusCmdContext(ctx, m.repo, task.ID, nextStatus)
}

func (m Model) moveSelectedTaskBackCmdContext(ctx context.Context) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	column := m.columns[m.active]
	if len(column.Tasks) == 0 {
		return nil
	}
	if column.Selected < 0 || column.Selected >= len(column.Tasks) {
		return nil
	}
	if m.active == 0 {
		return nil
	}

	task := column.Tasks[column.Selected]
	prevStatus := m.columns[m.active-1].Key
	return updateStatusCmdContext(ctx, m.repo, task.ID, prevStatus)
}

func (m Model) handleBoardActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		if m.boardActionFromBoard {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		return m, nil
	case "j":
		m.boardAction++
		m.boardAction = clampIndex(m.boardAction, len(boardActionItems()))
		return m, nil
	case "k":
		m.boardAction--
		m.boardAction = clampIndex(m.boardAction, len(boardActionItems()))
		return m, nil
	case "enter":
		return m.handleBoardActionSelection()
	default:
		return m, nil
	}
}

func (m Model) handleBoardActionSelection() (tea.Model, tea.Cmd) {
	board, ok := m.selectedBoard()
	if !ok {
		if m.boardActionFromBoard {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		m.boardActionFromBoard = false
		return m, nil
	}
	switch boardActionItems()[m.boardAction] {
	case "use":
		m.boardActionFromBoard = false
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return boardUseCmdContext(ctx, m.boardRepo, board.ID)
		})
	case "rename":
		m.screen = screenBoardEdit
		m.boardEditFromBoard = m.boardActionFromBoard
		m.boardActionFromBoard = false
		m.editMode = editRename
		m.editInput = board.Name
		m.editBoardID = board.ID
		return m, nil
	case "edit description":
		if board.ID != m.activeBoard {
			m.pendingBoardDescEdit = true
			m.boardActionFromBoard = false
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return boardUseCmdContext(ctx, m.boardRepo, board.ID)
			})
		}
		if m.repo == nil {
			return m, func() tea.Msg {
				return errMsg{err: fmt.Errorf("tui: repository unavailable")}
			}
		}
		path, err := m.repo.BoardDescriptionPath()
		if err != nil {
			return m, func() tea.Msg {
				return errMsg{err: err}
			}
		}
		m.screen = screenBoard
		m.boardActionFromBoard = false
		return m, openEditorCmd(m.repo, path, m.editor)
	case "archive":
		m.screen = screenConfirm
		m.confirmAction = confirmArchiveBoard
		m.confirmBoard = board.ID
		m.confirmFromBoard = m.boardActionFromBoard
		m.boardActionFromBoard = false
		return m, nil
	case "delete":
		m.screen = screenConfirm
		m.confirmAction = confirmDeleteBoard
		m.confirmBoard = board.ID
		m.confirmFromBoard = m.boardActionFromBoard
		m.boardActionFromBoard = false
		return m, nil
	default:
		if m.boardActionFromBoard {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		m.boardActionFromBoard = false
		return m, nil
	}
}

func (m Model) handleBoardEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.boardEditFromBoard {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		} else if m.sidebarWidth() > 0 {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		m.boardEditFromBoard = false
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.editInput)
		if name == "" {
			return m, nil
		}
		if m.boardEditFromBoard {
			m.screen = screenBoard
			m.boardFocus = focusBoards
			m.boardCommandReturn = true
		} else if m.sidebarWidth() > 0 {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		m.boardEditFromBoard = false
		if m.editMode == editCreate {
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return boardCreateCmdContext(ctx, m.boardRepo, name)
			})
		}
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return boardRenameCmdContext(ctx, m.boardRepo, m.editBoardID, name)
		})
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.editInput) > 0 {
			m.editInput = m.editInput[:len(m.editInput)-1]
		}
		return m, nil
	case tea.KeySpace:
		m.editInput += " "
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.editInput += string(msg.Runes)
		}
		return m, nil
	}
}

func (m Model) handleBoardDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "b":
		if m.sidebarWidth() > 0 {
			m.screen = screenBoard
			m.boardFocus = focusBoards
		}
		return m, nil
	case "ctrl+r", "f5":
		return m.startRefresh()
	case "x":
		m.screen = screenBoardActions
		m.boardAction = 0
		m.boardActionFromBoard = true
		return m, nil
	case "e":
		if m.repo == nil {
			return m, func() tea.Msg {
				return errMsg{err: fmt.Errorf("tui: repository unavailable")}
			}
		}
		path, err := m.repo.BoardDescriptionPath()
		if err != nil {
			return m, func() tea.Msg {
				return errMsg{err: err}
			}
		}
		return m, openEditorCmd(m.repo, path, m.editor)
	default:
		return m, nil
	}
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "y":
		switch m.confirmAction {
		case confirmArchiveBoard:
			if m.confirmFromBoard {
				m.screen = screenBoard
				m.boardFocus = focusBoards
			}
			m.confirmFromBoard = false
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return boardArchiveCmdContext(ctx, m.boardRepo, m.confirmBoard)
			})
		case confirmDeleteBoard:
			if m.confirmFromBoard {
				m.screen = screenBoard
				m.boardFocus = focusBoards
			}
			m.confirmFromBoard = false
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return boardDeleteCmdContext(ctx, m.boardRepo, m.confirmBoard)
			})
		case confirmArchiveTask:
			m.screen = screenBoard
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskArchiveCmdContext(ctx, m.repo, m.confirmTask)
			})
		case confirmDeleteTask:
			m.screen = screenBoard
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskDeleteCmdContext(ctx, m.repo, m.confirmTask)
			})
		case confirmDeleteADR:
			m.screen = screenADR
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return deleteADRCmdContext(ctx, m.adrRoot(), m.confirmADR)
			})
		default:
			m.screen = screenBoard
			return m, nil
		}
	case "n", "esc":
		if m.confirmAction == confirmArchiveTask || m.confirmAction == confirmDeleteTask || m.confirmAction == confirmDeleteADR {
			m.screen = screenBoard
			if m.confirmAction == confirmDeleteADR {
				m.screen = screenADR
			}
		} else {
			if m.confirmFromBoard {
				m.screen = screenBoard
				m.boardFocus = focusBoards
			}
		}
		m.confirmFromBoard = false
		return m, nil
	default:
		return m, nil
	}
}

func boardActionItems() []string {
	return []string{"use", "rename", "edit description", "archive", "delete", "cancel"}
}

func taskActionItems() []string {
	return []string{
		"move forward",
		"move back",
		"change status",
		"archive task",
		"delete task",
		"open in editor",
		"cancel",
	}
}

func detailFieldCount() detailField {
	return fieldDescription + 1
}

func (m Model) selectedBoard() (board.Board, bool) {
	if len(m.boards) == 0 || m.boardIndex < 0 || m.boardIndex >= len(m.boards) {
		return board.Board{}, false
	}
	return m.boards[m.boardIndex], true
}

func clampBoardIndex(boards []board.Board, current int, active string) int {
	if len(boards) == 0 {
		return 0
	}
	for i, board := range boards {
		if board.ID == active {
			return i
		}
	}
	return clampIndex(current, len(boards))
}

func clampIndex(value, length int) int {
	if length <= 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= length {
		return length - 1
	}
	return value
}

func sortTasksByReadiness(tasks []board.Task, index map[string]board.Task) {
	sort.SliceStable(tasks, func(a, b int) bool {
		readyA, _ := board.IsReady(tasks[a], index)
		readyB, _ := board.IsReady(tasks[b], index)
		if readyA != readyB {
			return readyA // ready tasks first
		}
		left := effectivePriority(tasks[a].Priority)
		right := effectivePriority(tasks[b].Priority)
		if left != right {
			return left < right
		}
		leftTitle := strings.ToLower(tasks[a].Title)
		rightTitle := strings.ToLower(tasks[b].Title)
		if leftTitle != rightTitle {
			return leftTitle < rightTitle
		}
		return tasks[a].ID < tasks[b].ID
	})
}

func taskIndex(tasks []board.Task, id string) int {
	for i, task := range tasks {
		if task.ID == id {
			return i
		}
	}
	return 0
}

func buildTaskIndex(columns []columnModel) map[string]board.Task {
	index := make(map[string]board.Task)
	for _, col := range columns {
		for _, task := range col.Tasks {
			index[task.ID] = task
		}
	}
	return index
}

func effectivePriority(value int) int {
	if value == 0 {
		return board.DefaultPriority
	}
	return value
}

func (m Model) currentTaskExists() bool {
	_, ok := m.currentTask()
	return ok
}

func (m Model) currentTask() (board.Task, bool) {
	if len(m.columns) == 0 {
		return board.Task{}, false
	}
	column := m.columns[m.active]
	if len(column.Tasks) == 0 {
		return board.Task{}, false
	}
	if column.Selected < 0 || column.Selected >= len(column.Tasks) {
		return board.Task{}, false
	}
	return column.Tasks[column.Selected], true
}

func (m *Model) captureSelection() {
	if task, ok := m.currentTask(); ok {
		m.selectedTaskID = task.ID
	} else {
		m.selectedTaskID = ""
	}
}

func (m *Model) restoreSelection() {
	if strings.TrimSpace(m.selectedTaskID) == "" || len(m.columns) == 0 {
		m.selectedTaskID = ""
		return
	}
	for colIdx, column := range m.columns {
		for taskIdx, task := range column.Tasks {
			if task.ID == m.selectedTaskID {
				m.active = colIdx
				m.columns[colIdx].Selected = taskIdx
				m.selectedTaskID = ""
				return
			}
		}
	}
	m.selectedTaskID = ""
}

func (m Model) handleTaskActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "j":
		m.taskAction++
		m.taskAction = clampIndex(m.taskAction, len(taskActionItems()))
		return m, nil
	case "k":
		m.taskAction--
		m.taskAction = clampIndex(m.taskAction, len(taskActionItems()))
		return m, nil
	case "enter":
		return m.handleTaskActionSelection()
	default:
		return m, nil
	}
}

func (m Model) handleTaskDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenBoard
		return m, nil
	case tea.KeyTab:
		m.detailField = (m.detailField + 1) % detailFieldCount()
		return m, nil
	case tea.KeyEnter:
		switch m.detailField {
		case fieldTitle:
			return m.startTaskEdit(editTitle)
		case fieldTags:
			return m.startTaskEdit(editTags)
		case fieldDescription:
			return m.startTaskEdit(editDescription)
		case fieldStatus:
			m.screen = screenStatusPicker
			m.statusIndex = m.active
			return m, nil
		case fieldPriority:
			return m.startTaskEdit(editPriority)
		default:
			return m, nil
		}
	}

	switch normalizedKey(msg) {
	case "x":
		m.screen = screenTaskActions
		m.taskAction = 0
		return m, nil
	case "ctrl+r", "f5":
		return m.startRefresh()
	case "a":
		if task, ok := m.currentTask(); ok {
			m.screen = screenConfirm
			m.confirmAction = confirmArchiveTask
			m.confirmTask = task.ID
		}
		return m, nil
	case "d":
		if task, ok := m.currentTask(); ok {
			m.screen = screenConfirm
			m.confirmAction = confirmDeleteTask
			m.confirmTask = task.ID
		}
		return m, nil
	case "e":
		if task, ok := m.currentTask(); ok {
			return m, openEditorCmd(m.repo, task.FilePath, m.editor)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleTaskActionSelection() (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		m.screen = screenBoard
		return m, nil
	}
	switch taskActionItems()[m.taskAction] {
	case "move forward":
		m.screen = screenBoard
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return m.moveSelectedTaskCmdContext(ctx)
		})
	case "move back":
		m.screen = screenBoard
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return m.moveSelectedTaskBackCmdContext(ctx)
		})
	case "change status":
		m.screen = screenStatusPicker
		m.statusIndex = m.active
		return m, nil
	case "archive task":
		m.screen = screenConfirm
		m.confirmAction = confirmArchiveTask
		m.confirmTask = task.ID
		return m, nil
	case "delete task":
		m.screen = screenConfirm
		m.confirmAction = confirmDeleteTask
		m.confirmTask = task.ID
		return m, nil
	case "open in editor":
		m.screen = screenBoard
		return m, openEditorCmd(m.repo, task.FilePath, m.editor)
	default:
		m.screen = screenBoard
		return m, nil
	}
}

func (m Model) handleStatusPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "j":
		m.statusIndex++
		m.statusIndex = clampIndex(m.statusIndex, len(m.columns))
		return m, nil
	case "k":
		m.statusIndex--
		m.statusIndex = clampIndex(m.statusIndex, len(m.columns))
		return m, nil
	case "enter":
		task, ok := m.currentTask()
		if !ok {
			m.screen = screenBoard
			return m, nil
		}
		if m.statusIndex < 0 || m.statusIndex >= len(m.columns) {
			m.screen = screenBoard
			return m, nil
		}
		status := m.columns[m.statusIndex].Key
		m.screen = screenBoard
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return updateStatusCmdContext(ctx, m.repo, task.ID, status)
		})
	default:
		return m, nil
	}
}

func (m Model) handleTaskEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenTaskDetail
		return m, nil
	case tea.KeyEnter:
		task, ok := m.currentTask()
		if !ok {
			m.screen = screenBoard
			return m, nil
		}
		input := strings.TrimSpace(m.taskEditInput)
		m.taskEditInput = ""
		m.screen = screenTaskDetail
		switch m.taskEditMode {
		case editTitle:
			if input == "" {
				return m, nil
			}
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskUpdateTitleCmdContext(ctx, m.repo, task.ID, input)
			})
		case editTags:
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskUpdateTagsCmdContext(ctx, m.repo, task.ID, board.ParseTags(input))
			})
		case editDescription:
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskUpdateContentCmdContext(ctx, m.repo, task.ID, input)
			})
		case editPriority:
			value, err := strconv.Atoi(input)
			if err != nil {
				return m, func() tea.Msg {
					return errMsg{err: fmt.Errorf("tui: invalid priority %q", input)}
				}
			}
			return m.withInFlight(func(ctx context.Context) tea.Cmd {
				return taskUpdatePriorityCmdContext(ctx, m.repo, task.ID, value)
			})
		default:
			return m, nil
		}
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.taskEditInput) > 0 {
			m.taskEditInput = m.taskEditInput[:len(m.taskEditInput)-1]
		}
		return m, nil
	case tea.KeySpace:
		m.taskEditInput += " "
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.taskEditInput += string(msg.Runes)
		}
		return m, nil
	}
}

func (m Model) startTaskEdit(mode taskEditMode) (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		return m, nil
	}
	m.taskEditMode = mode
	switch mode {
	case editTitle:
		m.taskEditInput = task.Title
	case editTags:
		m.taskEditInput = strings.Join(task.Tags, ", ")
	case editDescription:
		m.taskEditInput = task.Content
	case editPriority:
		m.taskEditInput = fmt.Sprintf("%d", effectivePriority(task.Priority))
	}
	m.screen = screenTaskEdit
	return m, nil
}

func (m Model) handleArchiveKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "j":
		m.archiveIndex++
		m.archiveIndex = clampIndex(m.archiveIndex, len(m.archived))
		return m, nil
	case "k":
		m.archiveIndex--
		m.archiveIndex = clampIndex(m.archiveIndex, len(m.archived))
		return m, nil
	case "enter", "r":
		if m.archiveIndex < 0 || m.archiveIndex >= len(m.archived) {
			return m, nil
		}
		task := m.archived[m.archiveIndex]
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return restoreArchiveCmdContext(ctx, m.repo, task.ID)
		})
	default:
		return m, nil
	}
}

func normalizedKey(msg tea.KeyMsg) string {
	key := strings.ToLower(msg.String())
	switch key {
	case "up":
		return "k"
	case "down":
		return "j"
	case "left":
		return "h"
	case "right":
		return "l"
	default:
		return key
	}
}

func (m Model) handleTaskCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenBoard
		return m, nil
	case tea.KeyEnter:
		title := strings.TrimSpace(m.taskTitle)
		if title == "" {
			return m, nil
		}
		tags := board.ParseTags(m.taskTags)
		status := m.taskStatus
		priority := m.taskPriority
		m.screen = screenBoard
		m.taskTitle = ""
		m.taskTags = ""
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return taskCreateCmdContext(ctx, m.repo, title, status, tags, priority)
		})
	case tea.KeyTab:
		m.taskField = (m.taskField + 1) % 3
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if m.taskField == 0 {
			if len(m.taskTitle) > 0 {
				m.taskTitle = m.taskTitle[:len(m.taskTitle)-1]
			}
			return m, nil
		}
		if m.taskField == 2 && len(m.taskTags) > 0 {
			m.taskTags = m.taskTags[:len(m.taskTags)-1]
		}
		return m, nil
	case tea.KeySpace:
		if m.taskField == 0 {
			m.taskTitle += " "
			return m, nil
		}
		if m.taskField == 2 {
			m.taskTags += " "
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			if m.taskField == 0 {
				m.taskTitle += string(msg.Runes)
				return m, nil
			}
			if m.taskField == 1 {
				for _, r := range msg.Runes {
					if r >= '1' && r <= '3' {
						m.taskPriority = int(r - '0')
						break
					}
				}
				return m, nil
			}
			m.taskTags += string(msg.Runes)
		}
		return m, nil
	}
}

func (m Model) applyStatusUpdate(id, status string) Model {
	fromCol, fromIndex, task := findTask(m.columns, id)
	if fromCol == -1 {
		return m
	}
	m.columns[fromCol].Tasks = append(m.columns[fromCol].Tasks[:fromIndex], m.columns[fromCol].Tasks[fromIndex+1:]...)
	clampSelection(&m.columns[fromCol])

	task.Status = status
	toCol := findColumn(m.columns, status)
	if toCol == -1 {
		return m
	}
	m.columns[toCol].Tasks = append(m.columns[toCol].Tasks, task)
	sortTasksByReadiness(m.columns[toCol].Tasks, buildTaskIndex(m.columns))
	m.columns[toCol].Selected = taskIndex(m.columns[toCol].Tasks, task.ID)
	m.active = toCol
	return m
}

func findTask(columns []columnModel, id string) (int, int, board.Task) {
	for colIndex, column := range columns {
		for taskIndex, task := range column.Tasks {
			if task.ID == id {
				return colIndex, taskIndex, task
			}
		}
	}
	return -1, -1, board.Task{}
}

func findColumn(columns []columnModel, status string) int {
	target := strings.ToLower(status)
	for i, column := range columns {
		if strings.ToLower(column.Key) == target {
			return i
		}
	}
	return -1
}

func clampSelection(column *columnModel) {
	if len(column.Tasks) == 0 {
		column.Selected = 0
		return
	}
	if column.Selected < 0 {
		column.Selected = 0
		return
	}
	if column.Selected >= len(column.Tasks) {
		column.Selected = len(column.Tasks) - 1
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
