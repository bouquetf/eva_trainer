package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type SplitHandler struct {
	splitRepo models.SplitRepository
}

func NewSplitHandler(splitRepo models.SplitRepository) *SplitHandler {
	return &SplitHandler{splitRepo: splitRepo}
}

type createSplitRequest struct {
	Name     string   `json:"name" binding:"required"`
	MapNames []string `json:"map_names" binding:"required"`
}

func (h *SplitHandler) CreateSplit(c *gin.Context) {
	var req createSplitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_body",
			"message": "invalid request body",
		})
		return
	}

	split, err := h.splitRepo.Create(req.Name, req.MapNames)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrSplitNameRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "split_name_required",
				"message": err.Error(),
			})
		case errors.Is(err, models.ErrSplitMapNamesRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "split_map_names_required",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to create split",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"split": split})
}

func (h *SplitHandler) ListSplits(c *gin.Context) {
	splits, err := h.splitRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list splits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"splits": splits})
}

func (h *SplitHandler) GetSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_split_id",
			"message": "invalid split id",
		})
		return
	}

	split, err := h.splitRepo.FindByID(id)
	if err == models.ErrSplitNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "split_not_found",
			"message": "split not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find split",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"split": split})
}

func (h *SplitHandler) UpdateSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_split_id",
			"message": "invalid split id",
		})
		return
	}

	var req createSplitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_body",
			"message": "invalid request body",
		})
		return
	}

	split, err := h.splitRepo.Update(id, req.Name, req.MapNames)
	if err != nil {
		if errors.Is(err, models.ErrSplitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "split_not_found",
				"message": "split not found",
			})
			return
		}
		switch {
		case errors.Is(err, models.ErrSplitNameRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "split_name_required",
				"message": err.Error(),
			})
		case errors.Is(err, models.ErrSplitMapNamesRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "split_map_names_required",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to update split",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"split": split})
}

func (h *SplitHandler) DeleteSplit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_split_id",
			"message": "invalid split id",
		})
		return
	}

	err = h.splitRepo.Delete(id)
	if err == models.ErrSplitNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "split_not_found",
			"message": "split not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to delete split",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "split deleted"})
}

func (h *SplitHandler) GetSplitMaps(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_split_id",
			"message": "invalid split id",
		})
		return
	}

	maps, err := h.splitRepo.FindMapsBySplitID(id)
	if err == models.ErrSplitNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "split_not_found",
			"message": "split not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find maps for split",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maps": maps})
}

func (h *SplitHandler) ListSplitsPublic(c *gin.Context) {
	splits, err := h.splitRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list splits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"splits": splits})
}

func (h *SplitHandler) GetSplitMapsPublic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_split_id",
			"message": "invalid split id",
		})
		return
	}

	maps, err := h.splitRepo.FindMapsBySplitID(id)
	if err == models.ErrSplitNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "split_not_found",
			"message": "split not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find maps for split",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maps": maps})
}
