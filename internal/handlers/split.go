package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type SplitHandler struct {
	splitRepo models.SplitRepository
	logger    services.Logger
}

func NewSplitHandler(splitRepo models.SplitRepository, logger services.Logger) *SplitHandler {
	return &SplitHandler{splitRepo: splitRepo, logger: logger}
}

type createSplitRequest struct {
	Name     string   `json:"name" binding:"required"`
	MapNames []string `json:"map_names" binding:"required"`
}

func (h *SplitHandler) CreateSplit(c *gin.Context) {
	var req createSplitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	split, err := h.splitRepo.Create(req.Name, req.MapNames)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrSplitNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("split_name_required", err.Error(), http.StatusBadRequest))
		case errors.Is(err, models.ErrSplitMapNamesRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("split_map_names_required", err.Error(), http.StatusBadRequest))
		default:
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to create split", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"split": split})
}

func (h *SplitHandler) ListSplits(c *gin.Context) {
	splits, err := h.splitRepo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list splits", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"splits": splits})
}

func (h *SplitHandler) GetSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_split_id", "invalid split id", http.StatusBadRequest))
		return
	}

	split, err := h.splitRepo.FindByID(id)
	if err == models.ErrSplitNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find split", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"split": split})
}

func (h *SplitHandler) UpdateSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_split_id", "invalid split id", http.StatusBadRequest))
		return
	}

	var req createSplitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	split, err := h.splitRepo.Update(id, req.Name, req.MapNames)
	if err != nil {
		if errors.Is(err, models.ErrSplitNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusNotFound))
			return
		}
		switch {
		case errors.Is(err, models.ErrSplitNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("split_name_required", err.Error(), http.StatusBadRequest))
		case errors.Is(err, models.ErrSplitMapNamesRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("split_map_names_required", err.Error(), http.StatusBadRequest))
		default:
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update split", http.StatusInternalServerError))
		}
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"split": split})
}

func (h *SplitHandler) DeleteSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_split_id", "invalid split id", http.StatusBadRequest))
		return
	}

	err = h.splitRepo.Delete(id)
	if err == models.ErrSplitNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to delete split", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "split deleted"})
}

func (h *SplitHandler) GetSplitMaps(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_split_id", "invalid split id", http.StatusBadRequest))
		return
	}

	maps, err := h.splitRepo.FindMapsBySplitID(id)
	if err == models.ErrSplitNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find maps for split", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"maps": maps})
}

func (h *SplitHandler) ListSplitsPublic(c *gin.Context) {
	splits, err := h.splitRepo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list splits", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"splits": splits})
}

func (h *SplitHandler) GetSplitMapsPublic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_split_id", "invalid split id", http.StatusBadRequest))
		return
	}

	maps, err := h.splitRepo.FindMapsBySplitID(id)
	if err == models.ErrSplitNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("split_not_found", "split not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find maps for split", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"maps": maps})
}
