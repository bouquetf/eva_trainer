package middleware

import (
	"net/http"
	"strings"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func RequireAuth(logger services.Logger, userRepo models.UserRepository, jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := extractAndValidateUser(c, logger, userRepo, jwtService)
		if err != nil {
			apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrorCode(err.Code), err.Message, http.StatusUnauthorized))
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

type authError struct {
	Code    string
	Message string
}

func extractAndValidateUser(c *gin.Context, logger services.Logger, userRepo models.UserRepository, jwtService *services.JWTService) (*models.User, *authError) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, &authError{"missing_authorization", "missing authorization header"}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, &authError{"invalid_authorization", "invalid authorization header format"}
	}

	claims, err := jwtService.ValidateAccessToken(parts[1])
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "token is expired") {
			return nil, &authError{"token_expired", "token expired"}
		}
		return nil, &authError{"invalid_token", "invalid token"}
	}

	user, err := userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, &authError{"user_not_found", "user not found"}
	}

	return user, nil
}

func OptionalAuth(logger services.Logger, userRepo models.UserRepository, jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := extractAndValidateUser(c, logger, userRepo, jwtService)
		if err != nil {
			c.Next()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func GetUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	u, ok := user.(*models.User)
	return u, ok
}

func MustGetUser(c *gin.Context) (*models.User, error) {
	user, exists := GetUser(c)
	if !exists {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return nil, &authError{"unauthorized", "unauthorized"}
	}
	return user, nil
}

func (e *authError) Error() string {
	return e.Message
}
