//go:build integration

package models

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func createTestUser(t *testing.T, repo UserRepository) *User {
	t.Helper()
	user, _, err := repo.FindOrCreateByDiscordID("test-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func hashTokenForTest(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func TestRefreshTokenRepository_CreateAndFind(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	plain, err := rtRepo.CreateRefreshToken(user.ID, expiresAt)
	if err != nil {
		t.Fatalf("CreateRefreshToken returned error: %v", err)
	}
	if plain == "" {
		t.Fatal("expected non-empty token string")
	}
	if len(plain) != 64 {
		t.Errorf("expected token of length 64 (32 bytes hex), got %d", len(plain))
	}

	tokenHash := hashTokenForTest(plain)
	found, err := rtRepo.FindRefreshTokenByHash(tokenHash)
	if err != nil {
		t.Fatalf("FindRefreshTokenByHash returned error: %v", err)
	}
	if found.UserID != user.ID {
		t.Errorf("expected user_id %d, got %d", user.ID, found.UserID)
	}
	if found.TokenHash != tokenHash {
		t.Errorf("token hash mismatch")
	}
	if found.RevokedAt != nil {
		t.Errorf("expected revoked_at to be nil for new token")
	}
}

func TestRefreshTokenRepository_Find_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	rtRepo := NewRefreshTokenRepository(db)

	_, err := rtRepo.FindRefreshTokenByHash("nonexistent_hash")
	if err != ErrRefreshTokenNotFound {
		t.Errorf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestRefreshTokenRepository_Rotate_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	oldPlain, err := rtRepo.CreateRefreshToken(user.ID, expiresAt)
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	oldHash := hashTokenForTest(oldPlain)

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	newPlain, err := rtRepo.RotateRefreshToken(oldHash, newExpiresAt)
	if err != nil {
		t.Fatalf("RotateRefreshToken failed: %v", err)
	}
	if newPlain == "" {
		t.Fatal("expected non-empty new token")
	}
	if newPlain == oldPlain {
		t.Errorf("new token should differ from old token")
	}

	newHash := hashTokenForTest(newPlain)

	oldToken, err := rtRepo.FindRefreshTokenByHash(oldHash)
	if err != nil {
		t.Fatalf("FindRefreshTokenByHash for old token failed: %v", err)
	}
	if oldToken.RevokedAt == nil {
		t.Errorf("expected old token to be revoked after rotation")
	}

	newToken, err := rtRepo.FindRefreshTokenByHash(newHash)
	if err != nil {
		t.Fatalf("FindRefreshTokenByHash for new token failed: %v", err)
	}
	if newToken.RevokedAt != nil {
		t.Errorf("expected new token revoked_at to be nil")
	}
}

func TestRefreshTokenRepository_Rotate_AlreadyRevoked(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	plain, err := rtRepo.CreateRefreshToken(user.ID, time.Now().Add(7*24*time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	tokenHash := hashTokenForTest(plain)

	if err := rtRepo.RevokeRefreshToken(tokenHash); err != nil {
		t.Fatalf("RevokeRefreshToken failed: %v", err)
	}

	_, err = rtRepo.RotateRefreshToken(tokenHash, time.Now().Add(7*24*time.Hour))
	if err != ErrRefreshTokenRevoked {
		t.Errorf("expected ErrRefreshTokenRevoked, got %v", err)
	}
}

func TestRefreshTokenRepository_Rotate_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	rtRepo := NewRefreshTokenRepository(db)

	_, err := rtRepo.RotateRefreshToken("nonexistent_hash", time.Now().Add(7*24*time.Hour))
	if err != ErrRefreshTokenNotFound {
		t.Errorf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestRefreshTokenRepository_Revoke_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	plain, err := rtRepo.CreateRefreshToken(user.ID, time.Now().Add(7*24*time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	tokenHash := hashTokenForTest(plain)

	if err := rtRepo.RevokeRefreshToken(tokenHash); err != nil {
		t.Fatalf("RevokeRefreshToken failed: %v", err)
	}

	token, err := rtRepo.FindRefreshTokenByHash(tokenHash)
	if err != nil {
		t.Fatalf("FindRefreshTokenByHash failed: %v", err)
	}
	if token.RevokedAt == nil {
		t.Errorf("expected token to be revoked")
	}
}

func TestRefreshTokenRepository_Revoke_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	rtRepo := NewRefreshTokenRepository(db)

	err := rtRepo.RevokeRefreshToken("nonexistent_hash")
	if err != ErrRefreshTokenNotFound {
		t.Errorf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestRefreshTokenRepository_RevokeAllUserTokens(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tokens := make([]string, 3)
	for i := range tokens {
		plain, err := rtRepo.CreateRefreshToken(user.ID, expiresAt)
		if err != nil {
			t.Fatalf("CreateRefreshToken %d failed: %v", i, err)
		}
		tokens[i] = hashTokenForTest(plain)
	}

	if err := rtRepo.RevokeAllUserTokens(user.ID); err != nil {
		t.Fatalf("RevokeAllUserTokens failed: %v", err)
	}

	for i, hash := range tokens {
		token, err := rtRepo.FindRefreshTokenByHash(hash)
		if err != nil {
			t.Fatalf("FindRefreshTokenByHash %d failed: %v", i, err)
		}
		if token.RevokedAt == nil {
			t.Errorf("expected token %d to be revoked after RevokeAllUserTokens", i)
		}
	}
}

func TestRefreshTokenRepository_ExpiredToken_CanBeFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	rtRepo := NewRefreshTokenRepository(db)

	user := createTestUser(t, userRepo)

	expiredAt := time.Now().Add(-24 * time.Hour)
	plain, err := rtRepo.CreateRefreshToken(user.ID, expiredAt)
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	tokenHash := hashTokenForTest(plain)
	token, err := rtRepo.FindRefreshTokenByHash(tokenHash)
	if err != nil {
		t.Fatalf("FindRefreshTokenByHash failed: %v", err)
	}

	if !token.ExpiresAt.Before(time.Now()) {
		t.Errorf("expected token to be expired (expiresAt=%v, now=%v)", token.ExpiresAt, time.Now())
	}
}
