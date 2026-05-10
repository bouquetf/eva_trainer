package handlers

import (
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

type LockResponse struct {
	ID               int       `json:"id"`
	TeamID           int       `json:"team_id"`
	MapID            int       `json:"map_id"`
	EditorID         int       `json:"editor_id"`
	EditorName       string    `json:"editor_name,omitempty"`
	LockedAt         time.Time `json:"locked_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RemainingSeconds int       `json:"remaining_seconds"`
}

func newLockResponse(lock *models.EditLock, editorName string) LockResponse {
	remaining := int(time.Until(lock.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return LockResponse{
		ID:               lock.ID,
		TeamID:           lock.TeamID,
		MapID:            lock.MapID,
		EditorID:         lock.EditorID,
		EditorName:       editorName,
		LockedAt:         lock.LockedAt,
		ExpiresAt:        lock.ExpiresAt,
		RemainingSeconds: remaining,
	}
}

type EditLockHandler struct {
	lockRepo models.EditLockRepository
	userRepo models.UserRepository
	logger   services.Logger
}

func NewEditLockHandler(lockRepo models.EditLockRepository, userRepo models.UserRepository, logger services.Logger) *EditLockHandler {
	return &EditLockHandler{lockRepo: lockRepo, userRepo: userRepo, logger: logger}
}

func (h *EditLockHandler) getUserDisplayName(userID int) string {
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return ""
	}
	return user.DisplayName
}

func (h *EditLockHandler) AcquireLock(c *gin.Context) {
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

	lock, err := h.lockRepo.Acquire(teamID, mapID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrLockAlreadyAcquired) {
			apierror.RespondWithError(c, apierror.NewAPIError("lock_already_acquired", "edit lock already acquired by another editor", http.StatusConflict))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to acquire lock", http.StatusInternalServerError))
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) ReleaseLock(c *gin.Context) {
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
		anyLocked, checkErr := h.lockRepo.IsLocked(teamID, mapID)
		if checkErr != nil {
			apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to check lock status", http.StatusInternalServerError))
			return
		}
		if anyLocked {
			apierror.RespondWithError(c, apierror.NewAPIError("lock_not_owned", "edit lock is owned by another editor", http.StatusForbidden))
			return
		}
	}

	if err := h.lockRepo.Release(teamID, mapID); err != nil {
		if errors.Is(err, models.ErrLockNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("lock_not_found", "edit lock not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to release lock", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "lock released"})
}

func (h *EditLockHandler) RefreshLock(c *gin.Context) {
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
		apierror.RespondWithError(c, apierror.NewAPIError("lock_not_owned", "you do not own the edit lock", http.StatusForbidden))
		return
	}

	lock, err := h.lockRepo.Refresh(teamID, mapID)
	if err != nil {
		if errors.Is(err, models.ErrLockNotFound) {
			apierror.RespondWithError(c, apierror.NewAPIError("lock_not_found", "edit lock not found", http.StatusNotFound))
			return
		}
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to refresh lock", http.StatusInternalServerError))
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) GetLockStatus(c *gin.Context) {
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

	lock, err := h.lockRepo.Get(teamID, mapID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to get lock status", http.StatusInternalServerError))
		return
	}

	if lock == nil {
		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"locked": false, "lock": nil})
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"locked": true, "lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) ReleaseLockForEditor(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid_map_id", http.StatusBadRequest))
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid_team_id", http.StatusBadRequest))
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	err = h.lockRepo.ReleaseByEditor(teamID, mapID, user.ID)
	if err != nil && !errors.Is(err, models.ErrLockNotFound) {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal_error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "lock released"})
}

func (h *EditLockHandler) CleanupExpiredLocks(c *gin.Context) {
	cleaned, err := h.lockRepo.CleanupExpired()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to cleanup expired locks", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"cleaned": cleaned})
}

func (h *EditLockHandler) ForceClearLocks(c *gin.Context) {
	count, err := h.lockRepo.ForceClearAll()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to force clear locks", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"cleared": true, "count": count})
}
