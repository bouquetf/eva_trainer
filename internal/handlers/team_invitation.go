package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type TeamInvitationHandler struct {
	teamInvitationRepo models.TeamInvitationRepository
	teamMemberRepo     models.TeamMemberRepository
	userRepo           models.UserRepository
	teamRepo           models.TeamRepository
	frontendURL        string
	logger             services.Logger
}

func NewTeamInvitationHandler(
	teamInvitationRepo models.TeamInvitationRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	teamRepo models.TeamRepository,
	frontendURL string,
	logger services.Logger,
) *TeamInvitationHandler {
	return &TeamInvitationHandler{
		teamInvitationRepo: teamInvitationRepo,
		teamMemberRepo:     teamMemberRepo,
		userRepo:           userRepo,
		teamRepo:           teamRepo,
		frontendURL:        frontendURL,
		logger:             logger,
	}
}

func (h *TeamInvitationHandler) InviteByDisplayName(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.DisplayName == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request", "display_name is required", http.StatusBadRequest))
		return
	}

	targetUser, err := h.userRepo.FindByDisplayNameOrDiscordUsername(req.DisplayName)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("user_not_found", "User not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	inv, err := h.teamInvitationRepo.CreateByUser(teamID, targetUser.ID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrCannotInviteSelf) {
			apierror.RespondWithError(c, apierror.NewAPIError("cannot_invite_self", "Vous ne pouvez pas vous inviter vous-même", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			apierror.RespondWithError(c, apierror.NewAPIError("already_member", "Vous êtes déjà membre de cette équipe", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"id":                 inv.ID,
		"team_id":            inv.TeamID,
		"invited_user_id":    inv.InvitedUserID,
		"invited_by_user_id": inv.InvitedByUserID,
		"status":             inv.Status,
		"expires_at":         inv.ExpiresAt,
		"created_at":         inv.CreatedAt,
	})
}

func (h *TeamInvitationHandler) GenerateInviteLink(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	token, err := models.GenerateInviteToken()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	inv, err := h.teamInvitationRepo.CreateLink(teamID, user.ID, token, expiresAt)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	inviteURL := h.frontendURL + "/invite/" + token

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"id":         inv.ID,
		"team_id":    inv.TeamID,
		"token":      inv.Token,
		"invite_url": inviteURL,
		"expires_at": inv.ExpiresAt,
		"created_at": inv.CreatedAt,
	})
}

func (h *TeamInvitationHandler) AcceptInvitation(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	invitationID, err := strconv.Atoi(c.Param("invitationId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_invitation_id", "invalid invitation id", http.StatusBadRequest))
		return
	}

	inv, err := h.teamInvitationRepo.FindByID(invitationID)
	if err != nil {
		if errors.Is(err, models.ErrInvitationNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_not_found", "Invitation not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if inv.InvitedUserID == nil || *inv.InvitedUserID != user.ID {
		apierror.RespondWithError(c, apierror.NewAPIError("not_your_invitation", "Cette invitation ne vous est pas destinée", http.StatusForbidden))
		return
	}

	if err := h.teamInvitationRepo.AcceptInvitation(invitationID); err != nil {
		if errors.Is(err, models.ErrInvitationExpired) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_expired", "Ce lien d'invitation a expiré.", http.StatusGone))
			return
		}
		if errors.Is(err, models.ErrInvitationAlreadyUsed) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_already_used", "Cette invitation a déjà été utilisée.", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			apierror.RespondWithError(c, apierror.NewAPIError("already_member", "Vous êtes déjà membre de cette équipe", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	team, err := h.teamRepo.FindByID(inv.TeamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message":   "Vous avez rejoint l'équipe",
		"team_id":   inv.TeamID,
		"team_name": team.Name,
	})
}

func (h *TeamInvitationHandler) DeclineInvitation(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	invitationID, err := strconv.Atoi(c.Param("invitationId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_invitation_id", "invalid invitation id", http.StatusBadRequest))
		return
	}

	inv, err := h.teamInvitationRepo.FindByID(invitationID)
	if err != nil {
		if errors.Is(err, models.ErrInvitationNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_not_found", "Invitation not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if inv.InvitedUserID == nil || *inv.InvitedUserID != user.ID {
		apierror.RespondWithError(c, apierror.NewAPIError("not_your_invitation", "Cette invitation ne vous est pas destinée", http.StatusForbidden))
		return
	}

	if err := h.teamInvitationRepo.DeclineInvitation(invitationID); err != nil {
		if errors.Is(err, models.ErrInvitationNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_not_found", "Invitation not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Invitation refusée"})
}

func (h *TeamInvitationHandler) AcceptInviteLink(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	token := c.Param("token")
	if token == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_token", "token is required", http.StatusBadRequest))
		return
	}

	if err := h.teamInvitationRepo.AcceptByToken(token, user.ID); err != nil {
		if errors.Is(err, models.ErrInvitationExpired) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_expired", "Ce lien d'invitation a expiré.", http.StatusGone))
			return
		}
		if errors.Is(err, models.ErrInvitationAlreadyUsed) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_already_used", "Cette invitation a déjà été utilisée.", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrInvitationNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_not_found", "Invitation not found", http.StatusNotFound))
			return
		}
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			apierror.RespondWithError(c, apierror.NewAPIError("already_member", "Vous êtes déjà membre de cette équipe", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	inv, err := h.teamInvitationRepo.FindByToken(token)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	team, err := h.teamRepo.FindByID(inv.TeamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message":   "Vous avez rejoint l'équipe",
		"team_id":   inv.TeamID,
		"team_name": team.Name,
	})
}

func (h *TeamInvitationHandler) GetInviteLinkInfo(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_token", "token is required", http.StatusBadRequest))
		return
	}

	inv, err := h.teamInvitationRepo.FindByToken(token)
	if err != nil {
		if errors.Is(err, models.ErrInvitationNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("invitation_not_found", "Invitation not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	team, err := h.teamRepo.FindByID(inv.TeamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	isExpired := time.Now().After(inv.ExpiresAt)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"team_id":    inv.TeamID,
		"team_name":  team.Name,
		"expires_at": inv.ExpiresAt,
		"is_expired": isExpired,
	})
}

func (h *TeamInvitationHandler) ListMyInvitations(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	invs, err := h.teamInvitationRepo.FindPendingByUser(user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type invitationItem struct {
		ID        int       `json:"id"`
		TeamID    int       `json:"team_id"`
		TeamName  string    `json:"team_name"`
		Status    string    `json:"status"`
		ExpiresAt time.Time `json:"expires_at"`
		CreatedAt time.Time `json:"created_at"`
	}

	items := make([]invitationItem, 0, len(invs))
	for _, inv := range invs {
		teamName := ""
		team, err := h.teamRepo.FindByID(inv.TeamID)
		if err == nil {
			teamName = team.Name
		}

		items = append(items, invitationItem{
			ID:        inv.ID,
			TeamID:    inv.TeamID,
			TeamName:  teamName,
			Status:    inv.Status,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, items)
}

func (h *TeamInvitationHandler) ListPendingByTeam(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	invs, err := h.teamInvitationRepo.FindPendingByTeam(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type invitationItem struct {
		ID                     int       `json:"id"`
		InvitedUserID          *int      `json:"invited_user_id,omitempty"`
		InvitedUserDisplayName string    `json:"invited_user_display_name"`
		InvitedByUserID        int       `json:"invited_by_user_id"`
		InvitedByDisplayName   string    `json:"invited_by_display_name"`
		Token                  *string   `json:"token,omitempty"`
		Status                 string    `json:"status"`
		ExpiresAt              time.Time `json:"expires_at"`
		CreatedAt              time.Time `json:"created_at"`
	}

	items := make([]invitationItem, 0, len(invs))
	for _, inv := range invs {
		item := invitationItem{
			ID:              inv.ID,
			InvitedUserID:   inv.InvitedUserID,
			InvitedByUserID: inv.InvitedByUserID,
			Token:           inv.Token,
			Status:          inv.Status,
			ExpiresAt:       inv.ExpiresAt,
			CreatedAt:       inv.CreatedAt,
		}

		if inv.InvitedUserID != nil {
			u, err := h.userRepo.FindByID(*inv.InvitedUserID)
			if err == nil {
				item.InvitedUserDisplayName = u.DisplayName
			}
		}

		invitedBy, err := h.userRepo.FindByID(inv.InvitedByUserID)
		if err == nil {
			item.InvitedByDisplayName = invitedBy.DisplayName
		}

		items = append(items, item)
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"invitations": items})
}

func (h *TeamInvitationHandler) SearchUsers(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	query := c.Query("q")
	if query == "" {
		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"users": []interface{}{}})
		return
	}

	users, err := h.userRepo.SearchUsers(query, 10)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type userItem struct {
		ID              int    `json:"id"`
		DisplayName     string `json:"display_name"`
		DiscordUsername string `json:"discord_username"`
	}

	items := make([]userItem, 0, len(users))
	for _, u := range users {
		if u.ID == user.ID {
			continue
		}
		items = append(items, userItem{
			ID:              u.ID,
			DisplayName:     u.DisplayName,
			DiscordUsername: u.DiscordUsername,
		})
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"users": items})
}
