package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type EditLockHandler struct {
	lockRepo models.EditLockRepository
}

func NewEditLockHandler(lockRepo models.EditLockRepository) *EditLockHandler {
	return &EditLockHandler{lockRepo: lockRepo}
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
			if getErr == nil && existing != nil {
				lockedBy = existing.EditorID
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":     "lock_already_acquired",
				"message":   "edit lock already acquired by another editor",
				"locked_by": lockedBy,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to acquire lock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lock": lock})
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

	c.JSON(http.StatusOK, gin.H{"lock": lock})
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

	c.JSON(http.StatusOK, gin.H{"locked": true, "lock": lock})
}
