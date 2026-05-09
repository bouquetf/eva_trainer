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
	setupTeamRoutes(apiV1, cfg)
	setupAdminRoutes(apiV1, cfg)

	apiV1.GET("/maps/:id/image", handlers.NewMapHandler(models.NewMapRepository(database.DB())).ServeImage)

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

func setupTeamRoutes(rg *gin.RouterGroup, cfg *config.AppConfig) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	permRepo := models.NewPermissionRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)

	teamHandler := handlers.NewTeamHandler(teamRepo, teamMemberRepo)
	teamRolesHandler := handlers.NewTeamRolesHandler(teamRepo, teamMemberRepo, userRepo, rolePermRepo, permRepo)

	invitationRepo := models.NewTeamInvitationRepository(db, teamMemberRepo)
	joinRequestRepo := models.NewTeamJoinRequestRepository(db, teamMemberRepo)
	teamInvitationHandler := handlers.NewTeamInvitationHandler(invitationRepo, teamMemberRepo, userRepo, teamRepo, cfg.Server.FrontendURL)
	joinRequestHandler := handlers.NewTeamJoinRequestHandler(joinRequestRepo, teamMemberRepo, teamRepo, userRepo)

	rg.GET("/myteams", middleware.RequireAuth(userRepo, jwtService), teamHandler.ListMyTeams)

	teams := rg.Group("/teams")
	teams.Use(middleware.RequireAuth(userRepo, jwtService))
	{
		teams.POST("", teamHandler.Create)
		teams.GET("", teamHandler.ListAllTeams)
		teams.GET("/:id", middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "member:view"), teamHandler.GetTeam)
		teams.PUT("/:id", middleware.RequireTeamRole(teamMemberRepo, "captain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:edit"), teamHandler.UpdateTeam)
		teams.DELETE("/:id", middleware.RequireTeamRole(teamMemberRepo, "captain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:delete"), teamHandler.DeleteTeam)
		teams.POST("/:id/leave", middleware.RequireTeamRole(teamMemberRepo, "member"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:leave"), teamHandler.LeaveTeam)

		teams.GET("/:id/members", middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "member:view"), teamRolesHandler.ListMembers)
		teams.PUT("/:id/members/:userId/role", middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "member:role:assign"), teamRolesHandler.UpdateRole)
		teams.POST("/:id/transfer-captaincy", middleware.RequireTeamRole(teamMemberRepo, "captain"), middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "member:role:transfer-captain"), teamRolesHandler.TransferCaptaincy)

		teams.POST("/:id/invitations",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
			teamInvitationHandler.InviteByDisplayName,
		)
		teams.POST("/:id/invitations/link",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
			teamInvitationHandler.GenerateInviteLink,
		)
		teams.GET("/:id/invitations/pending",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
			teamInvitationHandler.ListPendingByTeam,
		)

		teams.GET("/:id/join-requests/mine",
			middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.GetMyRequest,
		)
		teams.DELETE("/:id/join-requests/mine",
			middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.CancelMyRequest,
		)
		teams.POST("/:id/join-requests",
			middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.RequestJoin,
		)
		teams.GET("/:id/join-requests",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
			joinRequestHandler.ListPendingRequests,
		)
		teams.PUT("/:id/join-requests/:requestId/accept",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
			joinRequestHandler.AcceptRequest,
		)
		teams.PUT("/:id/join-requests/:requestId/decline",
			middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
			joinRequestHandler.DeclineRequest,
		)
	}

	invitations := rg.Group("/invitations")
	invitations.Use(middleware.RequireAuth(userRepo, jwtService))
	{
		invitations.GET("", teamInvitationHandler.ListMyInvitations)
		invitations.POST("/:invitationId/accept", teamInvitationHandler.AcceptInvitation)
		invitations.POST("/:invitationId/decline", teamInvitationHandler.DeclineInvitation)
	}

	inviteLinks := rg.Group("/invite-links")
	{
		inviteLinks.GET("/:token", teamInvitationHandler.GetInviteLinkInfo)
		inviteLinks.POST("/:token/accept",
			middleware.RequireAuth(userRepo, jwtService),
			teamInvitationHandler.AcceptInviteLink,
		)
	}

	users := rg.Group("/users")
	users.Use(middleware.RequireAuth(userRepo, jwtService))
	{
		users.GET("/search", teamInvitationHandler.SearchUsers)
	}
}

func setupAdminRoutes(rg *gin.RouterGroup, cfg *config.AppConfig) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	permRepo := models.NewPermissionRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
	mapRepo := models.NewMapRepository(db)
	mapHandler := handlers.NewMapHandler(mapRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, permRepo, rolePermRepo, teamRepo, teamMemberRepo)

	admin := rg.Group("/admin")
	admin.Use(middleware.RequireAuth(userRepo, jwtService))
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetUserDetail)
		admin.GET("/permissions", adminHandler.ListPermissions)
		admin.GET("/permissions/:role", adminHandler.GetRolePermissions)
		admin.GET("/teams", adminHandler.ListAllTeams)
		admin.PUT("/users/:id/admin", adminHandler.ToggleAdmin)
		admin.PUT("/users/:id/teams/:teamId/role", adminHandler.UpdateUserRole)
		admin.POST("/users/:id/teams", adminHandler.AddUserToTeam)
		admin.DELETE("/users/:id/teams/:teamId", adminHandler.RemoveUserFromTeam)

		admin.POST("/maps", mapHandler.UploadMap)
		admin.GET("/maps", mapHandler.ListMaps)
		admin.GET("/maps/:id", mapHandler.GetMapDetail)
		admin.POST("/maps/purge", mapHandler.PurgeObsoleteMaps)
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
		protected.PUT("/auth/display-name", authHandler.UpdateDisplayName)
	}
}
