package middleware

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func RequirePermission(
	logger services.Logger,
	teamMemberRepo models.TeamMemberRepository,
	rolePermRepo models.RolePermissionRepository,
	permRepo models.PermissionRepository,
	feature string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		if user.IsAdmin {
			c.Next()
			return
		}

		teamID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeBadRequest, "invalid team id", http.StatusBadRequest))
			return
		}

		member, exists := GetTeamMember(c)
		if !exists {
			member, err = teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you are not a member of this team", http.StatusForbidden))
				return
			}
		}

		hasPermission, err := models.HasPermission(member.Role, feature, rolePermRepo, permRepo)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}

		if !hasPermission {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required permission for this action", http.StatusForbidden))
			return
		}

		c.Next()
	}
}

func GetUserPermissions(
	userID int,
	teamID int,
	userRepo models.UserRepository,
	teamMemberRepo models.TeamMemberRepository,
	rolePermRepo models.RolePermissionRepository,
	permRepo models.PermissionRepository,
) ([]string, error) {
	user, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user.IsAdmin {
		allPerms, err := permRepo.FindAll()
		if err != nil {
			return nil, err
		}

		features := make([]string, 0, len(allPerms))
		for _, p := range allPerms {
			features = append(features, p.Feature)
		}
		return features, nil
	}

	member, err := teamMemberRepo.FindByUserAndTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	perms, err := rolePermRepo.FindByRole(member.Role)
	if err != nil {
		return nil, err
	}

	features := make([]string, 0, len(perms))
	for _, p := range perms {
		features = append(features, p.Feature)
	}
	return features, nil
}

func RequireGlobalPermission(
	logger services.Logger,
	permRepo models.PermissionRepository,
	rolePermRepo models.RolePermissionRepository,
	feature string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		if user.IsAdmin {
			c.Next()
			return
		}

		perm, err := permRepo.FindByFeature(feature)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required permission for this action", http.StatusForbidden))
			return
		}

		hasPerm, err := rolePermRepo.RoleHasPermission("member", perm.ID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}

		if !hasPerm {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required permission for this action", http.StatusForbidden))
			return
		}

		c.Next()
	}
}

func GetRequiredRole(
	feature string,
	rolePermRepo models.RolePermissionRepository,
) (string, error) {
	roles := []string{"member", "editor", "cocaptain", "captain"}

	for _, role := range roles {
		perms, err := rolePermRepo.FindByRole(role)
		if err != nil {
			return "", err
		}

		for _, p := range perms {
			if p.Feature == feature {
				return role, nil
			}
		}
	}

	return "", nil
}
