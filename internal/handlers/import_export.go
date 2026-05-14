package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type MapExport struct {
	MapName    string       `json:"map_name"`
	Team       string       `json:"team"`
	ExportDate string       `json:"export_date"`
	Zones      []ZoneExport `json:"zones"`
}

type ZoneExport struct {
	Name     string          `json:"name"`
	Polygons []PolygonExport `json:"polygons"`
}

type PolygonExport struct {
	Points []models.Point `json:"points"`
}

type ImportMapRequest struct {
	MapName string                  `json:"map_name"`
	Zones   []models.ImportZoneData `json:"zones"`
}

type ImportMapError struct {
	MapName string `json:"map_name"`
	Error   string `json:"error"`
}

type ImportTeamResult struct {
	ImportedCount int              `json:"imported_count"`
	MapsProcessed int              `json:"maps_processed"`
	Errors        []ImportMapError `json:"errors,omitempty"`
}

type ImportExportHandler struct {
	zoneRepo models.ZoneRepository
	mapRepo  models.MapRepository
	teamRepo models.TeamRepository
	logger   services.Logger
}

func NewImportExportHandler(
	zoneRepo models.ZoneRepository,
	mapRepo models.MapRepository,
	teamRepo models.TeamRepository,
	logger services.Logger,
) *ImportExportHandler {
	return &ImportExportHandler{
		zoneRepo: zoneRepo,
		mapRepo:  mapRepo,
		teamRepo: teamRepo,
		logger:   logger,
	}
}

func (h *ImportExportHandler) ExportTeamMap(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid map id", http.StatusBadRequest))
		return
	}

	team, err := h.teamRepo.FindByID(teamID)
	if err == models.ErrTeamNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find team", http.StatusInternalServerError))
		return
	}

	evaMap, err := h.mapRepo.FindByID(mapID)
	if err == models.ErrMapNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find map", http.StatusInternalServerError))
		return
	}

	zones, err := h.zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zones", http.StatusInternalServerError))
		return
	}

	export := buildMapExport(evaMap.Name, team.Name, zones)

	apierror.RespondWithSuccess(c, http.StatusOK, export)
}

func (h *ImportExportHandler) ExportTeamMaps(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	team, err := h.teamRepo.FindByID(teamID)
	if err == models.ErrTeamNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find team", http.StatusInternalServerError))
		return
	}

	zonesByMap, err := h.zoneRepo.FindWithDetailsByTeam(teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zones", http.StatusInternalServerError))
		return
	}

	exports := make([]MapExport, 0, len(zonesByMap))
	for mapID, zones := range zonesByMap {
		evaMap, err := h.mapRepo.FindByID(mapID)
		if err != nil {
			if err == sql.ErrNoRows || err == models.ErrMapNotFound {
				continue
			}
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find map", http.StatusInternalServerError))
			return
		}
		exports = append(exports, buildMapExport(evaMap.Name, team.Name, zones))
	}

	apierror.RespondWithSuccess(c, http.StatusOK, exports)
}

func buildMapExport(mapName, teamName string, zones []*models.ZoneWithDetails) MapExport {
	zoneExports := make([]ZoneExport, 0, len(zones))
	for _, z := range zones {
		polygonExports := make([]PolygonExport, 0, len(z.Polygons))
		for _, p := range z.Polygons {
			polygonExports = append(polygonExports, PolygonExport{
				Points: p.Points,
			})
		}
		zoneExports = append(zoneExports, ZoneExport{
			Name:     z.Name,
			Polygons: polygonExports,
		})
	}

	return MapExport{
		MapName:    mapName,
		Team:       teamName,
		ExportDate: time.Now().UTC().Format(time.RFC3339),
		Zones:      zoneExports,
	}
}

func (h *ImportExportHandler) ImportTeamZones(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	_, err = h.teamRepo.FindByID(teamID)
	if err == models.ErrTeamNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("team_not_found", "team not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find team", http.StatusInternalServerError))
		return
	}

	var requests []ImportMapRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_json", "invalid JSON: "+err.Error(), http.StatusBadRequest))
		return
	}

	result := ImportTeamResult{}

	for _, req := range requests {
		evaMap, err := h.mapRepo.FindByNameLatest(req.MapName)
		if err == models.ErrMapNotFound {
			result.Errors = append(result.Errors, ImportMapError{
				MapName: req.MapName,
				Error:   "map not found",
			})
			continue
		}
		if err != nil {
			result.Errors = append(result.Errors, ImportMapError{
				MapName: req.MapName,
				Error:   "internal error: " + err.Error(),
			})
			continue
		}

		created, err := h.zoneRepo.ImportZones(teamID, evaMap.ID, req.Zones)
		if err != nil {
			apiErr := "import failed"
			if errors.Is(err, models.ErrZoneNameTaken) {
				apiErr = "zone name already exists"
			} else if errors.Is(err, models.ErrPolygonHasTooFewPoints) {
				apiErr = "polygon must have at least 3 points"
			}
			result.Errors = append(result.Errors, ImportMapError{
				MapName: req.MapName,
				Error:   apiErr,
			})
			continue
		}

		result.ImportedCount += len(created)
		result.MapsProcessed++

		h.logger.Info("imported zones",
			"team_id", teamID,
			"map_name", req.MapName,
			"map_id", evaMap.ID,
			"count", len(created),
		)
	}

	apierror.RespondWithSuccess(c, http.StatusOK, result)
}
