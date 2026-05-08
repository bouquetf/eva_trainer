//go:build integration

package models

import (
	"testing"
)

func TestUserRepository_FindOrCreateByDiscordID_CreatesNew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, isNew, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("FindOrCreateByDiscordID returned error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", user.DiscordID)
	}
	if user.IsAdmin {
		t.Errorf("expected is_admin false for new user")
	}
	if user.ID <= 0 {
		t.Errorf("expected positive id, got %d", user.ID)
	}
	if !isNew {
		t.Errorf("expected isNew true, got false")
	}
}

func TestUserRepository_FindOrCreateByDiscordID_ReturnsExisting(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	user, isNew, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if user.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", user.DiscordID)
	}
	if isNew {
		t.Errorf("expected isNew false for existing user")
	}
}

func TestUserRepository_FindOrCreateByDiscordID_EmptyID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("")
	if err != ErrInvalidDiscordID {
		t.Errorf("expected ErrInvalidDiscordID, got %v", err)
	}
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	created, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", found.DiscordID)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, err := repo.FindByID(99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_IsFirstUser_True(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	first, err := repo.IsFirstUser()
	if err != nil {
		t.Fatalf("IsFirstUser returned error: %v", err)
	}
	if !first {
		t.Errorf("expected true for empty users table")
	}
}

func TestUserRepository_IsFirstUser_False(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	first, err := repo.IsFirstUser()
	if err != nil {
		t.Fatalf("IsFirstUser returned error: %v", err)
	}
	if first {
		t.Errorf("expected false when users exist")
	}
}

func TestUserRepository_SetAdmin_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := repo.SetAdmin(user.ID); err != nil {
		t.Fatalf("SetAdmin returned error: %v", err)
	}

	updated, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if !updated.IsAdmin {
		t.Errorf("expected user to be admin after SetAdmin")
	}
}

func TestUserRepository_SetAdmin_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	err := repo.SetAdmin(99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
