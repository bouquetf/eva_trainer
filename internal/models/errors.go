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
	ErrRoleNotFound            = errors.New("role not found")
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
	ErrLigueLocaleNameRequired      = errors.New("ligue locale name is required")
	ErrLigueLocaleNotFound          = errors.New("ligue locale not found")
	ErrLigueLocaleEmplacementExists = errors.New("a ligue locale with this emplacement already exists")
)

var (
	ErrLigueNameRequired = errors.New("ligue name is required")
	ErrLigueNotFound     = errors.New("ligue not found")
	ErrLigueInvalidDates = errors.New("ligue start date must be before end date")
)

var (
	ErrSplitNameRequired     = errors.New("split name is required")
	ErrSplitMapNamesRequired = errors.New("split must have at least one map name")
	ErrSplitNotFound         = errors.New("split not found")
)

var (
	ErrZoneNotFound           = errors.New("zone not found")
	ErrZoneNameRequired       = errors.New("zone name is required")
	ErrZoneNameTooLong        = errors.New("zone name exceeds maximum length of 250 characters")
	ErrZoneNameTaken          = errors.New("zone name already exists for this team on this map")
	ErrPolygonNotFound        = errors.New("polygon not found")
	ErrPolygonHasTooFewPoints = errors.New("polygon must have at least 3 points")
	ErrLockNotFound           = errors.New("edit lock not found")
	ErrLockAlreadyAcquired    = errors.New("edit lock already acquired by another editor")
	ErrLockNotOwned           = errors.New("edit lock is owned by another editor")

	ErrTemplateNotFound      = errors.New("template not found")
	ErrTemplateAlreadyExists = errors.New("template already exists for this map")

	ErrVersionNotFound   = errors.New("version not found")
	ErrInvalidFloorValue = errors.New("invalid floor value: must be 'haut', 'bas', or empty")
)

var (
	ErrLineOfFireNotFound     = errors.New("line of fire not found")
	ErrLineOfFireNameTooLong  = errors.New("line of fire name exceeds maximum length of 250 characters")
	ErrLineOfFireNameTaken    = errors.New("line of fire name already exists for this team on this map")
	ErrLineOfFireNameRequired = errors.New("line of fire name is required")
	ErrLineOfFireNoCones      = errors.New("line of fire must have at least one cone")
	ErrConeOriginRequired     = errors.New("cone must have an origin point")
)

var (
	ErrDiscordUnavailable  = errors.New("discord API is unavailable")
	ErrDatabaseUnavailable = errors.New("database is unavailable")
	ErrInternalServer      = errors.New("internal server error")
	ErrForbidden           = errors.New("forbidden")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConflict            = errors.New("resource conflict")
	ErrValidation          = errors.New("validation error")
	ErrBadRequest          = errors.New("bad request")
	ErrRateLimited         = errors.New("too many requests")
)
