package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type ArchiveHandler struct {
	ligueLocaleRepo     models.LigueLocaleRepository
	ligueRepo           models.LigueRepository
	splitRepo           models.SplitRepository
	evaTeamRepo         models.EvaTeamRepository
	teamRepo            models.TeamRepository
	divRepo             models.DivRepository
	boRepo              models.BORepository
	cardStatsRepo       models.CardStatsRepository
	headToHeadStatsRepo models.HeadToHeadStatsRepository
	matchRepo           models.MatchRepository
	banRepo             models.BanRepository
	pickRepo            models.PickRepository
	mapRepo             models.MapRepository
	divNoteRepo         models.DivNoteRepository
	logger              services.Logger
}

func NewArchiveHandler(
	ligueLocaleRepo models.LigueLocaleRepository,
	ligueRepo models.LigueRepository,
	splitRepo models.SplitRepository,
	evaTeamRepo models.EvaTeamRepository,
	teamRepo models.TeamRepository,
	divRepo models.DivRepository,
	boRepo models.BORepository,
	cardStatsRepo models.CardStatsRepository,
	headToHeadStatsRepo models.HeadToHeadStatsRepository,
	matchRepo models.MatchRepository,
	banRepo models.BanRepository,
	pickRepo models.PickRepository,
	mapRepo models.MapRepository,
	divNoteRepo models.DivNoteRepository,
	logger services.Logger,
) *ArchiveHandler {
	return &ArchiveHandler{
		ligueLocaleRepo:     ligueLocaleRepo,
		ligueRepo:           ligueRepo,
		splitRepo:           splitRepo,
		evaTeamRepo:         evaTeamRepo,
		teamRepo:            teamRepo,
		divRepo:             divRepo,
		boRepo:              boRepo,
		cardStatsRepo:       cardStatsRepo,
		headToHeadStatsRepo: headToHeadStatsRepo,
		matchRepo:           matchRepo,
		banRepo:             banRepo,
		pickRepo:            pickRepo,
		mapRepo:             mapRepo,
		divNoteRepo:         divNoteRepo,
		logger:              logger,
	}
}

type createLigueLocaleRequest struct {
	Nom         string `json:"nom" binding:"required"`
	Emplacement string `json:"emplacement" binding:"required"`
}

type updateLigueLocaleRequest struct {
	Nom         string `json:"nom" binding:"required"`
	Emplacement string `json:"emplacement" binding:"required"`
}

func (h *ArchiveHandler) ListLiguesLocales(c *gin.Context) {
	emplacement := c.Query("emplacement")

	locales, err := h.ligueLocaleRepo.FindAll(emplacement)
	if err != nil {
		h.logger.Error("failed to list ligues locales", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ligues_locales": locales})
}

type ligueDetail struct {
	ID        int    `json:"id"`
	Nom       string `json:"nom"`
	DateDebut string `json:"date_debut"`
	DateFin   string `json:"date_fin"`
	SplitName string `json:"split_name"`
	SplitID   int    `json:"split_id"`
}

func (h *ArchiveHandler) GetLigueLocale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_locale_id", "invalid ligue locale id", http.StatusBadRequest))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if perPage < 1 {
		perPage = 5
	}

	locale, err := h.ligueLocaleRepo.FindByID(id)
	if errors.Is(err, models.ErrLigueLocaleNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_not_found", "ligue locale not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find ligue locale", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	associatedLigues, err := h.ligueLocaleRepo.FindLiguesByLocaleID(id)
	if err != nil {
		h.logger.Error("failed to find associated ligues", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	ligueDetails := make([]ligueDetail, 0, len(associatedLigues))
	for _, l := range associatedLigues {
		splitName := ""
		split, err := h.splitRepo.FindByID(l.SplitID)
		if err == nil && split != nil {
			splitName = split.Name
		}
		ligueDetails = append(ligueDetails, ligueDetail{
			ID:        l.ID,
			Nom:       l.Nom,
			DateDebut: l.DateDebut.Format("2006-01-02"),
			DateFin:   l.DateFin.Format("2006-01-02"),
			SplitName: splitName,
			SplitID:   l.SplitID,
		})
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"ligue_locale": locale,
		"ligues":       ligueDetails,
	})
}

func (h *ArchiveHandler) CreateLigueLocale(c *gin.Context) {
	var req createLigueLocaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	locale, err := h.ligueLocaleRepo.Create(req.Nom, req.Emplacement)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueLocaleNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_name_required", "ligue locale name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueLocaleEmplacementExists):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_emplacement_exists", "a ligue locale with this emplacement already exists", http.StatusConflict))
		default:
			h.logger.Error("failed to create ligue locale", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"ligue_locale": locale})
}

func (h *ArchiveHandler) UpdateLigueLocale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_locale_id", "invalid ligue locale id", http.StatusBadRequest))
		return
	}

	var req updateLigueLocaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	locale, err := h.ligueLocaleRepo.Update(id, req.Nom, req.Emplacement)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueLocaleNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_name_required", "ligue locale name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueLocaleEmplacementExists):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_emplacement_exists", "a ligue locale with this emplacement already exists", http.StatusConflict))
		case errors.Is(err, models.ErrLigueLocaleNotFound):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_not_found", "ligue locale not found", http.StatusNotFound))
		default:
			h.logger.Error("failed to update ligue locale", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ligue_locale": locale})
}

func (h *ArchiveHandler) DeleteLigueLocale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_locale_id", "invalid ligue locale id", http.StatusBadRequest))
		return
	}

	err = h.ligueLocaleRepo.Delete(id)
	if errors.Is(err, models.ErrLigueLocaleNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_not_found", "ligue locale not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete ligue locale", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "ligue locale deleted"})
}

type ligueListItem struct {
	ID        int    `json:"id"`
	Nom       string `json:"nom"`
	DateDebut string `json:"date_debut"`
	DateFin   string `json:"date_fin"`
	SplitName string `json:"split_name"`
	SplitID   int    `json:"split_id"`
	NbLocales int    `json:"nb_locales"`
}

func (h *ArchiveHandler) ListLigues(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if perPage < 1 {
		perPage = 5
	}

	ligues, total, err := h.ligueRepo.FindAllPaginated(page, perPage)
	if err != nil {
		h.logger.Error("failed to list ligues", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	items := make([]ligueListItem, 0, len(ligues))
	for _, l := range ligues {
		splitName := ""
		split, err := h.splitRepo.FindByID(l.SplitID)
		if err == nil && split != nil {
			splitName = split.Name
		}

		_, nbLocales, _ := h.ligueLocaleRepo.FindByLigueID(l.ID, 1, 1)

		items = append(items, ligueListItem{
			ID:        l.ID,
			Nom:       l.Nom,
			DateDebut: l.DateDebut.Format("2006-01-02"),
			DateFin:   l.DateFin.Format("2006-01-02"),
			SplitName: splitName,
			SplitID:   l.SplitID,
			NbLocales: nbLocales,
		})
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

type createArchiveLigueRequest struct {
	DateDebut string `json:"date_debut" binding:"required"`
	DateFin   string `json:"date_fin" binding:"required"`
	SplitID   int    `json:"split_id" binding:"required"`
	Nom       string `json:"nom" binding:"required"`
}

func (h *ArchiveHandler) CreateLigue(c *gin.Context) {
	var req createArchiveLigueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	dateDebut, err := time.Parse("2006-01-02", req.DateDebut)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "date must be in YYYY-MM-DD format", http.StatusBadRequest))
		return
	}
	dateFin, err := time.Parse("2006-01-02", req.DateFin)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "date must be in YYYY-MM-DD format", http.StatusBadRequest))
		return
	}

	ligue, err := h.ligueRepo.Create(dateDebut, dateFin, req.SplitID, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_name_required", "ligue name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueInvalidDates):
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_dates", "start date must be before end date", http.StatusBadRequest))
		case strings.Contains(err.Error(), "foreign key constraint"):
			apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusBadRequest))
		default:
			h.logger.Error("failed to create ligue", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	associated, err := h.ligueRepo.AssociateAllLocales(ligue.ID)
	if err != nil {
		h.logger.Error("failed to associate locales", "error", err, "ligue_id", ligue.ID)
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"ligue":              ligue,
		"locales_associated": associated,
	})
}

func (h *ArchiveHandler) RefreshLigue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue id", http.StatusBadRequest))
		return
	}

	_, err = h.ligueRepo.FindByID(id)
	if errors.Is(err, models.ErrLigueNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	associated, err := h.ligueRepo.AssociateAllLocales(id)
	if err != nil {
		h.logger.Error("failed to refresh ligue associations", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "ligue refreshed", "associated": associated})
}

type updateArchiveLigueRequest struct {
	DateDebut string `json:"date_debut" binding:"required"`
	DateFin   string `json:"date_fin" binding:"required"`
	SplitID   int    `json:"split_id" binding:"required"`
	Nom       string `json:"nom" binding:"required"`
}

func (h *ArchiveHandler) UpdateLigue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue id", http.StatusBadRequest))
		return
	}

	var req updateArchiveLigueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	dateDebut, err := time.Parse("2006-01-02", req.DateDebut)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "date must be in YYYY-MM-DD format", http.StatusBadRequest))
		return
	}

	dateFin, err := time.Parse("2006-01-02", req.DateFin)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "date must be in YYYY-MM-DD format", http.StatusBadRequest))
		return
	}

	ligue, err := h.ligueRepo.Update(id, dateDebut, dateFin, req.SplitID, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_name_required", "ligue name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueInvalidDates):
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_dates", "start date must be before end date", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueNotFound):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusNotFound))
		case strings.Contains(err.Error(), "foreign key constraint"):
			apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusBadRequest))
		default:
			h.logger.Error("failed to update ligue", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ligue": ligue})
}

func (h *ArchiveHandler) DeleteLigue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue id", http.StatusBadRequest))
		return
	}

	err = h.ligueRepo.Delete(id)
	if errors.Is(err, models.ErrLigueNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "ligue deleted"})
}

type importLigueRequest struct {
	LigueID int `json:"ligue_id" binding:"required"`
}

func (h *ArchiveHandler) ImportLigue(c *gin.Context) {
	localeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_locale_id", "invalid ligue locale id", http.StatusBadRequest))
		return
	}

	var req importLigueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	_, err = h.ligueLocaleRepo.FindByID(localeID)
	if errors.Is(err, models.ErrLigueLocaleNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_locale_not_found", "ligue locale not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find ligue locale", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.ligueRepo.FindByID(req.LigueID)
	if errors.Is(err, models.ErrLigueNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	if err := h.ligueLocaleRepo.AssociateLigue(localeID, req.LigueID); err != nil {
		h.logger.Error("failed to import ligue for locale", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "ligue imported"})
}

type createEvaTeamRequest struct {
	Name             string `json:"name" binding:"required"`
	Tag              string `json:"tag" binding:"required"`
	EvaTrainerTeamID *int   `json:"eva_trainer_team_id,omitempty"`
}

type updateEvaTeamRequest struct {
	Name             string `json:"name" binding:"required"`
	Tag              string `json:"tag" binding:"required"`
	EvaTrainerTeamID *int   `json:"eva_trainer_team_id,omitempty"`
}

func (h *ArchiveHandler) ListEvaTeams(c *gin.Context) {
	teams, err := h.evaTeamRepo.FindAll()
	if err != nil {
		h.logger.Error("failed to list eva teams", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"teams": teams})
}

func (h *ArchiveHandler) GetEvaTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_eva_team_id", "invalid eva team id", http.StatusBadRequest))
		return
	}

	team, err := h.evaTeamRepo.FindByID(id)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find eva team", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"team": team})
}

func (h *ArchiveHandler) CreateEvaTeam(c *gin.Context) {
	var req createEvaTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if req.EvaTrainerTeamID != nil {
		_, err := h.teamRepo.FindByID(*req.EvaTrainerTeamID)
		if errors.Is(err, models.ErrTeamNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("eva_trainer_team_not_found", "Équipe eva-trainer non trouvée", http.StatusNotFound))
			return
		}
		if err != nil {
			h.logger.Error("failed to verify trainer team", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}
	}

	team, err := h.evaTeamRepo.Create(req.Name, req.Tag, req.EvaTrainerTeamID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEvaTeamNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_name_required", "Le nom est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrEvaTeamTagRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_tag_required", "Le tag est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrEvaTeamTagTooLong):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_tag_too_long", "Le tag ne peut pas dépasser 50 caractères", http.StatusBadRequest))
		case strings.Contains(err.Error(), "23505"):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_duplicate_association", "Cette équipe eva-trainer est déjà associée à une autre équipe eva", http.StatusConflict))
		default:
			h.logger.Error("failed to create eva team", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"team": team})
}

func (h *ArchiveHandler) UpdateEvaTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_eva_team_id", "invalid eva team id", http.StatusBadRequest))
		return
	}

	var req updateEvaTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if req.EvaTrainerTeamID != nil {
		_, err := h.teamRepo.FindByID(*req.EvaTrainerTeamID)
		if errors.Is(err, models.ErrTeamNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("eva_trainer_team_not_found", "Équipe eva-trainer non trouvée", http.StatusNotFound))
			return
		}
		if err != nil {
			h.logger.Error("failed to verify trainer team", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}
	}

	team, err := h.evaTeamRepo.Update(id, req.Name, req.Tag, req.EvaTrainerTeamID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEvaTeamNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_name_required", "Le nom est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrEvaTeamTagRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_tag_required", "Le tag est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrEvaTeamTagTooLong):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_tag_too_long", "Le tag ne peut pas dépasser 50 caractères", http.StatusBadRequest))
		case errors.Is(err, models.ErrEvaTeamNotFound):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
		case strings.Contains(err.Error(), "23505"):
			apierror.RespondWithError(c, apierror.NewAPIError("eva_team_duplicate_association", "Cette équipe eva-trainer est déjà associée à une autre équipe eva", http.StatusConflict))
		default:
			h.logger.Error("failed to update eva team", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"team": team})
}

func (h *ArchiveHandler) DeleteEvaTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_eva_team_id", "invalid eva team id", http.StatusBadRequest))
		return
	}

	hasRefs, err := h.evaTeamRepo.HasReferences(id)
	if err != nil {
		h.logger.Error("failed to check eva team references", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}
	if hasRefs {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_has_references", "Impossible de supprimer cette équipe eva : elle est référencée par des ligues ou matchs", http.StatusConflict))
		return
	}

	err = h.evaTeamRepo.Delete(id)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete eva team", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Équipe eva supprimée"})
}

type createDivRequest struct {
	LigueID int    `json:"ligue_id" binding:"required"`
	Nom     string `json:"nom" binding:"required"`
}

type updateDivRequest struct {
	Nom string `json:"nom" binding:"required"`
}

type assignTeamRequest struct {
	EquipeEvaID int `json:"equipe_eva_id" binding:"required"`
}

type upsertDivNoteRequest struct {
	ContentMD string `json:"content_md"`
}

type divWithTeams struct {
	ID        int               `json:"id"`
	LigueID   int               `json:"ligue_id"`
	Nom       string            `json:"nom"`
	ReplayURL string            `json:"replay_url"`
	Teams     []*models.EvaTeam `json:"teams"`
}

func (h *ArchiveHandler) ListDivs(c *gin.Context) {
	ligueIDStr := c.Query("ligue_id")
	if ligueIDStr == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_ligue_id", "ligue_id query parameter is required", http.StatusBadRequest))
		return
	}
	ligueID, err := strconv.Atoi(ligueIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue_id", http.StatusBadRequest))
		return
	}

	divs, err := h.divRepo.FindByLigueID(ligueID)
	if err != nil {
		h.logger.Error("failed to list divs", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	teamsByDiv, err := h.divRepo.FindTeamsByLigueID(ligueID)
	if err != nil {
		h.logger.Error("failed to list teams by ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	result := make([]divWithTeams, 0, len(divs))
	for _, d := range divs {
		teams := teamsByDiv[d.ID]
		if teams == nil {
			teams = []*models.EvaTeam{}
		}
		result = append(result, divWithTeams{
			ID:        d.ID,
			LigueID:   d.LigueID,
			Nom:       d.Nom,
			ReplayURL: d.ReplayURL,
			Teams:     teams,
		})
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"divs": result})
}

func (h *ArchiveHandler) GetDiv(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	div, err := h.divRepo.FindByID(id)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	teams, err := h.divRepo.FindTeamsByDivID(id)
	if err != nil {
		h.logger.Error("failed to find teams for div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"div":   div,
		"teams": teams,
	})
}

func (h *ArchiveHandler) CreateDiv(c *gin.Context) {
	var req createDivRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	div, err := h.divRepo.Create(req.LigueID, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDivNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("div_name_required", "Le nom de la div est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrDivNameExists):
			apierror.RespondWithError(c, apierror.NewAPIError("div_name_exists", "Une div avec ce nom existe déjà dans cette ligue", http.StatusConflict))
		default:
			h.logger.Error("failed to create div", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"div": div})
}

func (h *ArchiveHandler) UpdateDiv(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	var req updateDivRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	div, err := h.divRepo.Update(id, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDivNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("div_name_required", "Le nom de la div est requis", http.StatusBadRequest))
		case errors.Is(err, models.ErrDivNotFound):
			apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		case errors.Is(err, models.ErrDivNameExists):
			apierror.RespondWithError(c, apierror.NewAPIError("div_name_exists", "Une div avec ce nom existe déjà dans cette ligue", http.StatusConflict))
		default:
			h.logger.Error("failed to update div", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"div": div})
}

func (h *ArchiveHandler) DeleteDiv(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	err = h.divRepo.Delete(id)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Div supprimée"})
}

func (h *ArchiveHandler) AssignTeamToDiv(c *gin.Context) {
	divID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	var req assignTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	_, err = h.divRepo.FindByID(divID)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.evaTeamRepo.FindByID(req.EquipeEvaID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find eva team", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	err = h.divRepo.AssignTeam(divID, req.EquipeEvaID)
	if errors.Is(err, models.ErrDivTeamAlreadyAssigned) {
		apierror.RespondWithError(c, apierror.NewAPIError("team_already_assigned", "Cette équipe est déjà assignée à une autre div dans cette ligue", http.StatusConflict))
		return
	}
	if err != nil {
		h.logger.Error("failed to assign team to div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Équipe assignée à la div"})
}

func (h *ArchiveHandler) RemoveTeamFromDiv(c *gin.Context) {
	divID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("teamId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	err = h.divRepo.RemoveTeam(divID, teamID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("team_not_in_div", "Cette équipe n'est pas dans cette div", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to remove team from div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Équipe retirée de la div"})
}

func (h *ArchiveHandler) GetDivStandings(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	_, err = h.divRepo.FindByID(id)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	standings, err := h.divRepo.GetStandings(id)
	if err != nil {
		h.logger.Error("failed to get standings", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"standings": standings})
}

func (h *ArchiveHandler) SetDivReplayURL(c *gin.Context) {
	divID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	var req struct {
		ReplayURL string `json:"replay_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if !models.IsValidReplayURL(req.ReplayURL) {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_replay_url", "invalid replay URL: only YouTube and Twitch URLs are allowed", http.StatusBadRequest))
		return
	}

	div, err := h.divRepo.SetReplayURL(divID, req.ReplayURL)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to set replay url", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"div": div})
}

func (h *ArchiveHandler) DeleteDivReplayURL(c *gin.Context) {
	divID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	div, err := h.divRepo.SetReplayURL(divID, "")
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete replay url", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"div": div})
}

func (h *ArchiveHandler) GenerateRoundRobin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
		return
	}

	_, err = h.divRepo.FindByID(id)
	if errors.Is(err, models.ErrDivNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find div", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	bos, err := h.divRepo.GenerateRoundRobin(id)
	if errors.Is(err, models.ErrDivNotEnoughTeams) {
		apierror.RespondWithError(c, apierror.NewAPIError("not_enough_teams", "La division doit avoir au moins 2 équipes pour générer un calendrier", http.StatusBadRequest))
		return
	}
	if err != nil {
		h.logger.Error("failed to generate round-robin", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"bos": bos})
}

type createBORequest struct {
	LigueID   int               `json:"ligue_id" binding:"required"`
	Equipe1ID int               `json:"equipe1_id" binding:"required"`
	Equipe2ID int               `json:"equipe2_id" binding:"required"`
	Format    string            `json:"format" binding:"required"`
	Forfait   int               `json:"forfait"`
	Bans      []createBanItem   `json:"bans"`
	Picks     []createPickItem  `json:"picks"`
	Matchs    []createMatchItem `json:"matchs"`
}

type createBanItem struct {
	EquipeID int `json:"equipe_id" binding:"required"`
	MapID    int `json:"map_id" binding:"required"`
	Ordre    int `json:"ordre" binding:"required"`
}

type createPickItem struct {
	EquipeID *int `json:"equipe_id"`
	MapID    int  `json:"map_id" binding:"required"`
	Ordre    int  `json:"ordre" binding:"required"`
}

type createMatchItem struct {
	MapID           int     `json:"map_id" binding:"required"`
	ScoreEquipe1    int     `json:"score_equipe1"`
	ScoreEquipe2    int     `json:"score_equipe2"`
	Forfait         int     `json:"forfait"`
	Commentaires    *string `json:"commentaires"`
	HorsCompetition bool    `json:"hors_competition"`
}

type updateMatchRequest struct {
	HorsCompetition *bool `json:"hors_competition"`
}

type updateBORequest struct {
	Format  string            `json:"format" binding:"required"`
	Forfait int               `json:"forfait"`
	Bans    []createBanItem   `json:"bans"`
	Picks   []createPickItem  `json:"picks"`
	Matchs  []createMatchItem `json:"matchs"`
}

func (h *ArchiveHandler) GetLigueMaps(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue id", http.StatusBadRequest))
		return
	}

	ligue, err := h.ligueRepo.FindByID(id)
	if errors.Is(err, models.ErrLigueNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	maps, err := h.splitRepo.FindMapsBySplitID(ligue.SplitID)
	if err != nil {
		h.logger.Error("failed to find maps by split", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"maps": maps})
}

func (h *ArchiveHandler) CreateBO(c *gin.Context) {
	var req createBORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if req.Equipe1ID == req.Equipe2ID {
		apierror.RespondWithError(c, apierror.NewAPIError("same_team", "equipe1 and equipe2 must be different", http.StatusBadRequest))
		return
	}
	if req.Format != "BO3" && req.Format != "BO5" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_format", "format must be BO3 or BO5", http.StatusBadRequest))
		return
	}
	if req.Forfait < 0 || req.Forfait > 2 {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_forfait", "forfait must be 0, 1, or 2", http.StatusBadRequest))
		return
	}

	_, err := h.ligueRepo.FindByID(req.LigueID)
	if errors.Is(err, models.ErrLigueNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("ligue_not_found", "ligue not found", http.StatusBadRequest))
		return
	}
	if err != nil {
		h.logger.Error("failed to verify ligue", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.evaTeamRepo.FindByID(req.Equipe1ID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("equipe1_not_found", "equipe1 not found", http.StatusBadRequest))
		return
	}
	if err != nil {
		h.logger.Error("failed to verify equipe1", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.evaTeamRepo.FindByID(req.Equipe2ID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("equipe2_not_found", "equipe2 not found", http.StatusBadRequest))
		return
	}
	if err != nil {
		h.logger.Error("failed to verify equipe2", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	if req.Forfait == 0 {
		for _, m := range req.Matchs {
			if m.Forfait == 0 {
				if m.ScoreEquipe1 < 0 || m.ScoreEquipe1 > 100 || m.ScoreEquipe2 < 0 || m.ScoreEquipe2 > 100 {
					apierror.RespondWithError(c, apierror.NewAPIError("invalid_score", "scores must be between 0 and 100", http.StatusBadRequest))
					return
				}
			}
		}

		mapIDs := make(map[int]bool)
		for _, b := range req.Bans {
			mapIDs[b.MapID] = true
		}
		for _, p := range req.Picks {
			mapIDs[p.MapID] = true
		}
		for _, m := range req.Matchs {
			mapIDs[m.MapID] = true
		}
		for mid := range mapIDs {
			_, err := h.mapRepo.FindByID(mid)
			if errors.Is(err, models.ErrMapNotFound) {
				apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", fmt.Sprintf("map %d not found", mid), http.StatusBadRequest))
				return
			}
			if err != nil {
				h.logger.Error("failed to verify map", "error", err, "map_id", mid)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}
	}

	bo, err := h.boRepo.Create(req.LigueID, req.Equipe1ID, req.Equipe2ID, req.Format, req.Forfait)
	if err != nil {
		h.logger.Error("failed to create BO", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	if req.Forfait == 0 {
		for _, b := range req.Bans {
			_, err := h.banRepo.Create(bo.ID, b.EquipeID, b.MapID)
			if err != nil {
				h.logger.Error("failed to create ban", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}

		for _, p := range req.Picks {
			equipeID := 0
			if p.EquipeID != nil {
				equipeID = *p.EquipeID
			}
			_, err := h.pickRepo.Create(bo.ID, equipeID, p.MapID)
			if err != nil {
				h.logger.Error("failed to create pick", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}

		for _, m := range req.Matchs {
			_, err := h.matchRepo.Create(bo.ID, m.MapID, m.ScoreEquipe1, m.ScoreEquipe2, m.Forfait, m.Commentaires, m.HorsCompetition)
			if err != nil {
				h.logger.Error("failed to create match", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}
	}

	bo, err = h.boRepo.RecomputeWinner(bo.ID)
	if err != nil {
		h.logger.Error("failed to recompute BO winner", "error", err)
	}

	details, err := h.boRepo.FindWithDetails(bo.ID)
	if err != nil {
		h.logger.Error("failed to find BO details", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"bo": details})
}

func (h *ArchiveHandler) GetBO(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_bo_id", "invalid bo id", http.StatusBadRequest))
		return
	}

	details, err := h.boRepo.FindWithDetails(id)
	if errors.Is(err, models.ErrBONotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("bo_not_found", "BO not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find BO", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"bo": details})
}

func (h *ArchiveHandler) ListBOs(c *gin.Context) {
	ligueIDStr := c.Query("ligue_id")
	if ligueIDStr == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_ligue_id", "ligue_id query parameter is required", http.StatusBadRequest))
		return
	}
	ligueID, err := strconv.Atoi(ligueIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue_id", http.StatusBadRequest))
		return
	}

	details := c.Query("details") == "true"

	if details {
		bos, err := h.boRepo.FindByLigueID(ligueID)
		if err != nil {
			h.logger.Error("failed to list BOs", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}

		result := make([]*models.BOWithDetails, 0, len(bos))
		for _, bo := range bos {
			d, err := h.boRepo.FindWithDetails(bo.ID)
			if err != nil {
				h.logger.Error("failed to find BO details", "error", err, "bo_id", bo.ID)
				continue
			}
			result = append(result, d)
		}

		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"bos": result, "total": len(result)})
		return
	}

	items, err := h.boRepo.FindByLigueIDWithTeams(ligueID)
	if err != nil {
		h.logger.Error("failed to list BOs with teams", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"bos": items, "total": len(items)})
}

func (h *ArchiveHandler) UpdateBO(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_bo_id", "invalid bo id", http.StatusBadRequest))
		return
	}

	existing, err := h.boRepo.FindByID(id)
	if errors.Is(err, models.ErrBONotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("bo_not_found", "BO not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find BO", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	var req updateBORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if req.Format != "BO3" && req.Format != "BO5" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_format", "format must be BO3 or BO5", http.StatusBadRequest))
		return
	}
	if req.Forfait < 0 || req.Forfait > 2 {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_forfait", "forfait must be 0, 1, or 2", http.StatusBadRequest))
		return
	}

	if req.Forfait == 0 {
		for _, m := range req.Matchs {
			if m.Forfait == 0 {
				if m.ScoreEquipe1 < 0 || m.ScoreEquipe1 > 100 || m.ScoreEquipe2 < 0 || m.ScoreEquipe2 > 100 {
					apierror.RespondWithError(c, apierror.NewAPIError("invalid_score", "scores must be between 0 and 100", http.StatusBadRequest))
					return
				}
			}
		}

		mapIDs := make(map[int]bool)
		for _, b := range req.Bans {
			mapIDs[b.MapID] = true
		}
		for _, p := range req.Picks {
			mapIDs[p.MapID] = true
		}
		for _, m := range req.Matchs {
			mapIDs[m.MapID] = true
		}
		for mid := range mapIDs {
			_, err := h.mapRepo.FindByID(mid)
			if errors.Is(err, models.ErrMapNotFound) {
				apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", fmt.Sprintf("map %d not found", mid), http.StatusBadRequest))
				return
			}
			if err != nil {
				h.logger.Error("failed to verify map", "error", err, "map_id", mid)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}
	}

	_, err = h.boRepo.Update(id, existing.LigueID, existing.Equipe1ID, existing.Equipe2ID, req.Format, req.Forfait)
	if err != nil {
		h.logger.Error("failed to update BO", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	if err := h.matchRepo.DeleteByBOID(id); err != nil {
		h.logger.Error("failed to delete existing matchs", "error", fmt.Sprintf("%+v", err))
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}
	if err := h.banRepo.DeleteByBOID(id); err != nil {
		h.logger.Error("failed to delete existing bans", "error", fmt.Sprintf("%+v", err))
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}
	if err := h.pickRepo.DeleteByBOID(id); err != nil {
		h.logger.Error("failed to delete existing picks", "error", fmt.Sprintf("%+v", err))
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	if req.Forfait == 0 {
		for _, b := range req.Bans {
			_, err := h.banRepo.Create(id, b.EquipeID, b.MapID)
			if err != nil {
				h.logger.Error("failed to create ban", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}

		for _, p := range req.Picks {
			equipeID := 0
			if p.EquipeID != nil {
				equipeID = *p.EquipeID
			}
			_, err := h.pickRepo.Create(id, equipeID, p.MapID)
			if err != nil {
				h.logger.Error("failed to create pick", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}

		for _, m := range req.Matchs {
			_, err := h.matchRepo.Create(id, m.MapID, m.ScoreEquipe1, m.ScoreEquipe2, m.Forfait, m.Commentaires, m.HorsCompetition)
			if err != nil {
				h.logger.Error("failed to create match", "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
				return
			}
		}
	}

	_, err = h.boRepo.RecomputeWinner(id)
	if err != nil {
		h.logger.Error("failed to recompute BO winner", "error", err)
	}

	details, err := h.boRepo.FindWithDetails(id)
	if err != nil {
		h.logger.Error("failed to find BO details", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"bo": details})
}

func (h *ArchiveHandler) DeleteBO(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_bo_id", "invalid bo id", http.StatusBadRequest))
		return
	}

	err = h.boRepo.Delete(id)
	if errors.Is(err, models.ErrBONotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("bo_not_found", "BO not found", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to delete BO", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "BO supprimé"})
}

func (h *ArchiveHandler) GetCardStats(c *gin.Context) {
	teamIDStr := c.Query("team_id")
	if teamIDStr == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_team_id", "team_id query parameter is required", http.StatusBadRequest))
		return
	}
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team_id", http.StatusBadRequest))
		return
	}

	_, err = h.evaTeamRepo.FindByID(teamID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find eva team", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	stats, err := h.cardStatsRepo.Compute(teamID)
	if err != nil {
		h.logger.Error("failed to compute card stats", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"stats": stats})
}

func (h *ArchiveHandler) GetHeadToHeadStats(c *gin.Context) {
	team1IDStr := c.Query("team1_id")
	team2IDStr := c.Query("team2_id")

	if team1IDStr == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_team1_id", "team1_id query parameter is required", http.StatusBadRequest))
		return
	}
	if team2IDStr == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_team2_id", "team2_id query parameter is required", http.StatusBadRequest))
		return
	}

	team1ID, err := strconv.Atoi(team1IDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team1_id", "invalid team1_id", http.StatusBadRequest))
		return
	}
	team2ID, err := strconv.Atoi(team2IDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team2_id", "invalid team2_id", http.StatusBadRequest))
		return
	}

	if team1ID == team2ID {
		apierror.RespondWithError(c, apierror.NewAPIError("same_team", "team1 and team2 must be different", http.StatusBadRequest))
		return
	}

	_, err = h.evaTeamRepo.FindByID(team1ID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva team1 non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find eva team1", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	_, err = h.evaTeamRepo.FindByID(team2ID)
	if errors.Is(err, models.ErrEvaTeamNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva team2 non trouvée", http.StatusNotFound))
		return
	}
	if err != nil {
		h.logger.Error("failed to find eva team2", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	stats, err := h.headToHeadStatsRepo.Compute(team1ID, team2ID)
	if err != nil {
		h.logger.Error("failed to compute head-to-head stats", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"team1_id": team1ID,
		"team2_id": team2ID,
		"stats":    stats,
	})
}

func (h *ArchiveHandler) GetMyDivNote(c *gin.Context) {
	divIDStr := c.Param("id")
	divID, err := strconv.Atoi(divIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid division id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	note, err := h.divNoteRepo.FindByDivAndUser(divID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrDivNoteNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("note_not_found", "note not found", http.StatusNotFound))
			return
		}
		h.logger.Error("failed to get div note", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"note": note})
}

func (h *ArchiveHandler) UpsertMyDivNote(c *gin.Context) {
	divIDStr := c.Param("id")
	divID, err := strconv.Atoi(divIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid division id", http.StatusBadRequest))
		return
	}

	var req upsertDivNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_body", "invalid request body", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	note, err := h.divNoteRepo.Upsert(divID, user.ID, req.ContentMD)
	if err != nil {
		h.logger.Error("failed to upsert div note", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"note": note})
}

func (h *ArchiveHandler) DeleteMyDivNote(c *gin.Context) {
	divIDStr := c.Param("id")
	divID, err := strconv.Atoi(divIDStr)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid division id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	err = h.divNoteRepo.DeleteByDivAndUser(divID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrDivNoteNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("note_not_found", "note not found", http.StatusNotFound))
			return
		}
		h.logger.Error("failed to delete div note", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Note supprimée"})
}

func (h *ArchiveHandler) UpdateMatch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_match_id", "invalid match id", http.StatusBadRequest))
		return
	}

	var req updateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	var match *models.Match

	if req.HorsCompetition != nil {
		match, err = h.matchRepo.SetHorsCompetition(id, *req.HorsCompetition)
		if errors.Is(err, models.ErrMatchNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("match_not_found", "match not found", http.StatusNotFound))
			return
		}
		if err != nil {
			h.logger.Error("failed to set hors_competition", "error", err)
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
			return
		}
	}

	if match == nil {
		apierror.RespondWithError(c, apierror.NewAPIError("nothing_to_update", "no fields to update", http.StatusBadRequest))
		return
	}

	_, recErr := h.boRepo.RecomputeWinner(match.BoID)
	if recErr != nil {
		h.logger.Error("failed to recompute BO winner after match update", "error", recErr)
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"match": match})
}
