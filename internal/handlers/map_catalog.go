package handlers

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type MapCatalogHandler struct {
	repo   models.MapCatalogRepository
	logger services.Logger
}

func NewMapCatalogHandler(repo models.MapCatalogRepository, logger services.Logger) *MapCatalogHandler {
	return &MapCatalogHandler{repo: repo, logger: logger}
}

func (h *MapCatalogHandler) List(c *gin.Context) {
	entries, err := h.repo.FindAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list map catalog", http.StatusInternalServerError))
		return
	}
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"entries": entries})
}

func (h *MapCatalogHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "name is required", http.StatusBadRequest))
		return
	}

	entry, err := h.repo.Create(req.Name)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to create map catalog entry", http.StatusInternalServerError))
		return
	}
	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"entry": entry})
}

func (h *MapCatalogHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_id", "invalid id", http.StatusBadRequest))
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "name is required", http.StatusBadRequest))
		return
	}

	entry, err := h.repo.Update(id, req.Name)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("entry_not_found", "map catalog entry not found", http.StatusNotFound))
		return
	}
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"entry": entry})
}

func (h *MapCatalogHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_id", "invalid id", http.StatusBadRequest))
		return
	}

	if err := h.repo.Delete(id); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("entry_not_found", "map catalog entry not found", http.StatusNotFound))
		return
	}
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "entry deleted"})
}
