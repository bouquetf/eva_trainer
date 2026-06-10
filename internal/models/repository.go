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
	GetUserSettings(userID int) (*UserSettings, error)
	UpdateUserSettings(userID int, settings *UserSettings) error
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

type LigueLocaleRepository interface {
	Create(nom string, emplacement string) (*LigueLocale, error)
	FindByID(id int) (*LigueLocale, error)
	FindAll(emplacement string) ([]*LigueLocale, error)
	FindByLigueID(ligueID int, page, perPage int) ([]*LigueLocale, int, error)
	Update(id int, nom string, emplacement string) (*LigueLocale, error)
	AssociateLigue(localeID, ligueID int) error
	DisassociateLigue(localeID, ligueID int) error
	FindLiguesByLocaleID(localeID int) ([]*Ligue, error)
	Delete(id int) error
}

type LigueRepository interface {
	Create(dateDebut, dateFin time.Time, splitID int) (*Ligue, error)
	FindByID(id int) (*Ligue, error)
	FindAll() ([]*Ligue, error)
	FindAllPaginated(page, perPage int) ([]*Ligue, int, error)
	Update(id int, dateDebut, dateFin time.Time, splitID int) (*Ligue, error)
	Delete(id int) error
	AssociateAllLocales(ligueID int) (int64, error)
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

type ImportPolygonData struct {
	Points []Point `json:"points"`
}

type ImportZoneData struct {
	Name     string              `json:"name"`
	Polygons []ImportPolygonData `json:"polygons"`
}

type ZoneRepository interface {
	Create(teamID, mapID int) (*Zone, error)
	FindByID(zoneID int) (*Zone, error)
	FindByTeamAndMap(teamID, mapID int) ([]*Zone, error)
	FindWithDetailsByTeamAndMap(teamID, mapID int) ([]*ZoneWithDetails, error)
	FindWithDetailsByTeam(teamID int) (map[int][]*ZoneWithDetails, error)
	ImportZones(teamID, mapID int, zones []ImportZoneData) ([]*Zone, error)
	Delete(zoneID int) error

	AddPolygon(zoneID int, points []Point, orderIndex int, floor *string) (*ZonePolygon, error)
	GetPolygons(zoneID int) ([]ZonePolygon, error)
	UpdatePolygonPoints(polygonID int, points []Point) error
	RemovePolygon(polygonID int) error
	PolygonCount(zoneID int) (int, error)
	SetPolygonFloor(polygonID int, floor *string) error

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
	FindAllTemplates() ([]*TemplateWithZones, error)
	UpdateFromMap(templateID, teamID, mapID int) error
	Delete(templateID int) error
	DuplicateToTeam(teamID, templateID int) ([]*Zone, int, error)
	FindTemplateZonesByMapID(mapID int) ([]TemplateZone, error)
	FindTemplateLinesOfFireByMapID(mapID int) ([]TemplateLineOfFire, error)
	SetSyncStatus(templateID int, active bool) error
}

type ZoneVersionRepository interface {
	Save(zoneID int, editorID *int, snapshot VersionSnapshot) (*ZoneVersion, error)
	FindByMap(teamID, mapID int, page, limit int) ([]*ZoneVersion, int, error)
	FindByZone(zoneID int, page, limit int) ([]*ZoneVersion, int, error)
	FindByID(versionID int) (*ZoneVersion, error)
	GetLatestForZone(zoneID int) (*ZoneVersion, error)
	FindFirstOtherEditorVersion(zoneID, editorID int) (*int, error)
	CountByZone(zoneID int) (int, error)
	DeleteNewerThan(zoneID int, olderThan time.Time) error
}

type LineOfFireRepository interface {
	Create(teamID, mapID int, name string, data ConeData) (*LineOfFire, error)
	FindByID(id int) (*LineOfFire, error)
	FindByTeamAndMap(teamID, mapID int) ([]*LineOfFire, error)
	UpdateName(id int, name string) error
	UpdateOrder(id int, orderIndex int) error
	UpdatePolygonData(id int, data ConeData) error
	Delete(id int) error
	IsNameTaken(teamID, mapID int, name string) (bool, error)
}

type LineOfFireEditHistoryRepository interface {
	Log(lineOfFireID int, editorID *int, action string, oldData, newData json.RawMessage) error
}

type LineOfFireVersionRepository interface {
	Save(lineOfFireID int, editorID *int, snapshot json.RawMessage) (*LineOfFireVersion, error)
	FindByLineOfFire(lineOfFireID int) ([]*LineOfFireVersion, error)
	GetLatestForLineOfFire(lineOfFireID int) (*LineOfFireVersion, error)
	DeleteNewerThan(lineOfFireID int, olderThan time.Time) error
}

type GlobalRoleRepository interface {
	Add(userID int, role string) error
	Remove(userID int, role string) error
	UserHasRole(userID int, role string) (bool, error)
	FindUsersByRole(role string) ([]*UserGlobalRole, error)
}

type TeamRoleRepository interface {
	Add(teamID int, role string) error
	Remove(teamID int, role string) error
	TeamHasRole(teamID int, role string) (bool, error)
	FindTeamsByRole(role string) ([]*TeamWithRole, error)
	UserIsMemberOfTeamWithRole(userID int, role string) (bool, error)
}

type EvaTeamRepository interface {
	Create(name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
	FindByID(id int) (*EvaTeam, error)
	FindAll() ([]*EvaTeam, error)
	Update(id int, name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
	Delete(id int) error
	HasReferences(id int) (bool, error)
}

type DivRepository interface {
	Create(ligueID int, nom string) (*Div, error)
	FindByID(id int) (*Div, error)
	FindByLigueID(ligueID int) ([]*Div, error)
	Update(id int, nom string) (*Div, error)
	Delete(id int) error
	AssignTeam(divID, equipeEvaID int) error
	RemoveTeam(divID, equipeEvaID int) error
	FindTeamsByDivID(divID int) ([]*EvaTeam, error)
	FindTeamsByLigueID(ligueID int) (map[int][]*EvaTeam, error)
	GetStandings(divID int) ([]*StandingEntry, error)
	GenerateRoundRobin(divID int) ([]*BO, error)
}

type BORepository interface {
	Create(ligueID, equipe1ID, equipe2ID int, format string, forfait int) (*BO, error)
	Delete(id int) error
	FindByID(id int) (*BO, error)
	FindByLigueID(ligueID int) ([]*BO, error)
	FindWithDetails(id int) (*BOWithDetails, error)
	FindByLigueIDWithTeams(ligueID int) ([]*BOListItem, error)
	Update(id int, ligueID, equipe1ID, equipe2ID int, format string, forfait int) (*BO, error)
	RecomputeWinner(id int) (*BO, error)
}

type MatchRepository interface {
	Create(boID, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string) (*Match, error)
	FindByID(id int) (*Match, error)
	FindByBOID(boID int) ([]*Match, error)
	Update(id, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string) (*Match, error)
	Delete(id int) error
	DeleteByBOID(boID int) error
}

type CardStatsRepository interface {
	Compute(teamID int) ([]*CardStats, error)
}

type HeadToHeadStatsRepository interface {
	Compute(team1ID, team2ID int) ([]*HeadToHeadMapStats, error)
}

type BanRepository interface {
	Create(boID, equipeID, mapID int) (*Ban, error)
	FindByID(id int) (*Ban, error)
	FindByBOID(boID int) ([]*Ban, error)
	Delete(id int) error
	DeleteByBOID(boID int) error
}

type PickRepository interface {
	Create(boID, equipeID, mapID int) (*Pick, error)
	FindByID(id int) (*Pick, error)
	FindByBOID(boID int) ([]*Pick, error)
	Delete(id int) error
	DeleteByBOID(boID int) error
}

type EditHistoryRepository interface {
	Log(zoneID int, editorID *int, action string, oldData, newData json.RawMessage) (*EditHistory, error)
	FindByZone(zoneID int) ([]*EditHistory, error)
	FindByTeamAndMap(teamID, mapID int) ([]*EditHistory, error)
	FindByTeamAndMapWithFilters(teamID, mapID int, page, limit int) ([]*EditHistory, int, error)
	FindByZoneWithFilters(zoneID int, page, limit int) ([]*EditHistory, int, error)
}
