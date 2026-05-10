package middleware

import (
	"net/http"

	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(logger services.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		if !user.IsAdmin {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "admin access required", http.StatusForbidden))
			return
		}

		c.Next()
	}
}
