package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eva_trainer/api"
	"eva_trainer/db/seed"
	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/config"
	"eva_trainer/pkg/database"

	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if cfg.Server.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	subFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to get frontend sub filesystem: %v", err)
	}
	assetsSubFS, err := fs.Sub(subFS, "assets")
	if err != nil {
		assetsSubFS = subFS
	}

	dbCfg := config.NewDatabaseConfigFromApp(cfg)
	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	database.SetDB(db)

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := seed.SeedRolePermissions(db); err != nil {
		log.Fatalf("Failed to seed role permissions: %v", err)
	}

	if err := seed.SeedMaps(db, "maps"); err != nil {
		log.Printf("Warning: failed to seed maps: %v", err)
	}

	logger := services.NewLoggerFromConfig(cfg)

	lockRepo := models.NewEditLockRepository(db)
	lockCleanup := services.NewLockCleanupService(lockRepo, 5*time.Minute, logger)
	lockCleanup.Start()

	r := gin.New()
	r.Use(services.RequestLogger(logger))
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.ErrorHandlerMiddleware(logger))
	api.SetupRoutes(r, cfg, logger, http.FS(subFS), http.FS(assetsSubFS))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := r.Run(":" + cfg.Server.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	lockCleanup.Stop()
}
