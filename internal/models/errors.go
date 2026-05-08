package models

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidDiscordID     = errors.New("discord id cannot be empty")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
)

var (
	ErrTeamNotFound            = errors.New("team not found")
	ErrTeamNameRequired        = errors.New("team name is required")
	ErrInvalidTeamName         = errors.New("team name contains invalid characters")
	ErrTeamNameTooLong         = errors.New("team name exceeds maximum length of 250 characters")
	ErrMemberNotFound          = errors.New("member not found in team")
	ErrInvalidRole             = errors.New("invalid team role")
	ErrTeamAlreadyHasCocaptain = errors.New("team already has a cocaptain")
	ErrNotCaptain              = errors.New("user is not the team captain")
	ErrNotMember               = errors.New("user is not a team member")
	ErrDuplicateTeamMember     = errors.New("user is already a member of this team")
)
