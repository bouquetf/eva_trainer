package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type TeamRolesHandler struct {
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	userRepo       models.UserRepository
	rolePermRepo   models.RolePermissionRepository
	permRepo       models.PermissionRepository
}

func NewTeamRolesHandler(
	teamRepo models.TeamRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	rolePermRepo models.RolePermissionRepository,
	permRepo models.PermissionRepository,
) *TeamRolesHandler {
	return &TeamRolesHandler{
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
		rolePermRepo:   rolePermRepo,
		permRepo:       permRepo,
	}
}

func (h *TeamRolesHandler) ListMembers(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_id",
			"message": "invalid team id",
		})
		return
	}

	members, err := h.teamMemberRepo.FindByTeamIDWithDisplayNames(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
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

	c.JSON(http.StatusOK, gin.H{
		"members": items,
	})
}

func (h *TeamRolesHandler) UpdateRole(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_id",
			"message": "invalid team id",
		})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "invalid user id",
		})
		return
	}

	currentMember, ok := middleware.GetTeamMember(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_role",
			"message": "role is required",
		})
		return
	}

	if !models.ValidRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_role",
			"message": "invalid team role",
		})
		return
	}

	targetMember, err := h.teamMemberRepo.FindByUserAndTeam(targetUserID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "member_not_found",
				"message": "member not found in team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	targetUser, err := h.userRepo.FindByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if targetUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "cannot_downgrade_admin",
			"message": "cannot downgrade an admin",
		})
		return
	}

	if targetMember.Role == models.RoleCaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "cannot_downgrade_captain",
			"message": "cannot downgrade the captain, transfer captaincy first",
		})
		return
	}

	if currentMember.Role == models.RoleCocaptain && req.Role == models.RoleCocaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "only the captain can promote to cocaptain",
		})
		return
	}

	if currentMember.Role == models.RoleCocaptain && targetMember.Role == models.RoleCocaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "you cannot change the role of another cocaptain",
		})
		return
	}

	if err := h.teamMemberRepo.UpdateRole(teamID, targetUserID, req.Role); err != nil {
		if errors.Is(err, models.ErrTeamAlreadyHasCocaptain) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "team_already_has_cocaptain",
				"message": "team already has a cocaptain",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role updated",
		"user_id": targetUserID,
		"role":    req.Role,
	})
}

func (h *TeamRolesHandler) TransferCaptaincy(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_id",
			"message": "invalid team id",
		})
		return
	}

	currentMember, ok := middleware.GetTeamMember(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if currentMember.Role != models.RoleCaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "only the captain can transfer captaincy",
		})
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "user_id is required",
		})
		return
	}

	if req.UserID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "same_user",
			"message": "cannot transfer captaincy to yourself",
		})
		return
	}

	_, err = h.teamMemberRepo.FindByUserAndTeam(req.UserID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "member_not_found",
				"message": "target user is not a member of this team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	targetUser, err := h.userRepo.FindByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if targetUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "cannot_downgrade_admin",
			"message": "cannot transfer captaincy to an admin",
		})
		return
	}

	if err := h.teamRepo.TransferCaptaincy(teamID, user.ID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "captaincy transferred",
		"new_captain_id": req.UserID,
	})
}
