package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
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

	user, _ := middleware.MustGetUser(c)

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

	c.JSON(http.StatusOK, gin.H{
		"teams":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"my_teams": myTeams,
	})
}

func NewTeamHandler(teamRepo models.TeamRepository, teamMemberRepo models.TeamMemberRepository) *TeamHandler {
	return &TeamHandler{
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
	}
}

func (h *TeamHandler) Create(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_name",
			"message": "team name is required",
		})
		return
	}

	team, err := h.teamRepo.Create(req.Name, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidTeamName) || errors.Is(err, models.ErrTeamNameTooLong) || errors.Is(err, models.ErrTeamNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_team_name",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "team_name_exists",
				"message": "team name already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	teams, err := h.teamRepo.FindByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
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

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "team_not_found",
				"message": "team not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	_, err = h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_team_member",
				"message": "you are not a member of this team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
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

	c.JSON(http.StatusOK, gin.H{
		"team":    team,
		"members": members,
	})
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
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

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_team_name",
			"message": "team name is required",
		})
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_team_member",
				"message": "you are not a member of this team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if member.Role != models.RoleCaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "only the captain can update the team",
		})
		return
	}

	if err := h.teamRepo.UpdateName(teamID, req.Name); err != nil {
		if errors.Is(err, models.ErrInvalidTeamName) || errors.Is(err, models.ErrTeamNameTooLong) || errors.Is(err, models.ErrTeamNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_team_name",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "team_name_exists",
				"message": "team name already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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

	team, err := h.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "team_not_found",
				"message": "team not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_team_member",
				"message": "you are not a member of this team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if member.Role != models.RoleCaptain {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "only the captain can delete the team",
		})
		return
	}

	var req struct {
		ConfirmWithName string `json:"confirm_with_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.ConfirmWithName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "confirmation_name_mismatch",
			"message": "you must provide the team name to confirm deletion",
		})
		return
	}

	if req.ConfirmWithName != team.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "confirmation_name_mismatch",
			"message": "team name does not match, confirmation required",
		})
		return
	}

	if err := convertTeamZonesToTemplates(teamID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if err := h.teamRepo.Delete(teamID); err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "team_not_found",
				"message": "team not found",
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
		"message": "team deleted",
	})
}

func (h *TeamHandler) LeaveTeam(c *gin.Context) {
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

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		if errors.Is(err, models.ErrMemberNotFound) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_team_member",
				"message": "you are not a member of this team",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if member.Role == models.RoleCaptain {
		if err := convertTeamZonesToTemplates(teamID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "internal server error",
			})
			return
		}

		if err := h.teamRepo.Delete(teamID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "internal server error",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "team deleted (captain left without transfer)",
		})
		return
	}

	if err := h.teamMemberRepo.RemoveMember(teamID, user.ID); err != nil {
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

	c.JSON(http.StatusOK, gin.H{
		"message": "left team",
	})
}

func convertTeamZonesToTemplates(teamID int) error {
	_ = teamID
	return nil
}
