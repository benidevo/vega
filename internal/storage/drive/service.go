package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/benidevo/vega/internal/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	fileName        = "vega-ai-data.json.gz"
	fileDescription = "Vega AI User Data"
	mimeType        = "application/gzip"
	appDataFolder   = "appDataFolder" // Special folder in Google Drive
)

// DriveStorage implements UserStorage with Google Drive backend
type DriveStorage struct {
	client   *drive.Service
	cache    storage.UserStorage
	userID   string
	fileID   string // Cached file ID
	mu       sync.RWMutex
	syncMu   sync.Mutex
	lastSync time.Time
	isDirty  bool
}

// NewDriveStorage creates a new Google Drive storage instance
func NewDriveStorage(ctx context.Context, userID string, token *oauth2.Token, config *oauth2.Config, cache storage.UserStorage) (*DriveStorage, error) {
	client := config.Client(ctx, token)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	ds := &DriveStorage{
		client: service,
		cache:  cache,
		userID: userID,
	}

	// Find or create the data file
	if err := ds.initializeFile(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize file: %w", err)
	}

	// Initial sync from Drive to cache
	if err := ds.downloadAndSync(ctx); err != nil {
		// Log error but don't fail - cache will work offline
		fmt.Printf("Warning: initial sync failed: %v\n", err)
	}

	return ds, nil
}

// initializeFile finds or creates the data file in Google Drive
func (ds *DriveStorage) initializeFile(ctx context.Context) error {
	// Search for existing file
	query := fmt.Sprintf("name='%s' and 'appDataFolder' in parents and trashed=false", fileName)
	fileList, err := ds.client.Files.List().
		Context(ctx).
		Q(query).
		Spaces("appDataFolder").
		Fields("files(id, name, createdTime)").
		Do()

	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(fileList.Files) > 0 {
		// File exists, use it
		ds.fileID = fileList.Files[0].Id
		return nil
	}

	// Create new file
	file := &drive.File{
		Name:        fileName,
		Description: fileDescription,
		MimeType:    mimeType,
		Parents:     []string{appDataFolder},
	}

	// Create empty compressed JSON as initial content
	emptyDoc := storage.UserDocument{
		UpdatedAt: time.Now(),
		Data:      storage.UserDataCore{},
	}
	jsonData, _ := emptyDoc.ToJSON(false)
	compressed, err := compressJSON(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress initial data: %w", err)
	}

	createdFile, err := ds.client.Files.Create(file).
		Context(ctx).
		Media(bytes.NewReader(compressed)).
		Do()

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	ds.fileID = createdFile.Id
	return nil
}

// downloadAndSync downloads data from Drive and syncs to cache
func (ds *DriveStorage) downloadAndSync(ctx context.Context) error {
	if ds.fileID == "" {
		return fmt.Errorf("file ID not initialized")
	}

	// Download file
	resp, err := ds.client.Files.Get(ds.fileID).
		Context(ctx).
		Download()

	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Read compressed data
	compressed, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Decompress
	jsonData, err := decompressJSON(compressed)
	if err != nil {
		return fmt.Errorf("failed to decompress data: %w", err)
	}

	// Parse document
	doc, err := storage.FromJSON(jsonData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Sync to cache
	if err := ds.syncToCache(ctx, doc); err != nil {
		return fmt.Errorf("failed to sync to cache: %w", err)
	}

	ds.lastSync = time.Now()
	ds.isDirty = false

	return nil
}

// syncToCache updates the cache with data from the document
func (ds *DriveStorage) syncToCache(ctx context.Context, doc *storage.UserDocument) error {
	// Save profile
	if doc.Data.Profile != nil && doc.Data.Profile.ID != 0 {
		if err := ds.cache.SaveProfile(ctx, doc.Data.Profile); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}
	}

	// Save companies and jobs
	for _, company := range doc.Data.Companies {
		if err := ds.cache.SaveCompany(ctx, company); err != nil {
			return fmt.Errorf("failed to save company %d: %w", company.ID, err)
		}
	}

	for _, job := range doc.Data.Jobs {
		if err := ds.cache.SaveJob(ctx, job); err != nil {
			return fmt.Errorf("failed to save job %d: %w", job.ID, err)
		}
	}

	// Save matches
	for _, match := range doc.Data.Matches {
		if err := ds.cache.SaveMatchResult(ctx, match); err != nil {
			return fmt.Errorf("failed to save match %d: %w", match.ID, err)
		}
	}

	return nil
}

// uploadFromCache creates a document from cache and uploads to Drive
func (ds *DriveStorage) uploadFromCache(ctx context.Context) error {
	// Build document from cache
	doc := storage.UserDocument{
		UpdatedAt: time.Now(),
		Data:      storage.UserDataCore{},
	}

	// Get profile
	profile, err := ds.cache.GetProfile(ctx)
	if err == nil && profile != nil {
		doc.Data.Profile = profile
	}

	// Get companies
	companies, err := ds.cache.ListCompanies(ctx)
	if err == nil {
		doc.Data.Companies = companies
	}

	// Get all jobs
	for _, company := range companies {
		jobs, err := ds.cache.ListJobs(ctx, company.ID)
		if err == nil {
			doc.Data.Jobs = append(doc.Data.Jobs, jobs...)
		}
	}

	// Get matches
	matches, err := ds.cache.GetMatchHistory(ctx, 0) // 0 means get all
	if err == nil {
		doc.Data.Matches = matches
	}

	// Update checksum
	doc.UpdateChecksum()

	// Convert to JSON
	jsonData, err := doc.ToJSON(false)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Compress
	compressed, err := compressJSON(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	// Upload to Drive
	_, err = ds.client.Files.Update(ds.fileID, &drive.File{}).
		Context(ctx).
		Media(bytes.NewReader(compressed)).
		Do()

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	ds.lastSync = time.Now()
	ds.isDirty = false

	return nil
}

// markDirty marks the storage as needing sync
func (ds *DriveStorage) markDirty() {
	ds.mu.Lock()
	ds.isDirty = true
	ds.mu.Unlock()
}

// needsSync checks if sync is needed
func (ds *DriveStorage) needsSync() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.isDirty
}

// Initialize prepares the storage for use
func (ds *DriveStorage) Initialize(ctx context.Context, userID string) error {
	return ds.cache.Initialize(ctx, userID)
}

// Close cleans up resources
func (ds *DriveStorage) Close() error {
	// Final sync if needed
	if ds.needsSync() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = ds.Sync(ctx)
	}
	return ds.cache.Close()
}

// GetProfile retrieves the user profile
func (ds *DriveStorage) GetProfile(ctx context.Context) (*storage.Profile, error) {
	return ds.cache.GetProfile(ctx)
}

// SaveProfile saves the user profile
func (ds *DriveStorage) SaveProfile(ctx context.Context, profile *storage.Profile) error {
	if err := ds.cache.SaveProfile(ctx, profile); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// ListCompanies returns all companies
func (ds *DriveStorage) ListCompanies(ctx context.Context) ([]*storage.Company, error) {
	return ds.cache.ListCompanies(ctx)
}

// GetCompany retrieves a specific company
func (ds *DriveStorage) GetCompany(ctx context.Context, id int) (*storage.Company, error) {
	return ds.cache.GetCompany(ctx, id)
}

// SaveCompany saves a company
func (ds *DriveStorage) SaveCompany(ctx context.Context, company *storage.Company) error {
	if err := ds.cache.SaveCompany(ctx, company); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// DeleteCompany removes a company and its jobs
func (ds *DriveStorage) DeleteCompany(ctx context.Context, id int) error {
	if err := ds.cache.DeleteCompany(ctx, id); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// ListJobs returns jobs for a company
func (ds *DriveStorage) ListJobs(ctx context.Context, companyID int) ([]*storage.Job, error) {
	return ds.cache.ListJobs(ctx, companyID)
}

// GetJob retrieves a specific job
func (ds *DriveStorage) GetJob(ctx context.Context, id int) (*storage.Job, error) {
	return ds.cache.GetJob(ctx, id)
}

// SaveJob saves a job
func (ds *DriveStorage) SaveJob(ctx context.Context, job *storage.Job) error {
	if err := ds.cache.SaveJob(ctx, job); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// DeleteJob removes a job
func (ds *DriveStorage) DeleteJob(ctx context.Context, id int) error {
	if err := ds.cache.DeleteJob(ctx, id); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// SaveMatchResult saves a match result
func (ds *DriveStorage) SaveMatchResult(ctx context.Context, result *storage.MatchResult) error {
	if err := ds.cache.SaveMatchResult(ctx, result); err != nil {
		return err
	}
	ds.markDirty()
	return nil
}

// GetMatchHistory returns all match results
func (ds *DriveStorage) GetMatchHistory(ctx context.Context, limit int) ([]*storage.MatchResult, error) {
	return ds.cache.GetMatchHistory(ctx, limit)
}

// GetMatchResult retrieves a specific match result
func (ds *DriveStorage) GetMatchResult(ctx context.Context, id int) (*storage.MatchResult, error) {
	return ds.cache.GetMatchResult(ctx, id)
}

// Sync synchronizes data with Google Drive
func (ds *DriveStorage) Sync(ctx context.Context) error {
	ds.syncMu.Lock()
	defer ds.syncMu.Unlock()

	if !ds.needsSync() {
		return nil
	}

	return ds.uploadFromCache(ctx)
}

// GetLastSyncTime returns the last sync timestamp
func (ds *DriveStorage) GetLastSyncTime() time.Time {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.lastSync
}
