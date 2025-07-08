package vega

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benidevo/vega/internal/auth"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/db"
	"github.com/benidevo/vega/internal/storage"
	"github.com/benidevo/vega/internal/storage/badger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// App represents the core application structure, encapsulating configuration,
// HTTP router, database connection, HTTP server, and a channel for handling OS signals.
type App struct {
	config         config.Settings
	router         *gin.Engine
	db             *sql.DB
	storageFactory *storage.Factory
	server         *http.Server
	done           chan os.Signal
}

// New creates and returns a new instance of App with the provided configuration.
// It initializes the router using the Gin framework and sets up a channel to handle OS signals.
func New(cfg config.Settings) *App {
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	corsConfig.AllowCredentials = cfg.CORSAllowCredentials
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	router.Use(cors.New(corsConfig))

	// Only load templates in non-test environment
	if !cfg.IsTest {
		// Setup template functions
		router.SetFuncMap(templateFuncMap())
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

	if a.storageFactory != nil {
		storageErr := a.storageFactory.Close()
		if err == nil {
			err = storageErr
		}
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

	database.SetMaxOpenConns(a.config.DBMaxOpenConns)
	database.SetMaxIdleConns(a.config.DBMaxIdleConns)
	database.SetConnMaxLifetime(a.config.DBConnMaxLifetime)

	if err := database.Ping(); err != nil {
		return err
	}
	a.db = database

	if !a.config.IsTest {
		if err := a.runMigrations(); err != nil {
			return err
		}

		if err := auth.CreateAdminUserIfRequired(a.db, &a.config); err != nil {
			return err
		}
	}

	storageFactory, err := storage.NewFactory(&a.config, a.db)
	if err != nil {
		return err
	}
	a.storageFactory = storageFactory

	// Initialize Badger provider if multi-tenancy is enabled
	if a.config.MultiTenancyEnabled {
		cacheDir := "/app/data/cache"
		if a.config.IsDevelopment {
			cacheDir = "./data/cache"
		}

		provider := badger.NewProvider(cacheDir)
		a.storageFactory.SetProvider(provider)

		log.Info().Str("cache_dir", cacheDir).Msg("Initialized Badger storage provider")
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

// templateFuncMap returns a map of custom template functions
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"pageRange": func(current, total int) []int {
			// Show at most 5 page numbers around current page
			start := current - 2
			end := current + 2

			if start < 1 {
				start = 1
				end = start + 4
			}
			if end > total {
				end = total
				start = end - 4
			}
			if start < 1 {
				start = 1
			}

			pages := make([]int, 0)
			for i := start; i <= end; i++ {
				pages = append(pages, i)
			}
			return pages
		},
		"matchColors": func(score *int) string {
			if score == nil {
				return "bg-gray-500 bg-opacity-20 text-gray-400"
			}
			s := *score
			switch {
			case s >= 85:
				return "bg-green-600 bg-opacity-20 text-primary"
			case s >= 70:
				return "bg-green-500 bg-opacity-20 text-green-400"
			case s >= 55:
				return "bg-yellow-500 bg-opacity-20 text-yellow-400"
			case s >= 40:
				return "bg-orange-500 bg-opacity-20 text-orange-400"
			case s >= 25:
				return "bg-red-500 bg-opacity-20 text-red-400"
			default:
				return "bg-red-600 bg-opacity-20 text-red-400"
			}
		},
	}
}
