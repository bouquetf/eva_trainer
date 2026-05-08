package models

import "time"

type UserRepository interface {
	FindOrCreateByDiscordID(discordID string) (*User, bool, error)
	FindByID(userID int) (*User, error)
	SetAdmin(userID int) error
	UpdateDisplayName(userID int, displayName string) error
	SetDiscordUsername(userID int, discordUsername string) error
	IsFirstUser() (bool, error)
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
	FindByUserID(userID int) ([]*Team, error)
	UpdateName(id int, name string) error
	Delete(id int) error
}

type TeamMemberRepository interface {
	AddMember(teamID, userID int, role string) error
	RemoveMember(teamID, userID int) error
	UpdateRole(teamID, userID int, role string) error
	FindByTeamID(teamID int) ([]*TeamMember, error)
	FindByTeamIDWithDisplayNames(teamID int) ([]*TeamMemberWithDisplayName, error)
	FindByUserAndTeam(userID, teamID int) (*TeamMember, error)
	GetCaptainID(teamID int) (int, error)
	CountByRole(teamID int, role string) (int, error)
}
