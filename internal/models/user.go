package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID              int       `json:"id"`
	DiscordID       string    `json:"discord_id"`
	DiscordUsername string    `json:"discord_username"`
	DisplayName     string    `json:"display_name"`
	IsAdmin         bool      `json:"is_admin"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindOrCreateByDiscordID(discordID string) (*User, bool, error) {
	if discordID == "" {
		return nil, false, ErrInvalidDiscordID
	}

	var user User
	err := r.db.QueryRow(
		`SELECT id, discord_id, discord_username, display_name, is_admin, created_at, updated_at FROM users WHERE discord_id = $1`,
		discordID,
	).Scan(&user.ID, &user.DiscordID, &user.DiscordUsername, &user.DisplayName, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		return &user, false, nil
	}
	if err != sql.ErrNoRows {
		return nil, false, err
	}

	err = r.db.QueryRow(
		`INSERT INTO users (discord_id) VALUES ($1)
		 RETURNING id, discord_id, discord_username, display_name, is_admin, created_at, updated_at`,
		discordID,
	).Scan(&user.ID, &user.DiscordID, &user.DiscordUsername, &user.DisplayName, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, false, err
	}

	return &user, true, nil
}

func (r *userRepo) FindByID(userID int) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, discord_id, discord_username, display_name, is_admin, created_at, updated_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.DiscordID, &user.DiscordUsername, &user.DisplayName, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) SetAdmin(userID int) error {
	result, err := r.db.Exec(
		`UPDATE users SET is_admin = TRUE, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepo) UpdateDisplayName(userID int, displayName string) error {
	if displayName == "" {
		return fmt.Errorf("display name cannot be empty")
	}
	if len(displayName) > 250 {
		return fmt.Errorf("display name exceeds maximum length of 250 characters")
	}

	result, err := r.db.Exec(
		`UPDATE users SET display_name = $1, updated_at = NOW() WHERE id = $2`,
		displayName, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update display name: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepo) SetDiscordUsername(userID int, discordUsername string) error {
	if discordUsername == "" {
		return fmt.Errorf("discord username cannot be empty")
	}

	result, err := r.db.Exec(
		`UPDATE users SET discord_username = $1, updated_at = NOW() WHERE id = $2`,
		discordUsername, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to set discord username: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepo) IsFirstUser() (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
