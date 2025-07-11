package googledrive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/benidevo/vega/internal/common/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	appDataFolder = "appDataFolder"
	vegaDBFile    = "vega.db"
)

// Provider handles Google Drive operations for SQLite storage
type Provider struct {
	service *drive.Service
	logger  logger.PrivacyLogger
}

// NewProvider creates a new Google Drive provider with OAuth token
func NewProvider(ctx context.Context, oauthToken string, log logger.PrivacyLogger) (*Provider, error) {
	// Create an HTTP client with the OAuth token
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: oauthToken,
			}),
		},
	}

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	return &Provider{
		service: service,
		logger:  log,
	}, nil
}

// Download downloads the SQLite database from Google Drive to a temp file
func (p *Provider) Download(ctx context.Context, userID string) (string, error) {
	// Search for the vega.db file in appDataFolder
	query := fmt.Sprintf("name='%s' and parents in '%s' and trashed=false", vegaDBFile, appDataFolder)
	files, err := p.service.Files.List().
		Context(ctx).
		Q(query).
		Spaces(appDataFolder).
		Fields("files(id, name, modifiedTime)").
		Do()

	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	if len(files.Files) == 0 {
		// No file found, user is new
		p.logger.Info().
			Str("userID", userID).
			Msg("No existing database found for user")
		return "", nil
	}

	// Download the file
	file := files.Files[0]
	resp, err := p.service.Files.Get(file.Id).Context(ctx).Download()
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Create temp file
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("vega-%s-*.db", userID))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy content
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	p.logger.Info().
		Str("userID", userID).
		Str("fileID", file.Id).
		Str("tempFile", tempFile.Name()).
		Msg("Downloaded database from Google Drive")

	return tempFile.Name(), nil
}

// Upload uploads the SQLite database from a temp file to Google Drive
func (p *Provider) Upload(ctx context.Context, userID string, tempFilePath string) error {
	// Open the temp file
	file, err := os.Open(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer file.Close()

	// Check if file already exists
	query := fmt.Sprintf("name='%s' and parents in '%s' and trashed=false", vegaDBFile, appDataFolder)
	existingFiles, err := p.service.Files.List().
		Context(ctx).
		Q(query).
		Spaces(appDataFolder).
		Fields("files(id)").
		Do()

	if err != nil {
		return fmt.Errorf("failed to check existing files: %w", err)
	}

	if len(existingFiles.Files) > 0 {
		// Update existing file
		fileID := existingFiles.Files[0].Id
		_, err = p.service.Files.Update(fileID, &drive.File{
			ModifiedTime: time.Now().Format(time.RFC3339),
		}).Context(ctx).Media(file).Do()

		if err != nil {
			return fmt.Errorf("failed to update file: %w", err)
		}

		p.logger.Info().
			Str("userID", userID).
			Str("fileID", fileID).
			Msg("Updated database in Google Drive")
	} else {
		// Create new file
		driveFile := &drive.File{
			Name:    vegaDBFile,
			Parents: []string{appDataFolder},
		}

		createdFile, err := p.service.Files.Create(driveFile).
			Context(ctx).
			Media(file).
			Do()

		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		p.logger.Info().
			Str("userID", userID).
			Str("fileID", createdFile.Id).
			Msg("Created new database in Google Drive")
	}

	return nil
}

// Delete removes the SQLite database from Google Drive
func (p *Provider) Delete(ctx context.Context, userID string) error {
	// Search for the vega.db file
	query := fmt.Sprintf("name='%s' and parents in '%s' and trashed=false", vegaDBFile, appDataFolder)
	files, err := p.service.Files.List().
		Context(ctx).
		Q(query).
		Spaces(appDataFolder).
		Fields("files(id)").
		Do()

	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Delete all matching files
	for _, file := range files.Files {
		err = p.service.Files.Delete(file.Id).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to delete file %s: %w", file.Id, err)
		}
	}

	p.logger.Info().
		Str("userID", userID).
		Int("filesDeleted", len(files.Files)).
		Msg("Deleted database from Google Drive")

	return nil
}

// GetLastModified returns the last modified time of the database in Google Drive
func (p *Provider) GetLastModified(ctx context.Context, userID string) (time.Time, error) {
	// Search for the vega.db file
	query := fmt.Sprintf("name='%s' and parents in '%s' and trashed=false", vegaDBFile, appDataFolder)
	files, err := p.service.Files.List().
		Context(ctx).
		Q(query).
		Spaces(appDataFolder).
		Fields("files(modifiedTime)").
		Do()

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files.Files) == 0 {
		return time.Time{}, nil
	}

	modTime, err := time.Parse(time.RFC3339, files.Files[0].ModifiedTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse modified time: %w", err)
	}

	return modTime, nil
}

// CleanupTempFiles removes any temporary files created during sync
func CleanupTempFiles(userID string) {
	tempDir := os.TempDir()
	pattern := filepath.Join(tempDir, fmt.Sprintf("vega-%s-*.db", userID))

	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, file := range files {
		os.Remove(file)
	}
}
