package models

import (
	"testing"
)

func TestLigueLocaleRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	locale, err := repo.Create("Ligue Test", "Salle A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if locale.ID <= 0 {
		t.Errorf("expected valid id, got %d", locale.ID)
	}
	if locale.Nom != "Ligue Test" {
		t.Errorf("expected nom 'Ligue Test', got %q", locale.Nom)
	}
	if locale.Emplacement != "Salle A" {
		t.Errorf("expected emplacement 'Salle A', got %q", locale.Emplacement)
	}
}

func TestLigueLocaleRepository_Create_NomVide(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	_, err := repo.Create("", "Salle A")
	if err != ErrLigueLocaleNameRequired {
		t.Errorf("expected ErrLigueLocaleNameRequired, got %v", err)
	}
}

func TestLigueLocaleRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	created, err := repo.Create("Test League", "Salle A")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Nom != created.Nom {
		t.Errorf("expected nom %q, got %q", created.Nom, found.Nom)
	}
}

func TestLigueLocaleRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	_, err := repo.FindByID(9999)
	if err != ErrLigueLocaleNotFound {
		t.Errorf("expected ErrLigueLocaleNotFound, got %v", err)
	}
}

func TestLigueLocaleRepository_FindAll_NoFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	repo.Create("Alpha", "Grenoble")
	repo.Create("Beta", "Lyon")

	locales, err := repo.FindAll("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(locales) != 2 {
		t.Errorf("expected 2 locales, got %d", len(locales))
	}
	if locales[0].Nom != "Alpha" {
		t.Errorf("expected first locale to be 'Alpha', got %q", locales[0].Nom)
	}
	if locales[1].Nom != "Beta" {
		t.Errorf("expected second locale to be 'Beta', got %q", locales[1].Nom)
	}
}

func TestLigueLocaleRepository_FindAll_WithFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	repo.Create("Grenoble League", "Grenoble")
	repo.Create("Lyon League", "Lyon")

	locales, err := repo.FindAll("Grenoble")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(locales) != 1 {
		t.Errorf("expected 1 locale, got %d", len(locales))
	}
	if locales[0].Emplacement != "Grenoble" {
		t.Errorf("expected emplacement 'Grenoble', got %q", locales[0].Emplacement)
	}
}

func TestLigueLocaleRepository_FindAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	locales, err := repo.FindAll("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if locales == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(locales) != 0 {
		t.Errorf("expected 0 locales, got %d", len(locales))
	}
}

func TestLigueLocaleRepository_Update_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	created, err := repo.Create("Original", "Salle A")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	updated, err := repo.Update(created.ID, "Updated", "Salle B")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, updated.ID)
	}
	if updated.Nom != "Updated" {
		t.Errorf("expected nom 'Updated', got %q", updated.Nom)
	}
	if updated.Emplacement != "Salle B" {
		t.Errorf("expected emplacement 'Salle B', got %q", updated.Emplacement)
	}
}

func TestLigueLocaleRepository_Update_NomVide(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	_, err := repo.Update(1, "", "Salle A")
	if err != ErrLigueLocaleNameRequired {
		t.Errorf("expected ErrLigueLocaleNameRequired, got %v", err)
	}
}

func TestLigueLocaleRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	_, err := repo.Update(9999, "Test", "Salle A")
	if err != ErrLigueLocaleNotFound {
		t.Errorf("expected ErrLigueLocaleNotFound, got %v", err)
	}
}

func TestLigueLocaleRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	created, err := repo.Create("Test League", "Salle A")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	err = repo.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(created.ID)
	if err != ErrLigueLocaleNotFound {
		t.Errorf("expected ligue locale to be deleted, got %v", err)
	}
}

func TestLigueLocaleRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueLocaleRepository(db)

	err := repo.Delete(9999)
	if err != ErrLigueLocaleNotFound {
		t.Errorf("expected ErrLigueLocaleNotFound, got %v", err)
	}
}
