//go:build integration

package models

import (
	"strings"
	"testing"
)

func validConeData() ConeData {
	return ConeData{Cones: []Cone{
		{Origin: Point{X: 10, Y: 20}, LeftEnd: Point{X: 30, Y: 40}, RightEnd: Point{X: 50, Y: 60}},
	}}
}

func TestLineOfFireRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-test-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Test Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-test-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	l, err := repo.Create(team.ID, m.ID, "Couloir A", validConeData())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if l.ID <= 0 {
		t.Errorf("expected positive id, got %d", l.ID)
	}
	if l.Name != "Couloir A" {
		t.Errorf("expected name 'Couloir A', got '%s'", l.Name)
	}
	if l.TeamID != team.ID {
		t.Errorf("expected team_id %d, got %d", team.ID, l.TeamID)
	}
	if l.MapID != m.ID {
		t.Errorf("expected map_id %d, got %d", m.ID, l.MapID)
	}
}

func TestLineOfFireRepository_Create_NoName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)

	_, err := repo.Create(1, 1, "", validConeData())
	if err != ErrLineOfFireNameRequired {
		t.Errorf("expected ErrLineOfFireNameRequired, got %v", err)
	}
}

func TestLineOfFireRepository_Create_NameTooLong(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)

	longName := strings.Repeat("A", 251)
	_, err := repo.Create(1, 1, longName, validConeData())
	if err != ErrLineOfFireNameTooLong {
		t.Errorf("expected ErrLineOfFireNameTooLong, got %v", err)
	}
}

func TestLineOfFireRepository_Create_DuplicateName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-dup-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Dup Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-dup-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	_, err = repo.Create(team.ID, m.ID, "Unique Name", validConeData())
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = repo.Create(team.ID, m.ID, "Unique Name", validConeData())
	if err != ErrLineOfFireNameTaken {
		t.Errorf("expected ErrLineOfFireNameTaken, got %v", err)
	}
}

func TestLineOfFireRepository_Create_NoCones(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)

	_, err := repo.Create(1, 1, "No Cones", ConeData{Cones: []Cone{}})
	if err != ErrLineOfFireNoCones {
		t.Errorf("expected ErrLineOfFireNoCones, got %v", err)
	}
}

func TestLineOfFireRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)

	_, err := repo.FindByID(99999)
	if err != ErrLineOfFireNotFound {
		t.Errorf("expected ErrLineOfFireNotFound, got %v", err)
	}
}

func TestLineOfFireRepository_FindByTeamAndMap(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-find-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Find Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-find-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	cone1 := ConeData{Cones: []Cone{
		{Origin: Point{X: 1, Y: 2}, LeftEnd: Point{X: 3, Y: 4}, RightEnd: Point{X: 5, Y: 6}},
	}}
	cone2 := ConeData{Cones: []Cone{
		{Origin: Point{X: 7, Y: 8}, LeftEnd: Point{X: 9, Y: 10}, RightEnd: Point{X: 11, Y: 12}},
	}}

	_, err = repo.Create(team.ID, m.ID, "Line B", cone2)
	if err != nil {
		t.Fatalf("Create Line B failed: %v", err)
	}
	_, err = repo.Create(team.ID, m.ID, "Line A", cone1)
	if err != nil {
		t.Fatalf("Create Line A failed: %v", err)
	}

	lines, err := repo.FindByTeamAndMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("FindByTeamAndMap returned error: %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0].Name != "Line B" {
		t.Errorf("expected first line 'Line B' (order_index), got '%s'", lines[0].Name)
	}
	if lines[1].Name != "Line A" {
		t.Errorf("expected second line 'Line A' (order_index), got '%s'", lines[1].Name)
	}
}

func TestLineOfFireRepository_UpdateName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-rename-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Rename Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-rename-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	l, err := repo.Create(team.ID, m.ID, "Old Name", validConeData())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.UpdateName(l.ID, "New Name"); err != nil {
		t.Fatalf("UpdateName returned error: %v", err)
	}

	reloaded, err := repo.FindByID(l.ID)
	if err != nil {
		t.Fatalf("FindByID after rename failed: %v", err)
	}
	if reloaded.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", reloaded.Name)
	}
}

func TestLineOfFireRepository_UpdatePolygonData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-poly-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Poly Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-poly-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	l, err := repo.Create(team.ID, m.ID, "Poly Test", validConeData())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	newData := ConeData{Cones: []Cone{
		{Origin: Point{X: 100, Y: 200}, LeftEnd: Point{X: 300, Y: 400}, RightEnd: Point{X: 500, Y: 600}},
		{Origin: Point{X: 700, Y: 800}, LeftEnd: Point{X: 900, Y: 1000}, RightEnd: Point{X: 1100, Y: 1200}},
	}}

	if err := repo.UpdatePolygonData(l.ID, newData); err != nil {
		t.Fatalf("UpdatePolygonData returned error: %v", err)
	}

	reloaded, err := repo.FindByID(l.ID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if len(reloaded.PolygonData.Cones) != 2 {
		t.Errorf("expected 2 cones, got %d", len(reloaded.PolygonData.Cones))
	}
}

func TestLineOfFireRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)

	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-delete-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Delete Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-delete-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	l, err := repo.Create(team.ID, m.ID, "To Delete", validConeData())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(l.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err = repo.FindByID(l.ID)
	if err != ErrLineOfFireNotFound {
		t.Errorf("expected ErrLineOfFireNotFound after delete, got %v", err)
	}
}

func TestLineOfFireRepository_Delete_CascadeVersions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLineOfFireRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	versionRepo := NewLineOfFireVersionRepository(db)
	user, _, err := NewUserRepository(db).FindOrCreateByDiscordID("lof-cascade-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("LOF Cascade Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("lof-cascade-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	l, err := repo.Create(team.ID, m.ID, "Cascade Test", validConeData())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	raw := []byte(`{"name":"test"}`)
	if _, err := versionRepo.Save(l.ID, nil, raw); err != nil {
		t.Fatalf("Save version failed: %v", err)
	}

	if err := repo.Delete(l.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	versions, err := versionRepo.FindByLineOfFire(l.ID)
	if err != nil {
		t.Fatalf("FindByLineOfFire after delete failed: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected 0 versions after cascade delete, got %d", len(versions))
	}
}
