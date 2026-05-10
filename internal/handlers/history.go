package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	versionRepo    models.ZoneVersionRepository
	zoneRepo       models.ZoneRepository
	lockRepo       models.EditLockRepository
	teamMemberRepo models.TeamMemberRepository
	userRepo       models.UserRepository
	historyRepo    models.EditHistoryRepository
	logger         services.Logger
}

func NewHistoryHandler(
	versionRepo models.ZoneVersionRepository,
	zoneRepo models.ZoneRepository,
	lockRepo models.EditLockRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	historyRepo models.EditHistoryRepository,
	logger services.Logger,
) *HistoryHandler {
	return &HistoryHandler{
		versionRepo:    versionRepo,
		zoneRepo:       zoneRepo,
		lockRepo:       lockRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
		historyRepo:    historyRepo,
		logger:         logger,
	}
}

type versionResponseItem struct {
	ID          int                    `json:"id"`
	ZoneID      int                    `json:"zone_id"`
	ZoneName    string                 `json:"zone_name"`
	EditorID    *int                   `json:"editor_id,omitempty"`
	EditorName  string                 `json:"editor_name,omitempty"`
	VersionData models.VersionSnapshot `json:"version_data"`
	CreatedAt   time.Time              `json:"created_at"`
}

type listVersionsResponse struct {
	Versions []versionResponseItem `json:"versions"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	Limit    int                   `json:"limit"`
}

func (h *HistoryHandler) ListVersions(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}

	zoneIDStr := c.Query("zone_id")
	editorIDStr := c.Query("editor_id")

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
		return
	}

	var versions []*models.ZoneVersion
	var total int

	if zoneIDStr != "" {
		zoneID, err := strconv.Atoi(zoneIDStr)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
			return
		}
		versions, total, err = h.versionRepo.FindByZone(zoneID, page, limit)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zone versions", http.StatusInternalServerError))
			return
		}
	} else {
		versions, total, err = h.versionRepo.FindByMap(teamID, mapID, page, limit)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list versions", http.StatusInternalServerError))
			return
		}
	}

	if editorIDStr != "" && member.Role != "captain" && member.Role != "cocaptain" && !user.IsAdmin {
		apierror.RespondWithError(c, apierror.NewAPIError("insufficient_permission", "only captains can filter by editor", http.StatusForbidden))
		return
	}

	items := h.buildVersionResponseItems(versions)
	apierror.RespondWithSuccess(c, http.StatusOK, listVersionsResponse{
		Versions: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

func (h *HistoryHandler) ListZoneVersions(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
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

	if zone.MapID == 0 || zone.TeamID != teamID {
		apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found in this team", http.StatusNotFound))
		return
	}

	var visibilityCutoff *time.Time
	if member.Role != "captain" && member.Role != "cocaptain" && !user.IsAdmin {
		boundaryVersionID, err := h.versionRepo.FindFirstOtherEditorVersion(zoneID, user.ID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check version visibility", http.StatusInternalServerError))
			return
		}
		if boundaryVersionID != nil {
			boundaryVersion, err := h.versionRepo.FindByID(*boundaryVersionID)
			if err != nil && !errors.Is(err, models.ErrVersionNotFound) {
				apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get boundary version", http.StatusInternalServerError))
				return
			}
			if boundaryVersion != nil {
				visibilityCutoff = &boundaryVersion.CreatedAt
			}
		}
	}

	allVersions, total, err := h.versionRepo.FindByZone(zoneID, 1, 1000000)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list zone versions", http.StatusInternalServerError))
		return
	}

	var filtered []*models.ZoneVersion
	if visibilityCutoff != nil {
		for _, v := range allVersions {
			if v.CreatedAt.Before(*visibilityCutoff) {
				filtered = append(filtered, v)
			}
		}
	} else {
		filtered = allVersions
	}

	total = len(filtered)

	start := (page - 1) * limit
	if start >= len(filtered) {
		filtered = nil
	} else {
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	items := h.buildVersionResponseItems(filtered)
	apierror.RespondWithSuccess(c, http.StatusOK, listVersionsResponse{
		Versions: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

func (h *HistoryHandler) RestoreVersion(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_zone_id", "invalid zone id", http.StatusBadRequest))
		return
	}

	versionID, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_version_id", "invalid version id", http.StatusBadRequest))
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

	version, err := h.versionRepo.FindByID(versionID)
	if err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("version_not_found", "version not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get version", http.StatusInternalServerError))
		return
	}

	if version.ZoneID != zoneID {
		apierror.RespondWithError(c, apierror.NewAPIError("version_mismatch", "version does not belong to this zone", http.StatusBadRequest))
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

	if zone.TeamID != teamID {
		apierror.RespondWithError(c, apierror.NewAPIError("zone_not_found", "zone not found in this team", http.StatusNotFound))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before restoring a version", http.StatusLocked))
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "you are not a member of this team", http.StatusForbidden))
		return
	}

	isCaptainOrCocaptain := member.Role == "captain" || member.Role == "cocaptain" || user.IsAdmin

	if !isCaptainOrCocaptain {
		if version.EditorID == nil || *version.EditorID != user.ID {
			apierror.RespondWithError(c, apierror.NewAPIError("not_version_author", "you are not the author of this version", http.StatusForbidden))
			return
		}

		latestVersion, err := h.versionRepo.GetLatestForZone(zoneID)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check latest version", http.StatusInternalServerError))
			return
		}
		if latestVersion != nil &&
			latestVersion.ID != versionID &&
			(latestVersion.EditorID == nil || *latestVersion.EditorID != user.ID) {
			apierror.RespondWithError(c, apierror.NewAPIError("zone_modified_by_other", "the zone has been modified by another editor since this version", http.StatusForbidden))
			return
		}
	}

	currentPolygons, _ := h.zoneRepo.GetPolygons(zoneID)
	currentName := ""
	zn, _ := h.zoneRepo.GetName(zoneID)
	if zn != nil {
		currentName = zn.Name
	}

	oldData, _ := json.Marshal(gin.H{
		"zone_id":  zoneID,
		"polygons": currentPolygons,
		"name":     currentName,
	})

	for _, p := range currentPolygons {
		if err := h.zoneRepo.RemovePolygon(p.ID); err != nil {
			h.logger.Warn("failed to remove polygon during restore", "polygon", p.ID, "error", err)
		}
	}

	for _, sp := range version.VersionData.Polygons {
		_, err := h.zoneRepo.AddPolygon(zoneID, sp.Points, sp.OrderIndex)
		if err != nil {
			h.logger.Warn("failed to restore polygon for zone", "zone", zoneID, "error", err)
		}
	}

	if version.VersionData.Name != "" {
		if err := h.zoneRepo.SetName(zoneID, version.VersionData.Name); err != nil {
			h.logger.Warn("failed to restore name for zone", "zone", zoneID, "error", err)
		}
	}

	newData, _ := json.Marshal(gin.H{
		"zone_id":          zoneID,
		"restored_version": versionID,
		"version_data":     version.VersionData,
	})

	var editorIDPtr *int
	editorIDPtr = &user.ID
	h.historyRepo.Log(zoneID, editorIDPtr, "version_restore", json.RawMessage(oldData), json.RawMessage(newData))

	if err := h.versionRepo.DeleteNewerThan(zoneID, version.CreatedAt); err != nil {
		h.logger.Warn("failed to delete newer versions after restore", "zone", zoneID, "error", err)
	}

	restoredPolygons, _ := h.zoneRepo.GetPolygons(zoneID)
	restoredName := ""
	restoredZn, _ := h.zoneRepo.GetName(zoneID)
	if restoredZn != nil {
		restoredName = restoredZn.Name
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"zone": models.ZoneWithDetails{
			ID:        zone.ID,
			TeamID:    zone.TeamID,
			MapID:     zone.MapID,
			Polygons:  restoredPolygons,
			Name:      restoredName,
			CreatedAt: zone.CreatedAt,
			UpdatedAt: zone.UpdatedAt,
		},
	})
}

func (h *HistoryHandler) buildVersionResponseItems(versions []*models.ZoneVersion) []versionResponseItem {
	if len(versions) == 0 {
		return []versionResponseItem{}
	}

	zoneIDs := make([]int, 0, len(versions))
	editorIDs := make(map[int]bool)
	zoneIDSet := make(map[int]bool)

	for _, v := range versions {
		if !zoneIDSet[v.ZoneID] {
			zoneIDs = append(zoneIDs, v.ZoneID)
			zoneIDSet[v.ZoneID] = true
		}
		if v.EditorID != nil {
			editorIDs[*v.EditorID] = true
		}
	}

	zoneNames := make(map[int]string)
	for _, zid := range zoneIDs {
		zn, err := h.zoneRepo.GetName(zid)
		if err == nil && zn != nil {
			zoneNames[zid] = zn.Name
		}
	}

	editorNames := make(map[int]string)
	for eid := range editorIDs {
		u, err := h.userRepo.FindByID(eid)
		if err == nil && u != nil {
			editorNames[eid] = u.DisplayName
		}
	}

	items := make([]versionResponseItem, 0, len(versions))
	for _, v := range versions {
		zoneName := zoneNames[v.ZoneID]
		if zoneName == "" {
			zoneName = v.VersionData.Name
		}

		editorName := ""
		if v.EditorID != nil {
			editorName = editorNames[*v.EditorID]
			if editorName == "" {
				editorName = "Utilisateur supprimé"
			}
		}

		items = append(items, versionResponseItem{
			ID:          v.ID,
			ZoneID:      v.ZoneID,
			ZoneName:    zoneName,
			EditorID:    v.EditorID,
			EditorName:  editorName,
			VersionData: v.VersionData,
			CreatedAt:   v.CreatedAt,
		})
	}
	return items
}
