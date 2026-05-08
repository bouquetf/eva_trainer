package models

import "time"

type UserRepository interface {
	FindOrCreateByDiscordID(discordID string) (*User, bool, error)
	FindByID(userID int) (*User, error)
	SetAdmin(userID int) error
	IsFirstUser() (bool, error)
}

type RefreshTokenRepository interface {
	CreateRefreshToken(userID int, expiresAt time.Time) (tokenPlain string, err error)
	FindRefreshTokenByHash(tokenHash string) (*RefreshToken, error)
	RotateRefreshToken(oldTokenHash string, newExpiresAt time.Time) (newTokenPlain string, err error)
	RevokeRefreshToken(tokenHash string) error
	RevokeAllUserTokens(userID int) error
}
