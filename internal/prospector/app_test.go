package prospector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TEST_HOST = "http://localhost"

func TestNew(t *testing.T) {
	config := mockConfig(t)

	app := New(config)
	assert.NotNil(t, app, "Expected app to be initialized")
	assert.Equal(t, app.config, config, "Expected app config to match the provided config")
	assert.NotNil(t, app.router, "Expected app router to be initialized")
	assert.Nil(t, app.db, "Expected app db not to be initialized")
	assert.Nil(t, app.server, "Expected app server not to be initialized")
	assert.NotNil(t, app.done, "Expected app done channel to be initialized")
}

func TestSetup(t *testing.T) {
	config := mockConfig(t)
	app := New(config)

	err := app.Setup()
	require.NoError(t, err, "Expected Setup to succeed")

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	app.router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Expected status code to be 200 OK")

	assert.NotNil(t, app.db, "Expected app db to be initialized")

	ctx := context.Background()
	app.Shutdown(ctx)
}

func TestRun(t *testing.T) {
	config := mockConfig(t)
	app := New(config)

	err := app.Run()
	require.NoError(t, err, "Expected Run to succeed")
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, app.server, "Expected app server to be initialized")

	response, err := http.Get(TEST_HOST + config.ServerPort)
	if err == nil {
		defer response.Body.Close()
		assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status code to be 200 OK")
	}

	app.done <- os.Interrupt
	time.Sleep(100 * time.Millisecond)
}

func TestShutdown(t *testing.T) {
	config := mockConfig(t)
	app := New(config)

	err := app.Setup()
	require.NoError(t, err, "Expected Setup to succeed")

	app.server = &http.Server{
		Addr:    config.ServerPort,
		Handler: app.router,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Shutdown(ctx)
	require.NoError(t, err, "Expected Shutdown to succeed")
}

func TestWaitForShutdown(t *testing.T) {
	config := mockConfig(t)
	app := New(config)
	app.Setup()
	app.server = &http.Server{
		Addr:    config.ServerPort,
		Handler: app.router,
	}

	done := make(chan struct{})

	go func() {
		app.WaitForShutdown()
		done <- struct{}{}

	}()

	app.done <- syscall.SIGTERM

	select {
	case <-done:
		assert.Nil(t, app.server, "Expected app server to be nil after shutdown")
		assert.Nil(t, app.db, "Expected app db to be nil after shutdown")
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForShutdown did not complete in time")
	}
}

func mockConfig(t *testing.T) config.Settings {
	t.Helper()
	return config.Settings{
		ServerPort:         ":0",
		DBConnectionString: ":memory:",
		DBDriver:           "sqlite",
	}
}
