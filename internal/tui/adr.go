package tui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"

	tea "github.com/charmbracelet/bubbletea"
)

type adrStateMsg struct {
	columns []adr.Column
	adrs    []adr.ADR
}

type adrStatusUpdatedMsg struct {
	id     int
	status string
}

type adrCreatedMsg struct {
	record adr.ADR
}

func loadADRCmd(root string) tea.Cmd {
	return loadADRCmdContext(context.Background(), root)
}

func loadADRCmdContext(ctx context.Context, root string) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		repo, err := adr.NewRepository(root)
		if err != nil {
			return errMsg{err: err}
		}
		if err := repo.InitStoreContext(ctx); err != nil {
			return errMsg{err: err}
		}
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
		adrs, err := repo.ListADRsContext(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return adrStateMsg{columns: config.Columns, adrs: adrs}
	}
}

func updateADRStatusCmd(root string, id int, status string) tea.Cmd {
	return func() tea.Msg {
		repo, err := adr.NewRepository(root)
		if err != nil {
			return errMsg{err: err}
		}
		if err := repo.UpdateADRStatus(id, status); err != nil {
			return errMsg{err: err}
		}
		return adrStatusUpdatedMsg{id: id, status: status}
	}
}

func createADRMsgContext(ctx context.Context, root string, title string, status string, tags []string) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		repo, err := adr.NewRepository(root)
		if err != nil {
			return errMsg{err: err}
		}
		if err := repo.InitStoreContext(ctx); err != nil {
			return errMsg{err: err}
		}
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		record, err := repo.CreateADR(title, adr.CreateOptions{
			Status: status,
			Tags:   tags,
		})
		if err != nil {
			return errMsg{err: err}
		}
		return adrCreatedMsg{record: record}
	}
}

func deleteADRCmdContext(ctx context.Context, root string, id int) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return errMsg{err: ctx.Err()}
		default:
		}
		repo, err := adr.NewRepository(root)
		if err != nil {
			return errMsg{err: err}
		}
		if err := repo.DeleteADRContext(ctx, id); err != nil {
			return errMsg{err: err}
		}
		return loadADRCmdContext(ctx, root)()
	}
}

func openADREditorCmd(root, path string, editor string) tea.Cmd {
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
		return loadADRCmd(root)()
	})
}

func buildADRColumns(columns []adr.Column, adrs []adr.ADR) []adrColumnModel {
	if len(columns) == 0 {
		return nil
	}
	result := make([]adrColumnModel, len(columns))
	index := make(map[string]int, len(columns))
	for i, column := range columns {
		result[i] = adrColumnModel{
			Key:   column.Key,
			Title: column.Title,
		}
		index[strings.ToLower(column.Key)] = i
	}

	unknownIndex := -1
	for _, record := range adrs {
		key := strings.ToLower(strings.TrimSpace(record.Status))
		idx, ok := index[key]
		if !ok {
			if unknownIndex == -1 {
				result = append(result, adrColumnModel{
					Key:   "unknown",
					Title: "Unknown",
				})
				unknownIndex = len(result) - 1
			}
			idx = unknownIndex
		}
		result[idx].ADRs = append(result[idx].ADRs, record)
	}
	for i := range result {
		adr.SortADRs(result[i].ADRs)
	}
	return result
}

func (m *Model) captureADRSelection() {
	if record, ok := m.currentADR(); ok {
		m.selectedADRID = record.ID
		return
	}
	m.selectedADRID = 0
}

func (m *Model) restoreADRSelection() {
	if m.selectedADRID <= 0 || len(m.adrColumns) == 0 {
		m.selectedADRID = 0
		return
	}
	for colIdx, column := range m.adrColumns {
		for adrIdx, record := range column.ADRs {
			if record.ID == m.selectedADRID {
				m.adrActive = colIdx
				m.adrColumns[colIdx].Selected = adrIdx
				m.selectedADRID = 0
				return
			}
		}
	}
	m.selectedADRID = 0
}

func (m Model) currentADR() (adr.ADR, bool) {
	if len(m.adrColumns) == 0 {
		return adr.ADR{}, false
	}
	column := m.adrColumns[m.adrActive]
	if len(column.ADRs) == 0 {
		return adr.ADR{}, false
	}
	if column.Selected < 0 || column.Selected >= len(column.ADRs) {
		return adr.ADR{}, false
	}
	return column.ADRs[column.Selected], true
}

func clampADRSelection(column *adrColumnModel) {
	if len(column.ADRs) == 0 {
		column.Selected = 0
		return
	}
	if column.Selected < 0 {
		column.Selected = 0
		return
	}
	if column.Selected >= len(column.ADRs) {
		column.Selected = len(column.ADRs) - 1
	}
}

func (m Model) moveADRSelection(delta int) Model {
	if len(m.adrColumns) == 0 {
		return m
	}
	column := &m.adrColumns[m.adrActive]
	if len(column.ADRs) == 0 {
		return m
	}
	column.Selected += delta
	clampADRSelection(column)
	return m
}

func (m Model) moveSelectedADRCmd() tea.Cmd {
	if len(m.adrColumns) == 0 {
		return nil
	}
	column := m.adrColumns[m.adrActive]
	if len(column.ADRs) == 0 {
		return nil
	}
	if column.Selected < 0 || column.Selected >= len(column.ADRs) {
		return nil
	}
	if m.adrActive >= len(m.adrColumns)-1 {
		return nil
	}
	nextStatus := m.adrColumns[m.adrActive+1].Key
	if nextStatus == "unknown" {
		return nil
	}
	record := column.ADRs[column.Selected]
	return updateADRStatusCmd(m.adrRoot(), record.ID, nextStatus)
}

func (m Model) moveSelectedADRBackCmd() tea.Cmd {
	if len(m.adrColumns) == 0 {
		return nil
	}
	column := m.adrColumns[m.adrActive]
	if len(column.ADRs) == 0 {
		return nil
	}
	if column.Selected < 0 || column.Selected >= len(column.ADRs) {
		return nil
	}
	if m.adrActive == 0 {
		return nil
	}
	prevStatus := m.adrColumns[m.adrActive-1].Key
	if prevStatus == "unknown" {
		return nil
	}
	record := column.ADRs[column.Selected]
	return updateADRStatusCmd(m.adrRoot(), record.ID, prevStatus)
}

func (m Model) applyADRStatusUpdate(id int, status string) Model {
	fromCol, fromIndex, record := findADR(m.adrColumns, id)
	if fromCol == -1 {
		return m
	}
	m.adrColumns[fromCol].ADRs = append(m.adrColumns[fromCol].ADRs[:fromIndex], m.adrColumns[fromCol].ADRs[fromIndex+1:]...)
	clampADRSelection(&m.adrColumns[fromCol])

	record.Status = status
	toCol := findADRColumn(m.adrColumns, status)
	if toCol == -1 {
		return m
	}
	m.adrColumns[toCol].ADRs = append(m.adrColumns[toCol].ADRs, record)
	adr.SortADRs(m.adrColumns[toCol].ADRs)
	m.adrColumns[toCol].Selected = adrIndex(m.adrColumns[toCol].ADRs, record.ID)
	m.adrActive = toCol
	return m
}

func findADR(columns []adrColumnModel, id int) (int, int, adr.ADR) {
	for colIndex, column := range columns {
		for adrIndex, record := range column.ADRs {
			if record.ID == id {
				return colIndex, adrIndex, record
			}
		}
	}
	return -1, -1, adr.ADR{}
}

func findADRColumn(columns []adrColumnModel, status string) int {
	target := strings.ToLower(strings.TrimSpace(status))
	for i, column := range columns {
		if strings.ToLower(strings.TrimSpace(column.Key)) == target {
			return i
		}
	}
	return -1
}

func adrIndex(adrs []adr.ADR, id int) int {
	for i, record := range adrs {
		if record.ID == id {
			return i
		}
	}
	return 0
}

func (m Model) adrRoot() string {
	if m.repo != nil {
		return filepath.Join(m.repo.StorageRoot(), "adrs")
	}
	return filepath.Join(m.baseDir, ".sticky", "adrs")
}

func (m Model) handleADRKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "M" {
		return m, m.moveSelectedADRBackCmd()
	}
	switch normalizedKey(msg) {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "b":
		m.screen = screenBoard
		return m, nil
	case "ctrl+r", "f5":
		m.captureADRSelection()
		m.loading = true
		m.loadingMessage = "Loading ADRs..."
		return m, loadADRCmd(m.adrRoot())
	case "h":
		if m.adrActive > 0 {
			m.adrActive--
		}
		return m, nil
	case "l":
		if m.adrActive < len(m.adrColumns)-1 {
			m.adrActive++
		}
		return m, nil
	case "j":
		m = m.moveADRSelection(1)
		return m, nil
	case "k":
		m = m.moveADRSelection(-1)
		return m, nil
	case "enter", "i":
		if _, ok := m.currentADR(); ok {
			m.screen = screenADRDetail
		}
		return m, nil
	case "a":
		m.adrTitle = ""
		m.adrTags = ""
		m.adrField = 0
		m.adrStatus = m.currentADRCreateStatus()
		m.screen = screenADRCreate
		return m, nil
	case "m":
		return m, m.moveSelectedADRCmd()
	case "e":
		if record, ok := m.currentADR(); ok && strings.TrimSpace(record.FilePath) != "" {
			m.selectedADRID = record.ID
			m.loading = true
			m.loadingMessage = "Opening editor..."
			return m, openADREditorCmd(m.adrRoot(), record.FilePath, m.editor)
		}
		return m, nil
	case "x":
		m.screen = screenADRActions
		m.adrAction = 0
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) currentADRCreateStatus() string {
	if len(m.adrStatusColumns) == 0 {
		return "proposed"
	}
	if len(m.adrColumns) > 0 && m.adrActive >= 0 && m.adrActive < len(m.adrColumns) {
		key := strings.TrimSpace(m.adrColumns[m.adrActive].Key)
		if key != "" && key != "unknown" && findConfiguredADRStatusIndex(m.adrStatusColumns, key) >= 0 {
			return key
		}
	}
	return strings.TrimSpace(m.adrStatusColumns[0].Key)
}

func (m Model) handleADRActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenADR
		return m, nil
	case "j":
		m.adrAction++
		m.adrAction = clampIndex(m.adrAction, len(adrActionItems()))
		return m, nil
	case "k":
		m.adrAction--
		m.adrAction = clampIndex(m.adrAction, len(adrActionItems()))
		return m, nil
	case "enter":
		return m.handleADRActionSelection()
	default:
		return m, nil
	}
}

func adrActionItems() []string {
	return []string{
		"move forward",
		"move back",
		"change status",
		"open in editor",
		"delete adr",
		"create adr",
		"cancel",
	}
}

func (m Model) handleADRActionSelection() (tea.Model, tea.Cmd) {
	switch adrActionItems()[m.adrAction] {
	case "move forward":
		m.screen = screenADR
		return m, m.moveSelectedADRCmd()
	case "move back":
		m.screen = screenADR
		return m, m.moveSelectedADRBackCmd()
	case "change status":
		record, ok := m.currentADR()
		if !ok {
			m.screen = screenADR
			return m, nil
		}
		m.screen = screenADRStatusPicker
		m.adrStatusIndex = max(0, findConfiguredADRStatusIndex(m.adrStatusColumns, record.Status))
		return m, nil
	case "open in editor":
		record, ok := m.currentADR()
		if !ok || strings.TrimSpace(record.FilePath) == "" {
			m.screen = screenADR
			return m, nil
		}
		m.selectedADRID = record.ID
		m.screen = screenADR
		m.loading = true
		m.loadingMessage = "Opening editor..."
		return m, openADREditorCmd(m.adrRoot(), record.FilePath, m.editor)
	case "create adr":
		m.screen = screenADR
		m.adrTitle = ""
		m.adrTags = ""
		m.adrField = 0
		m.adrStatus = m.currentADRCreateStatus()
		m.screen = screenADRCreate
		return m, nil
	case "delete adr":
		record, ok := m.currentADR()
		if !ok {
			m.screen = screenADR
			return m, nil
		}
		m.screen = screenConfirm
		m.confirmAction = confirmDeleteADR
		m.confirmADR = record.ID
		return m, nil
	default:
		m.screen = screenADR
		return m, nil
	}
}

func (m Model) handleADRStatusPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "esc":
		m.screen = screenADR
		return m, nil
	case "j":
		m.adrStatusIndex++
		m.adrStatusIndex = clampIndex(m.adrStatusIndex, len(m.adrStatusColumns))
		return m, nil
	case "k":
		m.adrStatusIndex--
		m.adrStatusIndex = clampIndex(m.adrStatusIndex, len(m.adrStatusColumns))
		return m, nil
	case "enter":
		record, ok := m.currentADR()
		if !ok {
			m.screen = screenADR
			return m, nil
		}
		if m.adrStatusIndex < 0 || m.adrStatusIndex >= len(m.adrStatusColumns) {
			m.screen = screenADR
			return m, nil
		}
		status := m.adrStatusColumns[m.adrStatusIndex].Key
		m.screen = screenADR
		return m, updateADRStatusCmd(m.adrRoot(), record.ID, status)
	default:
		return m, nil
	}
}

func (m Model) handleADRCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenADR
		return m, nil
	case tea.KeyEnter:
		title := strings.TrimSpace(m.adrTitle)
		if title == "" {
			return m, nil
		}
		status := strings.TrimSpace(m.adrStatus)
		if status == "" {
			status = m.currentADRCreateStatus()
		}
		tags := board.ParseTags(m.adrTags)
		m.screen = screenADR
		m.adrTitle = ""
		m.adrTags = ""
		m.loading = true
		m.loadingMessage = "Creating ADR..."
		return m.withInFlight(func(ctx context.Context) tea.Cmd {
			return createADRMsgContext(ctx, m.adrRoot(), title, status, tags)
		})
	case tea.KeyTab:
		m.adrField = (m.adrField + 1) % 2
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if m.adrField == 0 {
			if len(m.adrTitle) > 0 {
				m.adrTitle = m.adrTitle[:len(m.adrTitle)-1]
			}
			return m, nil
		}
		if len(m.adrTags) > 0 {
			m.adrTags = m.adrTags[:len(m.adrTags)-1]
		}
		return m, nil
	case tea.KeySpace:
		if m.adrField == 0 {
			m.adrTitle += " "
			return m, nil
		}
		m.adrTags += " "
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			if m.adrField == 0 {
				m.adrTitle += string(msg.Runes)
				return m, nil
			}
			m.adrTags += string(msg.Runes)
		}
		return m, nil
	}
}

func (m Model) handleADRDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch normalizedKey(msg) {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "b":
		m.screen = screenADR
		return m, nil
	case "ctrl+r", "f5":
		m.captureADRSelection()
		m.loading = true
		m.loadingMessage = "Loading ADRs..."
		m.screen = screenADR
		return m, loadADRCmd(m.adrRoot())
	case "s":
		record, ok := m.currentADR()
		if !ok {
			return m, nil
		}
		m.screen = screenADRStatusPicker
		m.adrStatusIndex = max(0, findConfiguredADRStatusIndex(m.adrStatusColumns, record.Status))
		return m, nil
	case "e":
		record, ok := m.currentADR()
		if !ok || strings.TrimSpace(record.FilePath) == "" {
			return m, nil
		}
		m.selectedADRID = record.ID
		m.loading = true
		m.loadingMessage = "Opening editor..."
		m.screen = screenADR
		return m, openADREditorCmd(m.adrRoot(), record.FilePath, m.editor)
	case "x":
		m.screen = screenADRActions
		m.adrAction = 0
		return m, nil
	default:
		return m, nil
	}
}

func findConfiguredADRStatusIndex(columns []adr.Column, status string) int {
	needle := strings.ToLower(strings.TrimSpace(status))
	if needle == "" {
		return -1
	}
	for i, column := range columns {
		if strings.ToLower(strings.TrimSpace(column.Key)) == needle {
			return i
		}
	}
	return -1
}
