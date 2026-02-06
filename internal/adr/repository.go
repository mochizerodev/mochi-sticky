package adr

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"mochi-sticky/internal/shared"
)

// Repository manages ADR files on disk.
type Repository struct {
	mu           sync.RWMutex
	root         string
	configPath   string
	templatesDir string
	now          func() time.Time
}

// NewRepository creates an ADR repository rooted at root (e.g., `<storageRoot>/adrs`).
func NewRepository(root string) (*Repository, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("adr: root is required: %w", shared.ErrInvalidPath)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("adr: failed to resolve root: %w", err)
	}
	repo := &Repository{
		root:         abs,
		configPath:   filepath.Join(abs, ConfigFileName),
		templatesDir: filepath.Join(abs, TemplatesDirName),
		now:          time.Now,
	}
	if err := shared.EnsureInDir(abs, repo.configPath); err != nil {
		return nil, err
	}
	if err := shared.EnsureInDir(abs, repo.templatesDir); err != nil {
		return nil, err
	}
	return repo, nil
}

// Root returns the ADR root directory.
func (r *Repository) Root() string {
	return r.root
}

// InitStore ensures the ADR root, templates directory, and config exist.
func (r *Repository) InitStore() error {
	return r.InitStoreContext(context.Background())
}

// InitStoreContext ensures the ADR root, templates directory, and config exist, honoring ctx cancellation.
func (r *Repository) InitStoreContext(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(r.root, 0o755); err != nil {
		return fmt.Errorf("adr: failed to create adr root %s: %w", r.root, err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(r.templatesDir, 0o755); err != nil {
		return fmt.Errorf("adr: failed to create templates dir %s: %w", r.templatesDir, err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if _, err := os.Stat(r.configPath); err == nil {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("adr: failed to stat config %s: %w", r.configPath, err)
	}

	data, err := RenderConfig()
	if err != nil {
		return fmt.Errorf("adr: failed to render config: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(r.configPath, data, 0o644); err != nil {
		return fmt.Errorf("adr: failed to write config %s: %w", r.configPath, err)
	}
	return nil
}

// LoadConfig loads the ADR config file (or returns the default config when missing).
func (r *Repository) LoadConfig() (Config, error) {
	return r.LoadConfigContext(context.Background())
}

// LoadConfigContext loads the ADR config file (or returns the default config when missing), honoring ctx cancellation.
func (r *Repository) LoadConfigContext(ctx context.Context) (Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return loadConfigContext(ctx, r.root)
}

// SaveConfig writes the ADR config file to disk.
func (r *Repository) SaveConfig(cfg Config) error {
	return r.SaveConfigContext(context.Background(), cfg)
}

// SaveConfigContext writes the ADR config file to disk, honoring ctx cancellation.
func (r *Repository) SaveConfigContext(ctx context.Context, cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return saveConfigContext(ctx, r.root, cfg)
}

// ListADRs loads all ADRs from the root directory.
func (r *Repository) ListADRs() ([]ADR, error) {
	return r.ListADRsContext(context.Background())
}

// ListADRsContext loads all ADRs from the root directory, honoring ctx cancellation.
func (r *Repository) ListADRsContext(ctx context.Context) ([]ADR, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.listADRsLockedContext(ctx)
}

func (r *Repository) listADRsLockedContext(ctx context.Context) ([]ADR, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if _, err := os.Stat(r.root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("adr: failed to stat adr root %s: %w", r.root, err)
	}

	entries, err := os.ReadDir(r.root)
	if err != nil {
		return nil, fmt.Errorf("adr: failed to read adr root %s: %w", r.root, err)
	}

	adrs := make([]ADR, 0)
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(r.root, entry.Name())
		if err := shared.EnsureInDir(r.root, path); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		adr, err := LoadADR(path)
		if err != nil {
			return nil, err
		}
		adrs = append(adrs, adr)
	}
	SortADRs(adrs)
	return adrs, nil
}

// GetADRByID loads an ADR by its numeric ID.
func (r *Repository) GetADRByID(id int) (ADR, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path, err := r.findPathByIDLocked(id)
	if err != nil {
		return ADR{}, err
	}
	return LoadADR(path)
}

// CreateOptions controls ADR creation behaviour.
type CreateOptions struct {
	Status string
	Date   time.Time
	Tags   []string
	Links  []string
	Body   string
}

// CreateADR creates a new ADR file, allocating the next sequential ID via config.
func (r *Repository) CreateADR(title string, opts CreateOptions) (ADR, error) {
	return r.CreateADRContext(context.Background(), title, opts)
}

// CreateADRContext creates a new ADR file, allocating the next sequential ID via config,
// honoring ctx cancellation.
func (r *Repository) CreateADRContext(ctx context.Context, title string, opts CreateOptions) (ADR, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ADR{}, ctx.Err()
	default:
	}
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return ADR{}, fmt.Errorf("adr: %w", ErrInvalidTitle)
	}

	select {
	case <-ctx.Done():
		return ADR{}, ctx.Err()
	default:
	}
	if err := os.MkdirAll(r.root, 0o755); err != nil {
		return ADR{}, fmt.Errorf("adr: failed to create adr root %s: %w", r.root, err)
	}
	if err := os.MkdirAll(r.templatesDir, 0o755); err != nil {
		return ADR{}, fmt.Errorf("adr: failed to create templates dir %s: %w", r.templatesDir, err)
	}

	cfg, err := loadConfigContext(ctx, r.root)
	if err != nil {
		return ADR{}, err
	}
	maxID, err := r.maxIDLocked()
	if err != nil {
		return ADR{}, err
	}
	if cfg.NextID <= maxID {
		cfg.NextID = maxID + 1
	}

	id := cfg.NextID
	cfg.NextID++
	select {
	case <-ctx.Done():
		return ADR{}, ctx.Err()
	default:
	}
	if err := saveConfigContext(ctx, r.root, cfg); err != nil {
		return ADR{}, err
	}

	status := strings.TrimSpace(opts.Status)
	if status == "" {
		status = cfg.Columns[0].Key
	}
	when := opts.Date
	if when.IsZero() {
		when = r.now()
	}
	body := opts.Body
	if strings.TrimSpace(body) == "" {
		body = DefaultContent()
	}

	uid, err := newUID()
	if err != nil {
		return ADR{}, fmt.Errorf("adr: failed to generate uid: %w", err)
	}
	record := ADR{
		ID:      id,
		UID:     uid,
		Title:   trimmedTitle,
		Status:  status,
		Date:    Date{Time: when},
		Tags:    normalizeStrings(opts.Tags),
		Links:   normalizeStrings(opts.Links),
		Content: body,
	}

	slug := Slugify(trimmedTitle)
	if slug == "" {
		slug = "adr"
	}
	filename := fmt.Sprintf("%s-%s.md", FormatID(record.ID), slug)
	filePath := filepath.Join(r.root, filename)
	if err := shared.EnsureInDir(r.root, filePath); err != nil {
		return ADR{}, err
	}
	select {
	case <-ctx.Done():
		return ADR{}, ctx.Err()
	default:
	}
	if _, err := os.Stat(filePath); err == nil {
		return ADR{}, fmt.Errorf("adr: adr already exists: %s", filename)
	} else if err != nil && !os.IsNotExist(err) {
		return ADR{}, fmt.Errorf("adr: failed to stat adr file %s: %w", filePath, err)
	}
	select {
	case <-ctx.Done():
		return ADR{}, ctx.Err()
	default:
	}
	if err := SaveADR(filePath, record); err != nil {
		return ADR{}, err
	}
	record.FilePath = filePath
	return record, nil
}

// UpdateADRStatus updates an ADR status by ID.
func (r *Repository) UpdateADRStatus(id int, status string) error {
	return r.UpdateADRStatusContext(context.Background(), id, status)
}

// DeleteADR deletes an ADR by ID.
func (r *Repository) DeleteADR(id int) error {
	return r.DeleteADRContext(context.Background(), id)
}

// DeleteADRContext deletes an ADR by ID, honoring ctx cancellation.
func (r *Repository) DeleteADRContext(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := r.findPathByIDLocked(id)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("adr: %w", ErrADRNotFound)
		}
		return fmt.Errorf("adr: failed to delete adr %s: %w", path, err)
	}
	return nil
}

// UpdateADRStatusContext updates an ADR status by ID, honoring ctx cancellation.
func (r *Repository) UpdateADRStatusContext(ctx context.Context, id int, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := r.findPathByIDLocked(id)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	record, err := LoadADR(path)
	if err != nil {
		return err
	}
	record.Status = strings.TrimSpace(status)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := SaveADR(path, record); err != nil {
		return err
	}
	return nil
}

func (r *Repository) findPathByIDLocked(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("adr: %w", ErrInvalidID)
	}
	if _, err := os.Stat(r.root); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("adr: %w", ErrADRNotFound)
		}
		return "", fmt.Errorf("adr: failed to stat adr root %s: %w", r.root, err)
	}

	prefix := FormatID(id)
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return "", fmt.Errorf("adr: failed to read adr root %s: %w", r.root, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		name := entry.Name()
		if len(name) < len(prefix) {
			continue
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if len(name) > len(prefix) {
			next := name[len(prefix)]
			if next != '-' && next != '.' {
				continue
			}
		}
		path := filepath.Join(r.root, name)
		if err := shared.EnsureInDir(r.root, path); err != nil {
			return "", err
		}
		return path, nil
	}
	return "", fmt.Errorf("adr: %w", ErrADRNotFound)
}

func (r *Repository) maxIDLocked() (int, error) {
	if _, err := os.Stat(r.root); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("adr: failed to stat adr root %s: %w", r.root, err)
	}
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return 0, fmt.Errorf("adr: failed to read adr root %s: %w", r.root, err)
	}
	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		if id, ok := parseIDFromFilename(entry.Name()); ok && id > maxID {
			maxID = id
		}
	}
	return maxID, nil
}

func normalizeStrings(values []string) []string {
	clean := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		clean = append(clean, trimmed)
	}
	sort.Strings(clean)
	return clean
}

func newUID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(raw[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32]), nil
}
