package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type TeamJoinRequestHandler struct {
	joinRequestRepo models.TeamJoinRequestRepository
	teamMemberRepo  models.TeamMemberRepository
	teamRepo        models.TeamRepository
	userRepo        models.UserRepository
}

func NewTeamJoinRequestHandler(
	joinRequestRepo models.TeamJoinRequestRepository,
	teamMemberRepo models.TeamMemberRepository,
	teamRepo models.TeamRepository,
	userRepo models.UserRepository,
) *TeamJoinRequestHandler {
	return &TeamJoinRequestHandler{
		joinRequestRepo: joinRequestRepo,
		teamMemberRepo:  teamMemberRepo,
		teamRepo:        teamRepo,
		userRepo:        userRepo,
	}
}

func (h *TeamJoinRequestHandler) RequestJoin(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "unauthorized"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	jr, err := h.joinRequestRepo.Create(teamID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "already_member", "message": "Vous êtes déjà membre de cette équipe"})
			return
		}
		if errors.Is(err, models.ErrJoinRequestAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request_already_exists", "message": "Vous avez déjà une demande en attente pour cette équipe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         jr.ID,
		"team_id":    jr.TeamID,
		"user_id":    jr.UserID,
		"status":     jr.Status,
		"created_at": jr.CreatedAt,
	})
}

func (h *TeamJoinRequestHandler) ListPendingRequests(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	reqs, err := h.joinRequestRepo.FindPendingByTeam(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	type requestItem struct {
		ID          int       `json:"id"`
		UserID      int       `json:"user_id"`
		DisplayName string    `json:"display_name"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
	}

	items := make([]requestItem, 0, len(reqs))
	for _, r := range reqs {
		displayName := ""
		u, err := h.userRepo.FindByID(r.UserID)
		if err == nil {
			displayName = u.DisplayName
		}

		items = append(items, requestItem{
			ID:          r.ID,
			UserID:      r.UserID,
			DisplayName: displayName,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"requests": items})
}

func (h *TeamJoinRequestHandler) AcceptRequest(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	requestID, err := strconv.Atoi(c.Param("requestId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_id", "message": "invalid request id"})
		return
	}

	jr, err := h.joinRequestRepo.FindByID(requestID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	if jr.TeamID != teamID {
		c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
		return
	}

	if jr.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_already_processed", "message": "Cette demande a déjà été traitée"})
		return
	}

	if err := h.joinRequestRepo.AcceptRequest(requestID); err != nil {
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "already_member", "message": "Vous êtes déjà membre de cette équipe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Demande acceptée",
		"user_id": jr.UserID,
	})
}

func (h *TeamJoinRequestHandler) DeclineRequest(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	requestID, err := strconv.Atoi(c.Param("requestId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_id", "message": "invalid request id"})
		return
	}

	jr, err := h.joinRequestRepo.FindByID(requestID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	if jr.TeamID != teamID {
		c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
		return
	}

	if jr.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_already_processed", "message": "Cette demande a déjà été traitée"})
		return
	}

	if err := h.joinRequestRepo.DeclineRequest(requestID); err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Demande refusée",
		"user_id": jr.UserID,
	})
}

func (h *TeamJoinRequestHandler) GetMyRequest(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "unauthorized"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	jr, err := h.joinRequestRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Aucune demande d'adhésion trouvée"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         jr.ID,
		"team_id":    jr.TeamID,
		"user_id":    jr.UserID,
		"status":     jr.Status,
		"created_at": jr.CreatedAt,
	})
}

func (h *TeamJoinRequestHandler) CancelMyRequest(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "unauthorized"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	jr, err := h.joinRequestRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Aucune demande d'adhésion trouvée"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	if jr.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not_your_request", "message": "Vous n'êtes pas l'auteur de cette demande"})
		return
	}

	if jr.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_already_processed", "message": "Cette demande a déjà été traitée"})
		return
	}

	if err := h.joinRequestRepo.CancelRequest(jr.ID); err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found", "message": "Demande d'adhésion introuvable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Demande annulée"})
}
