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
	zoneRepo models.ZoneRepository
	mapRepo  models.MapRepository
	logger   services.Logger
}

func NewRevisionHandler(
	zoneRepo models.ZoneRepository,
	mapRepo models.MapRepository,
	logger services.Logger,
) *RevisionHandler {
	return &RevisionHandler{
		zoneRepo: zoneRepo,
		mapRepo:  mapRepo,
		logger:   logger,
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
