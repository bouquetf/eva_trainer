//go:build integration

package models

import (
	"testing"
)

func setupGlobalRoleTestDB(t *testing.T) (GlobalRoleRepository, UserRepository, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)
	globalRepo := NewGlobalRoleRepository(db)
	userRepo := NewUserRepository(db)

	return globalRepo, userRepo, cleanup
}

func TestGlobalRoleRepository_AddAndUserHasRole(t *testing.T) {
	globalRepo, userRepo, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("global-role-add-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := globalRepo.Add(user.ID, RoleArchivist); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	has, err := globalRepo.UserHasRole(user.ID, RoleArchivist)
	if err != nil {
		t.Fatalf("UserHasRole failed: %v", err)
	}
	if !has {
		t.Errorf("expected UserHasRole=true after Add")
	}
}

func TestGlobalRoleRepository_UserHasRole_FalseForNonArchivist(t *testing.T) {
	globalRepo, userRepo, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("global-role-none-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	has, err := globalRepo.UserHasRole(user.ID, RoleArchivist)
	if err != nil {
		t.Fatalf("UserHasRole failed: %v", err)
	}
	if has {
		t.Errorf("expected UserHasRole=false for non-archivist")
	}
}

func TestGlobalRoleRepository_Remove(t *testing.T) {
	globalRepo, userRepo, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("global-role-remove-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := globalRepo.Add(user.ID, RoleArchivist); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := globalRepo.Remove(user.ID, RoleArchivist); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	has, err := globalRepo.UserHasRole(user.ID, RoleArchivist)
	if err != nil {
		t.Fatalf("UserHasRole failed: %v", err)
	}
	if has {
		t.Errorf("expected UserHasRole=false after Remove")
	}
}

func TestGlobalRoleRepository_Remove_NotFound(t *testing.T) {
	globalRepo, _, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	err := globalRepo.Remove(99999, RoleArchivist)
	if err != ErrRoleNotFound {
		t.Errorf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestGlobalRoleRepository_FindUsersByRole(t *testing.T) {
	globalRepo, userRepo, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("global-role-find-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := globalRepo.Add(user.ID, RoleArchivist); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	users, err := globalRepo.FindUsersByRole(RoleArchivist)
	if err != nil {
		t.Fatalf("FindUsersByRole failed: %v", err)
	}

	found := false
	for _, u := range users {
		if u.UserID == user.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find user in archivists list")
	}
}

func TestGlobalRoleRepository_FindUsersByRole_ReturnsEmptyWhenNone(t *testing.T) {
	globalRepo, _, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	users, err := globalRepo.FindUsersByRole(RoleArchivist)
	if err != nil {
		t.Fatalf("FindUsersByRole failed: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected empty list, got %d items", len(users))
	}
}

func TestGlobalRoleRepository_Add_SameRoleDifferentUser(t *testing.T) {
	globalRepo, userRepo, cleanup := setupGlobalRoleTestDB(t)
	defer cleanup()

	user1, _, err := userRepo.FindOrCreateByDiscordID("global-role-u1-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	user2, _, err := userRepo.FindOrCreateByDiscordID("global-role-u2-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	if err := globalRepo.Add(user1.ID, RoleArchivist); err != nil {
		t.Fatalf("first Add failed: %v", err)
	}
	if err := globalRepo.Add(user2.ID, RoleArchivist); err != nil {
		t.Fatalf("second Add failed: %v", err)
	}

	users, err := globalRepo.FindUsersByRole(RoleArchivist)
	if err != nil {
		t.Fatalf("FindUsersByRole failed: %v", err)
	}

	if len(users) < 2 {
		t.Errorf("expected at least 2 archivists, got %d", len(users))
	}
}
