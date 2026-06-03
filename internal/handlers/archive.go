package handlers

import (
	"net/http"

	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
)

type ArchiveHandler struct {
	logger services.Logger
}

func NewArchiveHandler(logger services.Logger) *ArchiveHandler {
	return &ArchiveHandler{logger: logger}
}

func (h *ArchiveHandler) ListLeagues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})
}

func (h *ArchiveHandler) CreateLeague(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})
}

func (h *ArchiveHandler) ListEvaTeams(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})
}

func (h *ArchiveHandler) CreateEvaTeam(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})
}
