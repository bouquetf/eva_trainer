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

type TemplateHandler struct {
	templateRepo   models.TemplateRepository
	zoneRepo       models.ZoneRepository
	teamMemberRepo models.TeamMemberRepository
	logger         services.Logger
}

func NewTemplateHandler(
	templateRepo models.TemplateRepository,
	zoneRepo models.ZoneRepository,
	teamMemberRepo models.TeamMemberRepository,
	logger services.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		templateRepo:   templateRepo,
		zoneRepo:       zoneRepo,
		teamMemberRepo: teamMemberRepo,
		logger:         logger,
	}
}

func (h *TemplateHandler) PublishMap(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid map id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	existing, err := h.templateRepo.FindByMapID(mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check template status", http.StatusInternalServerError))
		return
	}
	if existing != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("already_published", "this map is already published", http.StatusConflict))
		return
	}

	template, err := h.templateRepo.CreateFromMap(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to publish map", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"template": template})
}

func (h *TemplateHandler) GetPublishStatus(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid map id", http.StatusBadRequest))
		return
	}

	template, err := h.templateRepo.FindByMapID(mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check publish status", http.StatusInternalServerError))
		return
	}

	if template == nil {
		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"published": false})
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"published": true, "template": template})
}

func (h *TemplateHandler) AdminDeleteTemplate(c *gin.Context) {
	templateID, err := strconv.Atoi(c.Param("templateId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_template_id", "invalid template id", http.StatusBadRequest))
		return
	}

	err = h.templateRepo.Delete(templateID)
	if err != nil {
		if errors.Is(err, models.ErrTemplateNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("template_not_found", "template not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to delete template", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "template deleted"})
}
