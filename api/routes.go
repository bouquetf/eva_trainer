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
	if fs == nil {
		c.Status(http.StatusNotFound)
		return
	}
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

func SetupRoutes(r *gin.Engine, cfg *config.AppConfig, logger services.Logger, frontendFS http.FileSystem, assetsFS http.FileSystem) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.GET("/health", handlers.HealthCheck)

	apiV1 := r.Group("/api/v1")

	deps := newCoreDeps(cfg)

	templateRepo := models.NewTemplateRepository(database.DB())
	zoneRepo := models.NewZoneRepository(database.DB())
	teamMemberRepo := models.NewTeamMemberRepository(database.DB())
	templateHandler := handlers.NewTemplateHandler(templateRepo, zoneRepo, teamMemberRepo, logger)

	setupAuthRoutes(apiV1, cfg, logger)
	setupTeamRoutes(apiV1, cfg, logger, templateHandler)
	setupAdminRoutes(apiV1, cfg, logger)
	setupArchiveRoutes(apiV1, cfg, logger)

	apiV1.GET("/maps/:id/image", handlers.NewMapHandler(models.NewMapRepository(database.DB()), logger).ServeImage)

	splitPublicRepo := models.NewSplitRepository(database.DB())
	splitPublicHandler := handlers.NewSplitHandler(splitPublicRepo, logger)

	splitsPublic := apiV1.Group("/splits")
	splitsPublic.Use(middleware.RequireAuth(logger, deps.userRepo, deps.jwtService))
	splitsPublic.Use(middleware.RequireGlobalPermission(logger, deps.permRepo, deps.rolePermRepo, "split:view"))
	{
		splitsPublic.GET("", splitPublicHandler.ListSplitsPublic)
		splitsPublic.GET("/:id/maps", splitPublicHandler.GetSplitMapsPublic)
	}

	apiV1.GET("/templates/public",
		middleware.RequireAuth(logger, deps.userRepo, deps.jwtService),
		templateHandler.ListAllTemplates,
	)

	r.StaticFS("/assets", assetsFS)
	r.GET("/", func(c *gin.Context) {
		serveFrontendFile(c, "index.html", frontendFS)
	})
	r.NoRoute(func(c *gin.Context) {
		serveFrontendFile(c, "index.html", frontendFS)
	})
}

func setupTeamRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger, templateHandler *handlers.TemplateHandler) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	permRepo := models.NewPermissionRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)

	teamHandler := handlers.NewTeamHandler(teamRepo, teamMemberRepo, logger)
	teamRolesHandler := handlers.NewTeamRolesHandler(teamRepo, teamMemberRepo, userRepo, rolePermRepo, permRepo, logger)

	invitationRepo := models.NewTeamInvitationRepository(db, teamMemberRepo)
	joinRequestRepo := models.NewTeamJoinRequestRepository(db, teamMemberRepo)
	teamInvitationHandler := handlers.NewTeamInvitationHandler(invitationRepo, teamMemberRepo, userRepo, teamRepo, cfg.Server.FrontendURL, logger)
	joinRequestHandler := handlers.NewTeamJoinRequestHandler(joinRequestRepo, teamMemberRepo, teamRepo, userRepo, logger)

	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	versionRepo := models.NewZoneVersionRepository(db)
	mapHandler := handlers.NewMapHandler(mapRepo, logger)
	templateRepo := models.NewTemplateRepository(db)
	templateSyncSvc := services.NewTemplateSyncService(templateRepo, zoneRepo, logger)
	zoneHandler := handlers.NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, templateSyncSvc, versionRepo, logger)
	lineOfFireRepo := models.NewLineOfFireRepository(db)
	lofVersionRepo := models.NewLineOfFireVersionRepository(db)
	lofHistoryRepo := models.NewLineOfFireEditHistoryRepository(db)
	lineOfFireHandler := handlers.NewLineOfFireHandler(lineOfFireRepo, mapRepo, lockRepo, lofHistoryRepo, teamMemberRepo, userRepo, templateSyncSvc, lofVersionRepo, logger)
	lockHandler := handlers.NewEditLockHandler(lockRepo, userRepo, logger)
	historyHandler := handlers.NewHistoryHandler(versionRepo, zoneRepo, lockRepo, teamMemberRepo, userRepo, historyRepo, logger)
	revisionHandler := handlers.NewRevisionHandler(zoneRepo, mapRepo, lineOfFireRepo, templateRepo, logger)

	rg.GET("/myteams", middleware.RequireAuth(logger, userRepo, jwtService), teamHandler.ListMyTeams)

	teams := rg.Group("/teams")
	teams.Use(middleware.RequireAuth(logger, userRepo, jwtService))
	{
		teams.POST("", teamHandler.Create)
		teams.GET("", teamHandler.ListAllTeams)
		teams.GET("/:id", middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "member:view"), teamHandler.GetTeam)
		teams.PUT("/:id", middleware.RequireTeamRole(logger, teamMemberRepo, "captain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:edit"), teamHandler.UpdateTeam)
		teams.DELETE("/:id", middleware.RequireTeamRole(logger, teamMemberRepo, "captain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:delete"), teamHandler.DeleteTeam)
		teams.POST("/:id/leave", middleware.RequireTeamRole(logger, teamMemberRepo, "member"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:leave"), teamHandler.LeaveTeam)

		teams.GET("/:id/maps", middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"), mapHandler.ListTeamMaps)
		teams.GET("/:id/members", middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "member:view"), teamRolesHandler.ListMembers)
		teams.PUT("/:id/members/:userId/role", middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "member:role:assign"), teamRolesHandler.UpdateRole)
		teams.POST("/:id/transfer-captaincy", middleware.RequireTeamRole(logger, teamMemberRepo, "captain"), middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "member:role:transfer-captain"), teamRolesHandler.TransferCaptaincy)

		teams.POST("/:id/invitations",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
			teamInvitationHandler.InviteByDisplayName,
		)
		teams.POST("/:id/invitations/link",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
			teamInvitationHandler.GenerateInviteLink,
		)

		teams.GET("/:id/maps/:mapId/zones",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:view"),
			zoneHandler.ListZones,
		)
		teams.POST("/:id/maps/:mapId/zones",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:create"),
			zoneHandler.CreateZone,
		)
		teams.POST("/:id/maps/:mapId/zones/batch",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.BatchActions,
		)

		teams.GET("/:id/maps/:mapId/lines-of-fire",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
			lineOfFireHandler.ListLinesOfFire,
		)
		teams.POST("/:id/maps/:mapId/lines-of-fire",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:create"),
			lineOfFireHandler.CreateLineOfFire,
		)
		teams.POST("/:id/maps/:mapId/lines-of-fire/batch",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
			lineOfFireHandler.BatchActions,
		)
		teams.POST("/:id/maps/:mapId/lines-of-fire/import",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:create"),
			lineOfFireHandler.ImportLinesOfFire,
		)
		teams.GET("/:id/maps/:mapId/lines-of-fire/export",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
			lineOfFireHandler.ExportLinesOfFire,
		)

		teams.POST("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:lock"),
			lockHandler.AcquireLock,
		)
		teams.POST("/:id/maps/:mapId/lock/release",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.ReleaseLockForEditor,
		)
		teams.DELETE("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.ReleaseLock,
		)
		teams.PUT("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			lockHandler.RefreshLock,
		)
		teams.GET("/:id/maps/:mapId/lock",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			lockHandler.GetLockStatus,
		)

		teams.POST("/:id/maps/:mapId/publish",
			middleware.RequireTeamRole(logger, teamMemberRepo, "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:publish"),
			templateHandler.PublishMap,
		)
		teams.POST("/:id/maps/:mapId/unpublish",
			middleware.RequireTeamRole(logger, teamMemberRepo, "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:publish"),
			templateHandler.UnpublishMap,
		)
		teams.GET("/:id/maps/:mapId/publish-status",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			templateHandler.GetPublishStatus,
		)
		teams.GET("/:id/maps/:mapId/revision-zones",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			revisionHandler.ListRevisionZones,
		)
		teams.POST("/:id/maps/:mapId/revision-items",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			revisionHandler.ListRevisionItems,
		)

		teams.POST("/:id/templates/:templateId/duplicate",
			middleware.RequireTeamRole(logger, teamMemberRepo, "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "template:duplicate"),
			templateHandler.DuplicateTemplate,
		)

		teams.GET("/:id/zones/:zoneId",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:view"),
			zoneHandler.GetZone,
		)
		teams.DELETE("/:id/zones/:zoneId",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:delete"),
			zoneHandler.DeleteZone,
		)
		teams.GET("/:id/lines-of-fire/:lineOfFireId",
			middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
			lineOfFireHandler.GetLineOfFire,
		)
		teams.DELETE("/:id/lines-of-fire/:lineOfFireId",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:delete"),
			lineOfFireHandler.DeleteLineOfFire,
		)
		teams.PUT("/:id/lines-of-fire/:lineOfFireId/name",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
			lineOfFireHandler.SetLineOfFireName,
		)
		teams.POST("/:id/lines-of-fire/:lineOfFireId/cones",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
			lineOfFireHandler.AddLineOfFireCone,
		)
		teams.PUT("/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
			lineOfFireHandler.UpdateLineOfFireCone,
		)
		teams.DELETE("/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
			lineOfFireHandler.RemoveLineOfFireCone,
		)
		teams.GET("/:id/lines-of-fire/:lineOfFireId/versions",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
			lineOfFireHandler.ListLineOfFireVersions,
		)

		teams.PUT("/:id/zones/:zoneId/name",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.SetZoneName,
		)
		teams.POST("/:id/zones/:zoneId/polygons",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.AddPolygon,
		)
		teams.DELETE("/:id/zones/:zoneId/polygons/:polygonId",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.RemovePolygon,
		)
		teams.PUT("/:id/zones/:zoneId/polygons/:polygonId/points",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.UpdatePolygonPoints,
		)
		teams.POST("/:id/zones/:zoneId/polygons/:polygonId/points",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.AddPolygonPoint,
		)
		teams.DELETE("/:id/zones/:zoneId/polygons/:polygonId/points/:pointIndex",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:edit"),
			zoneHandler.RemovePolygonPoint,
		)
		teams.GET("/:id/maps/:mapId/versions",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
			historyHandler.ListVersions,
		)
		teams.GET("/:id/zones/:zoneId/versions",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
			historyHandler.ListZoneVersions,
		)
		teams.POST("/:id/zones/:zoneId/versions/:versionId/restore",
			middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
			historyHandler.RestoreVersion,
		)

		teams.GET("/:id/invitations/pending",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
			teamInvitationHandler.ListPendingByTeam,
		)

		teams.GET("/:id/join-requests/mine",
			middleware.RequireAuthOrMember(logger, teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.GetMyRequest,
		)
		teams.DELETE("/:id/join-requests/mine",
			middleware.RequireAuthOrMember(logger, teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.CancelMyRequest,
		)
		teams.POST("/:id/join-requests",
			middleware.RequireAuthOrMember(logger, teamMemberRepo, userRepo, jwtService),
			joinRequestHandler.RequestJoin,
		)
		teams.GET("/:id/join-requests",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
			joinRequestHandler.ListPendingRequests,
		)
		teams.PUT("/:id/join-requests/:requestId/accept",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
			joinRequestHandler.AcceptRequest,
		)
		teams.PUT("/:id/join-requests/:requestId/decline",
			middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain"),
			middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
			joinRequestHandler.DeclineRequest,
		)
	}

	invitations := rg.Group("/invitations")
	invitations.Use(middleware.RequireAuth(logger, userRepo, jwtService))
	{
		invitations.GET("", teamInvitationHandler.ListMyInvitations)
		invitations.POST("/:invitationId/accept", teamInvitationHandler.AcceptInvitation)
		invitations.POST("/:invitationId/decline", teamInvitationHandler.DeclineInvitation)
	}

	inviteLinks := rg.Group("/invite-links")
	{
		inviteLinks.GET("/:token", teamInvitationHandler.GetInviteLinkInfo)
		inviteLinks.POST("/:token/accept",
			middleware.RequireAuth(logger, userRepo, jwtService),
			teamInvitationHandler.AcceptInviteLink,
		)
	}

	users := rg.Group("/users")
	users.Use(middleware.RequireAuth(logger, userRepo, jwtService))
	{
		users.GET("/search", teamInvitationHandler.SearchUsers)
	}
}

func setupAdminRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	permRepo := models.NewPermissionRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	globalRoleRepo := models.NewGlobalRoleRepository(db)
	teamRoleRepo := models.NewTeamRoleRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
	mapRepo := models.NewMapRepository(db)
	mapHandler := handlers.NewMapHandler(mapRepo, logger)
	adminHandler := handlers.NewAdminHandler(userRepo, permRepo, rolePermRepo, teamRepo, teamMemberRepo, globalRoleRepo, teamRoleRepo, logger)

	lockRepoAdmin := models.NewEditLockRepository(db)
	lockHandlerAdmin := handlers.NewEditLockHandler(lockRepoAdmin, userRepo, logger)

	templateRepoAdmin := models.NewTemplateRepository(db)
	zoneRepoAdmin := models.NewZoneRepository(db)
	templateHandlerAdmin := handlers.NewTemplateHandler(templateRepoAdmin, zoneRepoAdmin, teamMemberRepo, logger)

	admin := rg.Group("/admin")
	admin.Use(middleware.RequireAuth(logger, userRepo, jwtService))
	admin.Use(middleware.RequireAdmin(logger))
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

		admin.DELETE("/templates/:templateId", templateHandlerAdmin.AdminDeleteTemplate)

		splitRepo := models.NewSplitRepository(db)
		splitHandler := handlers.NewSplitHandler(splitRepo, logger)

		admin.POST("/splits", splitHandler.CreateSplit)
		admin.GET("/splits", splitHandler.ListSplits)
		admin.GET("/splits/:id", splitHandler.GetSplit)
		admin.PUT("/splits/:id", splitHandler.UpdateSplit)
		admin.DELETE("/splits/:id", splitHandler.DeleteSplit)
		admin.GET("/splits/:id/maps", splitHandler.GetSplitMaps)

		mapCatalogRepo := models.NewMapCatalogRepository(db)
		mapCatalogHandler := handlers.NewMapCatalogHandler(mapCatalogRepo, logger)

		admin.GET("/map-catalog", mapCatalogHandler.List)
		admin.POST("/map-catalog", mapCatalogHandler.Create)
		admin.PUT("/map-catalog/:id", mapCatalogHandler.Update)
		admin.DELETE("/map-catalog/:id", mapCatalogHandler.Delete)

		zoneRepoExport := models.NewZoneRepository(db)
		mapRepoExport := models.NewMapRepository(db)
		teamRepoExport := models.NewTeamRepository(db)
		importExportHandler := handlers.NewImportExportHandler(zoneRepoExport, mapRepoExport, teamRepoExport, logger)

		admin.GET("/teams/:id/export", importExportHandler.ExportTeamMaps)
		admin.GET("/teams/:id/export/:mapId", importExportHandler.ExportTeamMap)
		admin.POST("/teams/:id/import", importExportHandler.ImportTeamZones)

		archiveGroup := admin.Group("/archive")
		{
			archiveGroup.GET("/archivists", adminHandler.ListArchivists)
			archiveGroup.POST("/archivists", adminHandler.AddArchivist)
			archiveGroup.DELETE("/archivists/:userId", adminHandler.RemoveArchivist)
			archiveGroup.GET("/reader-teams", adminHandler.ListReaderTeams)
			archiveGroup.POST("/reader-teams", adminHandler.AddReaderTeam)
			archiveGroup.DELETE("/reader-teams/:teamId", adminHandler.RemoveReaderTeam)
		}

		ligueRepoAdmin := models.NewLigueRepository(db)
		ligueLocaleRepoAdmin := models.NewLigueLocaleRepository(db)
		adminLigueHandler := handlers.NewAdminLigueHandler(ligueRepoAdmin, ligueLocaleRepoAdmin, logger)

		adminLigues := admin.Group("/ligues")
		{
			adminLigues.GET("", adminLigueHandler.ListLigues)
			adminLigues.GET("/:id", adminLigueHandler.GetLigue)
			adminLigues.POST("", adminLigueHandler.CreateLigue)
			adminLigues.PUT("/:id", adminLigueHandler.UpdateLigue)
			adminLigues.DELETE("/:id", adminLigueHandler.DeleteLigue)
		}
	}
}

func setupArchiveRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	globalRoleRepo := models.NewGlobalRoleRepository(db)
	teamRoleRepo := models.NewTeamRoleRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)

	ligueLocaleRepo := models.NewLigueLocaleRepository(db)
	ligueRepo := models.NewLigueRepository(db)
	splitRepo := models.NewSplitRepository(db)
	evaTeamRepo := models.NewEvaTeamRepository(db)
	teamRepo := models.NewTeamRepository(db)
	archivePermissionsHandler := handlers.NewArchivePermissionsHandler(globalRoleRepo, teamRoleRepo, logger)
	archiveHandler := handlers.NewArchiveHandler(ligueLocaleRepo, ligueRepo, splitRepo, evaTeamRepo, teamRepo, logger)

	archive := rg.Group("/archive")
	{
		archive.GET("/permissions",
			middleware.OptionalAuth(logger, userRepo, jwtService),
			archivePermissionsHandler.GetPermissions,
		)

		archive.GET("/leagues-locales",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
			archiveHandler.ListLiguesLocales,
		)
		archive.GET("/leagues-locales/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
			archiveHandler.GetLigueLocale,
		)
		archive.POST("/leagues-locales",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.CreateLigueLocale,
		)
		archive.PUT("/leagues-locales/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.UpdateLigueLocale,
		)
		archive.DELETE("/leagues-locales/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.DeleteLigueLocale,
		)
		archive.POST("/leagues-locales/:id/import-ligue",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.ImportLigue,
		)
		archive.GET("/ligues",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
			archiveHandler.ListLigues,
		)
		archive.POST("/ligues",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.CreateLigue,
		)
		archive.PUT("/ligues/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.UpdateLigue,
		)
		archive.DELETE("/ligues/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.DeleteLigue,
		)
		archive.POST("/ligues/:id/refresh",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.RefreshLigue,
		)
		archive.GET("/teams",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
			archiveHandler.ListEvaTeams,
		)
		archive.GET("/teams/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
			archiveHandler.GetEvaTeam,
		)
		archive.POST("/teams",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.CreateEvaTeam,
		)
		archive.PUT("/teams/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.UpdateEvaTeam,
		)
		archive.DELETE("/teams/:id",
			middleware.RequireAuth(logger, userRepo, jwtService),
			middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
			archiveHandler.DeleteEvaTeam,
		)
	}
}

func setupAuthRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger) {
	db := database.DB()
	userRepo := models.NewUserRepository(db)
	refreshRepo := models.NewRefreshTokenRepository(db)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
	oauthCfg := services.DiscordOAuthConfig(cfg.Discord.ClientID, cfg.Discord.ClientSecret, cfg.Discord.RedirectURL)
	authHandler := handlers.NewAuthHandler(userRepo, refreshRepo, oauthCfg, jwtService, cfg.Server.FrontendURL, logger)

	rg.GET("/auth/discord/login", authHandler.LoginRedirect)
	rg.GET("/auth/discord/callback", authHandler.Callback)
	rg.POST("/auth/refresh", authHandler.Refresh)

	protected := rg.Group("")
	protected.Use(middleware.RequireAuth(logger, userRepo, jwtService))
	{
		protected.GET("/auth/me", authHandler.Me)
		protected.PUT("/auth/display-name", authHandler.UpdateDisplayName)
		protected.GET("/auth/settings", authHandler.GetSettings)
		protected.PUT("/auth/settings", authHandler.UpdateSettings)
	}
}
