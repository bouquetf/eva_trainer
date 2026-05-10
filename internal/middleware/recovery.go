package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(logger services.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%v", v)
				}

				logger.With(
					"panic", err.Error(),
					"stack", stack,
				).Error("recovered from panic")

				apierror.RespondWithError(c, apierror.NewAPIError(
					apierror.ErrCodeInternal,
					"internal server error",
					http.StatusInternalServerError,
				))
			}
		}()

		c.Next()
	}
}
