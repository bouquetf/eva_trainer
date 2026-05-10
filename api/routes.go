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

type coreDeps struct {
	userRepo     models.UserRepository
	jwtService   *services.JWTService
	permRepo     models.PermissionRepository
	rolePermRepo models.RolePermissionRepository
}

func newCoreDeps(cfg *config.AppConfig) coreDeps {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
	permRepo := models.NewPermissionRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	return coreDeps{userRepo: userRepo, jwtService: jwtService, permRepo: permRepo, rolePermRepo: rolePermRepo}
}

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

	deps := newCoreDeps(cfg)
	splitPublicRepo := models.NewSplitRepository(database.DB())
	splitPublicHandler := handlers.NewSplitHandler(splitPublicRepo)

	splitsPublic := apiV1.Group("/splits")
	splitsPublic.Use(middleware.RequireAuth(deps.userRepo, deps.jwtService))
	splitsPublic.Use(middleware.RequireGlobalPermission(deps.permRepo, deps.rolePermRepo, "split:view"))
	{
		splitsPublic.GET("", splitPublicHandler.ListSplitsPublic)
		splitsPublic.GET("/:id/maps", splitPublicHandler.GetSplitMapsPublic)
	}

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

	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	mapHandler := handlers.NewMapHandler(mapRepo)
	zoneHandler := handlers.NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo)
	lockHandler := handlers.NewEditLockHandler(lockRepo, userRepo)

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

		teams.GET("/:id/maps", middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"), mapHandler.ListTeamMaps)
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

		teams.GET("/:id/maps/:mapId/zones",
			middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:view"),
			zoneHandler.ListZones,
		)
		teams.POST("/:id/maps/:mapId/zones",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:create"),
			zoneHandler.CreateZone,
		)
		teams.POST("/:id/maps/:mapId/zones/batch",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.BatchActions,
		)

		teams.POST("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:lock"),
			lockHandler.AcquireLock,
		)
		teams.POST("/:id/maps/:mapId/lock/release",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.ReleaseLockForEditor,
		)
		teams.DELETE("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.ReleaseLock,
		)
		teams.PUT("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.RefreshLock,
		)
		teams.GET("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			lockHandler.GetLockStatus,
		)

		teams.GET("/:id/zones/:zoneId",
			middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:view"),
			zoneHandler.GetZone,
		)
		teams.DELETE("/:id/zones/:zoneId",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:delete"),
			zoneHandler.DeleteZone,
		)
		teams.PUT("/:id/zones/:zoneId/name",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.SetZoneName,
		)
		teams.POST("/:id/zones/:zoneId/polygons",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.AddPolygon,
		)
		teams.DELETE("/:id/zones/:zoneId/polygons/:polygonId",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.RemovePolygon,
		)
		teams.PUT("/:id/zones/:zoneId/polygons/:polygonId/points",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.UpdatePolygonPoints,
		)
		teams.POST("/:id/zones/:zoneId/polygons/:polygonId/points",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.AddPolygonPoint,
		)
		teams.DELETE("/:id/zones/:zoneId/polygons/:polygonId/points/:pointIndex",
			middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.RemovePolygonPoint,
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

	lockRepoAdmin := models.NewEditLockRepository(db)
	lockHandlerAdmin := handlers.NewEditLockHandler(lockRepoAdmin, userRepo)

	admin := rg.Group("/admin")
	admin.Use(middleware.RequireAuth(userRepo, jwtService))
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetUserDetail)
		admin.GET("/permissions", adminHandler.ListPermissions)
		admin.GET("/permissions/:role", adminHandler.GetRolePermissions)
		admin.GET("/teams", adminHandler.ListAllTeams)
		admin.POST("/users/:id/promote", adminHandler.PromoteToAdmin)
		admin.PUT("/users/:id/teams/:teamId/role", adminHandler.UpdateUserRole)
		admin.POST("/users/:id/teams", adminHandler.AddUserToTeam)
		admin.DELETE("/users/:id/teams/:teamId", adminHandler.RemoveUserFromTeam)

		admin.POST("/maps", mapHandler.UploadMap)
		admin.GET("/maps", mapHandler.ListMaps)
		admin.GET("/maps/:id", mapHandler.GetMapDetail)
		admin.POST("/maps/purge", mapHandler.PurgeObsoleteMaps)

		admin.POST("/locks/cleanup", lockHandlerAdmin.CleanupExpiredLocks)
		admin.DELETE("/locks", lockHandlerAdmin.ForceClearLocks)

		splitRepo := models.NewSplitRepository(db)
		splitHandler := handlers.NewSplitHandler(splitRepo)

		admin.POST("/splits", splitHandler.CreateSplit)
		admin.GET("/splits", splitHandler.ListSplits)
		admin.GET("/splits/:id", splitHandler.GetSplit)
		admin.PUT("/splits/:id", splitHandler.UpdateSplit)
		admin.DELETE("/splits/:id", splitHandler.DeleteSplit)
		admin.GET("/splits/:id/maps", splitHandler.GetSplitMaps)

		mapCatalogRepo := models.NewMapCatalogRepository(db)
		mapCatalogHandler := handlers.NewMapCatalogHandler(mapCatalogRepo)

		admin.GET("/map-catalog", mapCatalogHandler.List)
		admin.POST("/map-catalog", mapCatalogHandler.Create)
		admin.PUT("/map-catalog/:id", mapCatalogHandler.Update)
		admin.DELETE("/map-catalog/:id", mapCatalogHandler.Delete)
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
