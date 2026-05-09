package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	mapRepo models.MapRepository
}

func NewMapHandler(mapRepo models.MapRepository) *MapHandler {
	return &MapHandler{mapRepo: mapRepo}
}

func (h *MapHandler) UploadMap(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "file_too_large",
			"message": "file exceeds maximum size of 50 MB",
		})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "map_name_required",
			"message": "map name is required",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "file_required",
			"message": "map file is required",
		})
		return
	}

	if ext := filepath.Ext(file.Filename); !strings.EqualFold(ext, ".webp") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file_format",
			"message": "only .webp files are accepted",
		})
		return
	}

	if file.Header.Get("Content-Type") != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file_format",
			"message": "Content-Type must be image/webp",
		})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to read uploaded file",
		})
		return
	}
	defer f.Close()

	blobData, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to read file content",
		})
		return
	}

	existing, _ := h.mapRepo.FindByNameLatest(name)
	var previousVersion *previousVersionInfo
	if existing != nil && !existing.IsObsolete {
		previousVersion = &previousVersionInfo{ID: existing.ID, Name: existing.Name}
	}

	m, err := h.mapRepo.Create(name, blobData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": fmt.Sprintf("failed to create map: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"map":                        m,
		"previous_version_obsoleted": previousVersion,
	})
}

type previousVersionInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *MapHandler) ListTeamMaps(c *gin.Context) {
	maps, err := h.mapRepo.FindAllActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list maps",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maps": maps})
}

func (h *MapHandler) ListMaps(c *gin.Context) {
	name := c.Query("name")
	var maps []*models.Map
	var err error
	if name != "" {
		maps, err = h.mapRepo.FindByName(name)
	} else {
		maps, err = h.mapRepo.FindAll()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list maps",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maps": maps})
}

func (h *MapHandler) GetMapDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_map_id",
			"message": "invalid map id",
		})
		return
	}

	m, err := h.mapRepo.FindByID(id)
	if err == models.ErrMapNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "map_not_found",
			"message": "map not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find map",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"map": m})
}

func (h *MapHandler) PurgeObsoleteMaps(c *gin.Context) {
	obsoleteMaps, err := h.mapRepo.FindObsoleteWithoutZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to find obsolete maps",
		})
		return
	}

	if len(obsoleteMaps) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    "no obsolete maps to purge",
			"purged_ids": []int{},
		})
		return
	}

	var purgedIDs []int
	for _, m := range obsoleteMaps {
		if err := h.mapRepo.Delete(m.ID); err == nil {
			purgedIDs = append(purgedIDs, m.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("purged %d obsolete maps without zones", len(purgedIDs)),
		"purged_ids": purgedIDs,
	})
}

func (h *MapHandler) ServeImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_map_id",
			"message": "invalid map id",
		})
		return
	}

	blob, err := h.mapRepo.GetBlob(id)
	if err == models.ErrMapNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "map_not_found",
			"message": "map not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to retrieve map image",
		})
		return
	}

	if blob == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "map_not_found",
			"message": "map has no image data",
		})
		return
	}

	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Data(http.StatusOK, "image/webp", blob)
}
