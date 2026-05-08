package handlers

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo       models.UserRepository
	permRepo       models.PermissionRepository
	rolePermRepo   models.RolePermissionRepository
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
}

func NewAdminHandler(
	userRepo models.UserRepository,
	permRepo models.PermissionRepository,
	rolePermRepo models.RolePermissionRepository,
	teamRepo models.TeamRepository,
	teamMemberRepo models.TeamMemberRepository,
) *AdminHandler {
	return &AdminHandler{
		userRepo:       userRepo,
		permRepo:       permRepo,
		rolePermRepo:   rolePermRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
	}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err == models.ErrUserNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "user not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find user",
		})
		return
	}

	memberships, err := h.teamMemberRepo.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list memberships",
		})
		return
	}

	type teamMembership struct {
		TeamID   int    `json:"team_id"`
		TeamName string `json:"team_name"`
		Role     string `json:"role"`
	}

	result := make([]teamMembership, 0, len(memberships))
	for _, m := range memberships {
		team, err := h.teamRepo.FindByID(m.TeamID)
		if err != nil {
			continue
		}
		result = append(result, teamMembership{
			TeamID:   m.TeamID,
			TeamName: team.Name,
			Role:     m.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"memberships": result,
	})
}

func (h *AdminHandler) ListAllTeams(c *gin.Context) {
	teams, err := h.teamRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list teams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	teamID, err := strconv.Atoi(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_id",
			"message": "invalid team id",
		})
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_body",
			"message": "role is required",
		})
		return
	}

	if err := h.teamMemberRepo.UpdateRole(teamID, userID, req.Role); err != nil {
		if err == models.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "member_not_found",
				"message": "user is not a member of this team",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_role",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) AddUserToTeam(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		TeamID int    `json:"team_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TeamID == 0 || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_body",
			"message": "team_id and role are required",
		})
		return
	}

	if err := h.teamMemberRepo.AddMember(req.TeamID, userID, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "add_member_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *AdminHandler) RemoveUserFromTeam(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	teamID, err := strconv.Atoi(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_id",
			"message": "invalid team id",
		})
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(userID, teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "member_not_found",
			"message": "user is not a member of this team",
		})
		return
	}

	if member.Role == models.RoleCaptain {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "cannot_remove_captain",
			"message": "cannot remove a captain from a team. transfer captaincy first",
		})
		return
	}

	if err := h.teamMemberRepo.RemoveMember(teamID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to remove member",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) ListPermissions(c *gin.Context) {
	perms, err := h.permRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list permissions",
		})
		return
	}

	roles, err := h.rolePermRepo.FindAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list roles",
		})
		return
	}

	type permissionWithRoles struct {
		ID          int      `json:"id"`
		Feature     string   `json:"feature"`
		Description string   `json:"description"`
		Roles       []string `json:"roles"`
	}

	result := make([]permissionWithRoles, 0, len(perms))
	for _, p := range perms {
		rps, err := h.rolePermRepo.FindByPermissionID(p.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to find role permissions",
			})
			return
		}

		assignedRoles := make([]string, 0, len(rps))
		for _, rp := range rps {
			assignedRoles = append(assignedRoles, rp.Role)
		}

		result = append(result, permissionWithRoles{
			ID:          p.ID,
			Feature:     p.Feature,
			Description: p.Description,
			Roles:       assignedRoles,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": result,
		"roles":       roles,
	})
}

func (h *AdminHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_role",
			"message": "role is required",
		})
		return
	}

	perms, err := h.rolePermRepo.FindByRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": perms,
	})
}

func (h *AdminHandler) ToggleAdmin(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		IsAdmin bool `json:"is_admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_body",
			"message": "invalid request body",
		})
		return
	}

	adminUser, _ := middleware.MustGetUser(c)
	if adminUser != nil && adminUser.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "cannot_self_demote",
			"message": "you cannot change your own admin status",
		})
		return
	}

	if req.IsAdmin {
		err = h.userRepo.SetAdmin(userID)
	} else {
		err = h.userRepo.RemoveAdmin(userID)
	}

	if err == models.ErrUserNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "user not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
