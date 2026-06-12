package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type AdminLigueHandler struct {
	ligueRepo       models.LigueRepository
	ligueLocaleRepo models.LigueLocaleRepository
	logger          services.Logger
}

func NewAdminLigueHandler(
	ligueRepo models.LigueRepository,
	ligueLocaleRepo models.LigueLocaleRepository,
	logger services.Logger,
) *AdminLigueHandler {
	return &AdminLigueHandler{
		ligueRepo:       ligueRepo,
		ligueLocaleRepo: ligueLocaleRepo,
		logger:          logger,
	}
}

type createLigueRequest struct {
	DateDebut string `json:"date_debut" binding:"required"`
	DateFin   string `json:"date_fin" binding:"required"`
	SplitID   int    `json:"split_id" binding:"required"`
	Nom       string `json:"nom" binding:"required"`
}

type updateLigueRequest struct {
	DateDebut string `json:"date_debut" binding:"required"`
	DateFin   string `json:"date_fin" binding:"required"`
	SplitID   int    `json:"split_id" binding:"required"`
	Nom       string `json:"nom" binding:"required"`
}

func parseLigueDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func (h *AdminLigueHandler) CreateLigue(c *gin.Context) {
	var req createLigueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	dateDebut, err := parseLigueDate(req.DateDebut)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	dateFin, err := parseLigueDate(req.DateFin)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	ligue, err := h.ligueRepo.Create(dateDebut, dateFin, req.SplitID, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_name_required", "ligue name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueInvalidDates):
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_dates", "ligue start date must be before end date", http.StatusBadRequest))
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

func (h *AdminLigueHandler) ListLigues(c *gin.Context) {
	ligues, err := h.ligueRepo.FindAll()
	if err != nil {
		h.logger.Error("failed to list ligues", "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ligues": ligues})
}

func (h *AdminLigueHandler) GetLigue(c *gin.Context) {
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

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"ligue": ligue})
}

func (h *AdminLigueHandler) UpdateLigue(c *gin.Context) {
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

	var req updateLigueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	dateDebut, err := parseLigueDate(req.DateDebut)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	dateFin, err := parseLigueDate(req.DateFin)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_date_format", "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	ligue, err := h.ligueRepo.Update(id, dateDebut, dateFin, req.SplitID, req.Nom)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrLigueNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("ligue_name_required", "ligue name is required", http.StatusBadRequest))
		case errors.Is(err, models.ErrLigueInvalidDates):
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_dates", "ligue start date must be before end date", http.StatusBadRequest))
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

func (h *AdminLigueHandler) DeleteLigue(c *gin.Context) {
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
