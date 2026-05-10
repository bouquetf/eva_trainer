package middleware

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func RequireTeamRole(logger services.Logger, teamMemberRepo models.TeamMemberRepository, allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		teamID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeBadRequest, "invalid team id", http.StatusBadRequest))
			return
		}

		member, err := teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you are not a member of this team", http.StatusForbidden))
			return
		}

		if len(roleSet) > 0 && !roleSet[member.Role] {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "you don't have the required role", http.StatusForbidden))
			return
		}

		c.Set("team_member", member)
		c.Set("team_id", teamID)
		c.Next()
	}
}

func GetTeamMember(c *gin.Context) (*models.TeamMember, bool) {
	member, exists := c.Get("team_member")
	if !exists {
		return nil, false
	}
	m, ok := member.(*models.TeamMember)
	return m, ok
}

func RequireAuthOrMember(logger services.Logger, teamMemberRepo models.TeamMemberRepository, userRepo models.UserRepository, jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		teamID, err := strconv.Atoi(c.Param("id"))
		if err == nil {
			member, err := teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
			if err == nil {
				c.Set("team_member", member)
			}
		}

		c.Next()
	}
}
