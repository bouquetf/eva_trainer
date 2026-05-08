package models

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidDiscordID     = errors.New("discord id cannot be empty")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
)
