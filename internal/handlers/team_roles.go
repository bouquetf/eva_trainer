package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type TeamRolesHandler struct {
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	userRepo       models.UserRepository
	rolePermRepo   models.RolePermissionRepository
	permRepo       models.PermissionRepository
	logger         services.Logger
}

func NewTeamRolesHandler(
	teamRepo models.TeamRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	rolePermRepo models.RolePermissionRepository,
	permRepo models.PermissionRepository,
	logger services.Logger,
) *TeamRolesHandler {
	return &TeamRolesHandler{
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
		rolePermRepo:   rolePermRepo,
		permRepo:       permRepo,
		logger:         logger,
	}
}

func (h *TeamRolesHandler) ListMembers(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	members, err := h.teamMemberRepo.FindByTeamIDWithDisplayNames(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type memberItem struct {
		UserID      int    `json:"user_id"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		JoinedAt    string `json:"joined_at"`
	}

	items := make([]memberItem, 0, len(members))
	for _, m := range members {
		items = append(items, memberItem{
			UserID:      m.UserID,
			DisplayName: m.DisplayName,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"members": items})
}

func (h *TeamRolesHandler) UpdateRole(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_user_id", "invalid user id", http.StatusBadRequest))
		return
	}

	currentMember, ok := middleware.GetTeamMember(c)
	if !ok {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Role == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_role", "role is required", http.StatusBadRequest))
		return
	}

	if !models.ValidRoles[req.Role] {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_role", "invalid team role", http.StatusBadRequest))
		return
	}

	targetMember, err := h.teamMemberRepo.FindByUserAndTeam(targetUserID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("member_not_found", "member not found in team", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	targetUser, err := h.userRepo.FindByID(targetUserID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if targetUser.IsAdmin {
		apierror.RespondWithError(c, apierror.NewAPIError("cannot_downgrade_admin", "cannot downgrade an admin", http.StatusForbidden))
		return
	}

	if targetMember.Role == models.RoleCaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("cannot_downgrade_captain", "cannot downgrade the captain, transfer captaincy first", http.StatusForbidden))
		return
	}

	if currentMember.Role == models.RoleCocaptain && req.Role == models.RoleCocaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_role", "only the captain can promote to cocaptain", http.StatusForbidden))
		return
	}

	if currentMember.Role == models.RoleCocaptain && targetMember.Role == models.RoleCocaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_role", "you cannot change the role of another cocaptain", http.StatusForbidden))
		return
	}

	if err := h.teamMemberRepo.UpdateRole(teamID, targetUserID, req.Role); err != nil {
		if errors.Is(err, models.ErrTeamAlreadyHasCocaptain) {
			apierror.RespondWithError(c, apierror.NewAPIError("team_already_has_cocaptain", "team already has a cocaptain", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message": "role updated",
		"user_id": targetUserID,
		"role":    req.Role,
	})
}

func (h *TeamRolesHandler) TransferCaptaincy(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	currentMember, ok := middleware.GetTeamMember(c)
	if !ok {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if currentMember.Role != models.RoleCaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_role", "only the captain can transfer captaincy", http.StatusForbidden))
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request", "user_id is required", http.StatusBadRequest))
		return
	}

	if req.UserID == user.ID {
		apierror.RespondWithError(c, apierror.NewAPIError("same_user", "cannot transfer captaincy to yourself", http.StatusBadRequest))
		return
	}

	_, err = h.teamMemberRepo.FindByUserAndTeam(req.UserID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("member_not_found", "target user is not a member of this team", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	targetUser, err := h.userRepo.FindByID(req.UserID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if targetUser.IsAdmin {
		apierror.RespondWithError(c, apierror.NewAPIError("cannot_downgrade_admin", "cannot transfer captaincy to an admin", http.StatusForbidden))
		return
	}

	if err := h.teamRepo.TransferCaptaincy(teamID, user.ID, req.UserID); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message":        "captaincy transferred",
		"new_captain_id": req.UserID,
	})
}
