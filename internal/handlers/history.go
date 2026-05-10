package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	versionRepo    models.ZoneVersionRepository
	zoneRepo       models.ZoneRepository
	lockRepo       models.EditLockRepository
	teamMemberRepo models.TeamMemberRepository
	userRepo       models.UserRepository
	historyRepo    models.EditHistoryRepository
}

func NewHistoryHandler(
	versionRepo models.ZoneVersionRepository,
	zoneRepo models.ZoneRepository,
	lockRepo models.EditLockRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	historyRepo models.EditHistoryRepository,
) *HistoryHandler {
	return &HistoryHandler{
		versionRepo:    versionRepo,
		zoneRepo:       zoneRepo,
		lockRepo:       lockRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
		historyRepo:    historyRepo,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id", "message": "invalid map id"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "not_team_member", "message": "you are not a member of this team"})
		return
	}

	var versions []*models.ZoneVersion
	var total int

	if zoneIDStr != "" {
		zoneID, err := strconv.Atoi(zoneIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_zone_id", "message": "invalid zone id"})
			return
		}
		versions, total, err = h.versionRepo.FindByZone(zoneID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to list zone versions"})
			return
		}
	} else {
		versions, total, err = h.versionRepo.FindByMap(teamID, mapID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to list versions"})
			return
		}
	}

	if editorIDStr != "" && member.Role != "captain" && member.Role != "cocaptain" && !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient_permission", "message": "only captains can filter by editor"})
		return
	}

	items := h.buildVersionResponseItems(versions)
	c.JSON(http.StatusOK, listVersionsResponse{
		Versions: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

func (h *HistoryHandler) ListZoneVersions(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_zone_id", "message": "invalid zone id"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "not_team_member", "message": "you are not a member of this team"})
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone_not_found", "message": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get zone"})
		return
	}

	if zone.MapID == 0 || zone.TeamID != teamID {
		c.JSON(http.StatusNotFound, gin.H{"error": "zone_not_found", "message": "zone not found in this team"})
		return
	}

	var visibilityCutoff *time.Time
	if member.Role != "captain" && member.Role != "cocaptain" && !user.IsAdmin {
		boundaryVersionID, err := h.versionRepo.FindFirstOtherEditorVersion(zoneID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check version visibility"})
			return
		}
		if boundaryVersionID != nil {
			boundaryVersion, err := h.versionRepo.FindByID(*boundaryVersionID)
			if err != nil && !errors.Is(err, models.ErrVersionNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get boundary version"})
				return
			}
			if boundaryVersion != nil {
				visibilityCutoff = &boundaryVersion.CreatedAt
			}
		}
	}

	allVersions, total, err := h.versionRepo.FindByZone(zoneID, 1, 1000000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to list zone versions"})
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
	c.JSON(http.StatusOK, listVersionsResponse{
		Versions: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

func (h *HistoryHandler) RestoreVersion(c *gin.Context) {
	zoneID, err := strconv.Atoi(c.Param("zoneId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_zone_id", "message": "invalid zone id"})
		return
	}

	versionID, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_version_id", "message": "invalid version id"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id", "message": "invalid team id"})
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	version, err := h.versionRepo.FindByID(versionID)
	if err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "version_not_found", "message": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get version"})
		return
	}

	if version.ZoneID != zoneID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version_mismatch", "message": "version does not belong to this zone"})
		return
	}

	zone, err := h.zoneRepo.FindByID(zoneID)
	if err != nil {
		if errors.Is(err, models.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone_not_found", "message": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get zone"})
		return
	}

	if zone.TeamID != teamID {
		c.JSON(http.StatusNotFound, gin.H{"error": "zone_not_found", "message": "zone not found in this team"})
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(zone.TeamID, zone.MapID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check lock status"})
		return
	}
	if !locked {
		c.JSON(http.StatusLocked, gin.H{"error": "lock_required", "message": "you must acquire the edit lock before restoring a version"})
		return
	}

	member, err := h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not_team_member", "message": "you are not a member of this team"})
		return
	}

	isCaptainOrCocaptain := member.Role == "captain" || member.Role == "cocaptain" || user.IsAdmin

	if !isCaptainOrCocaptain {
		if version.EditorID == nil || *version.EditorID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not_version_author", "message": "you are not the author of this version"})
			return
		}

		latestVersion, err := h.versionRepo.GetLatestForZone(zoneID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check latest version"})
			return
		}
		if latestVersion != nil &&
			latestVersion.ID != versionID &&
			(latestVersion.EditorID == nil || *latestVersion.EditorID != user.ID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "zone_modified_by_other", "message": "the zone has been modified by another editor since this version"})
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
			log.Printf("warning: failed to remove polygon %d during restore: %v", p.ID, err)
		}
	}

	for _, sp := range version.VersionData.Polygons {
		_, err := h.zoneRepo.AddPolygon(zoneID, sp.Points, sp.OrderIndex)
		if err != nil {
			log.Printf("warning: failed to restore polygon for zone %d: %v", zoneID, err)
		}
	}

	if version.VersionData.Name != "" {
		if err := h.zoneRepo.SetName(zoneID, version.VersionData.Name); err != nil {
			log.Printf("warning: failed to restore name for zone %d: %v", zoneID, err)
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
		log.Printf("warning: failed to delete newer versions after restore (zone=%d): %v", zoneID, err)
	}

	restoredPolygons, _ := h.zoneRepo.GetPolygons(zoneID)
	restoredName := ""
	restoredZn, _ := h.zoneRepo.GetName(zoneID)
	if restoredZn != nil {
		restoredName = restoredZn.Name
	}

	c.JSON(http.StatusOK, gin.H{
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
