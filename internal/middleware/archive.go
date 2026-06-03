package middleware

import (
	"net/http"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type ArchivePermissions struct {
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
}

func CheckArchivePermissions(
	user *models.User,
	globalRoleRepo models.GlobalRoleRepository,
	teamRoleRepo models.TeamRoleRepository,
) (*ArchivePermissions, error) {
	if user.IsAdmin {
		return &ArchivePermissions{CanRead: true, CanWrite: true}, nil
	}

	isArchivist, err := globalRoleRepo.UserHasRole(user.ID, models.RoleArchivist)
	if err != nil {
		return nil, err
	}

	if isArchivist {
		return &ArchivePermissions{CanRead: true, CanWrite: true}, nil
	}

	hasReaderRole, err := teamRoleRepo.UserIsMemberOfTeamWithRole(user.ID, models.RoleReader)
	if err != nil {
		return nil, err
	}

	return &ArchivePermissions{CanRead: hasReaderRole, CanWrite: false}, nil
}

func RequireArchiveAccess(
	logger services.Logger,
	globalRoleRepo models.GlobalRoleRepository,
	teamRoleRepo models.TeamRoleRepository,
	accessType string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			return
		}

		perms, err := CheckArchivePermissions(user, globalRoleRepo, teamRoleRepo)
		if err != nil {
			logger.Error("failed to check archive permissions", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}

		if accessType == "write" && !perms.CanWrite {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required permission for this action", http.StatusForbidden))
			return
		}

		if accessType == "read" && !perms.CanRead {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required permission for this action", http.StatusForbidden))
			return
		}

		c.Next()
	}
}
