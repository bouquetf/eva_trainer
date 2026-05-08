package api

import (
	"net/http"

	"eva_trainer/internal/handlers"
	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/config"
	"eva_trainer/pkg/database"

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

func SetupRoutes(r *gin.Engine, cfg *config.AppConfig, frontendFS http.FileSystem, assetsFS http.FileSystem) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.GET("/health", handlers.HealthCheck)

	apiV1 := r.Group("/api/v1")

	setupAuthRoutes(apiV1, cfg)

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

func setupAuthRoutes(rg *gin.RouterGroup, cfg *config.AppConfig) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	refreshRepo := models.NewRefreshTokenRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
	oauthCfg := services.DiscordOAuthConfig(cfg.Discord.ClientID, cfg.Discord.ClientSecret, cfg.Discord.RedirectURL)
	authHandler := handlers.NewAuthHandler(userRepo, refreshRepo, oauthCfg, jwtService, cfg.Server.FrontendURL)

	rg.GET("/auth/discord/login", authHandler.LoginRedirect)
	rg.GET("/auth/discord/callback", authHandler.Callback)
	rg.POST("/auth/refresh", authHandler.Refresh)

	protected := rg.Group("")
	protected.Use(middleware.RequireAuth(userRepo, jwtService))
	{
		protected.GET("/auth/me", authHandler.Me)
	}
}
