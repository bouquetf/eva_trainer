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
	ErrPermissionNotFound      = errors.New("permission not found")
)

var (
	ErrMapNameRequired = errors.New("map name is required")
	ErrMapNotFound     = errors.New("map not found")
)

var (
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed = errors.New("invitation has already been used")
	ErrCannotInviteSelf      = errors.New("cannot invite yourself")
)

var (
	ErrJoinRequestNotFound      = errors.New("join request not found")
	ErrJoinRequestAlreadyExists = errors.New("you already have a pending join request for this team")
	ErrJoinRequestAlreadyMember = errors.New("you are already a member of this team")
)

var (
	ErrSplitNameRequired     = errors.New("split name is required")
	ErrSplitMapNamesRequired = errors.New("split must have at least one map name")
	ErrSplitNotFound         = errors.New("split not found")
)
