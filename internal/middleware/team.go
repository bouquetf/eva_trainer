package middleware

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

func RequireTeamRole(teamMemberRepo models.TeamMemberRepository, allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		teamID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_team_id",
				"message": "invalid team id",
			})
			c.Abort()
			return
		}

		member, err := teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_team_member",
				"message": "you are not a member of this team",
			})
			c.Abort()
			return
		}

		if len(roleSet) > 0 && !roleSet[member.Role] {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_role",
				"message": "you don't have the required role",
			})
			c.Abort()
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
