package middleware

import (
	"errors"
	"net/http"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func MapError(err error) *apierror.APIError {
	if err == nil {
		return nil
	}

	switch {
	case isNotFound(err):
		return apierror.NewAPIError(apierror.ErrCodeNotFound, err.Error(), http.StatusNotFound)
	case isBadRequest(err):
		return apierror.NewAPIError(apierror.ErrCodeBadRequest, err.Error(), http.StatusBadRequest)
	case isUnauthorized(err):
		return apierror.NewAPIError(apierror.ErrCodeUnauthorized, err.Error(), http.StatusUnauthorized)
	case isForbidden(err):
		return apierror.NewAPIError(apierror.ErrCodeForbidden, err.Error(), http.StatusForbidden)
	case isConflict(err):
		return apierror.NewAPIError(apierror.ErrCodeConflict, err.Error(), http.StatusConflict)
	case isValidation(err):
		return apierror.NewAPIError(apierror.ErrCodeValidation, err.Error(), http.StatusUnprocessableEntity)
	case isDiscordError(err):
		return apierror.NewAPIError(apierror.ErrCodeDiscordAPI, err.Error(), http.StatusBadGateway)
	case isRateLimited(err):
		return apierror.NewAPIError(apierror.ErrCodeRateLimited, err.Error(), http.StatusTooManyRequests)
	default:
		return apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError)
	}
}

func MapErrorWithMessage(err error, message string) *apierror.APIError {
	apiErr := MapError(err)
	if apiErr == nil {
		return nil
	}
	apiErr.Message = message
	return apiErr
}

func isNotFound(err error) bool {
	return errors.Is(err, models.ErrUserNotFound) ||
		errors.Is(err, models.ErrTeamNotFound) ||
		errors.Is(err, models.ErrMapNotFound) ||
		errors.Is(err, models.ErrSplitNotFound) ||
		errors.Is(err, models.ErrZoneNotFound) ||
		errors.Is(err, models.ErrInvitationNotFound) ||
		errors.Is(err, models.ErrJoinRequestNotFound) ||
		errors.Is(err, models.ErrTemplateNotFound) ||
		errors.Is(err, models.ErrVersionNotFound) ||
		errors.Is(err, models.ErrMemberNotFound) ||
		errors.Is(err, models.ErrPermissionNotFound) ||
		errors.Is(err, models.ErrLockNotFound) ||
		errors.Is(err, models.ErrPolygonNotFound)
}

func isBadRequest(err error) bool {
	return errors.Is(err, models.ErrInvalidDiscordID) ||
		errors.Is(err, models.ErrTeamNameRequired) ||
		errors.Is(err, models.ErrInvalidTeamName) ||
		errors.Is(err, models.ErrTeamNameTooLong) ||
		errors.Is(err, models.ErrMapNameRequired) ||
		errors.Is(err, models.ErrZoneNameRequired) ||
		errors.Is(err, models.ErrZoneNameTooLong) ||
		errors.Is(err, models.ErrPolygonHasTooFewPoints) ||
		errors.Is(err, models.ErrInvalidRole) ||
		errors.Is(err, models.ErrInvitationExpired) ||
		errors.Is(err, models.ErrCannotInviteSelf) ||
		errors.Is(err, models.ErrSplitNameRequired) ||
		errors.Is(err, models.ErrSplitMapNamesRequired) ||
		errors.Is(err, models.ErrBadRequest)
}

func isUnauthorized(err error) bool {
	return errors.Is(err, models.ErrRefreshTokenNotFound) ||
		errors.Is(err, models.ErrRefreshTokenExpired) ||
		errors.Is(err, models.ErrRefreshTokenRevoked) ||
		errors.Is(err, models.ErrUnauthorized)
}

func isForbidden(err error) bool {
	return errors.Is(err, models.ErrNotCaptain) ||
		errors.Is(err, models.ErrNotMember) ||
		errors.Is(err, models.ErrLockNotOwned) ||
		errors.Is(err, models.ErrForbidden)
}

func isConflict(err error) bool {
	return errors.Is(err, models.ErrZoneNameTaken) ||
		errors.Is(err, models.ErrTeamAlreadyHasCocaptain) ||
		errors.Is(err, models.ErrDuplicateTeamMember) ||
		errors.Is(err, models.ErrInvitationAlreadyUsed) ||
		errors.Is(err, models.ErrJoinRequestAlreadyExists) ||
		errors.Is(err, models.ErrJoinRequestAlreadyMember) ||
		errors.Is(err, models.ErrLockAlreadyAcquired) ||
		errors.Is(err, models.ErrTemplateAlreadyExists) ||
		errors.Is(err, models.ErrConflict)
}

func isValidation(err error) bool {
	return errors.Is(err, models.ErrValidation)
}

func isDiscordError(err error) bool {
	return errors.Is(err, models.ErrDiscordUnavailable)
}

func isRateLimited(err error) bool {
	return errors.Is(err, models.ErrRateLimited)
}

func ErrorHandlerMiddleware(logger services.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("handler error", "error", e.Err.Error())
			}

			if !c.Writer.Written() {
				lastErr := c.Errors.Last().Err
				apiErr := MapError(lastErr)
				apierror.RespondWithError(c, apiErr)
			}
		}
	}
}
