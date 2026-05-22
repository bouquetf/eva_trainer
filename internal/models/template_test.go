//go:build integration

package models

import (
	"database/sql"
	"testing"
)

func TestTemplateRepository_CreateFromMap_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-create-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)
	zoneRepo.SetName(zone.ID, "Test Zone")

	tmpl, err := templateRepo.CreateFromMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}
	if tmpl.OriginalTeamID != team.ID {
		t.Errorf("expected team_id %d, got %d", team.ID, tmpl.OriginalTeamID)
	}
	if tmpl.MapID != m.ID {
		t.Errorf("expected map_id %d, got %d", m.ID, tmpl.MapID)
	}
}

func TestTemplateRepository_CreateFromMap_NoZones(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-nozones-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zoneRepo.Create(team.ID, m.ID)

	tmpl, err := templateRepo.CreateFromMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}
	if tmpl.ID <= 0 {
		t.Errorf("expected positive id, got %d", tmpl.ID)
	}
}

func TestTemplateRepository_CreateFromMap_AlreadyExists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-already-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)

	first, err := templateRepo.CreateFromMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("first CreateFromMap failed: %v", err)
	}

	second, err := templateRepo.CreateFromMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("second CreateFromMap failed: %v", err)
	}
	if second.ID <= first.ID {
		t.Errorf("expected second id > first id (%d), got %d", first.ID, second.ID)
	}
}

func TestTemplateRepository_FindByMapID_Found(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-findbymap-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)

	templateRepo.CreateFromMap(team.ID, m.ID)

	found, err := templateRepo.FindByMapID(m.ID)
	if err != nil {
		t.Fatalf("FindByMapID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected template, got nil")
	}
	if found.MapID != m.ID {
		t.Errorf("expected map_id %d, got %d", m.ID, found.MapID)
	}
}

func TestTemplateRepository_FindByMapID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	templateRepo := NewTemplateRepository(db)

	found, err := templateRepo.FindByMapID(99999)
	if err != nil {
		t.Fatalf("FindByMapID failed: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %v", found)
	}
}

func TestTemplateRepository_UpdateFromMap_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-update-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)

	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)
	zoneRepo.SetName(zone.ID, "Initial Zone")

	tmpl, _ := templateRepo.CreateFromMap(team.ID, m.ID)

	zone2, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone2.ID, []Point{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 10, Y: 20}}, 0, nil)
	zoneRepo.SetName(zone2.ID, "Added Zone")

	err := templateRepo.UpdateFromMap(tmpl.ID, team.ID, m.ID)
	if err != nil {
		t.Fatalf("UpdateFromMap failed: %v", err)
	}
}

func TestTemplateRepository_FindAllTemplates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-findall-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)

	m1, _ := mapRepo.Create("test-map-1-"+t.Name(), nil)
	m2, _ := mapRepo.Create("test-map-2-"+t.Name(), nil)

	zoneRepo.Create(team.ID, m1.ID)
	zoneRepo.Create(team.ID, m2.ID)

	templateRepo.CreateFromMap(team.ID, m1.ID)
	templateRepo.CreateFromMap(team.ID, m2.ID)

	all, err := templateRepo.FindAllTemplates()
	if err != nil {
		t.Fatalf("FindAllTemplates failed: %v", err)
	}
	if len(all) < 2 {
		t.Errorf("expected at least 2 templates, got %d", len(all))
	}
}

func TestTemplateRepository_DuplicateToTeam_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user1, _, _ := userRepo.FindOrCreateByDiscordID("test-template-dup-1-" + t.Name())
	user2, _, _ := userRepo.FindOrCreateByDiscordID("test-template-dup-2-" + t.Name())

	sourceTeam, _ := teamRepo.Create("Source Team "+t.Name(), user1.ID)
	targetTeam, _ := teamRepo.Create("Target Team "+t.Name(), user2.ID)

	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)

	zone, _ := zoneRepo.Create(sourceTeam.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)
	zoneRepo.SetName(zone.ID, "Duplicated Zone")

	tmpl, _ := templateRepo.CreateFromMap(sourceTeam.ID, m.ID)

	createdZones, mapID, err := templateRepo.DuplicateToTeam(targetTeam.ID, tmpl.ID)
	if err != nil {
		t.Fatalf("DuplicateToTeam failed: %v", err)
	}
	if mapID != m.ID {
		t.Errorf("expected map_id %d, got %d", m.ID, mapID)
	}
	if len(createdZones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(createdZones))
	}
	if createdZones[0].TeamID != targetTeam.ID {
		t.Errorf("expected team_id %d, got %d", targetTeam.ID, createdZones[0].TeamID)
	}
}

func TestTemplateRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	templateRepo := NewTemplateRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-template-delete-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)

	tmpl, _ := templateRepo.CreateFromMap(team.ID, m.ID)

	err := templateRepo.Delete(tmpl.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, _ := templateRepo.FindByMapID(m.ID)
	if found != nil {
		t.Errorf("expected nil after delete, got %v", found)
	}
}

func TestTemplateRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	templateRepo := NewTemplateRepository(db)

	err := templateRepo.Delete(99999)
	if err != ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}
