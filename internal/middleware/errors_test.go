package middleware

import (
	"errors"
	"net/http"
	"testing"

	"eva_trainer/internal/models"
	"eva_trainer/pkg/apierror"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apierror.ErrorCode
		wantHTTP int
	}{
		// Not found errors
		{"user not found", models.ErrUserNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"team not found", models.ErrTeamNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"map not found", models.ErrMapNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"split not found", models.ErrSplitNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"zone not found", models.ErrZoneNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"invitation not found", models.ErrInvitationNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"join request not found", models.ErrJoinRequestNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"template not found", models.ErrTemplateNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"version not found", models.ErrVersionNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"member not found", models.ErrMemberNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"permission not found", models.ErrPermissionNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"lock not found", models.ErrLockNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},
		{"polygon not found", models.ErrPolygonNotFound, apierror.ErrCodeNotFound, http.StatusNotFound},

		// Bad request errors
		{"invalid discord id", models.ErrInvalidDiscordID, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"team name required", models.ErrTeamNameRequired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"invalid team name", models.ErrInvalidTeamName, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"team name too long", models.ErrTeamNameTooLong, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"map name required", models.ErrMapNameRequired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"zone name required", models.ErrZoneNameRequired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"zone name too long", models.ErrZoneNameTooLong, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"polygon has too few points", models.ErrPolygonHasTooFewPoints, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"invalid role", models.ErrInvalidRole, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"invitation expired", models.ErrInvitationExpired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"cannot invite self", models.ErrCannotInviteSelf, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"split name required", models.ErrSplitNameRequired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"split map names required", models.ErrSplitMapNamesRequired, apierror.ErrCodeBadRequest, http.StatusBadRequest},
		{"bad request generic", models.ErrBadRequest, apierror.ErrCodeBadRequest, http.StatusBadRequest},

		// Unauthorized errors
		{"refresh token not found", models.ErrRefreshTokenNotFound, apierror.ErrCodeUnauthorized, http.StatusUnauthorized},
		{"refresh token expired", models.ErrRefreshTokenExpired, apierror.ErrCodeUnauthorized, http.StatusUnauthorized},
		{"refresh token revoked", models.ErrRefreshTokenRevoked, apierror.ErrCodeUnauthorized, http.StatusUnauthorized},
		{"unauthorized generic", models.ErrUnauthorized, apierror.ErrCodeUnauthorized, http.StatusUnauthorized},

		// Forbidden errors
		{"not captain", models.ErrNotCaptain, apierror.ErrCodeForbidden, http.StatusForbidden},
		{"not member", models.ErrNotMember, apierror.ErrCodeForbidden, http.StatusForbidden},
		{"lock not owned", models.ErrLockNotOwned, apierror.ErrCodeForbidden, http.StatusForbidden},
		{"forbidden generic", models.ErrForbidden, apierror.ErrCodeForbidden, http.StatusForbidden},

		// Conflict errors
		{"zone name taken", models.ErrZoneNameTaken, apierror.ErrCodeConflict, http.StatusConflict},
		{"team already has cocaptain", models.ErrTeamAlreadyHasCocaptain, apierror.ErrCodeConflict, http.StatusConflict},
		{"duplicate team member", models.ErrDuplicateTeamMember, apierror.ErrCodeConflict, http.StatusConflict},
		{"invitation already used", models.ErrInvitationAlreadyUsed, apierror.ErrCodeConflict, http.StatusConflict},
		{"join request already exists", models.ErrJoinRequestAlreadyExists, apierror.ErrCodeConflict, http.StatusConflict},
		{"join request already member", models.ErrJoinRequestAlreadyMember, apierror.ErrCodeConflict, http.StatusConflict},
		{"lock already acquired", models.ErrLockAlreadyAcquired, apierror.ErrCodeConflict, http.StatusConflict},
		{"template already exists", models.ErrTemplateAlreadyExists, apierror.ErrCodeConflict, http.StatusConflict},
		{"conflict generic", models.ErrConflict, apierror.ErrCodeConflict, http.StatusConflict},

		// Validation
		{"validation error", models.ErrValidation, apierror.ErrCodeValidation, http.StatusUnprocessableEntity},

		// Discord error
		{"discord unavailable", models.ErrDiscordUnavailable, apierror.ErrCodeDiscordAPI, http.StatusBadGateway},

		// Rate limited
		{"rate limited", models.ErrRateLimited, apierror.ErrCodeRateLimited, http.StatusTooManyRequests},

		// Default / internal errors
		{"database unavailable", models.ErrDatabaseUnavailable, apierror.ErrCodeInternal, http.StatusInternalServerError},
		{"internal server error", models.ErrInternalServer, apierror.ErrCodeInternal, http.StatusInternalServerError},
		{"unknown error", errors.New("unknown error"), apierror.ErrCodeInternal, http.StatusInternalServerError},

		// Nil
		{"nil error", nil, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapError(tt.err)

			if tt.err == nil {
				if result != nil {
					t.Errorf("expected nil for nil input, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if result.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", result.Code, tt.wantCode)
			}
			if result.HTTPCode != tt.wantHTTP {
				t.Errorf("HTTPCode = %d, want %d", result.HTTPCode, tt.wantHTTP)
			}

			expectedMsg := tt.err.Error()
			if result.Code == apierror.ErrCodeInternal && tt.err != models.ErrInternalServer {
				expectedMsg = "internal server error"
			}
			if result.Message != expectedMsg {
				t.Errorf("Message = %q, want %q", result.Message, expectedMsg)
			}
		})
	}
}

func TestMapErrorWithMessage(t *testing.T) {
	customMsg := "custom error message"
	result := MapErrorWithMessage(models.ErrTeamNotFound, customMsg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Message != customMsg {
		t.Errorf("Message = %q, want %q", result.Message, customMsg)
	}
	if result.Code != apierror.ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", result.Code, apierror.ErrCodeNotFound)
	}

	nilResult := MapErrorWithMessage(nil, "should be nil")
	if nilResult != nil {
		t.Errorf("expected nil for nil input, got %v", nilResult)
	}
}
