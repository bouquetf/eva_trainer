package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	templateRepo   models.TemplateRepository
	zoneRepo       models.ZoneRepository
	teamMemberRepo models.TeamMemberRepository
}

func NewTemplateHandler(
	templateRepo models.TemplateRepository,
	zoneRepo models.ZoneRepository,
	teamMemberRepo models.TeamMemberRepository,
) *TemplateHandler {
	return &TemplateHandler{
		templateRepo:   templateRepo,
		zoneRepo:       zoneRepo,
		teamMemberRepo: teamMemberRepo,
	}
}

func (h *TemplateHandler) PublishMap(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id", "message": "invalid map id"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	existing, err := h.templateRepo.FindByMapID(mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check template status"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already_published", "message": "this map is already published"})
		return
	}

	template, err := h.templateRepo.CreateFromMap(teamID, mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to publish map"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"template": template})
}

func (h *TemplateHandler) GetPublishStatus(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id", "message": "invalid map id"})
		return
	}

	template, err := h.templateRepo.FindByMapID(mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check publish status"})
		return
	}

	if template == nil {
		c.JSON(http.StatusOK, gin.H{"published": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"published": true, "template": template})
}

func (h *TemplateHandler) AdminDeleteTemplate(c *gin.Context) {
	templateID, err := strconv.Atoi(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_template_id", "message": "invalid template id"})
		return
	}

	err = h.templateRepo.Delete(templateID)
	if err != nil {
		if errors.Is(err, models.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template_not_found", "message": "template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}
