package api

import (
	"eva_trainer/internal/handlers"
	"eva_trainer/pkg/config"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func serveFrontendFile(c *gin.Context, name string, fs http.FileSystem) {
	f, err := fs.Open(name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), f)
}

func SetupRoutes(r *gin.Engine, corsCfg *config.CORSConfig, frontendFS http.FileSystem, assetsFS http.FileSystem) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsCfg.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.GET("/health", handlers.HealthCheck)

	apiV1 := r.Group("/api/v1")
	_ = apiV1

	if gin.Mode() == gin.ReleaseMode {
		r.StaticFS("/assets", assetsFS)
		r.GET("/", func(c *gin.Context) {
			serveFrontendFile(c, "index.html", frontendFS)
		})
		r.NoRoute(func(c *gin.Context) {
			serveFrontendFile(c, "index.html", frontendFS)
		})
	}
}
