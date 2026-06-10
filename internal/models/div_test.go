//go:build integration

package models

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func createTestLigueForDiv(t *testing.T, db *sql.DB) int {
	t.Helper()
	splitRepo := NewSplitRepository(db)
	split, err := splitRepo.Create("Test Split", []string{"cp", "mistral"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	now := time.Now().Truncate(24 * time.Hour)
	var ligueID int
	err = db.QueryRow(
		`INSERT INTO ligue (date_debut, date_fin, split_id, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		now, now.Add(30*24*time.Hour), split.ID,
	).Scan(&ligueID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}
	return ligueID
}

func TestDiv_SetReplayURL_ValidYouTube(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ligueID := createTestLigueForDiv(t, db)
	divRepo := NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	updated, err := divRepo.SetReplayURL(div.ID, "https://www.youtube.com/watch?v=test123")
	if err != nil {
		t.Fatalf("SetReplayURL failed: %v", err)
	}
	if updated.ReplayURL != "https://www.youtube.com/watch?v=test123" {
		t.Errorf("expected replay_url 'https://www.youtube.com/watch?v=test123', got '%s'", updated.ReplayURL)
	}
}

func TestDiv_SetReplayURL_EmptyString(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ligueID := createTestLigueForDiv(t, db)
	divRepo := NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	_, err = divRepo.SetReplayURL(div.ID, "https://www.youtube.com/watch?v=test")
	if err != nil {
		t.Fatalf("failed to set initial replay url: %v", err)
	}

	updated, err := divRepo.SetReplayURL(div.ID, "")
	if err != nil {
		t.Fatalf("SetReplayURL with empty failed: %v", err)
	}
	if updated.ReplayURL != "" {
		t.Errorf("expected empty replay_url, got '%s'", updated.ReplayURL)
	}
}

func TestDiv_SetReplayURL_DivNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	divRepo := NewDivRepository(db)
	_, err := divRepo.SetReplayURL(99999, "https://www.youtube.com/watch?v=test")
	if !errors.Is(err, ErrDivNotFound) {
		t.Errorf("expected ErrDivNotFound, got %v", err)
	}
}
