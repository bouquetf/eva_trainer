package api

import (
	"eva_trainer/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/health", handlers.HealthCheck)

	apiV1 := r.Group("/api/v1")
	_ = apiV1
}
