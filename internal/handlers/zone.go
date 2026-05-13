package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type ZoneHandler struct {
	zoneRepo        models.ZoneRepository
	mapRepo         models.MapRepository
	lockRepo        models.EditLockRepository
	historyRepo     models.EditHistoryRepository
	teamMemberRepo  models.TeamMemberRepository
	templateSyncSvc *services.TemplateSyncService
	versionRepo     models.ZoneVersionRepository
	logger          services.Logger
}

func NewZoneHandler(
	zoneRepo models.ZoneRepository,
	mapRepo models.MapRepository,
	lockRepo models.EditLockRepository,
	historyRepo models.EditHistoryRepository,
	teamMemberRepo models.TeamMemberRepository,
	templateSyncSvc *services.TemplateSyncService,
	versionRepo models.ZoneVersionRepository,
	logger services.Logger,
) *ZoneHandler {
	return &ZoneHandler{
		zoneRepo:        zoneRepo,
		mapRepo:         mapRepo,
		lockRepo:        lockRepo,
		historyRepo:     historyRepo,
		teamMemberRepo:  teamMemberRepo,
		templateSyncSvc: templateSyncSvc,
		versionRepo:     versionRepo,
		logger:          logger,
	}
}

func (h *ZoneHandler) ListZones(c *gin.Context) {
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

	zones, err := h.zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zones", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"zones": zones})
}

type createZoneRequest struct {
	Polygons []createZonePolygon `json:"polygons"`
	Name     string              `json:"name"`
}

type createZonePolygon struct {
	Points []models.Point `json:"points"`
}

func (h *ZoneHandler) CreateZone(c *gin.Context) {
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

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	obsolete, err := h.mapRepo.IsMapObsolete(mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check map status", http.StatusInternalServerError))
		return
	}
	if obsolete {
		apierror.RespondWithError(c, apierror.NewAPIError("map_obsolete", "this map is obsolete and cannot be edited", http.StatusBadRequest))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(teamID, mapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req createZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	zone, err := h.zoneRepo.Create(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to create zone", http.StatusInternalServerError))
		return
	}

	for i, p := range req.Polygons {
		_, err := h.zoneRepo.AddPolygon(zone.ID, p.Points, i)
		if err != nil {
			if errors.Is(err, models.ErrPolygonHasTooFewPoints) {
				h.zoneRepo.Delete(zone.ID)
				apierror.RespondWithError(c, apierror.NewAPIError("polygon_too_few_points", "each polygon must have at least 3 points", http.StatusBadRequest))
				return
			}
			h.zoneRepo.Delete(zone.ID)
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to add polygon", http.StatusInternalServerError))
			return
		}
	}

	if req.Name != "" {
		taken, err := h.zoneRepo.IsNameTaken(teamID, mapID, req.Name)
		if err != nil {
			h.zoneRepo.Delete(zone.ID)
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check name uniqueness", http.StatusInternalServerError))
			return
		}
		if taken {
			h.zoneRepo.Delete(zone.ID)
			apierror.RespondWithError(c, apierror.NewAPIError("zone_name_taken", "zone name already exists for this team on this map", http.StatusConflict))
			return
		}
		if len(req.Name) > 250 {
			h.zoneRepo.Delete(zone.ID)
			apierror.RespondWithError(c, apierror.NewAPIError("zone_name_too_long", "zone name exceeds maximum length of 250 characters", http.StatusBadRequest))
			return
		}
		if err := h.zoneRepo.SetName(zone.ID, req.Name); err != nil {
			h.zoneRepo.Delete(zone.ID)
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to set zone name", http.StatusInternalServerError))
			return
		}
	}

	polygons, _ := h.zoneRepo.GetPolygons(zone.ID)
	zoneName := ""
	zn, _ := h.zoneRepo.GetName(zone.ID)
	if zn != nil {
		zoneName = zn.Name
	}

	h.syncAfterZoneChange(teamID, mapID)
	h.saveZoneSnapshot(zone.ID, user.ID)

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"zone": models.ZoneWithDetails{
			ID:        zone.ID,
			TeamID:    zone.TeamID,
			MapID:     zone.MapID,
			Polygons:  polygons,
			Name:      zoneName,
			CreatedAt: zone.CreatedAt,
			UpdatedAt: zone.UpdatedAt,
		},
	})
}

func (h *ZoneHandler) GetZone(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	polygons, _ := h.zoneRepo.GetPolygons(zoneID)
	zoneName := ""
	zn, _ := h.zoneRepo.GetName(zoneID)
	if zn != nil {
		zoneName = zn.Name
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"zone": models.ZoneWithDetails{
			ID:        zone.ID,
			TeamID:    zone.TeamID,
			MapID:     zone.MapID,
			Polygons:  polygons,
			Name:      zoneName,
			CreatedAt: zone.CreatedAt,
			UpdatedAt: zone.UpdatedAt,
		},
	})
}

func (h *ZoneHandler) DeleteZone(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	polygons, _ := h.zoneRepo.GetPolygons(zoneID)
	oldData, _ := json.Marshal(gin.H{
		"zone_id":  zoneID,
		"polygons": polygons,
	})
	var editorID *int
	editorID = &user.ID

	h.historyRepo.Log(zoneID, editorID, "delete", json.RawMessage(oldData), nil)

	if err := h.zoneRepo.Delete(zoneID); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to delete zone", http.StatusInternalServerError))
		return
	}

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "zone deleted"})
}

func (h *ZoneHandler) AddPolygon(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req struct {
		Points []models.Point `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if len(req.Points) < 3 {
		apierror.RespondWithError(c, apierror.NewAPIError("polygon_too_few_points", "polygon must have at least 3 points", http.StatusBadRequest))
		return
	}

	count, err := h.zoneRepo.PolygonCount(zoneID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check polygon count", http.StatusInternalServerError))
		return
	}

	if count == 0 {
		zn, nameErr := h.zoneRepo.GetName(zoneID)
		if nameErr != nil && !errors.Is(nameErr, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check zone name", http.StatusInternalServerError))
			return
		}
		if zn == nil || zn.Name == "" {
			apierror.RespondWithError(c, apierror.NewAPIError("name_required", "zone must have a name before adding the first polygon", http.StatusBadRequest))
			return
		}
	}

	polygon, err := h.zoneRepo.AddPolygon(zoneID, req.Points, count)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to add polygon", http.StatusInternalServerError))
		return
	}

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)
	h.saveZoneSnapshot(zone.ID, user.ID)

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"polygon": polygon})
}

func (h *ZoneHandler) RemovePolygon(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	polygonID, err := strconv.Atoi(c.Param("polygonId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_polygon_id", "invalid polygon id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	if err := h.zoneRepo.RemovePolygon(polygonID); err != nil {
		if errors.Is(err, models.ErrPolygonNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("polygon_not_found", "polygon not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to remove polygon", http.StatusInternalServerError))
		return
	}

	zoneDeleted := false
	count, err := h.zoneRepo.PolygonCount(zoneID)
	if err == nil && count == 0 {
		h.zoneRepo.Delete(zoneID)
		zoneDeleted = true
	}

	var editorID *int
	editorID = &user.ID
	h.historyRepo.Log(zoneID, editorID, "remove_polygon", nil, nil)

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)
	if !zoneDeleted {
		h.saveZoneSnapshot(zone.ID, user.ID)
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"polygon_deleted": true,
		"zone_deleted":    zoneDeleted,
	})
}

func (h *ZoneHandler) UpdatePolygonPoints(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	polygonID, err := strconv.Atoi(c.Param("polygonId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_polygon_id", "invalid polygon id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req struct {
		Points []models.Point `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	if len(req.Points) < 3 {
		apierror.RespondWithError(c, apierror.NewAPIError("polygon_too_few_points", "polygon must have at least 3 points", http.StatusBadRequest))
		return
	}

	polygons, _ := h.zoneRepo.GetPolygons(zoneID)
	var oldPoints []models.Point
	for _, p := range polygons {
		if p.ID == polygonID {
			oldPoints = p.Points
			break
		}
	}
	oldData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": oldPoints})
	newData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": req.Points})

	if err := h.zoneRepo.UpdatePolygonPoints(polygonID, req.Points); err != nil {
		if errors.Is(err, models.ErrPolygonNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("polygon_not_found", "polygon not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update polygon points", http.StatusInternalServerError))
		return
	}

	var editorID *int
	editorID = &user.ID
	h.historyRepo.Log(zoneID, editorID, "update", json.RawMessage(oldData), json.RawMessage(newData))

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)
	h.saveZoneSnapshot(zone.ID, user.ID)

	polygon := getPolygonByID(polygons, polygonID)
	if polygon != nil {
		polygon.Points = req.Points
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"polygon": polygon})
}

func (h *ZoneHandler) syncAfterZoneChange(teamID, mapID int) {
	if h.templateSyncSvc == nil {
		return
	}
	if err := h.templateSyncSvc.SyncOnZoneChange(teamID, mapID); err != nil {
		h.logger.Warn("failed to sync template after zone change", "team", teamID, "map", mapID, "error", err)
	}
}

func (h *ZoneHandler) saveZoneSnapshot(zoneID, editorID int) {
	if h.versionRepo == nil {
		return
	}
	polygons, err := h.zoneRepo.GetPolygons(zoneID)
	if err != nil {
		h.logger.Warn("failed to get polygons for snapshot", "zone", zoneID, "error", err)
		return
	}

	zoneName := ""
	zn, err := h.zoneRepo.GetName(zoneID)
	if err == nil && zn != nil {
		zoneName = zn.Name
	}

	snapshot := models.VersionSnapshot{
		Name: zoneName,
	}
	for _, p := range polygons {
		snapshot.Polygons = append(snapshot.Polygons, models.SnapshotPolygon{
			Points:     p.Points,
			OrderIndex: p.OrderIndex,
		})
	}

	editorIDPtr := &editorID
	if _, err := h.versionRepo.Save(zoneID, editorIDPtr, snapshot); err != nil {
		h.logger.Warn("failed to save zone snapshot", "zone", zoneID, "error", err)
	}
}

func getPolygonByID(polygons []models.ZonePolygon, id int) *models.ZonePolygon {
	for _, p := range polygons {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

func (h *ZoneHandler) AddPolygonPoint(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	polygonID, err := strconv.Atoi(c.Param("polygonId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_polygon_id", "invalid polygon id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req struct {
		Point      models.Point `json:"point"`
		IndexAfter int          `json:"index_after"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	polygons, err := h.zoneRepo.GetPolygons(zoneID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get polygons", http.StatusInternalServerError))
		return
	}

	var targetPolygon *models.ZonePolygon
	for _, p := range polygons {
		if p.ID == polygonID {
			targetPolygon = &p
			break
		}
	}
	if targetPolygon == nil {
		apierror.RespondWithError(c, apierror.NewAPIError("polygon_not_found", "polygon not found", http.StatusNotFound))
		return
	}

	oldData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": targetPolygon.Points})

	newPoints := make([]models.Point, 0, len(targetPolygon.Points)+1)
	insertAt := req.IndexAfter + 1
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(targetPolygon.Points) {
		insertAt = len(targetPolygon.Points)
	}
	newPoints = append(newPoints, targetPolygon.Points[:insertAt]...)
	newPoints = append(newPoints, req.Point)
	newPoints = append(newPoints, targetPolygon.Points[insertAt:]...)

	if err := h.zoneRepo.UpdatePolygonPoints(polygonID, newPoints); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update polygon points", http.StatusInternalServerError))
		return
	}

	newData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": newPoints})
	var editorID *int
	editorID = &user.ID
	h.historyRepo.Log(zoneID, editorID, "add_point", json.RawMessage(oldData), json.RawMessage(newData))

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)
	h.saveZoneSnapshot(zone.ID, user.ID)

	targetPolygon.Points = newPoints
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"polygon": targetPolygon})
}

func (h *ZoneHandler) RemovePolygonPoint(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	polygonID, err := strconv.Atoi(c.Param("polygonId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_polygon_id", "invalid polygon id", http.StatusBadRequest))
		return
	}

	pointIndex, err := strconv.Atoi(c.Param("pointIndex"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_point_index", "invalid point index", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	polygons, err := h.zoneRepo.GetPolygons(zoneID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get polygons", http.StatusInternalServerError))
		return
	}

	var targetPolygon *models.ZonePolygon
	for _, p := range polygons {
		if p.ID == polygonID {
			targetPolygon = &p
			break
		}
	}
	if targetPolygon == nil {
		apierror.RespondWithError(c, apierror.NewAPIError("polygon_not_found", "polygon not found", http.StatusNotFound))
		return
	}

	if pointIndex < 0 || pointIndex >= len(targetPolygon.Points) {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_point_index", "point index out of range", http.StatusBadRequest))
		return
	}

	oldData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": targetPolygon.Points})

	newPoints := make([]models.Point, 0, len(targetPolygon.Points)-1)
	newPoints = append(newPoints, targetPolygon.Points[:pointIndex]...)
	newPoints = append(newPoints, targetPolygon.Points[pointIndex+1:]...)

	if len(newPoints) < 3 {
		apierror.RespondWithError(c, apierror.NewAPIError("polygon_too_few_points", "polygon must have at least 3 points", http.StatusBadRequest))
		return
	}

	if err := h.zoneRepo.UpdatePolygonPoints(polygonID, newPoints); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update polygon points", http.StatusInternalServerError))
		return
	}

	newData, _ := json.Marshal(gin.H{"polygon_id": polygonID, "points": newPoints})
	var editorID *int
	editorID = &user.ID
	h.historyRepo.Log(zoneID, editorID, "remove_point", json.RawMessage(oldData), json.RawMessage(newData))

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)
	h.saveZoneSnapshot(zone.ID, user.ID)

	targetPolygon.Points = newPoints
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"polygon": targetPolygon})
}

type setZoneNameRequest struct {
	Name string `json:"name"`
}

func (h *ZoneHandler) SetZoneName(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get zone", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req setZoneNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	oldZn, _ := h.zoneRepo.GetName(zoneID)
	var oldName string
	if oldZn != nil {
		oldName = oldZn.Name
	}
	oldData, _ := json.Marshal(gin.H{"zone_id": zoneID, "name": oldName})

	if err := h.zoneRepo.SetName(zoneID, req.Name); err != nil {
		switch {
		case errors.Is(err, models.ErrZoneNameRequired):
			apierror.RespondWithError(c, apierror.NewAPIError("zone_name_required", err.Error(), http.StatusBadRequest))
		case errors.Is(err, models.ErrZoneNameTooLong):
			apierror.RespondWithError(c, apierror.NewAPIError("zone_name_too_long", err.Error(), http.StatusBadRequest))
		case errors.Is(err, models.ErrZoneNameTaken):
			apierror.RespondWithError(c, apierror.NewAPIError("zone_name_taken", err.Error(), http.StatusConflict))
		default:
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to set zone name", http.StatusInternalServerError))
		}
		return
	}

	zn, _ := h.zoneRepo.GetName(zoneID)

	newData, _ := json.Marshal(gin.H{"zone_id": zoneID, "name": req.Name})
	var editorID *int
	editorID = &user.ID

	h.historyRepo.Log(zoneID, editorID, "rename", json.RawMessage(oldData), json.RawMessage(newData))

	h.saveZoneSnapshot(zone.ID, user.ID)

	h.syncAfterZoneChange(zone.TeamID, zone.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"name": zn})
}

type batchAction struct {
	Type       string         `json:"type"`
	TempID     string         `json:"temp_id,omitempty"`
	ZoneID     *int           `json:"zone_id,omitempty"`
	PolygonID  *int           `json:"polygon_id,omitempty"`
	Points     []models.Point `json:"points,omitempty"`
	Point      *models.Point  `json:"point,omitempty"`
	PointIndex *int           `json:"point_index,omitempty"`
	IndexAfter *int           `json:"index_after,omitempty"`
	Polygons   []batchPolygon `json:"polygons,omitempty"`
	ZoneName   string         `json:"name,omitempty"`
}

type batchPolygon struct {
	Points []models.Point `json:"points"`
}

type batchActionResult struct {
	TempID           string               `json:"temp_id,omitempty"`
	ZoneID           *int                 `json:"zone_id,omitempty"`
	Polygon          *models.ZonePolygon  `json:"polygon,omitempty"`
	Polygons         []models.ZonePolygon `json:"polygons,omitempty"`
	Name             *models.ZoneName     `json:"name,omitempty"`
	Deleted          *bool                `json:"deleted,omitempty"`
	ZoneDeleted      *bool                `json:"zone_deleted,omitempty"`
	DeletedPolygonID *int                 `json:"deleted_polygon_id,omitempty"`
}

type batchRequest struct {
	Actions []batchAction `json:"actions"`
}

type batchResponse struct {
	Results []batchActionResult `json:"results"`
}

func (h *ZoneHandler) BatchActions(c *gin.Context) {
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

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(teamID, mapID, user.ID)
	if err != nil {
		h.logger.Warn("lock check failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		h.logger.Warn("lock required but not held", "user", user.ID, "team", teamID, "map", mapID)
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	obsolete, err := h.mapRepo.IsMapObsolete(mapID)
	if err != nil {
		h.logger.Warn("map obsolete check failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check map status", http.StatusInternalServerError))
		return
	}
	if obsolete {
		apierror.RespondWithError(c, apierror.NewAPIError("map_obsolete", "this map is obsolete and cannot be edited", http.StatusBadRequest))
		return
	}

	results := make([]batchActionResult, 0, len(req.Actions))
	modifiedZoneIDs := make(map[int]bool)
	deletedZoneIDs := make(map[int]bool)
	tempZoneMap := make(map[string]int)

	for _, action := range req.Actions {
		switch action.Type {
		case "create_zone":
			result, err := h.batchCreateZone(teamID, mapID, user.ID, action)
			if err != nil {
				h.logger.Warn("create_zone failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			if action.TempID != "" {
				tempZoneMap[action.TempID] = *result.ZoneID
			}
			results = append(results, result)

		case "add_polygon":
			result, err := h.batchAddPolygon(user.ID, action)
			if err != nil {
				h.logger.Warn("add_polygon failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "remove_polygon":
			result, err := h.batchRemovePolygon(user.ID, action)
			if err != nil {
				h.logger.Warn("remove_polygon failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "update_points":
			result, err := h.batchUpdatePoints(user.ID, action)
			if err != nil {
				h.logger.Warn("update_points failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "add_point":
			result, err := h.batchAddPoint(user.ID, action)
			if err != nil {
				h.logger.Warn("add_point failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "remove_point":
			result, err := h.batchRemovePoint(user.ID, action)
			if err != nil {
				h.logger.Warn("remove_point failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "rename":
			result, err := h.batchRename(user.ID, action)
			if err != nil {
				h.logger.Warn("rename failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		case "delete_zone":
			result, err := h.batchDeleteZone(user.ID, action)
			if err != nil {
				h.logger.Warn("delete_zone failed", "user", user.ID, "team", teamID, "map", mapID, "error", err)
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			results = append(results, result)

		default:
			apierror.RespondWithError(c, apierror.NewAPIError("unknown_action_type", "unknown action type: "+action.Type, http.StatusBadRequest))
			return
		}

		if action.Type == "remove_polygon" && action.ZoneID != nil {
			lastResult := results[len(results)-1]
			if lastResult.ZoneDeleted != nil && *lastResult.ZoneDeleted {
				deletedZoneIDs[*action.ZoneID] = true
			} else {
				modifiedZoneIDs[*action.ZoneID] = true
			}
		} else if action.Type == "create_zone" && len(results) > 0 {
			lastResult := results[len(results)-1]
			if lastResult.ZoneID != nil {
				modifiedZoneIDs[*lastResult.ZoneID] = true
			}
		} else if action.Type != "delete_zone" && action.ZoneID != nil {
			modifiedZoneIDs[*action.ZoneID] = true
		}
	}

	for zoneID := range modifiedZoneIDs {
		if deletedZoneIDs[zoneID] {
			continue
		}
		h.saveZoneSnapshot(zoneID, user.ID)
	}

	h.syncAfterZoneChange(teamID, mapID)

	apierror.RespondWithSuccess(c, http.StatusOK, batchResponse{Results: results})
}

func (h *ZoneHandler) batchCreateZone(teamID, mapID, editorID int, action batchAction) (batchActionResult, error) {
	zone, err := h.zoneRepo.Create(teamID, mapID)
	if err != nil {
		return batchActionResult{}, err
	}

	for i, p := range action.Polygons {
		_, err := h.zoneRepo.AddPolygon(zone.ID, p.Points, i)
		if err != nil {
			h.zoneRepo.Delete(zone.ID)
			return batchActionResult{}, err
		}
	}

	if action.ZoneName != "" {
		if err := h.zoneRepo.SetName(zone.ID, action.ZoneName); err != nil {
			h.zoneRepo.Delete(zone.ID)
			return batchActionResult{}, err
		}
	}

	polygons, _ := h.zoneRepo.GetPolygons(zone.ID)

	var editorIDPtr *int
	editorIDPtr = &editorID
	h.historyRepo.Log(zone.ID, editorIDPtr, "create", nil, nil)

	return batchActionResult{
		TempID:   action.TempID,
		ZoneID:   &zone.ID,
		Polygons: polygons,
	}, nil
}

func (h *ZoneHandler) batchAddPolygon(editorID int, action batchAction) (batchActionResult, error) {
	if action.ZoneID == nil {
		return batchActionResult{}, errors.New("zone_id is required for add_polygon")
	}

	zone, err := h.zoneRepo.FindByID(*action.ZoneID)
	if err != nil {
		return batchActionResult{}, err
	}

	count, err := h.zoneRepo.PolygonCount(*action.ZoneID)
	if err != nil {
		return batchActionResult{}, err
	}

	polygon, err := h.zoneRepo.AddPolygon(*action.ZoneID, action.Points, count)
	if err != nil {
		return batchActionResult{}, err
	}

	var editorIDPtr *int
	editorIDPtr = &editorID
	h.historyRepo.Log(zone.ID, editorIDPtr, "add_polygon", nil, nil)

	return batchActionResult{
		Polygon: polygon,
	}, nil
}

func (h *ZoneHandler) batchRemovePolygon(editorID int, action batchAction) (batchActionResult, error) {
	if action.PolygonID == nil {
		return batchActionResult{}, errors.New("polygon_id is required for remove_polygon")
	}

	if err := h.zoneRepo.RemovePolygon(*action.PolygonID); err != nil {
		return batchActionResult{}, err
	}

	zoneDeleted := false
	if action.ZoneID != nil {
		count, err := h.zoneRepo.PolygonCount(*action.ZoneID)
		if err == nil && count == 0 {
			h.zoneRepo.Delete(*action.ZoneID)
			zoneDeleted = true
		}
	}

	zoneIDForLog := 0
	if action.ZoneID != nil {
		zoneIDForLog = *action.ZoneID
	}
	var editorIDPtr *int
	editorIDPtr = &editorID
	h.historyRepo.Log(zoneIDForLog, editorIDPtr, "remove_polygon", nil, nil)

	zoneDeletedVal := zoneDeleted
	return batchActionResult{
		ZoneDeleted:      &zoneDeletedVal,
		DeletedPolygonID: action.PolygonID,
	}, nil
}

func (h *ZoneHandler) batchUpdatePoints(editorID int, action batchAction) (batchActionResult, error) {
	if action.PolygonID == nil {
		return batchActionResult{}, errors.New("polygon_id is required for update_points")
	}

	if err := h.zoneRepo.UpdatePolygonPoints(*action.PolygonID, action.Points); err != nil {
		return batchActionResult{}, err
	}

	var editorIDPtr *int
	editorIDPtr = &editorID
	zoneIDForLog := 0
	if action.ZoneID != nil {
		zoneIDForLog = *action.ZoneID
	}
	h.historyRepo.Log(zoneIDForLog, editorIDPtr, "update", nil, nil)

	return batchActionResult{
		Polygon: &models.ZonePolygon{
			ID:     *action.PolygonID,
			Points: action.Points,
		},
	}, nil
}

func (h *ZoneHandler) batchAddPoint(editorID int, action batchAction) (batchActionResult, error) {
	if action.PolygonID == nil {
		return batchActionResult{}, errors.New("polygon_id is required for add_point")
	}

	zoneIDForLog := 0
	if action.ZoneID != nil {
		zoneIDForLog = *action.ZoneID
	}

	zone, err := h.zoneRepo.FindByID(zoneIDForLog)
	if err == nil {
		polygons, getErr := h.zoneRepo.GetPolygons(zone.ID)
		if getErr == nil {
			for _, p := range polygons {
				if p.ID == *action.PolygonID {
					newPoints := make([]models.Point, 0, len(p.Points)+1)
					insertAt := 0
					if action.IndexAfter != nil {
						insertAt = *action.IndexAfter + 1
					}
					if insertAt < 0 {
						insertAt = 0
					}
					if insertAt > len(p.Points) {
						insertAt = len(p.Points)
					}
					newPoints = append(newPoints, p.Points[:insertAt]...)
					if action.Point != nil {
						newPoints = append(newPoints, *action.Point)
					}
					newPoints = append(newPoints, p.Points[insertAt:]...)

					if err := h.zoneRepo.UpdatePolygonPoints(*action.PolygonID, newPoints); err != nil {
						return batchActionResult{}, err
					}

					var editorIDPtr *int
					editorIDPtr = &editorID
					h.historyRepo.Log(zoneIDForLog, editorIDPtr, "add_point", nil, nil)

					return batchActionResult{
						Polygon: &models.ZonePolygon{
							ID:     *action.PolygonID,
							Points: newPoints,
						},
					}, nil
				}
			}
		}
	}

	return batchActionResult{}, errors.New("polygon not found or invalid data")
}

func (h *ZoneHandler) batchRemovePoint(editorID int, action batchAction) (batchActionResult, error) {
	if action.PolygonID == nil || action.PointIndex == nil {
		return batchActionResult{}, errors.New("polygon_id and point_index are required for remove_point")
	}

	zoneIDForLog := 0
	if action.ZoneID != nil {
		zoneIDForLog = *action.ZoneID
	}

	zone, err := h.zoneRepo.FindByID(zoneIDForLog)
	if err == nil {
		polygons, getErr := h.zoneRepo.GetPolygons(zone.ID)
		if getErr == nil {
			for _, p := range polygons {
				if p.ID == *action.PolygonID {
					if *action.PointIndex < 0 || *action.PointIndex >= len(p.Points) {
						return batchActionResult{}, errors.New("point index out of range")
					}

					newPoints := make([]models.Point, 0, len(p.Points)-1)
					newPoints = append(newPoints, p.Points[:*action.PointIndex]...)
					newPoints = append(newPoints, p.Points[*action.PointIndex+1:]...)

					if len(newPoints) < 3 {
						return batchActionResult{}, errors.New("polygon must have at least 3 points")
					}

					if err := h.zoneRepo.UpdatePolygonPoints(*action.PolygonID, newPoints); err != nil {
						return batchActionResult{}, err
					}

					var editorIDPtr *int
					editorIDPtr = &editorID
					h.historyRepo.Log(zoneIDForLog, editorIDPtr, "remove_point", nil, nil)

					return batchActionResult{
						Polygon: &models.ZonePolygon{
							ID:     *action.PolygonID,
							Points: newPoints,
						},
					}, nil
				}
			}
		}
	}

	return batchActionResult{}, errors.New("polygon not found or invalid data")
}

func (h *ZoneHandler) batchRename(editorID int, action batchAction) (batchActionResult, error) {
	if action.ZoneID == nil {
		return batchActionResult{}, errors.New("zone_id is required for rename")
	}

	if err := h.zoneRepo.SetName(*action.ZoneID, action.ZoneName); err != nil {
		return batchActionResult{}, err
	}

	zn, _ := h.zoneRepo.GetName(*action.ZoneID)

	var editorIDPtr *int
	editorIDPtr = &editorID
	h.historyRepo.Log(*action.ZoneID, editorIDPtr, "rename", nil, nil)

	return batchActionResult{
		Name: zn,
	}, nil
}

func (h *ZoneHandler) batchDeleteZone(editorID int, action batchAction) (batchActionResult, error) {
	if action.ZoneID == nil {
		return batchActionResult{}, errors.New("zone_id is required for delete_zone")
	}

	var editorIDPtr *int
	editorIDPtr = &editorID
	h.historyRepo.Log(*action.ZoneID, editorIDPtr, "delete", nil, nil)

	if err := h.zoneRepo.Delete(*action.ZoneID); err != nil {
		return batchActionResult{}, err
	}

	deleted := true
	return batchActionResult{
		Deleted: &deleted,
	}, nil
}
