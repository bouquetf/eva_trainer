package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type RevisionHandler struct {
	zoneRepo       models.ZoneRepository
	mapRepo        models.MapRepository
	lineOfFireRepo models.LineOfFireRepository
	templateRepo   models.TemplateRepository
	logger         services.Logger
}

func NewRevisionHandler(
	zoneRepo models.ZoneRepository,
	mapRepo models.MapRepository,
	lineOfFireRepo models.LineOfFireRepository,
	templateRepo models.TemplateRepository,
	logger services.Logger,
) *RevisionHandler {
	return &RevisionHandler{
		zoneRepo:       zoneRepo,
		mapRepo:        mapRepo,
		lineOfFireRepo: lineOfFireRepo,
		templateRepo:   templateRepo,
		logger:         logger,
	}
}

type RevisionZone struct {
	ID       int                  `json:"id"`
	Source   string               `json:"source"`
	MapID    int                  `json:"map_id"`
	Polygons []models.ZonePolygon `json:"polygons"`
	Name     string               `json:"name"`
}

func (h *RevisionHandler) ListRevisionZones(c *gin.Context) {
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

	if _, err := middleware.MustGetUser(c); err != nil {
		return
	}

	originalMap, err := h.mapRepo.FindByID(mapID)
	if err != nil {
		if errors.Is(err, models.ErrMapNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find map", http.StatusInternalServerError))
		return
	}

	allMaps, err := h.mapRepo.FindByName(originalMap.Name)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list map versions", http.StatusInternalServerError))
		return
	}

	zones := make([]RevisionZone, 0)

	for _, mv := range allMaps {
		teamZones, err := h.zoneRepo.FindWithDetailsByTeamAndMap(teamID, mv.ID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zones", http.StatusInternalServerError))
			return
		}

		for _, tz := range teamZones {
			zones = append(zones, RevisionZone{
				ID:       tz.ID,
				Source:   "team",
				MapID:    tz.MapID,
				Polygons: tz.Polygons,
				Name:     tz.Name,
			})
		}
	}

	if zones == nil {
		zones = []RevisionZone{}
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"zones": zones})
}

type RevisionLineOfFire struct {
	ID          int             `json:"id"`
	Source      string          `json:"source"`
	MapID       int             `json:"map_id"`
	PolygonData models.ConeData `json:"polygon_data"`
	Name        string          `json:"name"`
}

type revisionItemsRequest struct {
	Zones *bool `json:"zones"`
	Lines *bool `json:"lines"`
}

func (h *RevisionHandler) ListRevisionItems(c *gin.Context) {
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

	if _, err := middleware.MustGetUser(c); err != nil {
		return
	}

	var req revisionItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	includeZones := req.Zones == nil || *req.Zones
	includeLines := req.Lines != nil && *req.Lines

	if !includeZones && !includeLines {
		apierror.RespondWithError(c, apierror.NewAPIError("revision_filter_empty", "at least one of zones or lines must be true", http.StatusBadRequest))
		return
	}

	result := gin.H{}

	if includeZones {
		originalMap, err := h.mapRepo.FindByID(mapID)
		if err != nil {
			if errors.Is(err, models.ErrMapNotFound) {
				apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map not found", http.StatusNotFound))
				return
			}
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find map", http.StatusInternalServerError))
			return
		}

		allMaps, err := h.mapRepo.FindByName(originalMap.Name)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list map versions", http.StatusInternalServerError))
			return
		}

		zones := make([]RevisionZone, 0)
		for _, mv := range allMaps {
			teamZones, err := h.zoneRepo.FindWithDetailsByTeamAndMap(teamID, mv.ID)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zones", http.StatusInternalServerError))
				return
			}
			for _, tz := range teamZones {
				zones = append(zones, RevisionZone{
					ID:       tz.ID,
					Source:   "team",
					MapID:    tz.MapID,
					Polygons: tz.Polygons,
					Name:     tz.Name,
				})
			}
		}
		if zones == nil {
			zones = []RevisionZone{}
		}
		result["zones"] = zones
	}

	if includeLines {
		lines, err := h.lineOfFireRepo.FindByTeamAndMap(teamID, mapID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list lines of fire", http.StatusInternalServerError))
			return
		}

		revisionLines := make([]RevisionLineOfFire, 0, len(lines))
		for _, l := range lines {
			revisionLines = append(revisionLines, RevisionLineOfFire{
				ID:          l.ID,
				Source:      "team",
				MapID:       l.MapID,
				PolygonData: l.PolygonData,
				Name:        l.Name,
			})
		}

		templateLines, err := h.templateRepo.FindTemplateLinesOfFireByMapID(mapID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list template lines of fire", http.StatusInternalServerError))
			return
		}
		for _, tl := range templateLines {
			revisionLines = append(revisionLines, RevisionLineOfFire{
				ID:          tl.ID,
				Source:      "template",
				MapID:       mapID,
				PolygonData: tl.PolygonData,
				Name:        tl.Name,
			})
		}

		if revisionLines == nil {
			revisionLines = []RevisionLineOfFire{}
		}
		result["lines"] = revisionLines
	}

	apierror.RespondWithSuccess(c, http.StatusOK, result)
}
