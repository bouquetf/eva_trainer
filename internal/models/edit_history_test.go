//go:build integration

package models

import (
	"encoding/json"
	"testing"
)

func TestEditHistoryRepository_Log_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-log-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	oldData := json.RawMessage(`{"name": "old_name"}`)
	newData := json.RawMessage(`{"name": "new_name"}`)

	entry, err := historyRepo.Log(zone.ID, &user.ID, "rename", oldData, newData)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	if entry.ZoneID != zone.ID {
		t.Errorf("expected zone_id %d, got %d", zone.ID, entry.ZoneID)
	}
	if entry.Action != "rename" {
		t.Errorf("expected action 'rename', got '%s'", entry.Action)
	}
	if entry.EditorID == nil || *entry.EditorID != user.ID {
		t.Errorf("expected editor_id %d, got %v", user.ID, entry.EditorID)
	}
	if entry.ID <= 0 {
		t.Errorf("expected positive id, got %d", entry.ID)
	}
}

func TestEditHistoryRepository_Log_WithoutData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-log-nodata-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	entry, err := historyRepo.Log(zone.ID, &user.ID, "delete", nil, nil)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	if entry.OldData != nil {
		t.Errorf("expected nil OldData, got %v", entry.OldData)
	}
	if entry.NewData != nil {
		t.Errorf("expected nil NewData, got %v", entry.NewData)
	}
}

func TestEditHistoryRepository_FindByZone_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-findzone-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	historyRepo.Log(zone.ID, &user.ID, "create", nil, json.RawMessage(`{"name": "new"}`))
	historyRepo.Log(zone.ID, &user.ID, "rename", json.RawMessage(`{"name": "old"}`), json.RawMessage(`{"name": "new"}`))

	entries, err := historyRepo.FindByZone(zone.ID)
	if err != nil {
		t.Fatalf("FindByZone failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Action != "rename" {
		t.Errorf("expected first entry action 'rename' (most recent), got '%s'", entries[0].Action)
	}
	if entries[1].Action != "create" {
		t.Errorf("expected second entry action 'create', got '%s'", entries[1].Action)
	}
}

func TestEditHistoryRepository_FindByZone_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-pagination-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	for i := 0; i < 5; i++ {
		historyRepo.Log(zone.ID, &user.ID, "edit", nil, nil)
	}

	page1, total, err := historyRepo.FindByZoneWithFilters(zone.ID, 1, 2)
	if err != nil {
		t.Fatalf("FindByZoneWithFilters page 1 failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(page1) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(page1))
	}

	page3, _, err := historyRepo.FindByZoneWithFilters(zone.ID, 3, 2)
	if err != nil {
		t.Fatalf("FindByZoneWithFilters page 3 failed: %v", err)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 result on page 3, got %d", len(page3))
	}
}

func TestEditHistoryRepository_FindByTeamAndMap_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-findteam-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)

	zone1, _ := zoneRepo.Create(team.ID, m.ID)
	zone2, _ := zoneRepo.Create(team.ID, m.ID)

	historyRepo.Log(zone1.ID, &user.ID, "create", nil, nil)
	historyRepo.Log(zone2.ID, &user.ID, "rename", nil, nil)

	entries, err := historyRepo.FindByTeamAndMap(team.ID, m.ID)
	if err != nil {
		t.Fatalf("FindByTeamAndMap failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestEditHistoryRepository_FindByTeamAndMap_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-team-page-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	for i := 0; i < 3; i++ {
		historyRepo.Log(zone.ID, &user.ID, "edit", nil, nil)
	}

	entries, total, err := historyRepo.FindByTeamAndMapWithFilters(team.ID, m.ID, 1, 2)
	if err != nil {
		t.Fatalf("FindByTeamAndMapWithFilters failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestEditHistoryRepository_Log_PreservesData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-preserve-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	oldData := json.RawMessage(`{"name":"before","value":42}`)
	newData := json.RawMessage(`{"name":"after","value":99}`)

	entry, err := historyRepo.Log(zone.ID, &user.ID, "update", oldData, newData)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if entry.OldData == nil {
		t.Fatal("expected OldData to be non-nil")
	}
	var oldParsed map[string]interface{}
	if err := json.Unmarshal(*entry.OldData, &oldParsed); err != nil {
		t.Fatalf("failed to unmarshal OldData: %v", err)
	}
	if oldParsed["name"] != "before" {
		t.Errorf("expected OldData.name 'before', got '%v'", oldParsed["name"])
	}
	if oldParsed["value"] != float64(42) {
		t.Errorf("expected OldData.value 42, got %v", oldParsed["value"])
	}

	if entry.NewData == nil {
		t.Fatal("expected NewData to be non-nil")
	}
	var newParsed map[string]interface{}
	if err := json.Unmarshal(*entry.NewData, &newParsed); err != nil {
		t.Fatalf("failed to unmarshal NewData: %v", err)
	}
	if newParsed["name"] != "after" {
		t.Errorf("expected NewData.name 'after', got '%v'", newParsed["name"])
	}
	if newParsed["value"] != float64(99) {
		t.Errorf("expected NewData.value 99, got %v", newParsed["value"])
	}
}

func TestEditHistoryRepository_Log_WithoutEditor(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	mapRepo := NewMapRepository(db)
	zoneRepo := NewZoneRepository(db)
	historyRepo := NewEditHistoryRepository(db)

	user, _, _ := userRepo.FindOrCreateByDiscordID("test-edit-noeditor-" + t.Name())
	team, _ := teamRepo.Create("Test Team "+t.Name(), user.ID)
	m, _ := mapRepo.Create("test-map-"+t.Name(), nil)
	zone, _ := zoneRepo.Create(team.ID, m.ID)

	entry, err := historyRepo.Log(zone.ID, nil, "system", nil, nil)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	if entry.EditorID != nil {
		t.Errorf("expected nil editor_id, got %v", entry.EditorID)
	}
}
