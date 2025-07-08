package middleware

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         *config.Settings
		setupContext   func(*gin.Context)
		expectStorage  bool
		expectContinue bool
		expectError    bool
	}{
		{
			name: "skip when not in cloud mode",
			config: &config.Settings{
				IsCloudMode:         false,
				MultiTenancyEnabled: false,
			},
			setupContext:   func(c *gin.Context) {},
			expectStorage:  false,
			expectContinue: true,
			expectError:    false,
		},
		{
			name: "skip when multi-tenancy disabled",
			config: &config.Settings{
				IsCloudMode:         true,
				MultiTenancyEnabled: false,
			},
			setupContext:   func(c *gin.Context) {},
			expectStorage:  false,
			expectContinue: true,
			expectError:    false,
		},
		{
			name: "continue without storage for unauthenticated requests",
			config: &config.Settings{
				IsCloudMode:         true,
				MultiTenancyEnabled: true,
			},
			setupContext:   func(c *gin.Context) {},
			expectStorage:  false,
			expectContinue: true,
			expectError:    false,
		},
		{
			name: "continue without storage for invalid username type",
			config: &config.Settings{
				IsCloudMode:         true,
				MultiTenancyEnabled: true,
			},
			setupContext: func(c *gin.Context) {
				c.Set("username", 123) // Invalid type
			},
			expectStorage:  false,
			expectContinue: true,
			expectError:    false,
		},
		{
			name: "continue without storage for empty username",
			config: &config.Settings{
				IsCloudMode:         true,
				MultiTenancyEnabled: true,
			},
			setupContext: func(c *gin.Context) {
				c.Set("username", "")
			},
			expectStorage:  false,
			expectContinue: true,
			expectError:    false,
		},
		{
			name: "set storage for valid authenticated user",
			config: &config.Settings{
				IsCloudMode:         true,
				MultiTenancyEnabled: true,
			},
			setupContext: func(c *gin.Context) {
				c.Set("username", "user@example.com")
			},
			expectStorage:  true,
			expectContinue: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, err := storage.NewFactory(tt.config, nil)
			require.NoError(t, err)

			middleware := TenantIsolation(factory, tt.config)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			tt.setupContext(c)

			router := gin.New()
			handlerCalled := false

			router.Use(middleware)
			router.GET("/test", func(c *gin.Context) {
				handlerCalled = true
			})

			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectContinue, handlerCalled, "Handler called expectation mismatch")

			storageWasSet := false
			if tt.expectStorage {
				router2 := gin.New()
				router2.Use(func(ctx *gin.Context) {
					tt.setupContext(ctx)
					ctx.Next()
				})
				router2.Use(middleware)
				router2.GET("/test", func(ctx *gin.Context) {
					_, storageWasSet = GetUserStorage(ctx)
				})
				req := httptest.NewRequest("GET", "/test", nil)
				w2 := httptest.NewRecorder()
				router2.ServeHTTP(w2, req)
			}
			exists := storageWasSet
			assert.Equal(t, tt.expectStorage, exists, "Storage existence expectation mismatch")

			if tt.expectError {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			} else if tt.expectContinue {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestGetUserStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setupContext  func(*gin.Context)
		expectStorage bool
		expectOK      bool
	}{
		{
			name:          "no storage in context",
			setupContext:  func(c *gin.Context) {},
			expectStorage: false,
			expectOK:      false,
		},
		{
			name: "invalid type in context",
			setupContext: func(c *gin.Context) {
				c.Set(UserStorageKey, "not a storage")
			},
			expectStorage: false,
			expectOK:      false,
		},
		{
			name: "valid storage in context",
			setupContext: func(c *gin.Context) {
				cfg := &config.Settings{}
				factory, _ := storage.NewFactory(cfg, nil)
				userStorage, _ := factory.GetUserStorage(context.Background(), "test@example.com")
				c.Set(UserStorageKey, userStorage)
			},
			expectStorage: true,
			expectOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			storage, ok := GetUserStorage(c)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectStorage {
				assert.NotNil(t, storage)
			} else {
				assert.Nil(t, storage)
			}
		})
	}
}

func TestRequireUserStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		expectAbort  bool
		expectStatus int
	}{
		{
			name:         "abort when no storage",
			setupContext: func(c *gin.Context) {},
			expectAbort:  true,
			expectStatus: http.StatusInternalServerError,
		},
		{
			name: "continue when storage exists",
			setupContext: func(c *gin.Context) {
				cfg := &config.Settings{}
				factory, _ := storage.NewFactory(cfg, nil)
				userStorage, _ := factory.GetUserStorage(context.Background(), "test@example.com")
				c.Set(UserStorageKey, userStorage)
			},
			expectAbort:  false,
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Create a router with the middleware
			router := gin.New()

			// simple HTML template to prevent panic
			router.SetHTMLTemplate(template.Must(template.New("").Parse(`{{define "layouts/base.html"}}Error{{end}}`)))

			handlerCalled := false

			// Apply setup to the router's context
			router.Use(func(ctx *gin.Context) {
				tt.setupContext(ctx)
				ctx.Next()
			})

			router.Use(RequireUserStorage())
			router.GET("/test", func(c *gin.Context) {
				handlerCalled = true
			})

			router.ServeHTTP(w, c.Request)

			assert.Equal(t, !tt.expectAbort, handlerCalled)
			if tt.expectAbort {
				assert.Equal(t, tt.expectStatus, w.Code)
			}
		})
	}
}
