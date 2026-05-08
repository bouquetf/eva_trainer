package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type RefreshToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type refreshTokenRepo struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *refreshTokenRepo) CreateRefreshToken(userID int, expiresAt time.Time) (string, error) {
	plain, err := generateToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(plain)

	_, err = r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return plain, nil
}

func (r *refreshTokenRepo) FindRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt)

	if err == sql.ErrNoRows {
		return nil, ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

func (r *refreshTokenRepo) RotateRefreshToken(oldTokenHash string, newExpiresAt time.Time) (string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int
	var revokedAt *time.Time
	err = tx.QueryRow(
		`SELECT user_id, revoked_at FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE`,
		oldTokenHash,
	).Scan(&userID, &revokedAt)

	if err == sql.ErrNoRows {
		return "", ErrRefreshTokenNotFound
	}
	if err != nil {
		return "", err
	}
	if revokedAt != nil {
		return "", ErrRefreshTokenRevoked
	}

	_, err = tx.Exec(
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`,
		oldTokenHash,
	)
	if err != nil {
		return "", fmt.Errorf("failed to revoke old token: %w", err)
	}

	plain, err := generateToken()
	if err != nil {
		return "", err
	}

	newHash := hashToken(plain)

	_, err = tx.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, newHash, newExpiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert new token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return plain, nil
}

func (r *refreshTokenRepo) RevokeRefreshToken(tokenHash string) error {
	result, err := r.db.Exec(
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepo) RevokeAllUserTokens(userID int) error {
	_, err := r.db.Exec(
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}
