package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

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
}

func NewEditLockHandler(lockRepo models.EditLockRepository, userRepo models.UserRepository) *EditLockHandler {
	return &EditLockHandler{lockRepo: lockRepo, userRepo: userRepo}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id", "message": "invalid map id"})
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

	lock, err := h.lockRepo.Acquire(teamID, mapID, user.ID)
	if err != nil {
		if errors.Is(err, models.ErrLockAlreadyAcquired) {
			existing, getErr := h.lockRepo.Get(teamID, mapID)
			lockedBy := 0
			editorName := ""
			if getErr == nil && existing != nil {
				lockedBy = existing.EditorID
				editorName = h.getUserDisplayName(existing.EditorID)
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":       "lock_already_acquired",
				"message":     "edit lock already acquired by another editor",
				"locked_by":   lockedBy,
				"editor_name": editorName,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to acquire lock"})
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	c.JSON(http.StatusOK, gin.H{"lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) ReleaseLock(c *gin.Context) {
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

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(teamID, mapID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check lock status"})
		return
	}
	if !locked {
		anyLocked, checkErr := h.lockRepo.IsLocked(teamID, mapID)
		if checkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check lock status"})
			return
		}
		if anyLocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "lock_not_owned", "message": "edit lock is owned by another editor"})
			return
		}
	}

	if err := h.lockRepo.Release(teamID, mapID); err != nil {
		if errors.Is(err, models.ErrLockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lock_not_found", "message": "edit lock not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to release lock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "lock released"})
}

func (h *EditLockHandler) RefreshLock(c *gin.Context) {
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

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	locked, err := h.lockRepo.IsLockedByEditor(teamID, mapID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check lock status"})
		return
	}
	if !locked {
		c.JSON(http.StatusForbidden, gin.H{"error": "lock_not_owned", "message": "you do not own the edit lock"})
		return
	}

	lock, err := h.lockRepo.Refresh(teamID, mapID)
	if err != nil {
		if errors.Is(err, models.ErrLockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lock_not_found", "message": "edit lock not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to refresh lock"})
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	c.JSON(http.StatusOK, gin.H{"lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) GetLockStatus(c *gin.Context) {
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

	lock, err := h.lockRepo.Get(teamID, mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get lock status"})
		return
	}

	if lock == nil {
		c.JSON(http.StatusOK, gin.H{"locked": false, "lock": nil})
		return
	}

	editorName := h.getUserDisplayName(lock.EditorID)
	c.JSON(http.StatusOK, gin.H{"locked": true, "lock": newLockResponse(lock, editorName)})
}

func (h *EditLockHandler) ReleaseLockForEditor(c *gin.Context) {
	mapID, err := strconv.Atoi(c.Param("mapId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id"})
		return
	}

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id"})
		return
	}

	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	err = h.lockRepo.ReleaseByEditor(teamID, mapID, user.ID)
	if err != nil && !errors.Is(err, models.ErrLockNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "lock released"})
}

func (h *EditLockHandler) CleanupExpiredLocks(c *gin.Context) {
	cleaned, err := h.lockRepo.CleanupExpired()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to cleanup expired locks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cleaned": cleaned})
}

func (h *EditLockHandler) ForceClearLocks(c *gin.Context) {
	count, err := h.lockRepo.ForceClearAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to force clear locks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cleared": true, "count": count})
}
