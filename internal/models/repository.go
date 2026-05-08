package models

import "time"

type UserRepository interface {
	FindOrCreateByDiscordID(discordID string) (*User, bool, error)
	FindByID(userID int) (*User, error)
	SetAdmin(userID int) error
	UpdateDisplayName(userID int, displayName string) error
	SetDiscordUsername(userID int, discordUsername string) error
	FindAll() ([]*User, error)
	RemoveAdmin(userID int) error
	IsFirstUser() (bool, error)
	AdminCount() (int, error)
}

type RefreshTokenRepository interface {
	CreateRefreshToken(userID int, expiresAt time.Time) (tokenPlain string, err error)
	FindRefreshTokenByHash(tokenHash string) (*RefreshToken, error)
	RotateRefreshToken(oldTokenHash string, newExpiresAt time.Time) (newTokenPlain string, err error)
	RevokeRefreshToken(tokenHash string) error
	RevokeAllUserTokens(userID int) error
}

type TeamRepository interface {
	Create(name string, captainID int) (*Team, error)
	FindByID(id int) (*Team, error)
	FindAll() ([]*Team, error)
	FindByUserID(userID int) ([]*Team, error)
	UpdateName(id int, name string) error
	UpdateCaptain(teamID, newCaptainID int) error
	TransferCaptaincy(teamID, oldCaptainID, newCaptainID int) error
	Delete(id int) error
}

type TeamMemberRepository interface {
	AddMember(teamID, userID int, role string) error
	RemoveMember(teamID, userID int) error
	UpdateRole(teamID, userID int, role string) error
	FindByTeamID(teamID int) ([]*TeamMember, error)
	FindByTeamIDWithDisplayNames(teamID int) ([]*TeamMemberWithDisplayName, error)
	FindByUserID(userID int) ([]*TeamMember, error)
	FindByUserAndTeam(userID, teamID int) (*TeamMember, error)
	GetCaptainID(teamID int) (int, error)
	CountByRole(teamID int, role string) (int, error)
}

type PermissionRepository interface {
	FindAll() ([]*Permission, error)
	FindByID(id int) (*Permission, error)
	FindByFeature(feature string) (*Permission, error)
	Create(feature, description string) (*Permission, error)
}

type MapRepository interface {
	Create(name string, blobData []byte) (*Map, error)
	FindByID(id int) (*Map, error)
	FindAll() ([]*Map, error)
	FindByName(name string) ([]*Map, error)
	FindByNameLatest(name string) (*Map, error)
	MarkObsolete(mapID int, replacedByMapID int) error
	IsMapObsolete(mapID int) (bool, error)
	FindObsoleteWithoutZones() ([]*Map, error)
	Delete(mapID int) error
	GetBlob(mapID int) ([]byte, error)
}

type RolePermissionRepository interface {
	FindByRole(role string) ([]*Permission, error)
	FindByPermissionID(permissionID int) ([]*RolePermission, error)
	AssignPermissionToRole(role string, permissionID int) error
	RemovePermissionFromRole(role string, permissionID int) error
	FindAllRoles() ([]string, error)
	RoleHasPermission(role string, permissionID int) (bool, error)
}
