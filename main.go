package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"eva_trainer/api"
	"eva_trainer/pkg/config"

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

	r := gin.Default()
	api.SetupRoutes(r, &cfg.CORS, http.FS(subFS), http.FS(assetsSubFS))

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
