package prospector

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

// App represents the core application structure, containing configuration,
// HTTP router, and database connection.
type App struct {
	config Config
	router *gin.Engine
	db     *sql.DB
}

func New(config Config) *App {
	return &App{
		config: config,
		router: gin.Default(),
	}
}

// Run initializes the application by setting up dependencies and routes,
// and starts the HTTP server on the configured port.
func (a *App) Run() {
	a.setupDependencies()
	a.setupRoutes()
	if err := a.router.Run(a.config.ServerPort); err != nil {
		log.Fatal(err)
	}
}

func (a *App) setupRoutes() {
	a.router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	authGroup := a.router.Group("/auth")
	{
		authGroup.GET("/login", func(c *gin.Context) {
			c.String(200, "Login")
		})
	}
}

func (a *App) setupDependencies() {
	db, err := sql.Open(a.config.DBDriver, a.config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to establish a database connection:", err)
	}
	a.db = db
}

// Shutdown gracefully shuts down the application by closing the database
// connection if it is open.
func (a *App) Shutdown() {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Fatal("Failed to close database:", err)
		}
	}
}
