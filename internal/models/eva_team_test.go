//go:build integration

package models

import (
	"testing"
)

func TestEvaTeamRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	team, err := repo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team.ID <= 0 {
		t.Errorf("expected valid id, got %d", team.ID)
	}
	if team.Name != "Team A" {
		t.Errorf("expected name 'Team A', got %q", team.Name)
	}
	if team.Tag != "TA" {
		t.Errorf("expected tag 'TA', got %q", team.Tag)
	}
	if team.EvaTrainerTeamID != nil {
		t.Errorf("expected nil trainer team id, got %v", *team.EvaTrainerTeamID)
	}
}

func TestEvaTeamRepository_Create_EmptyName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	_, err := repo.Create("", "TA", nil)
	if err != ErrEvaTeamNameRequired {
		t.Errorf("expected ErrEvaTeamNameRequired, got %v", err)
	}
}

func TestEvaTeamRepository_Create_EmptyTag(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	_, err := repo.Create("Team A", "", nil)
	if err != ErrEvaTeamTagRequired {
		t.Errorf("expected ErrEvaTeamTagRequired, got %v", err)
	}
}

func TestEvaTeamRepository_Create_TagTooLong(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	tag := ""
	for i := 0; i < 51; i++ {
		tag += "x"
	}
	_, err := repo.Create("Team A", tag, nil)
	if err != ErrEvaTeamTagTooLong {
		t.Errorf("expected ErrEvaTeamTagTooLong, got %v", err)
	}
}

func TestEvaTeamRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	created, err := repo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Name != created.Name {
		t.Errorf("expected name %q, got %q", created.Name, found.Name)
	}
	if found.Tag != created.Tag {
		t.Errorf("expected tag %q, got %q", created.Tag, found.Tag)
	}
}

func TestEvaTeamRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	_, err := repo.FindByID(9999)
	if err != ErrEvaTeamNotFound {
		t.Errorf("expected ErrEvaTeamNotFound, got %v", err)
	}
}

func TestEvaTeamRepository_FindAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	teams, err := repo.FindAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("expected empty list, got %d items", len(teams))
	}
}

func TestEvaTeamRepository_FindAll_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	repo.Create("Beta", "BE", nil)
	repo.Create("Alpha", "AL", nil)

	teams, err := repo.FindAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	// ordered by name ASC
	if teams[0].Name != "Alpha" {
		t.Errorf("expected first team 'Alpha', got %q", teams[0].Name)
	}
	if teams[1].Name != "Beta" {
		t.Errorf("expected second team 'Beta', got %q", teams[1].Name)
	}
}

func TestEvaTeamRepository_Update_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	created, err := repo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	updated, err := repo.Update(created.ID, "Team A Updated", "TAU", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Team A Updated" {
		t.Errorf("expected name 'Team A Updated', got %q", updated.Name)
	}
	if updated.Tag != "TAU" {
		t.Errorf("expected tag 'TAU', got %q", updated.Tag)
	}
}

func TestEvaTeamRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	_, err := repo.Update(9999, "Name", "TAG", nil)
	if err != ErrEvaTeamNotFound {
		t.Errorf("expected ErrEvaTeamNotFound, got %v", err)
	}
}

func TestEvaTeamRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	created, err := repo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	err = repo.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(created.ID)
	if err != ErrEvaTeamNotFound {
		t.Errorf("expected ErrEvaTeamNotFound, got %v", err)
	}
}

func TestEvaTeamRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	err := repo.Delete(9999)
	if err != ErrEvaTeamNotFound {
		t.Errorf("expected ErrEvaTeamNotFound, got %v", err)
	}
}

func TestEvaTeamRepository_Create_WithAssociation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	user, _, err := userRepo.FindOrCreateByDiscordID("discord_assoc")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	teamRepo := NewTeamRepository(db)
	trainerTeam, err := teamRepo.Create("Trainer Team", user.ID)
	if err != nil {
		t.Fatalf("failed to create trainer team: %v", err)
	}

	evaRepo := NewEvaTeamRepository(db)
	team, err := evaRepo.Create("Team Assoc", "TAG", &trainerTeam.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team.EvaTrainerTeamID == nil {
		t.Fatal("expected non-nil eva_trainer_team_id")
	}
	if *team.EvaTrainerTeamID != trainerTeam.ID {
		t.Errorf("expected eva_trainer_team_id %d, got %d", trainerTeam.ID, *team.EvaTrainerTeamID)
	}
}

func TestEvaTeamRepository_Update_Dissociate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	user, _, err := userRepo.FindOrCreateByDiscordID("discord_dissoc")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	teamRepo := NewTeamRepository(db)
	trainerTeam, err := teamRepo.Create("Trainer Dissoc", user.ID)
	if err != nil {
		t.Fatalf("failed to create trainer team: %v", err)
	}

	evaRepo := NewEvaTeamRepository(db)
	created, err := evaRepo.Create("Team Dissoc", "DIS", &trainerTeam.ID)
	if err != nil {
		t.Fatalf("failed to create eva team with association: %v", err)
	}

	updated, err := evaRepo.Update(created.ID, "Team Dissoc Updated", "DISU", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.EvaTrainerTeamID != nil {
		t.Errorf("expected nil eva_trainer_team_id after dissociation, got %v", *updated.EvaTrainerTeamID)
	}
}

func TestEvaTeamRepository_HasReferences_NoRefs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewEvaTeamRepository(db)

	created, err := repo.Create("Team NoRefs", "NOR", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	hasRefs, err := repo.HasReferences(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hasRefs {
		t.Error("expected HasReferences to be false, got true")
	}
}
