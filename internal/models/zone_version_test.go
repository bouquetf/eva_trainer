//go:build integration

package models

import (
	"testing"
)

func TestZoneVersionRepository_Save_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-zone-version-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)
	zoneRepo.AddPolygon(zone.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}, 0, nil)

	snapshot := VersionSnapshot{
		Name: "Test Zone",
		Polygons: []SnapshotPolygon{{
			Points:     []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}},
			OrderIndex: 0,
		}},
	}

	version, err := versionRepo.Save(zone.ID, &user.ID, snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if version.ZoneID != zone.ID {
		t.Errorf("expected zone_id %d, got %d", zone.ID, version.ZoneID)
	}
	if version.VersionData.Name != "Test Zone" {
		t.Errorf("expected name 'Test Zone', got '%s'", version.VersionData.Name)
	}
	if version.EditorID == nil || *version.EditorID != user.ID {
		t.Errorf("expected editor_id %d, got %v", user.ID, version.EditorID)
	}
	if version.ID <= 0 {
		t.Errorf("expected positive id, got %d", version.ID)
	}
	if len(version.VersionData.Polygons) != 1 {
		t.Errorf("expected 1 polygon, got %d", len(version.VersionData.Polygons))
	}
}

func TestZoneVersionRepository_Save_WithoutEditor(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-zone-version-no-editor-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	snapshot := VersionSnapshot{
		Name: "No Editor Zone",
	}

	version, err := versionRepo.Save(zone.ID, nil, snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if version.EditorID != nil {
		t.Errorf("expected nil editor_id, got %v", version.EditorID)
	}
}

func TestZoneVersionRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-findbyid-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	snapshot := VersionSnapshot{Name: "Find By ID"}
	saved, _ := versionRepo.Save(zone.ID, &user.ID, snapshot)

	found, err := versionRepo.FindByID(saved.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ID != saved.ID {
		t.Errorf("expected id %d, got %d", saved.ID, found.ID)
	}
	if found.VersionData.Name != "Find By ID" {
		t.Errorf("expected name 'Find By ID', got '%s'", found.VersionData.Name)
	}
}

func TestZoneVersionRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	versionRepo := NewZoneVersionRepository(db)
	_, err := versionRepo.FindByID(99999)
	if err != ErrVersionNotFound {
		t.Errorf("expected ErrVersionNotFound, got %v", err)
	}
}

func TestZoneVersionRepository_FindByZone_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-pagination-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	for i := 0; i < 5; i++ {
		snapshot := VersionSnapshot{Name: "Version"}
		versionRepo.Save(zone.ID, &user.ID, snapshot)
	}

	page1, total, err := versionRepo.FindByZone(zone.ID, 1, 2)
	if err != nil {
		t.Fatalf("FindByZone page 1 failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(page1) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(page1))
	}

	page3, _, err := versionRepo.FindByZone(zone.ID, 3, 2)
	if err != nil {
		t.Fatalf("FindByZone page 3 failed: %v", err)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 result on page 3, got %d", len(page3))
	}
}

func TestZoneVersionRepository_FindByMap_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-findbymap-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)

	zone1, _ := zoneRepo.Create(team.ID, m.ID)
	zone2, _ := zoneRepo.Create(team.ID, m.ID)

	versionRepo.Save(zone1.ID, &user.ID, VersionSnapshot{Name: "Zone 1"})
	versionRepo.Save(zone2.ID, &user.ID, VersionSnapshot{Name: "Zone 2"})

	versions, total, err := versionRepo.FindByMap(team.ID, m.ID, 1, 20)
	if err != nil {
		t.Fatalf("FindByMap failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestZoneVersionRepository_GetLatestForZone_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-latest-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	versionRepo.Save(zone.ID, &user.ID, VersionSnapshot{Name: "Older"})
	versionRepo.Save(zone.ID, &user.ID, VersionSnapshot{Name: "Newer"})

	latest, err := versionRepo.GetLatestForZone(zone.ID)
	if err != nil {
		t.Fatalf("GetLatestForZone failed: %v", err)
	}
	if latest == nil {
		t.Fatal("expected a version, got nil")
	}
	if latest.VersionData.Name != "Newer" {
		t.Errorf("expected 'Newer', got '%s'", latest.VersionData.Name)
	}
}

func TestZoneVersionRepository_GetLatestForZone_NoVersions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-latest-none-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	latest, err := versionRepo.GetLatestForZone(zone.ID)
	if err != nil {
		t.Fatalf("GetLatestForZone failed: %v", err)
	}
	if latest != nil {
		t.Errorf("expected nil, got %v", latest)
	}
}

func TestZoneVersionRepository_CountByZone_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-version-count-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	count, err := versionRepo.CountByZone(zone.ID)
	if err != nil {
		t.Fatalf("CountByZone failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	versionRepo.Save(zone.ID, &user.ID, VersionSnapshot{Name: "V1"})
	versionRepo.Save(zone.ID, &user.ID, VersionSnapshot{Name: "V2"})

	count, err = versionRepo.CountByZone(zone.ID)
	if err != nil {
		t.Fatalf("CountByZone failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestZoneVersionRepository_FindFirstOtherEditorVersion_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user1, _, _ := userRepo.FindOrCreateByDiscordID("test-first-other-1-" + t.Name())
	user2, _, _ := userRepo.FindOrCreateByDiscordID("test-first-other-2-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user1.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	v1, _ := versionRepo.Save(zone.ID, &user1.ID, VersionSnapshot{Name: "User1 V1"})
	versionRepo.Save(zone.ID, &user2.ID, VersionSnapshot{Name: "User2 V1"})

	otherID, err := versionRepo.FindFirstOtherEditorVersion(zone.ID, user1.ID)
	if err != nil {
		t.Fatalf("FindFirstOtherEditorVersion failed: %v", err)
	}
	if otherID == nil {
		t.Fatal("expected a version id, got nil")
	}
	if *otherID != v1.ID+1 {
		t.Errorf("expected version id %d, got %d", v1.ID+1, *otherID)
	}
}

func TestZoneVersionRepository_FindFirstOtherEditorVersion_None(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	versionRepo := NewZoneVersionRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-first-other-none-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	versionRepo.Save(zone.ID, &user.ID, VersionSnapshot{Name: "V1"})

	otherID, err := versionRepo.FindFirstOtherEditorVersion(zone.ID, user.ID)
	if err != nil {
		t.Fatalf("FindFirstOtherEditorVersion failed: %v", err)
	}
	if otherID != nil {
		t.Errorf("expected nil, got %d", *otherID)
	}
}
