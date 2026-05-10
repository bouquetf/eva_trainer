package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	mapRepo models.MapRepository
	logger  services.Logger
}

func NewMapHandler(mapRepo models.MapRepository, logger services.Logger) *MapHandler {
	return &MapHandler{mapRepo: mapRepo, logger: logger}
}

func (h *MapHandler) UploadMap(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("file_too_large", "file exceeds maximum size of 50 MB", http.StatusBadRequest))
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("map_name_required", "map name is required", http.StatusBadRequest))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("file_required", "map file is required", http.StatusBadRequest))
		return
	}

	if ext := filepath.Ext(file.Filename); !strings.EqualFold(ext, ".webp") {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_file_format", "only .webp files are accepted", http.StatusBadRequest))
		return
	}

	if file.Header.Get("Content-Type") != "image/webp" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_file_format", "Content-Type must be image/webp", http.StatusBadRequest))
		return
	}

	f, err := file.Open()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to read uploaded file", http.StatusInternalServerError))
		return
	}
	defer f.Close()

	blobData, err := io.ReadAll(f)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to read file content", http.StatusInternalServerError))
		return
	}

	existing, _ := h.mapRepo.FindByNameLatest(name)
	var previousVersion *previousVersionInfo
	if existing != nil && !existing.IsObsolete {
		previousVersion = &previousVersionInfo{ID: existing.ID, Name: existing.Name}
	}

	m, err := h.mapRepo.Create(name, blobData)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", fmt.Sprintf("failed to create map: %v", err), http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{
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
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list maps", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"maps": maps})
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
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to list maps", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"maps": maps})
}

func (h *MapHandler) GetMapDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid map id", http.StatusBadRequest))
		return
	}

	m, err := h.mapRepo.FindByID(id)
	if err == models.ErrMapNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find map", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"map": m})
}

func (h *MapHandler) PurgeObsoleteMaps(c *gin.Context) {
	obsoleteMaps, err := h.mapRepo.FindObsoleteWithoutZones()
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to find obsolete maps", http.StatusInternalServerError))
		return
	}

	if len(obsoleteMaps) == 0 {
		apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
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

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"message":    fmt.Sprintf("purged %d obsolete maps without zones", len(purgedIDs)),
		"purged_ids": purgedIDs,
	})
}

func (h *MapHandler) ServeImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_map_id", "invalid map id", http.StatusBadRequest))
		return
	}

	blob, err := h.mapRepo.GetBlob(id)
	if err == models.ErrMapNotFound {
		apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map not found", http.StatusNotFound))
		return
	}
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "failed to retrieve map image", http.StatusInternalServerError))
		return
	}

	if blob == nil {
		apierror.RespondWithError(c, apierror.NewAPIError("map_not_found", "map has no image data", http.StatusNotFound))
		return
	}

	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Data(http.StatusOK, "image/webp", blob)
}
