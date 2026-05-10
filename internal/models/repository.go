package models

import (
	"encoding/json"
	"time"
)

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
	FindByDisplayNameOrDiscordUsername(query string) (*User, error)
	SearchUsers(query string, limit int) ([]*User, error)
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
	FindAllPaginated(q string, page, perPage int) ([]*Team, int, error)
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

type TeamInvitationRepository interface {
	CreateByUser(teamID, invitedUserID, invitedByUserID int) (*TeamInvitation, error)
	CreateLink(teamID, invitedByUserID int, token string, expiresAt time.Time) (*TeamInvitation, error)
	FindByID(id int) (*TeamInvitation, error)
	FindByToken(token string) (*TeamInvitation, error)
	FindPendingByTeam(teamID int) ([]*TeamInvitation, error)
	FindPendingByUser(userID int) ([]*TeamInvitation, error)
	AcceptInvitation(id int) error
	AcceptByToken(token string, userID int) error
	DeclineInvitation(id int) error
	CancelInvitation(id int) error
	ExpireOverdue() (int64, error)
}

type TeamJoinRequestRepository interface {
	Create(teamID, userID int) (*TeamJoinRequest, error)
	FindByID(id int) (*TeamJoinRequest, error)
	FindPendingByTeam(teamID int) ([]*TeamJoinRequest, error)
	FindPendingByUser(userID int) ([]*TeamJoinRequest, error)
	FindByUserAndTeam(userID, teamID int) (*TeamJoinRequest, error)
	AcceptRequest(id int) error
	DeclineRequest(id int) error
	CancelRequest(id int) error
}

type MapRepository interface {
	Create(name string, blobData []byte) (*Map, error)
	FindByID(id int) (*Map, error)
	FindAll() ([]*Map, error)
	FindAllActive() ([]*Map, error)
	FindByName(name string) ([]*Map, error)
	FindByNameLatest(name string) (*Map, error)
	MarkObsolete(mapID int, replacedByMapID int) error
	IsMapObsolete(mapID int) (bool, error)
	FindObsoleteWithoutZones() ([]*Map, error)
	Delete(mapID int) error
	GetBlob(mapID int) ([]byte, error)
}

type SplitRepository interface {
	Create(name string, mapNames []string) (*Split, error)
	FindByID(id int) (*Split, error)
	FindAll() ([]*Split, error)
	Update(id int, name string, mapNames []string) (*Split, error)
	Delete(id int) error
	FindMapsBySplitID(splitID int) ([]*Map, error)
}

type MapCatalogRepository interface {
	FindAll() ([]*MapCatalogEntry, error)
	Create(name string) (*MapCatalogEntry, error)
	Update(id int, name string) (*MapCatalogEntry, error)
	Delete(id int) error
}

type RolePermissionRepository interface {
	FindByRole(role string) ([]*Permission, error)
	FindByPermissionID(permissionID int) ([]*RolePermission, error)
	AssignPermissionToRole(role string, permissionID int) error
	RemovePermissionFromRole(role string, permissionID int) error
	FindAllRoles() ([]string, error)
	RoleHasPermission(role string, permissionID int) (bool, error)
}

type ZoneRepository interface {
	Create(teamID, mapID int) (*Zone, error)
	FindByID(zoneID int) (*Zone, error)
	FindByTeamAndMap(teamID, mapID int) ([]*Zone, error)
	FindWithDetailsByTeamAndMap(teamID, mapID int) ([]*ZoneWithDetails, error)
	Delete(zoneID int) error

	AddPolygon(zoneID int, points []Point, orderIndex int) (*ZonePolygon, error)
	GetPolygons(zoneID int) ([]ZonePolygon, error)
	UpdatePolygonPoints(polygonID int, points []Point) error
	RemovePolygon(polygonID int) error
	PolygonCount(zoneID int) (int, error)

	SetName(zoneID int, name string) error
	GetName(zoneID int) (*ZoneName, error)
	IsNameTaken(teamID, mapID int, name string) (bool, error)
}

type EditLockRepository interface {
	Acquire(teamID, mapID, editorID int) (*EditLock, error)
	Release(teamID, mapID int) error
	ReleaseByEditor(teamID, mapID, editorID int) error
	Get(teamID, mapID int) (*EditLock, error)
	Refresh(teamID, mapID int) (*EditLock, error)
	IsLocked(teamID, mapID int) (bool, error)
	IsLockedByEditor(teamID, mapID, editorID int) (bool, error)
	CleanupExpired() (int64, error)
	ForceClearAll() (int64, error)
}

type TemplateRepository interface {
	CreateFromMap(teamID, mapID int) (*Template, error)
	FindByMapID(mapID int) (*Template, error)
	FindAllPublic() ([]*TemplateWithZones, error)
	UpdateFromMap(templateID, teamID, mapID int) error
	Delete(templateID int) error
}

type EditHistoryRepository interface {
	Log(zoneID int, editorID *int, action string, oldData, newData json.RawMessage) (*EditHistory, error)
	FindByZone(zoneID int) ([]*EditHistory, error)
	FindByTeamAndMap(teamID, mapID int) ([]*EditHistory, error)
}
