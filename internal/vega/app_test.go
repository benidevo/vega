package vega

import (
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TEST_HOST = "http://localhost"

func mockConfig(t *testing.T) config.Settings {
	t.Helper()
	return config.NewTestSettings()
}

func TestAppLifecycle(t *testing.T) {
	t.Run("should_initialize_app_with_config", func(t *testing.T) {
		config := mockConfig(t)
		app := New(config)

		assert.NotNil(t, app, "Expected app to be initialized")
		assert.Equal(t, app.config, config, "Expected app config to match")
		assert.NotNil(t, app.router, "Expected router to be initialized")
		assert.Nil(t, app.db, "Expected db not to be initialized yet")
		assert.Nil(t, app.server, "Expected server not to be initialized")
		assert.NotNil(t, app.done, "Expected done channel to be initialized")
	})

	t.Run("should_start_server_when_run_is_called", func(t *testing.T) {
		config := mockConfig(t)
		app := New(config)

		err := app.Run()
		require.NoError(t, err, "Expected Run to succeed")
		time.Sleep(100 * time.Millisecond)
		assert.NotNil(t, app.server, "Expected server to be initialized")

		response, err := http.Get(TEST_HOST + config.ServerPort)
		if err == nil {
			defer response.Body.Close()
			assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status code 200 OK")
		}

		app.done <- os.Interrupt
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("should_shutdown_app_when_signal_received", func(t *testing.T) {
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
			assert.Nil(t, app.server, "Expected server to be nil after shutdown")
			assert.Nil(t, app.db, "Expected db to be nil after shutdown")
		case <-time.After(2 * time.Second):
			t.Fatal("WaitForShutdown did not complete in time")
		}
	})
}
