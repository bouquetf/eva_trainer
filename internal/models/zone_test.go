//go:build integration

package models

import (
	"database/sql"
	"strings"
	"testing"
)

func createTestTeamForZone(t *testing.T, db *sql.DB, userRepo UserRepository, suffix string) (*User, *Team) {
	t.Helper()
	user, _, err := userRepo.FindOrCreateByDiscordID("zone-test-user-" + suffix)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	teamRepo := NewTeamRepository(db)
	team, err := teamRepo.Create("Zone Test Team "+suffix, user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	return user, team
}

func createTestMapForZone(t *testing.T, db *sql.DB) *Map {
	t.Helper()
	mapRepo := NewMapRepository(db)
	m, err := mapRepo.Create("zone-test-map", nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	return m
}

func TestZoneRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "create")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if zone.ID <= 0 {
		t.Errorf("expected positive id, got %d", zone.ID)
	}
	if zone.TeamID != team.ID {
		t.Errorf("expected team_id %d, got %d", team.ID, zone.TeamID)
	}
	if zone.MapID != m.ID {
		t.Errorf("expected map_id %d, got %d", m.ID, zone.MapID)
	}
	if zone.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if zone.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestZoneRepository_CreateWithPolygonAndName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "poly-name")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)

	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	points := []Point{{X: 100, Y: 100}, {X: 200, Y: 200}, {X: 300, Y: 100}}
	polygon, err := zoneRepo.AddPolygon(zone.ID, points, 0, nil)
	if err != nil {
		t.Fatalf("AddPolygon failed: %v", err)
	}
	if polygon.ID <= 0 {
		t.Errorf("expected positive polygon id, got %d", polygon.ID)
	}
	if len(polygon.Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(polygon.Points))
	}

	if err := zoneRepo.SetName(zone.ID, "top"); err != nil {
		t.Fatalf("SetName failed: %v", err)
	}

	details, err := zoneRepo.FindWithDetailsByTeamAndMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("FindWithDetailsByTeamAndMap failed: %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(details))
	}
	if len(details[0].Polygons) != 1 {
		t.Errorf("expected 1 polygon, got %d", len(details[0].Polygons))
	}
	if details[0].Name != "top" {
		t.Errorf("expected name 'top', got '%s'", details[0].Name)
	}
}

func TestZoneRepository_AddMultiplePolygons(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "multi-poly")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		points := []Point{
			{X: float64(i * 100), Y: float64(i * 100)},
			{X: float64(i*100 + 100), Y: float64(i * 100)},
			{X: float64(i*100 + 50), Y: float64(i*100 + 100)},
		}
		_, err := zoneRepo.AddPolygon(zone.ID, points, i, nil)
		if err != nil {
			t.Fatalf("AddPolygon %d failed: %v", i, err)
		}
	}

	polygons, err := zoneRepo.GetPolygons(zone.ID)
	if err != nil {
		t.Fatalf("GetPolygons failed: %v", err)
	}
	if len(polygons) != 3 {
		t.Fatalf("expected 3 polygons, got %d", len(polygons))
	}
	for i, p := range polygons {
		if p.OrderIndex != i {
			t.Errorf("polygon %d: expected order_index %d, got %d", p.ID, i, p.OrderIndex)
		}
	}
}

func TestZoneRepository_RemoveLastPolygonDeletesZone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "remove-last")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	points := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	polygon, err := zoneRepo.AddPolygon(zone.ID, points, 0, nil)
	if err != nil {
		t.Fatalf("AddPolygon failed: %v", err)
	}

	if err := zoneRepo.RemovePolygon(polygon.ID); err != nil {
		t.Fatalf("RemovePolygon failed: %v", err)
	}

	count, err := zoneRepo.PolygonCount(zone.ID)
	if err != nil {
		t.Fatalf("PolygonCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected polygon count 0, got %d", count)
	}

	if err := zoneRepo.Delete(zone.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestZoneRepository_IsNameTaken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "name-taken")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)

	zoneA, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create zone A failed: %v", err)
	}

	points := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := zoneRepo.AddPolygon(zoneA.ID, points, 0, nil); err != nil {
		t.Fatalf("AddPolygon for zone A failed: %v", err)
	}

	if err := zoneRepo.SetName(zoneA.ID, "top"); err != nil {
		t.Fatalf("SetName for zone A failed: %v", err)
	}

	zoneB, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create zone B failed: %v", err)
	}
	if _, err := zoneRepo.AddPolygon(zoneB.ID, points, 0, nil); err != nil {
		t.Fatalf("AddPolygon for zone B failed: %v", err)
	}

	taken, err := zoneRepo.IsNameTaken(team.ID, m.ID, "top")
	if err != nil {
		t.Fatalf("IsNameTaken failed: %v", err)
	}
	if !taken {
		t.Error("expected IsNameTaken to be true, got false")
	}

	notTaken, err := zoneRepo.IsNameTaken(team.ID, m.ID, "nonexistent")
	if err != nil {
		t.Fatalf("IsNameTaken failed: %v", err)
	}
	if notTaken {
		t.Error("expected IsNameTaken to be false for nonexistent name, got true")
	}
}

func TestZoneRepository_NameTooLong(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "name-long")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	points := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := zoneRepo.AddPolygon(zone.ID, points, 0, nil); err != nil {
		t.Fatalf("AddPolygon failed: %v", err)
	}

	longName := strings.Repeat("a", 300)
	err = zoneRepo.SetName(zone.ID, longName)
	if err != ErrZoneNameTooLong {
		t.Errorf("expected ErrZoneNameTooLong, got %v", err)
	}
}

func TestZoneRepository_PolygonTooFewPoints(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "too-few")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	points := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}}
	_, err = zoneRepo.AddPolygon(zone.ID, points, 0, nil)
	if err != ErrPolygonHasTooFewPoints {
		t.Errorf("expected ErrPolygonHasTooFewPoints, got %v", err)
	}
}

func TestZoneRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	zoneRepo := NewZoneRepository(db)
	_, err := zoneRepo.FindByID(99999)
	if err != ErrZoneNotFound {
		t.Errorf("expected ErrZoneNotFound, got %v", err)
	}
}

func TestZoneRepository_SetName_RequiresName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "req-name")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = zoneRepo.SetName(zone.ID, "  ")
	if err != ErrZoneNameRequired {
		t.Errorf("expected ErrZoneNameRequired, got %v", err)
	}
}

func TestZoneRepository_DuplicateNameOnSameTeamMapFails(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "dup-name")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)

	zoneA, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create zone A failed: %v", err)
	}
	pts := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := zoneRepo.AddPolygon(zoneA.ID, pts, 0, nil); err != nil {
		t.Fatalf("AddPolygon zone A failed: %v", err)
	}
	if err := zoneRepo.SetName(zoneA.ID, "dup-name"); err != nil {
		t.Fatalf("SetName zone A failed: %v", err)
	}

	zoneB, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create zone B failed: %v", err)
	}
	if _, err := zoneRepo.AddPolygon(zoneB.ID, pts, 0, nil); err != nil {
		t.Fatalf("AddPolygon zone B failed: %v", err)
	}

	err = zoneRepo.SetName(zoneB.ID, "dup-name")
	if err != ErrZoneNameTaken {
		t.Errorf("expected ErrZoneNameTaken, got %v", err)
	}
}

func TestZoneRepository_GetName_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	zoneRepo := NewZoneRepository(db)
	_, err := zoneRepo.GetName(99999)
	if err != ErrZoneNotFound {
		t.Errorf("expected ErrZoneNotFound, got %v", err)
	}
}

func TestZoneRepository_RemovePolygon_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	zoneRepo := NewZoneRepository(db)
	err := zoneRepo.RemovePolygon(99999)
	if err != ErrPolygonNotFound {
		t.Errorf("expected ErrPolygonNotFound, got %v", err)
	}
}

func TestZoneRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	zoneRepo := NewZoneRepository(db)
	err := zoneRepo.Delete(99999)
	if err != ErrZoneNotFound {
		t.Errorf("expected ErrZoneNotFound, got %v", err)
	}
}

func ptr(s string) *string {
	return &s
}

func TestValidateFloor(t *testing.T) {
	tests := []struct {
		name  string
		floor *string
		want  error
	}{
		{"nil", nil, nil},
		{"haut", ptr("haut"), nil},
		{"bas", ptr("bas"), nil},
		{"empty string", ptr(""), nil},
		{"middle", ptr("middle"), ErrInvalidFloorValue},
		{"HAUT uppercase", ptr("HAUT"), ErrInvalidFloorValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateFloor(tt.floor)
			if got != tt.want {
				t.Errorf("validateFloor(%v) = %v, want %v", tt.floor, got, tt.want)
			}
		})
	}
}

func TestZoneRepository_UpdatePolygonPoints_TooFew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	_, team := createTestTeamForZone(t, db, userRepo, "upd-too-few")
	m := createTestMapForZone(t, db)

	zoneRepo := NewZoneRepository(db)
	zone, err := zoneRepo.Create(team.ID, m.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	pts := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	polygon, err := zoneRepo.AddPolygon(zone.ID, pts, 0, nil)
	if err != nil {
		t.Fatalf("AddPolygon failed: %v", err)
	}

	err = zoneRepo.UpdatePolygonPoints(polygon.ID, []Point{{X: 0, Y: 0}, {X: 10, Y: 0}})
	if err != ErrPolygonHasTooFewPoints {
		t.Errorf("expected ErrPolygonHasTooFewPoints, got %v", err)
	}
}
