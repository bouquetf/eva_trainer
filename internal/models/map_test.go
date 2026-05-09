//go:build integration

package models

import (
	"testing"
)

func TestMapRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	m, err := repo.Create("Atlantis", []byte("fake-webp-binary"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if m.ID <= 0 {
		t.Errorf("expected positive id, got %d", m.ID)
	}
	if m.Name != "Atlantis" {
		t.Errorf("expected name 'Atlantis', got '%s'", m.Name)
	}
	if m.IsObsolete {
		t.Errorf("expected is_obsolete false, got true")
	}
}

func TestMapRepository_Create_WithExistingName_ObsoletesOld(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	v1, err := repo.Create("Atlantis", []byte("version1"))
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	v2, err := repo.Create("Atlantis", []byte("version2"))
	if err != nil {
		t.Fatalf("second Create failed: %v", err)
	}

	if v2.IsObsolete {
		t.Errorf("expected v2 is_obsolete false, got true")
	}

	v1Reloaded, err := repo.FindByID(v1.ID)
	if err != nil {
		t.Fatalf("FindByID(v1) failed: %v", err)
	}
	if !v1Reloaded.IsObsolete {
		t.Errorf("expected v1 is_obsolete true after replacement")
	}
	if v1Reloaded.ObsoleteOf == nil {
		t.Fatalf("expected v1 obsolete_of to be set")
	}
	if *v1Reloaded.ObsoleteOf != v2.ID {
		t.Errorf("expected v1 obsolete_of %d, got %d", v2.ID, *v1Reloaded.ObsoleteOf)
	}
}

func TestMapRepository_Create_EmptyName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	_, err := repo.Create("", []byte("data"))
	if err != ErrMapNameRequired {
		t.Errorf("expected ErrMapNameRequired, got %v", err)
	}
}

func TestMapRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	created, err := repo.Create("Polaris", []byte("blob-data"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Name != "Polaris" {
		t.Errorf("expected name 'Polaris', got '%s'", found.Name)
	}
}

func TestMapRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	_, err := repo.FindByID(99999)
	if err != ErrMapNotFound {
		t.Errorf("expected ErrMapNotFound, got %v", err)
	}
}

func TestMapRepository_FindAll_Multiple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	repo.Create("Alpha", []byte("a"))
	repo.Create("Beta", []byte("b"))
	repo.Create("Gamma", []byte("c"))

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 maps, got %d", len(all))
	}
}

func TestMapRepository_FindAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if all == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(all) != 0 {
		t.Errorf("expected 0 maps, got %d", len(all))
	}
}

func TestMapRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	repo.Create("Atlantis", []byte("v1"))
	repo.Create("Atlantis", []byte("v2"))
	repo.Create("Polaris", []byte("v1"))

	result, err := repo.FindByName("Atlantis")
	if err != nil {
		t.Fatalf("FindByName returned error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 maps named Atlantis, got %d", len(result))
	}
}

func TestMapRepository_MarkObsolete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	m, err := repo.Create("Original", []byte("data"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	replacement, err := repo.Create("Replacement", []byte("data"))
	if err != nil {
		t.Fatalf("Create replacement failed: %v", err)
	}

	if err := repo.MarkObsolete(m.ID, replacement.ID); err != nil {
		t.Fatalf("MarkObsolete returned error: %v", err)
	}

	reloaded, err := repo.FindByID(m.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if !reloaded.IsObsolete {
		t.Error("expected is_obsolete true")
	}
	if reloaded.ObsoleteOf == nil || *reloaded.ObsoleteOf != replacement.ID {
		t.Errorf("expected obsolete_of %d, got %v", replacement.ID, reloaded.ObsoleteOf)
	}
}

func TestMapRepository_IsMapObsolete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	active, _ := repo.Create("Active", []byte("data"))
	obsolete, _ := repo.Create("ObsoleteMap", []byte("data"))
	replacement, _ := repo.Create("Replacement", []byte("data"))
	repo.MarkObsolete(obsolete.ID, replacement.ID)

	ob, err := repo.IsMapObsolete(active.ID)
	if err != nil {
		t.Fatalf("IsMapObsolete(active) failed: %v", err)
	}
	if ob {
		t.Error("expected active map to be not obsolete")
	}

	ob, err = repo.IsMapObsolete(obsolete.ID)
	if err != nil {
		t.Fatalf("IsMapObsolete(obsolete) failed: %v", err)
	}
	if !ob {
		t.Error("expected obsolete map to be obsolete")
	}

	ob, err = repo.IsMapObsolete(99999)
	if err != nil {
		t.Fatalf("IsMapObsolete(99999) failed: %v", err)
	}
	if ob {
		t.Error("expected non-existent map to return false")
	}
}

func TestMapRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	m, err := repo.Create("ToDelete", []byte("data"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(m.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err = repo.FindByID(m.ID)
	if err != ErrMapNotFound {
		t.Errorf("expected ErrMapNotFound after delete, got %v", err)
	}
}

func TestMapRepository_GetBlob(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMapRepository(db)

	blob := []byte("fake-webp-binary-content")
	m, err := repo.Create("WithBlob", blob)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := repo.GetBlob(m.ID)
	if err != nil {
		t.Fatalf("GetBlob returned error: %v", err)
	}
	if string(retrieved) != string(blob) {
		t.Errorf("expected blob %q, got %q", string(blob), string(retrieved))
	}
}
