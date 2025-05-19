package ascentio

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/db"
	"github.com/benidevo/ascentio/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// App represents the core application structure, encapsulating configuration,
// HTTP router, database connection, HTTP server, and a channel for handling OS signals.
type App struct {
	config config.Settings
	router *gin.Engine
	db     *sql.DB
	server *http.Server
	done   chan os.Signal
}

// New creates and returns a new instance of App with the provided configuration.
// It initializes the router using the Gin framework and sets up a channel to handle OS signals.
func New(cfg config.Settings) *App {
	router := gin.Default()

	// Only load templates in non-test environment
	if !cfg.IsTest {
		router.LoadHTMLGlob("templates/**/*.html")
	}

	return &App{
		config: cfg,
		router: router,
		done:   make(chan os.Signal, 1),
	}
}

// Setup initializes the application by setting up dependencies and routes.
func (a *App) Setup() error {
	logger.Initialize(
		a.config.IsDevelopment,
		a.config.LogLevel,
	)

	log.Info().Msg("Starting application setup")

	if err := a.setupDependencies(); err != nil {
		log.Error().Err(err).Msg("Failed to setup dependencies")
		return err
	}
	SetupRoutes(a)

	log.Info().Msg("Application setup completed successfully")
	return nil
}

// Run initializes the application, sets up the HTTP server, and starts it in a separate goroutine.
// It also listens for system interrupt signals to handle graceful shutdown.
func (a *App) Run() error {
	if err := a.Setup(); err != nil {
		return err
	}

	a.server = &http.Server{
		Addr:    a.config.ServerPort,
		Handler: a.router,
	}

	signal.Notify(a.done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s\n", a.config.ServerPort)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v\n", err)
		}
	}()

	return nil
}

// WaitForShutdown waits for a shutdown signal, gracefully shuts down the server,
// It uses a context with a 10-second timeout
// to ensure the shutdown completes within the specified time.
func (a *App) WaitForShutdown() {
	<-a.done
	log.Info().Msg("Received shutdown signal, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error during shutdown")
	}

	log.Info().Msg("Server shut down gracefully")
}

// Shutdown gracefully shuts down the application by stopping the server
// and closing the database connection.
func (a *App) Shutdown(ctx context.Context) error {
	var err error

	if a.server != nil {
		err = a.server.Shutdown(ctx)
	}

	if a.db != nil {
		dbErr := a.db.Close()
		if err == nil {
			err = dbErr
		}
	}

	a.server = nil
	a.db = nil

	return err
}

func (a *App) setupDependencies() error {
	database, err := sql.Open(a.config.DBDriver, a.config.DBConnectionString)
	if err != nil {
		return err
	}

	if err := database.Ping(); err != nil {
		return err
	}
	a.db = database

	if !a.config.IsTest {
		if err := a.runMigrations(); err != nil {
			return err
		}
	}

	return nil
}

// runMigrations applies database migrations from the configured migrations directory
func (a *App) runMigrations() error {
	dbPath := a.config.DBConnectionString
	for i, char := range dbPath {
		if char == '?' {
			dbPath = dbPath[:i]
			break
		}
	}

	if err := db.MigrateDatabase(dbPath, a.config.MigrationsDir); err != nil {
		log.Error().Err(err).Msg("Database migration failed")
		return err
	}

	return nil
}
