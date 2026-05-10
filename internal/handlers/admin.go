package handlers

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo       models.UserRepository
	permRepo       models.PermissionRepository
	rolePermRepo   models.RolePermissionRepository
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	logger         services.Logger
}

func NewAdminHandler(
	userRepo models.UserRepository,
	permRepo models.PermissionRepository,
	rolePermRepo models.RolePermissionRepository,
	teamRepo models.TeamRepository,
	teamMemberRepo models.TeamMemberRepository,
	logger services.Logger,
) *AdminHandler {
	return &AdminHandler{
		userRepo:       userRepo,
		permRepo:       permRepo,
		rolePermRepo:   rolePermRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		logger:         logger,
	}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list users", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err == models.ErrUserNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("user_not_found", "user not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find user", http.StatusInternalServerError))
		return
	}

	memberships, err := h.teamMemberRepo.FindByUserID(userID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list memberships", http.StatusInternalServerError))
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

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"user":        user,
		"memberships": result,
	})
}

func (h *AdminHandler) ListAllTeams(c *gin.Context) {
	teams, err := h.teamRepo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list teams", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"teams": teams})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("teamId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Role == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_body", "role is required", http.StatusBadRequest))
		return
	}

	if err := h.teamMemberRepo.UpdateRole(teamID, userID, req.Role); err != nil {
		if err == models.ErrMemberNotFound {
			apierror.RespondWithError(c, apierror.NewAPIError("member_not_found", "user is not a member of this team", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_role", err.Error(), http.StatusBadRequest))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) AddUserToTeam(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	var req struct {
		TeamID int    `json:"team_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TeamID == 0 || req.Role == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_body", "team_id and role are required", http.StatusBadRequest))
		return
	}

	if err := h.teamMemberRepo.AddMember(req.TeamID, userID, req.Role); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("add_member_failed", err.Error(), http.StatusBadRequest))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"ok": true})
}

func (h *AdminHandler) RemoveUserFromTeam(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("teamId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(userID, teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("member_not_found", "user is not a member of this team", http.StatusNotFound))
		return
	}

	if member.Role == models.RoleCaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("cannot_remove_captain", "cannot remove a captain from a team. transfer captaincy first", http.StatusBadRequest))
		return
	}

	if err := h.teamMemberRepo.RemoveMember(teamID, userID); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to remove member", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) ListPermissions(c *gin.Context) {
	perms, err := h.permRepo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list permissions", http.StatusInternalServerError))
		return
	}

	roles, err := h.rolePermRepo.FindAllRoles()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list roles", http.StatusInternalServerError))
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
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find role permissions", http.StatusInternalServerError))
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

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"permissions": result,
		"roles":       roles,
	})
}

func (h *AdminHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_role", "role is required", http.StatusBadRequest))
		return
	}

	perms, err := h.rolePermRepo.FindByRole(role)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list permissions", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"role":        role,
		"permissions": perms,
	})
}

func (h *AdminHandler) PromoteToAdmin(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	targetUser, err := h.userRepo.FindByID(userID)
	if err == models.ErrUserNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("user_not_found", "user not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find user", http.StatusInternalServerError))
		return
	}

	if targetUser.IsAdmin {
		apierror.RespondWithError(c, apierror.NewAPIError("already_admin", "user is already an administrator", http.StatusBadRequest))
		return
	}

	if err := h.userRepo.SetAdmin(userID); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to promote user", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message":  "user promoted to admin",
		"user_id":  userID,
		"is_admin": true,
	})
}
