//go:build integration

package models

import (
	"database/sql"
	"testing"
	"time"
)

func createTestUserForLock(t *testing.T, db *sql.DB, suffix string) *User {
	t.Helper()
	userRepo := NewUserRepository(db)
	user, _, err := userRepo.FindOrCreateByDiscordID("lock-test-user-" + suffix)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func createTestTeamForLock(t *testing.T, db *sql.DB, user *User, suffix string) *Team {
	t.Helper()
	teamRepo := NewTeamRepository(db)
	team, err := teamRepo.Create("Lock Test Team "+suffix, user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	return team
}

func createTestMapForLock(t *testing.T, db *sql.DB) *Map {
	t.Helper()
	mapRepo := NewMapRepository(db)
	m, err := mapRepo.Create("lock-test-map", nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	return m
}

func TestEditLockRepository_Acquire_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "acquire")
	team := createTestTeamForLock(t, db, user, "acquire")
	m := createTestMapForLock(t, db)

	lock, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	if lock.ID <= 0 {
		t.Errorf("expected positive id, got %d", lock.ID)
	}
	if lock.EditorID != user.ID {
		t.Errorf("expected editor_id %d, got %d", user.ID, lock.EditorID)
	}
	if !lock.ExpiresAt.After(lock.LockedAt) {
		t.Error("expected expires_at to be after locked_at")
	}
	diff := lock.ExpiresAt.Sub(lock.LockedAt)
	if diff < 14*time.Minute || diff > 16*time.Minute {
		t.Errorf("expected expiry ~15min from locked_at, got %v", diff)
	}
}

func TestEditLockRepository_DoubleAcquire_Fails(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user1 := createTestUserForLock(t, db, "double1")
	user2 := createTestUserForLock(t, db, "double2")
	team := createTestTeamForLock(t, db, user1, "double")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user1.ID)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	_, err = lockRepo.Acquire(team.ID, m.ID, user2.ID)
	if err != ErrLockAlreadyAcquired {
		t.Errorf("expected ErrLockAlreadyAcquired, got %v", err)
	}
}

func TestEditLockRepository_Refresh_UpdatesExpiry(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "refresh")
	team := createTestTeamForLock(t, db, user, "refresh")
	m := createTestMapForLock(t, db)

	lock, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	refreshed, err := lockRepo.Refresh(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if !refreshed.ExpiresAt.After(lock.ExpiresAt) {
		t.Error("expected refreshed expires_at to be later than original")
	}
}

func TestEditLockRepository_Release_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "release")
	team := createTestTeamForLock(t, db, user, "release")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if err := lockRepo.Release(team.ID, m.ID); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	lock, err := lockRepo.Get(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if lock != nil {
		t.Error("expected lock to be nil after release")
	}
}

func TestEditLockRepository_Get_NoLock(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "get-nil")
	team := createTestTeamForLock(t, db, user, "get-nil")
	m := createTestMapForLock(t, db)

	lock, err := lockRepo.Get(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if lock != nil {
		t.Error("expected nil lock when no lock exists")
	}
}

func TestEditLockRepository_Release_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	err := lockRepo.Release(999, 999)
	if err != ErrLockNotFound {
		t.Errorf("expected ErrLockNotFound, got %v", err)
	}
}

func TestEditLockRepository_Refresh_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	_, err := lockRepo.Refresh(999, 999)
	if err != ErrLockNotFound {
		t.Errorf("expected ErrLockNotFound, got %v", err)
	}
}

func TestEditLockRepository_IsLocked_True(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "is-locked")
	team := createTestTeamForLock(t, db, user, "is-locked")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	locked, err := lockRepo.IsLocked(team.ID, m.ID)
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if !locked {
		t.Error("expected IsLocked to be true after acquire")
	}
}

func TestEditLockRepository_IsLockedByEditor_CorrectEditor(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "is-locked-by")
	team := createTestTeamForLock(t, db, user, "is-locked-by")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	locked, err := lockRepo.IsLockedByEditor(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("IsLockedByEditor failed: %v", err)
	}
	if !locked {
		t.Error("expected IsLockedByEditor to be true for the correct editor")
	}
}

func TestEditLockRepository_IsLockedByEditor_WrongEditor(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user1 := createTestUserForLock(t, db, "wrong-editor1")
	user2 := createTestUserForLock(t, db, "wrong-editor2")
	team := createTestTeamForLock(t, db, user1, "wrong-editor")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user1.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	locked, err := lockRepo.IsLockedByEditor(team.ID, m.ID, user2.ID)
	if err != nil {
		t.Fatalf("IsLockedByEditor failed: %v", err)
	}
	if locked {
		t.Error("expected IsLockedByEditor to be false for wrong editor")
	}
}

func TestEditLockRepository_ReleaseByEditor_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	user := createTestUserForLock(t, db, "rel-by-editor")
	team := createTestTeamForLock(t, db, user, "rel-by-editor")
	m := createTestMapForLock(t, db)

	_, err := lockRepo.Acquire(team.ID, m.ID, user.ID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if err := lockRepo.ReleaseByEditor(team.ID, m.ID, user.ID); err != nil {
		t.Fatalf("ReleaseByEditor failed: %v", err)
	}

	lock, err := lockRepo.Get(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if lock != nil {
		t.Error("expected lock to be nil after ReleaseByEditor")
	}
}

func TestEditLockRepository_ReleaseByEditor_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lockRepo := NewEditLockRepository(db)
	err := lockRepo.ReleaseByEditor(999, 999, 999)
	if err != ErrLockNotFound {
		t.Errorf("expected ErrLockNotFound, got %v", err)
	}
}
