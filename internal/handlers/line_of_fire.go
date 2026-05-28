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

type LineOfFireHandler struct {
	lineOfFireRepo  models.LineOfFireRepository
	mapRepo         models.MapRepository
	lockRepo        models.EditLockRepository
	lofHistoryRepo  models.LineOfFireEditHistoryRepository
	teamMemberRepo  models.TeamMemberRepository
	userRepo        models.UserRepository
	templateSyncSvc *services.TemplateSyncService
	versionRepo     models.LineOfFireVersionRepository
	logger          services.Logger
}

func NewLineOfFireHandler(
	lineOfFireRepo models.LineOfFireRepository,
	mapRepo models.MapRepository,
	lockRepo models.EditLockRepository,
	lofHistoryRepo models.LineOfFireEditHistoryRepository,
	teamMemberRepo models.TeamMemberRepository,
	userRepo models.UserRepository,
	templateSyncSvc *services.TemplateSyncService,
	versionRepo models.LineOfFireVersionRepository,
	logger services.Logger,
) *LineOfFireHandler {
	return &LineOfFireHandler{
		lineOfFireRepo:  lineOfFireRepo,
		mapRepo:         mapRepo,
		lockRepo:        lockRepo,
		lofHistoryRepo:  lofHistoryRepo,
		teamMemberRepo:  teamMemberRepo,
		userRepo:        userRepo,
		templateSyncSvc: templateSyncSvc,
		versionRepo:     versionRepo,
		logger:          logger,
	}
}

func (h *LineOfFireHandler) ListLinesOfFire(c *gin.Context) {
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

	lines, err := h.lineOfFireRepo.FindByTeamAndMap(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list lines of fire", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"lines_of_fire": lines})
}

type createLineOfFireRequest struct {
	Name        string          `json:"name"`
	PolygonData models.ConeData `json:"polygon_data"`
}

func (h *LineOfFireHandler) CreateLineOfFire(c *gin.Context) {
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

	var req createLineOfFireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	line, err := h.lineOfFireRepo.Create(teamID, mapID, req.Name, req.PolygonData)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNameRequired) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_required", "line of fire name is required", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrLineOfFireNameTooLong) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_too_long", "line of fire name exceeds maximum length of 250 characters", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrLineOfFireNameTaken) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_taken", "line of fire name already exists for this team on this map", http.StatusConflict))
			return
		}
		if errors.Is(err, models.ErrLineOfFireNoCones) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_no_cones", "line of fire must have at least one cone", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrConeOriginRequired) {
			apierror.RespondWithError(c, apierror.NewAPIError("cone_origin_required", "each cone must have an origin point", http.StatusBadRequest))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to create line of fire", http.StatusInternalServerError))
		return
	}

	h.logEditHistory(line.ID, user.ID, "create", nil, nil)
	h.saveLineOfFireSnapshot(line.ID, user.ID)
	h.syncAfterLineOfFireChange(teamID, mapID)

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"line_of_fire": line})
}

func (h *LineOfFireHandler) GetLineOfFire(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"line_of_fire": line})
}

func (h *LineOfFireHandler) DeleteLineOfFire(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(line.TeamID, line.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	h.logEditHistory(lineOfFireID, user.ID, "delete", nil, nil)

	if err := h.lineOfFireRepo.Delete(lineOfFireID); err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to delete line of fire", http.StatusInternalServerError))
		return
	}

	h.syncAfterLineOfFireChange(line.TeamID, line.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "line of fire deleted"})
}

type setLineOfFireNameRequest struct {
	Name string `json:"name"`
}

func (h *LineOfFireHandler) SetLineOfFireName(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(line.TeamID, line.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req setLineOfFireNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	oldData, _ := json.Marshal(gin.H{"name": line.Name})
	newData, _ := json.Marshal(gin.H{"name": req.Name})

	if err := h.lineOfFireRepo.UpdateName(lineOfFireID, req.Name); err != nil {
		if errors.Is(err, models.ErrLineOfFireNameRequired) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_required", "line of fire name is required", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrLineOfFireNameTooLong) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_too_long", "line of fire name exceeds maximum length of 250 characters", http.StatusBadRequest))
			return
		}
		if errors.Is(err, models.ErrLineOfFireNameTaken) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_name_taken", "line of fire name already exists for this team on this map", http.StatusConflict))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update line of fire name", http.StatusInternalServerError))
		return
	}

	h.logEditHistory(lineOfFireID, user.ID, "rename", oldData, newData)
	h.saveLineOfFireSnapshot(lineOfFireID, user.ID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"name": gin.H{
		"id": lineOfFireID, "line_of_fire_id": lineOfFireID, "name": req.Name,
	}})
}

type coneRequest struct {
	Origin   models.Point `json:"origin"`
	LeftEnd  models.Point `json:"left_end"`
	RightEnd models.Point `json:"right_end"`
}

func (h *LineOfFireHandler) UpdateLineOfFireCone(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	coneIndex, err := strconv.Atoi(c.Param("coneIndex"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "invalid cone index", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(line.TeamID, line.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	if coneIndex < 0 || coneIndex >= len(line.PolygonData.Cones) {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "cone index out of range", http.StatusBadRequest))
		return
	}

	var req coneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	oldCone, _ := json.Marshal(line.PolygonData.Cones[coneIndex])
	newCone, _ := json.Marshal(req)

	line.PolygonData.Cones[coneIndex] = models.Cone{
		Origin:   req.Origin,
		LeftEnd:  req.LeftEnd,
		RightEnd: req.RightEnd,
	}

	if err := h.lineOfFireRepo.UpdatePolygonData(lineOfFireID, line.PolygonData); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to update cone", http.StatusInternalServerError))
		return
	}

	h.logEditHistory(lineOfFireID, user.ID, "update_cone", oldCone, newCone)
	h.saveLineOfFireSnapshot(lineOfFireID, user.ID)
	h.syncAfterLineOfFireChange(line.TeamID, line.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "cone updated"})
}

func (h *LineOfFireHandler) AddLineOfFireCone(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(line.TeamID, line.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	var req coneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	newCone := models.Cone{
		Origin:   req.Origin,
		LeftEnd:  req.LeftEnd,
		RightEnd: req.RightEnd,
	}

	line.PolygonData.Cones = append(line.PolygonData.Cones, newCone)

	if err := h.lineOfFireRepo.UpdatePolygonData(lineOfFireID, line.PolygonData); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to add cone", http.StatusInternalServerError))
		return
	}

	newConeData, _ := json.Marshal(newCone)
	h.logEditHistory(lineOfFireID, user.ID, "add_cone", nil, newConeData)
	h.saveLineOfFireSnapshot(lineOfFireID, user.ID)
	h.syncAfterLineOfFireChange(line.TeamID, line.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "cone added"})
}

func (h *LineOfFireHandler) RemoveLineOfFireCone(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "invalid line of fire id", http.StatusBadRequest))
		return
	}

	coneIndex, err := strconv.Atoi(c.Param("coneIndex"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "invalid cone index", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "line of fire not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get line of fire", http.StatusInternalServerError))
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(line.TeamID, line.MapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
		return
	}

	if coneIndex < 0 || coneIndex >= len(line.PolygonData.Cones) {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "cone index out of range", http.StatusBadRequest))
		return
	}

	removedCone, _ := json.Marshal(line.PolygonData.Cones[coneIndex])

	if len(line.PolygonData.Cones) == 1 {
		if err := h.lineOfFireRepo.Delete(lineOfFireID); err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to delete line of fire", http.StatusInternalServerError))
			return
		}
		h.logEditHistory(lineOfFireID, user.ID, "delete", removedCone, nil)
		h.syncAfterLineOfFireChange(line.TeamID, line.MapID)
		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "line of fire deleted (last cone removed)"})
		return
	}

	line.PolygonData.Cones = append(line.PolygonData.Cones[:coneIndex], line.PolygonData.Cones[coneIndex+1:]...)

	if err := h.lineOfFireRepo.UpdatePolygonData(lineOfFireID, line.PolygonData); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to remove cone", http.StatusInternalServerError))
		return
	}

	h.logEditHistory(lineOfFireID, user.ID, "remove_cone", removedCone, nil)
	h.saveLineOfFireSnapshot(lineOfFireID, user.ID)
	h.syncAfterLineOfFireChange(line.TeamID, line.MapID)

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "cone removed"})
}

type lofBatchAction struct {
	Type         string           `json:"type"`
	TempID       string           `json:"temp_id,omitempty"`
	LineOfFireID *int             `json:"line_of_fire_id,omitempty"`
	Cone         *models.Cone     `json:"cone,omitempty"`
	ConeIndex    *int             `json:"cone_index,omitempty"`
	Name         string           `json:"name,omitempty"`
	PolygonData  *models.ConeData `json:"polygon_data,omitempty"`
}

type lofBatchActionResult struct {
	TempID       string `json:"temp_id,omitempty"`
	LineOfFireID *int   `json:"line_of_fire_id,omitempty"`
	Deleted      *bool  `json:"deleted,omitempty"`
}

type lofBatchRequest struct {
	Actions []lofBatchAction `json:"actions"`
}

type lofBatchResponse struct {
	Results []lofBatchActionResult `json:"results"`
}

func (h *LineOfFireHandler) BatchActions(c *gin.Context) {
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

	locked, err := h.lockRepo.IsLockedByEditor(teamID, mapID, user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
		return
	}
	if !locked {
		apierror.RespondWithError(c, apierror.NewAPIError("lock_required", "you must acquire the edit lock before making changes", http.StatusLocked))
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

	var req lofBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	results := make([]lofBatchActionResult, 0, len(req.Actions))

	for _, action := range req.Actions {
		switch action.Type {
		case "create_line_of_fire":
			if action.PolygonData == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "polygon_data required for create_line_of_fire", http.StatusBadRequest))
				return
			}
			line, err := h.lineOfFireRepo.Create(teamID, mapID, action.Name, *action.PolygonData)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			h.logEditHistory(line.ID, user.ID, "create", nil, nil)
			h.saveLineOfFireSnapshot(line.ID, user.ID)
			results = append(results, lofBatchActionResult{
				TempID:       action.TempID,
				LineOfFireID: &line.ID,
			})

		case "add_cone":
			if action.LineOfFireID == nil || action.Cone == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id and cone required for add_cone", http.StatusBadRequest))
				return
			}
			line, err := h.lineOfFireRepo.FindByID(*action.LineOfFireID)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			line.PolygonData.Cones = append(line.PolygonData.Cones, *action.Cone)
			if err := h.lineOfFireRepo.UpdatePolygonData(*action.LineOfFireID, line.PolygonData); err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			h.logEditHistory(*action.LineOfFireID, user.ID, "add_cone", nil, nil)
			h.saveLineOfFireSnapshot(*action.LineOfFireID, user.ID)
			results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID})

		case "update_cone":
			if action.LineOfFireID == nil || action.Cone == nil || action.ConeIndex == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id, cone, and cone_index required for update_cone", http.StatusBadRequest))
				return
			}
			line, err := h.lineOfFireRepo.FindByID(*action.LineOfFireID)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			if *action.ConeIndex < 0 || *action.ConeIndex >= len(line.PolygonData.Cones) {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "cone index out of range", http.StatusBadRequest))
				return
			}
			line.PolygonData.Cones[*action.ConeIndex] = *action.Cone
			if err := h.lineOfFireRepo.UpdatePolygonData(*action.LineOfFireID, line.PolygonData); err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			h.saveLineOfFireSnapshot(*action.LineOfFireID, user.ID)
			results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID})

		case "remove_cone":
			if action.LineOfFireID == nil || action.ConeIndex == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id and cone_index required for remove_cone", http.StatusBadRequest))
				return
			}
			line, err := h.lineOfFireRepo.FindByID(*action.LineOfFireID)
			if err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			if *action.ConeIndex < 0 || *action.ConeIndex >= len(line.PolygonData.Cones) {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_cone_index", "cone index out of range", http.StatusBadRequest))
				return
			}
			if len(line.PolygonData.Cones) == 1 {
				if err := h.lineOfFireRepo.Delete(*action.LineOfFireID); err != nil {
					apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
					return
				}
				deleted := true
				results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID, Deleted: &deleted})
			} else {
				line.PolygonData.Cones = append(line.PolygonData.Cones[:*action.ConeIndex], line.PolygonData.Cones[*action.ConeIndex+1:]...)
				if err := h.lineOfFireRepo.UpdatePolygonData(*action.LineOfFireID, line.PolygonData); err != nil {
					apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
					return
				}
				h.saveLineOfFireSnapshot(*action.LineOfFireID, user.ID)
				results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID})
			}

		case "rename_line_of_fire":
			if action.LineOfFireID == nil || action.Name == "" {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id and name required for rename_line_of_fire", http.StatusBadRequest))
				return
			}
			if err := h.lineOfFireRepo.UpdateName(*action.LineOfFireID, action.Name); err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			h.saveLineOfFireSnapshot(*action.LineOfFireID, user.ID)
			results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID})

		case "delete_line_of_fire":
			if action.LineOfFireID == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id required for delete_line_of_fire", http.StatusBadRequest))
				return
			}
			if err := h.lineOfFireRepo.Delete(*action.LineOfFireID); err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			deleted := true
			results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID, Deleted: &deleted})

		case "update_polygon_data":
			if action.LineOfFireID == nil || action.PolygonData == nil {
				apierror.RespondWithError(c, apierror.NewAPIError("invalid_action", "line_of_fire_id and polygon_data required for update_polygon_data", http.StatusBadRequest))
				return
			}
			if err := h.lineOfFireRepo.UpdatePolygonData(*action.LineOfFireID, *action.PolygonData); err != nil {
				apierror.RespondWithError(c, apierror.NewAPIError("batch_action_failed", err.Error(), http.StatusInternalServerError))
				return
			}
			h.saveLineOfFireSnapshot(*action.LineOfFireID, user.ID)
			results = append(results, lofBatchActionResult{LineOfFireID: action.LineOfFireID})

		default:
			apierror.RespondWithError(c, apierror.NewAPIError("unknown_action_type", "unknown action type: "+action.Type, http.StatusBadRequest))
			return
		}
	}

	h.syncAfterLineOfFireChange(teamID, mapID)

	apierror.RespondWithSuccess(c, http.StatusOK, lofBatchResponse{Results: results})
}

type importLinesOfFireItem struct {
	Name        string          `json:"name"`
	PolygonData models.ConeData `json:"polygon_data"`
}

type importLinesOfFireRequest struct {
	LinesOfFire []importLinesOfFireItem `json:"lines_of_fire"`
}

func (h *LineOfFireHandler) ImportLinesOfFire(c *gin.Context) {
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

	var req importLinesOfFireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
		return
	}

	var created []*models.LineOfFire
	for _, item := range req.LinesOfFire {
		line, err := h.lineOfFireRepo.Create(teamID, mapID, item.Name, item.PolygonData)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("import_failed", err.Error(), http.StatusInternalServerError))
			return
		}
		h.logEditHistory(line.ID, user.ID, "create", nil, nil)
		h.saveLineOfFireSnapshot(line.ID, user.ID)
		created = append(created, line)
	}

	h.syncAfterLineOfFireChange(teamID, mapID)

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"lines_of_fire": created})
}

func (h *LineOfFireHandler) ExportLinesOfFire(c *gin.Context) {
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

	lines, err := h.lineOfFireRepo.FindByTeamAndMap(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to export lines of fire", http.StatusInternalServerError))
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"lines_of_fire": lines})
}

func (h *LineOfFireHandler) logEditHistory(lineOfFireID, editorID int, action string, oldData, newData json.RawMessage) {
	if h.lofHistoryRepo == nil {
		return
	}
	editorIDPtr := &editorID
	if err := h.lofHistoryRepo.Log(lineOfFireID, editorIDPtr, action, oldData, newData); err != nil {
		h.logger.Warn("failed to log line of fire edit history", "line_of_fire", lineOfFireID, "action", action, "error", err)
	}
}

type listLineOfFireVersionsResponseItem struct {
	ID           int             `json:"id"`
	LineOfFireID int             `json:"line_of_fire_id"`
	EditorID     *int            `json:"editor_id,omitempty"`
	EditorName   string          `json:"editor_name,omitempty"`
	VersionData  json.RawMessage `json:"version_data"`
	CreatedAt    time.Time       `json:"created_at"`
}

type listLineOfFireVersionsResponse struct {
	Versions []listLineOfFireVersionsResponseItem `json:"versions"`
	Total    int                                  `json:"total"`
}

func (h *LineOfFireHandler) ListLineOfFireVersions(c *gin.Context) {
	lineOfFireID, err := strconv.Atoi(c.Param("lineOfFireId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_line_of_fire_id", "id de ligne de tir invalide", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "id d'équipe invalide", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	_, err = h.teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("not_team_member", "vous n'êtes pas membre de cette équipe", http.StatusForbidden))
		return
	}

	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		if errors.Is(err, models.ErrLineOfFireNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "ligne de tir introuvable", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "erreur lors de la récupération de la ligne de tir", http.StatusInternalServerError))
		return
	}

	if line.TeamID != teamID {
		apierror.RespondWithError(c, apierror.NewAPIError("line_of_fire_not_found", "ligne de tir introuvable dans cette équipe", http.StatusNotFound))
		return
	}

	allVersions, err := h.versionRepo.FindByLineOfFire(lineOfFireID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "erreur lors de la récupération des versions", http.StatusInternalServerError))
		return
	}

	total := len(allVersions)

	editorIDs := make(map[int]string)
	for _, v := range allVersions {
		if v.EditorID != nil {
			if _, ok := editorIDs[*v.EditorID]; !ok {
				u, err := h.userRepo.FindByID(*v.EditorID)
				if err == nil {
					editorIDs[*v.EditorID] = u.DisplayName
				}
			}
		}
	}

	items := make([]listLineOfFireVersionsResponseItem, len(allVersions))
	for i, v := range allVersions {
		item := listLineOfFireVersionsResponseItem{
			ID:           v.ID,
			LineOfFireID: v.LineOfFireID,
			EditorID:     v.EditorID,
			VersionData:  v.VersionData,
			CreatedAt:    v.CreatedAt,
		}
		if v.EditorID != nil {
			item.EditorName = editorIDs[*v.EditorID]
		}
		items[i] = item
	}

	apierror.RespondWithSuccess(c, http.StatusOK, listLineOfFireVersionsResponse{
		Versions: items,
		Total:    total,
	})
}

func (h *LineOfFireHandler) saveLineOfFireSnapshot(lineOfFireID, editorID int) {
	if h.versionRepo == nil {
		return
	}
	line, err := h.lineOfFireRepo.FindByID(lineOfFireID)
	if err != nil {
		h.logger.Warn("failed to get line of fire for snapshot", "line_of_fire", lineOfFireID, "error", err)
		return
	}

	raw, err := json.Marshal(gin.H{
		"name":         line.Name,
		"polygon_data": line.PolygonData,
	})
	if err != nil {
		h.logger.Warn("failed to marshal snapshot", "line_of_fire", lineOfFireID, "error", err)
		return
	}

	editorIDPtr := &editorID
	if _, err := h.versionRepo.Save(lineOfFireID, editorIDPtr, raw); err != nil {
		h.logger.Warn("failed to save snapshot", "line_of_fire", lineOfFireID, "error", err)
	}
}

func (h *LineOfFireHandler) syncAfterLineOfFireChange(teamID, mapID int) {
	if h.templateSyncSvc == nil {
		return
	}
	if err := h.templateSyncSvc.SyncAfterLineOfFireChange(teamID, mapID); err != nil {
		h.logger.Warn("failed to sync template after line of fire change", "team", teamID, "map", mapID, "error", err)
	}
}
