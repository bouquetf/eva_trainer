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

type TeamJoinRequestHandler struct {
	joinRequestRepo models.TeamJoinRequestRepository
	teamMemberRepo  models.TeamMemberRepository
	teamRepo        models.TeamRepository
	userRepo        models.UserRepository
	logger          services.Logger
}

func NewTeamJoinRequestHandler(
	joinRequestRepo models.TeamJoinRequestRepository,
	teamMemberRepo models.TeamMemberRepository,
	teamRepo models.TeamRepository,
	userRepo models.UserRepository,
	logger services.Logger,
) *TeamJoinRequestHandler {
	return &TeamJoinRequestHandler{
		joinRequestRepo: joinRequestRepo,
		teamMemberRepo:  teamMemberRepo,
		teamRepo:        teamRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

func (h *TeamJoinRequestHandler) RequestJoin(c *gin.Context) {
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

	jr, err := h.joinRequestRepo.Create(teamID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			apierror.RespondWithError(c, apierror.NewAPIError("already_member", "Vous êtes déjà membre de cette équipe", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrJoinRequestAlreadyExists) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_already_exists", "Vous avez déjà une demande en attente pour cette équipe", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
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
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	reqs, err := h.joinRequestRepo.FindPendingByTeam(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
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

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"requests": items})
}

func (h *TeamJoinRequestHandler) AcceptRequest(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	requestID, err := strconv.Atoi(c.Param("requestId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_id", "invalid request id", http.StatusBadRequest))
		return
	}

	jr, err := h.joinRequestRepo.FindByID(requestID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if jr.TeamID != teamID {
		apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
		return
	}

	if jr.Status != "pending" {
		apierror.RespondWithError(c, apierror.NewAPIError("request_already_processed", "Cette demande a déjà été traitée", http.StatusBadRequest))
		return
	}

	if err := h.joinRequestRepo.AcceptRequest(requestID); err != nil {
		if errors.Is(err, models.ErrJoinRequestAlreadyMember) {
			apierror.RespondWithError(c, apierror.NewAPIError("already_member", "Vous êtes déjà membre de cette équipe", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Demande acceptée",
		"user_id": jr.UserID,
	})
}

func (h *TeamJoinRequestHandler) DeclineRequest(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	requestID, err := strconv.Atoi(c.Param("requestId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_id", "invalid request id", http.StatusBadRequest))
		return
	}

	jr, err := h.joinRequestRepo.FindByID(requestID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if jr.TeamID != teamID {
		apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
		return
	}

	if jr.Status != "pending" {
		apierror.RespondWithError(c, apierror.NewAPIError("request_already_processed", "Cette demande a déjà été traitée", http.StatusBadRequest))
		return
	}

	if err := h.joinRequestRepo.DeclineRequest(requestID); err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Demande refusée",
		"user_id": jr.UserID,
	})
}

func (h *TeamJoinRequestHandler) GetMyRequest(c *gin.Context) {
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

	jr, err := h.joinRequestRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Aucune demande d'adhésion trouvée", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
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
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	jr, err := h.joinRequestRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Aucune demande d'adhésion trouvée", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if jr.UserID != user.ID {
		apierror.RespondWithError(c, apierror.NewAPIError("not_your_request", "Vous n'êtes pas l'auteur de cette demande", http.StatusForbidden))
		return
	}

	if jr.Status != "pending" {
		apierror.RespondWithError(c, apierror.NewAPIError("request_already_processed", "Cette demande a déjà été traitée", http.StatusBadRequest))
		return
	}

	if err := h.joinRequestRepo.CancelRequest(jr.ID); err != nil {
		if errors.Is(err, models.ErrJoinRequestNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("request_not_found", "Demande d'adhésion introuvable", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Demande annulée"})
}
