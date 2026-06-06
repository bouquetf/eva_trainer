package models

import (
	"database/sql"
	"testing"
	"time"
)

func createTestSplit2(t *testing.T, db *sql.DB, repo SplitRepository) *Split {
	t.Helper()
	split, err := repo.Create("Test Split 2026", []string{"Atlantis"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	return split
}

func TestLigueRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	dateDebut := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dateFin := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	ligue, err := repo.Create(dateDebut, dateFin, split.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ligue.ID <= 0 {
		t.Errorf("expected valid id, got %d", ligue.ID)
	}
	if !ligue.DateDebut.Equal(dateDebut) {
		t.Errorf("expected date_debut %v, got %v", dateDebut, ligue.DateDebut)
	}
	if !ligue.DateFin.Equal(dateFin) {
		t.Errorf("expected date_fin %v, got %v", dateFin, ligue.DateFin)
	}
	if ligue.SplitID != split.ID {
		t.Errorf("expected split_id %d, got %d", split.ID, ligue.SplitID)
	}
}

func TestLigueRepository_Create_DatesInvalides(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueRepository(db)

	_, err := repo.Create(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 1)
	if err != ErrLigueInvalidDates {
		t.Errorf("expected ErrLigueInvalidDates, got %v", err)
	}
}

func TestLigueRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	created, err := repo.Create(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, found.ID)
	}
}

func TestLigueRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueRepository(db)

	_, err := repo.FindByID(9999)
	if err != ErrLigueNotFound {
		t.Errorf("expected ErrLigueNotFound, got %v", err)
	}
}

func TestLigueRepository_FindAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	repo.Create(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	repo.Create(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)

	ligues, err := repo.FindAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ligues) != 2 {
		t.Errorf("expected 2 ligues, got %d", len(ligues))
	}
}

func TestLigueRepository_Update_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	created, err := repo.Create(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}

	newDate := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	updated, err := repo.Update(created.ID, newDate, newDate.AddDate(0, 6, 0), split.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, updated.ID)
	}
	if !updated.DateDebut.Equal(newDate) {
		t.Errorf("expected date_debut %v, got %v", newDate, updated.DateDebut)
	}
}

func TestLigueRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	_, err := repo.Update(9999, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	if err != ErrLigueNotFound {
		t.Errorf("expected ErrLigueNotFound, got %v", err)
	}
}

func TestLigueRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	repo := NewLigueRepository(db)

	created, err := repo.Create(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}

	err = repo.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(created.ID)
	if err != ErrLigueNotFound {
		t.Errorf("expected ligue to be deleted, got %v", err)
	}
}

func TestLigueRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLigueRepository(db)

	err := repo.Delete(9999)
	if err != ErrLigueNotFound {
		t.Errorf("expected ErrLigueNotFound, got %v", err)
	}
}

func TestLigueRepository_AssociateAllLocales(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	splitRepo := NewSplitRepository(db)
	split := createTestSplit2(t, db, splitRepo)
	ligueRepo := NewLigueRepository(db)
	localeRepo := NewLigueLocaleRepository(db)

	ligue, err := ligueRepo.Create(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}

	localeRepo.Create("Locale A", "Grenoble")
	localeRepo.Create("Locale B", "Lyon")

	count, err := ligueRepo.AssociateAllLocales(ligue.ID)
	if err != nil {
		t.Fatalf("failed to associate locales: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 associated locales, got %d", count)
	}

	ligues, err := localeRepo.FindLiguesByLocaleID(1)
	if err != nil {
		t.Fatalf("failed to find ligues for locale: %v", err)
	}
	if len(ligues) != 1 || ligues[0].ID != ligue.ID {
		t.Errorf("expected locale 1 to be associated with ligue %d, got %v", ligue.ID, ligues)
	}
}
