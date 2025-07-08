package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// mockDriveService is a mock implementation of Google Drive for testing
type mockDriveService struct {
	files map[string]*mockFile
	mu    sync.RWMutex
}

type mockFile struct {
	id      string
	name    string
	content []byte
	parents []string
}

func newMockDriveService() *mockDriveService {
	return &mockDriveService{
		files: make(map[string]*mockFile),
	}
}

// mockFilesService implements a subset of drive.FilesService
type mockFilesService struct {
	service *mockDriveService
}

func (m *mockDriveService) Files() *mockFilesService {
	return &mockFilesService{service: m}
}

// List returns a mock file list call
func (f *mockFilesService) List() *mockListCall {
	return &mockListCall{service: f.service}
}

// Get returns a mock file get call
func (f *mockFilesService) Get(fileId string) *mockGetCall {
	return &mockGetCall{
		service: f.service,
		fileId:  fileId,
	}
}

// Create returns a mock file create call
func (f *mockFilesService) Create(file *drive.File) *mockCreateCall {
	return &mockCreateCall{
		service: f.service,
		file:    file,
	}
}

// Update returns a mock file update call
func (f *mockFilesService) Update(fileId string, file *drive.File) *mockUpdateCall {
	return &mockUpdateCall{
		service: f.service,
		fileId:  fileId,
		file:    file,
	}
}

// Mock call implementations
type mockListCall struct {
	service *mockDriveService
	query   string
	spaces  string
	fields  string
}

func (c *mockListCall) Context(ctx context.Context) *mockListCall { return c }
func (c *mockListCall) Q(q string) *mockListCall                  { c.query = q; return c }
func (c *mockListCall) Spaces(spaces string) *mockListCall        { c.spaces = spaces; return c }
func (c *mockListCall) Fields(fields googleapi.Field) *mockListCall {
	c.fields = string(fields)
	return c
}

func (c *mockListCall) Do(opts ...googleapi.CallOption) (*drive.FileList, error) {
	c.service.mu.RLock()
	defer c.service.mu.RUnlock()

	var files []*drive.File
	for _, f := range c.service.files {
		if f.name == fileName && contains(f.parents, appDataFolder) {
			files = append(files, &drive.File{
				Id:   f.id,
				Name: f.name,
			})
		}
	}

	return &drive.FileList{Files: files}, nil
}

type mockGetCall struct {
	service *mockDriveService
	fileId  string
}

func (c *mockGetCall) Context(ctx context.Context) *mockGetCall { return c }

func (c *mockGetCall) Download(opts ...googleapi.CallOption) (*http.Response, error) {
	c.service.mu.RLock()
	defer c.service.mu.RUnlock()

	file, exists := c.service.files[c.fileId]
	if !exists {
		return nil, fmt.Errorf("file not found")
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(file.content)),
	}, nil
}

type mockCreateCall struct {
	service *mockDriveService
	file    *drive.File
	content []byte
}

func (c *mockCreateCall) Context(ctx context.Context) *mockCreateCall { return c }

func (c *mockCreateCall) Media(r io.Reader, opts ...googleapi.MediaOption) *mockCreateCall {
	content, _ := io.ReadAll(r)
	c.content = content
	return c
}

func (c *mockCreateCall) Do(opts ...googleapi.CallOption) (*drive.File, error) {
	c.service.mu.Lock()
	defer c.service.mu.Unlock()

	id := fmt.Sprintf("file_%d", len(c.service.files)+1)

	mockFile := &mockFile{
		id:      id,
		name:    c.file.Name,
		content: c.content,
		parents: c.file.Parents,
	}

	c.service.files[id] = mockFile

	return &drive.File{
		Id:   id,
		Name: c.file.Name,
	}, nil
}

type mockUpdateCall struct {
	service *mockDriveService
	fileId  string
	file    *drive.File
	content []byte
}

func (c *mockUpdateCall) Context(ctx context.Context) *mockUpdateCall { return c }

func (c *mockUpdateCall) Media(r io.Reader, opts ...googleapi.MediaOption) *mockUpdateCall {
	content, _ := io.ReadAll(r)
	c.content = content
	return c
}

func (c *mockUpdateCall) Do(opts ...googleapi.CallOption) (*drive.File, error) {
	c.service.mu.Lock()
	defer c.service.mu.Unlock()

	file, exists := c.service.files[c.fileId]
	if !exists {
		return nil, fmt.Errorf("file not found")
	}

	file.content = c.content

	return &drive.File{
		Id:   file.id,
		Name: file.name,
	}, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
