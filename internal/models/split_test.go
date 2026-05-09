//go:build integration

package models

import (
	"testing"
)

func TestSplitRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	split, err := repo.Create("Season 1", []string{"Atlantis", "Lunar Outpost"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if split.ID <= 0 {
		t.Errorf("expected valid id, got %d", split.ID)
	}
	if split.Name != "Season 1" {
		t.Errorf("expected name 'Season 1', got %q", split.Name)
	}
	if len(split.MapNames) != 2 || split.MapNames[0] != "Atlantis" || split.MapNames[1] != "Lunar Outpost" {
		t.Errorf("expected map names [Atlantis Lunar Outpost], got %v", split.MapNames)
	}
}

func TestSplitRepository_Create_ValidationErrors(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	_, err := repo.Create("", []string{"Atlantis"})
	if err != ErrSplitNameRequired {
		t.Errorf("expected ErrSplitNameRequired, got %v", err)
	}

	_, err = repo.Create("Empty", []string{})
	if err != ErrSplitMapNamesRequired {
		t.Errorf("expected ErrSplitMapNamesRequired, got %v", err)
	}
}

func TestSplitRepository_Create_WithMultipleMapNames(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	mapNames := []string{"Atlantis", "Lunar Outpost", "Polaris", "Eclipse", "Nebula"}
	split, err := repo.Create("Full Season", mapNames)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(split.MapNames) != len(mapNames) {
		t.Errorf("expected %d map names, got %d", len(mapNames), len(split.MapNames))
	}
	for i, name := range mapNames {
		if split.MapNames[i] != name {
			t.Errorf("expected map_names[%d] = %q, got %q", i, name, split.MapNames[i])
		}
	}
}

func TestSplitRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	created, err := repo.Create("Test Split", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, found.ID)
	}
	if found.Name != created.Name {
		t.Errorf("expected name %q, got %q", created.Name, found.Name)
	}
}

func TestSplitRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	_, err := repo.FindByID(99999)
	if err != ErrSplitNotFound {
		t.Errorf("expected ErrSplitNotFound, got %v", err)
	}
}

func TestSplitRepository_FindAll_Sorted(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	_, err := repo.Create("Z Split", []string{"Polaris"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	_, err = repo.Create("A Split", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	_, err = repo.Create("M Split", []string{"Eclipse"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	splits, err := repo.FindAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(splits) != 3 {
		t.Fatalf("expected 3 splits, got %d", len(splits))
	}
	if splits[0].Name != "A Split" {
		t.Errorf("expected first split 'A Split', got %q", splits[0].Name)
	}
	if splits[1].Name != "M Split" {
		t.Errorf("expected second split 'M Split', got %q", splits[1].Name)
	}
	if splits[2].Name != "Z Split" {
		t.Errorf("expected third split 'Z Split', got %q", splits[2].Name)
	}
}

func TestSplitRepository_FindAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	splits, err := repo.FindAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if splits == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(splits) != 0 {
		t.Errorf("expected 0 splits, got %d", len(splits))
	}
}

func TestSplitRepository_Update_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	created, err := repo.Create("Original Name", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	updated, err := repo.Update(created.ID, "Updated Name", []string{"Polaris", "Eclipse"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", updated.Name)
	}
	if len(updated.MapNames) != 2 || updated.MapNames[0] != "Polaris" {
		t.Errorf("expected map names [Polaris Eclipse], got %v", updated.MapNames)
	}
}

func TestSplitRepository_Update_ValidationErrors(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	created, err := repo.Create("Valid Split", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	_, err = repo.Update(created.ID, "", []string{"Atlantis"})
	if err != ErrSplitNameRequired {
		t.Errorf("expected ErrSplitNameRequired, got %v", err)
	}

	_, err = repo.Update(created.ID, "Valid Name", []string{})
	if err != ErrSplitMapNamesRequired {
		t.Errorf("expected ErrSplitMapNamesRequired, got %v", err)
	}
}

func TestSplitRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	_, err := repo.Update(99999, "Ghost", []string{"Atlantis"})
	if err != ErrSplitNotFound {
		t.Errorf("expected ErrSplitNotFound, got %v", err)
	}
}

func TestSplitRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	created, err := repo.Create("To Delete", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	err = repo.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(created.ID)
	if err != ErrSplitNotFound {
		t.Errorf("expected ErrSplitNotFound after delete, got %v", err)
	}
}

func TestSplitRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSplitRepository(db)

	err := repo.Delete(99999)
	if err != ErrSplitNotFound {
		t.Errorf("expected ErrSplitNotFound, got %v", err)
	}
}

func TestSplitRepository_FindMapsBySplitID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	mapRepo := NewMapRepository(db)

	atlantis1, err := mapRepo.Create("Atlantis", []byte("v1"))
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	atlantis2, err := mapRepo.Create("Atlantis", []byte("v2"))
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	polaris, err := mapRepo.Create("Polaris", []byte("v1"))
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	split, err := splitRepo.Create("Test Split", []string{"Atlantis", "Polaris"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	maps, err := splitRepo.FindMapsBySplitID(split.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(maps) != 3 {
		t.Fatalf("expected 3 maps (including obsolete), got %d", len(maps))
	}

	foundIDs := make(map[int]bool)
	for _, m := range maps {
		foundIDs[m.ID] = true
	}
	if !foundIDs[atlantis1.ID] {
		t.Error("expected atlantis1 (obsolete) to be included")
	}
	if !foundIDs[atlantis2.ID] {
		t.Error("expected atlantis2 to be included")
	}
	if !foundIDs[polaris.ID] {
		t.Error("expected polaris to be included")
	}
}

func TestSplitRepository_FindMapsBySplitID_NoMatches(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)

	split, err := splitRepo.Create("No Match", []string{"NonExistentMap"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	maps, err := splitRepo.FindMapsBySplitID(split.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if maps == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(maps) != 0 {
		t.Errorf("expected 0 maps, got %d", len(maps))
	}
}

func TestSplitRepository_FindMapsBySplitID_PartialMatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	mapRepo := NewMapRepository(db)

	_, err := mapRepo.Create("Atlantis", []byte("v1"))
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	split, err := splitRepo.Create("Partial", []string{"Atlantis", "NonExistentMap"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	maps, err := splitRepo.FindMapsBySplitID(split.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(maps) != 1 {
		t.Fatalf("expected 1 map, got %d", len(maps))
	}
	if maps[0].Name != "Atlantis" {
		t.Errorf("expected 'Atlantis', got %q", maps[0].Name)
	}
}

func TestSplitRepository_FindMapsBySplitID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)

	_, err := splitRepo.FindMapsBySplitID(99999)
	if err != ErrSplitNotFound {
		t.Errorf("expected ErrSplitNotFound, got %v", err)
	}
}
