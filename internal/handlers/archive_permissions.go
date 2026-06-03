package handlers

import (
	"net/http"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type ArchivePermissionsHandler struct {
	globalRoleRepo models.GlobalRoleRepository
	teamRoleRepo   models.TeamRoleRepository
	logger         services.Logger
}

func NewArchivePermissionsHandler(
	globalRoleRepo models.GlobalRoleRepository,
	teamRoleRepo models.TeamRoleRepository,
	logger services.Logger,
) *ArchivePermissionsHandler {
	return &ArchivePermissionsHandler{
		globalRoleRepo: globalRoleRepo,
		teamRoleRepo:   teamRoleRepo,
		logger:         logger,
	}
}

func (h *ArchivePermissionsHandler) GetPermissions(c *gin.Context) {
	user, ok := middleware.GetUser(c)
	if !ok {
		apierror.RespondWithSuccess(c, http.StatusOK, middleware.ArchivePermissions{CanRead: false, CanWrite: false})
		return
	}

	perms, err := middleware.CheckArchivePermissions(user, h.globalRoleRepo, h.teamRoleRepo)
	if err != nil {
		h.logger.Error("failed to check archive permissions", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, perms)
}
