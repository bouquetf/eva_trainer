package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := MustGetUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
