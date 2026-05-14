package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type UserSettings struct {
	RevisionCorrectTimer   int `json:"revision_correct_timer"`
	RevisionIncorrectTimer int `json:"revision_incorrect_timer"`
	RevisionQuestionCount  int `json:"revision_question_count"`
}

func DefaultUserSettings() UserSettings {
	return UserSettings{
		RevisionCorrectTimer:   3,
		RevisionIncorrectTimer: 5,
		RevisionQuestionCount:  20,
	}
}

func MergeUserSettings(overrides *UserSettings) UserSettings {
	if overrides == nil {
		return DefaultUserSettings()
	}
	result := DefaultUserSettings()
	if overrides.RevisionCorrectTimer > 0 {
		result.RevisionCorrectTimer = overrides.RevisionCorrectTimer
	}
	if overrides.RevisionIncorrectTimer > 0 {
		result.RevisionIncorrectTimer = overrides.RevisionIncorrectTimer
	}
	if overrides.RevisionQuestionCount > 0 {
		result.RevisionQuestionCount = overrides.RevisionQuestionCount
	}
	return result
}

type User struct {
	ID              int          `json:"id"`
	DiscordID       string       `json:"discord_id"`
	DiscordUsername string       `json:"discord_username"`
	DisplayName     string       `json:"display_name"`
	IsAdmin         bool         `json:"is_admin"`
	Settings        UserSettings `json:"settings"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
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
	if !errors.Is(err, sql.ErrNoRows) {
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

func (r *userRepo) FindAll() ([]*User, error) {
	rows, err := r.db.Query(`SELECT id, discord_id, discord_username, display_name, is_admin, created_at, updated_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.DiscordID, &u.DiscordUsername, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *userRepo) RemoveAdmin(userID int) error {
	result, err := r.db.Exec(`UPDATE users SET is_admin = FALSE, updated_at = NOW() WHERE id = $1`, userID)
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

func (r *userRepo) AdminCount() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = TRUE`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepo) IsFirstUser() (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *userRepo) FindByDisplayNameOrDiscordUsername(query string) (*User, error) {
	if query == "" {
		return nil, ErrUserNotFound
	}

	var user User
	err := r.db.QueryRow(
		`SELECT id, discord_id, discord_username, display_name, is_admin, created_at, updated_at
		 FROM users
		 WHERE display_name ILIKE $1 OR discord_username ILIKE $1
		 LIMIT 1`,
		query,
	).Scan(&user.ID, &user.DiscordID, &user.DiscordUsername, &user.DisplayName, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) SearchUsers(query string, limit int) ([]*User, error) {
	if query == "" {
		return []*User{}, nil
	}

	rows, err := r.db.Query(
		`SELECT id, discord_id, discord_username, display_name, is_admin, created_at, updated_at
		 FROM users
		 WHERE display_name ILIKE $1 OR discord_username ILIKE $1
		 ORDER BY display_name
		 LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.DiscordID, &u.DiscordUsername, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *userRepo) GetUserSettings(userID int) (*UserSettings, error) {
	var rawSettings sql.NullString
	err := r.db.QueryRow(`SELECT settings FROM users WHERE id = $1`, userID).Scan(&rawSettings)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	if !rawSettings.Valid || rawSettings.String == "" {
		defaults := DefaultUserSettings()
		return &defaults, nil
	}

	var settings UserSettings
	if err := json.Unmarshal([]byte(rawSettings.String), &settings); err != nil {
		return nil, fmt.Errorf("failed to parse user settings: %w", err)
	}

	merged := MergeUserSettings(&settings)
	return &merged, nil
}

func (r *userRepo) UpdateUserSettings(userID int, settings *UserSettings) error {
	jsonBytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal user settings: %w", err)
	}

	result, err := r.db.Exec(
		`UPDATE users SET settings = $1::jsonb, updated_at = NOW() WHERE id = $2`,
		string(jsonBytes), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
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
