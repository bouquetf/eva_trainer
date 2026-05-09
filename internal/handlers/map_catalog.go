package handlers

import (
	"net/http"
	"strconv"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type MapCatalogHandler struct {
	repo models.MapCatalogRepository
}

func NewMapCatalogHandler(repo models.MapCatalogRepository) *MapCatalogHandler {
	return &MapCatalogHandler{repo: repo}
}

func (h *MapCatalogHandler) List(c *gin.Context) {
	entries, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list map catalog",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *MapCatalogHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_body",
			"message": "name is required",
		})
		return
	}

	entry, err := h.repo.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to create map catalog entry",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entry": entry})
}

func (h *MapCatalogHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "invalid id",
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request_body",
			"message": "name is required",
		})
		return
	}

	entry, err := h.repo.Update(id, req.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "entry_not_found",
			"message": "map catalog entry not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *MapCatalogHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "invalid id",
		})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "entry_not_found",
			"message": "map catalog entry not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "entry deleted"})
}
