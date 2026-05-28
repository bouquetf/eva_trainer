package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	logger         services.Logger
}

func (h *TeamHandler) ListAllTeams(c *gin.Context) {
	q := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 20
	}

	teams, total, err := h.teamRepo.FindAllPaginated(q, page, perPage)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type teamItem struct {
		Team        *models.Team `json:"team"`
		MemberCount int          `json:"member_count"`
	}

	items := make([]teamItem, 0, len(teams))
	for _, t := range teams {
		members, err := h.teamMemberRepo.FindByTeamID(t.ID)
		memberCount := 0
		if err == nil {
			memberCount = len(members)
		}
		items = append(items, teamItem{
			Team:        t,
			MemberCount: memberCount,
		})
	}

	user, _ := middleware.GetUser(c)

	type myTeamStatus struct {
		TeamID int    `json:"team_id"`
		Role   string `json:"role"`
	}
	var myTeams []myTeamStatus
	if user != nil {
		memberList, err := h.teamMemberRepo.FindByUserID(user.ID)
		if err == nil {
			for _, m := range memberList {
				myTeams = append(myTeams, myTeamStatus{TeamID: m.TeamID, Role: m.Role})
			}
		}
	}
	if myTeams == nil {
		myTeams = []myTeamStatus{}
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"teams":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"my_teams": myTeams,
	})
}

func NewTeamHandler(teamRepo models.TeamRepository, teamMemberRepo models.TeamMemberRepository, logger services.Logger) *TeamHandler {
	return &TeamHandler{
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		logger:         logger,
	}
}

func (h *TeamHandler) Create(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_name", "team name is required", http.StatusBadRequest))
		return
	}

	team, err := h.teamRepo.Create(req.Name, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidTeamName) || errors.Is(err, models.ErrTeamNameTooLong) || errors.Is(err, models.ErrTeamNameRequired) {
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_name", err.Error(), http.StatusBadRequest))
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			apierror.RespondWithError(c, apierror.NewAPIError("team_name_exists", "team name already exists", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"id":         team.ID,
		"name":       team.Name,
		"captain_id": team.CaptainID,
		"created_at": team.CreatedAt,
		"updated_at": team.UpdatedAt,
	})
}

func (h *TeamHandler) ListMyTeams(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	teams, err := h.teamRepo.FindByUserID(user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	type teamItem struct {
		Team        *models.Team `json:"team"`
		MemberCount int          `json:"member_count"`
		MyRole      string       `json:"my_role"`
	}

	items := make([]teamItem, 0, len(teams))
	for _, t := range teams {
		members, err := h.teamMemberRepo.FindByTeamID(t.ID)
		memberCount := 0
		if err == nil {
			memberCount = len(members)
		}

		myRole := ""
		member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, t.ID)
		if err == nil {
			myRole = member.Role
		}

		items = append(items, teamItem{
			Team:        t,
			MemberCount: memberCount,
			MyRole:      myRole,
		})
	}

	c.JSON(http.StatusOK, items)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
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

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	members, err := h.teamMemberRepo.FindByTeamIDWithDisplayNames(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"team":    team,
		"members": members,
	})
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
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

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_name", "team name is required", http.StatusBadRequest))
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if member.Role != models.RoleCaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_role", "only the captain can update the team", http.StatusForbidden))
		return
	}

	if err := h.teamRepo.UpdateName(teamID, req.Name); err != nil {
		if errors.Is(err, models.ErrInvalidTeamName) || errors.Is(err, models.ErrTeamNameTooLong) || errors.Is(err, models.ErrTeamNameRequired) {
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_name", err.Error(), http.StatusBadRequest))
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			apierror.RespondWithError(c, apierror.NewAPIError("team_name_exists", "team name already exists", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"id":         team.ID,
		"name":       team.Name,
		"captain_id": team.CaptainID,
		"created_at": team.CreatedAt,
		"updated_at": team.UpdatedAt,
	})
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
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

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if member.Role != models.RoleCaptain {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_role", "only the captain can delete the team", http.StatusForbidden))
		return
	}

	var req struct {
		ConfirmWithName string `json:"confirm_with_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.ConfirmWithName == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("confirmation_name_mismatch", "you must provide the team name to confirm deletion", http.StatusBadRequest))
		return
	}

	if req.ConfirmWithName != team.Name {
		apierror.RespondWithError(c, apierror.NewAPIError("confirmation_name_mismatch", "team name does not match, confirmation required", http.StatusBadRequest))
		return
	}

	if err := convertTeamZonesToTemplates(teamID); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if err := h.teamRepo.Delete(teamID); err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message": "team deleted",
	})
}

func (h *TeamHandler) LeaveTeam(c *gin.Context) {
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

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	if member.Role == models.RoleCaptain {
		if err := convertTeamZonesToTemplates(teamID); err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
			return
		}

		if err := h.teamRepo.Delete(teamID); err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
			return
		}

		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
			"message": "team deleted (captain left without transfer)",
		})
		return
	}

	if err := h.teamMemberRepo.RemoveMember(teamID, user.ID); err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("member_not_found", "member not found in team", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message": "left team",
	})
}

func convertTeamZonesToTemplates(teamID int) error {
	_ = teamID
	return nil
}
