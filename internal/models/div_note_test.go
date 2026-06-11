//go:build integration

package models

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func createTestUserForDivNote(t *testing.T, db *sql.DB) int {
	t.Helper()
	var userID int
	err := db.QueryRow(
		`INSERT INTO users (discord_id, discord_username, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		"test-div-note-"+t.Name(), "test_user", "Test User",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return userID
}

func createTestDivForNote(t *testing.T, db *sql.DB) (int, int) {
	t.Helper()
	ligueID := createTestLigueForDiv(t, db)
	divRepo := NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Test Div Note")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}
	return ligueID, div.ID
}

func TestDivNote_Upsert_InsertsNewNote(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	note, err := repo.Upsert(divID, userID, "Hello")
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if note.DivID != divID {
		t.Errorf("expected div_id %d, got %d", divID, note.DivID)
	}
	if note.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, note.UserID)
	}
	if note.ContentMD != "Hello" {
		t.Errorf("expected content_md 'Hello', got '%s'", note.ContentMD)
	}
	if note.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if note.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestDivNote_Upsert_UpdatesExistingNote(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	_, err := repo.Upsert(divID, userID, "Hello")
	if err != nil {
		t.Fatalf("initial Upsert failed: %v", err)
	}

	time.Sleep(time.Millisecond * 10)

	note, err := repo.Upsert(divID, userID, "World")
	if err != nil {
		t.Fatalf("second Upsert failed: %v", err)
	}
	if note.ContentMD != "World" {
		t.Errorf("expected content_md 'World', got '%s'", note.ContentMD)
	}
	if !note.UpdatedAt.After(note.CreatedAt) {
		t.Errorf("expected updated_at to be after created_at")
	}
}

func TestDivNote_FindByDivAndUser_ReturnsNote(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	_, err := repo.Upsert(divID, userID, "My note content")
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	note, err := repo.FindByDivAndUser(divID, userID)
	if err != nil {
		t.Fatalf("FindByDivAndUser failed: %v", err)
	}
	if note.ContentMD != "My note content" {
		t.Errorf("expected content_md 'My note content', got '%s'", note.ContentMD)
	}
	if note.DivID != divID {
		t.Errorf("expected div_id %d, got %d", divID, note.DivID)
	}
	if note.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, note.UserID)
	}
}

func TestDivNote_FindByDivAndUser_ReturnsErrNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	_, err := repo.FindByDivAndUser(divID, userID)
	if !errors.Is(err, ErrDivNoteNotFound) {
		t.Errorf("expected ErrDivNoteNotFound, got %v", err)
	}
}

func TestDivNote_DeleteByDivAndUser_RemovesNote(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	_, err := repo.Upsert(divID, userID, "To delete")
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	err = repo.DeleteByDivAndUser(divID, userID)
	if err != nil {
		t.Fatalf("DeleteByDivAndUser failed: %v", err)
	}

	_, err = repo.FindByDivAndUser(divID, userID)
	if !errors.Is(err, ErrDivNoteNotFound) {
		t.Errorf("expected ErrDivNoteNotFound after delete, got %v", err)
	}
}

func TestDivNote_DeleteByDivAndUser_ReturnsErrNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userID := createTestUserForDivNote(t, db)
	_, divID := createTestDivForNote(t, db)
	repo := NewDivNoteRepository(db)

	err := repo.DeleteByDivAndUser(divID, userID)
	if !errors.Is(err, ErrDivNoteNotFound) {
		t.Errorf("expected ErrDivNoteNotFound, got %v", err)
	}
}
