//go:build integration

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
)

func TestEditLockHandler_AcquireAndRejectSecond(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Give Editor A a display name so editor_name appears in 409 response
	if err := f.userRepo.UpdateDisplayName(f.user.ID, "Editor A"); err != nil {
		t.Fatalf("failed to set display name for editor A: %v", err)
	}

	user2, _, err := f.userRepo.FindOrCreateByDiscordID("lock-editor-b-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor B: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	// Step 1: Editor A acquires lock → 200
	w := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for editor A, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	lock, ok := resp["lock"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lock object, got %T", resp["lock"])
	}
	if lock["editor_id"].(float64) != float64(f.user.ID) {
		t.Errorf("expected editor_id %d, got %v", f.user.ID, lock["editor_id"])
	}

	// Step 2: Editor B tries to acquire → 409
	jwtServiceB := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	tokenB, _, err := jwtServiceB.GenerateAccessToken(user2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w2 := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", user2, tokenB, params)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for editor B, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var errResp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errResp["error"] != "lock_already_acquired" {
		t.Errorf("expected error 'lock_already_acquired', got '%v'", errResp["error"])
	}
	if editorName, ok := errResp["editor_name"]; !ok || editorName == "" {
		t.Errorf("expected editor_name in 409 response, got '%v'", editorName)
	}
	if editorID, ok := errResp["editor_id"]; !ok || editorID == nil {
		t.Errorf("expected editor_id in 409 response, got '%v'", editorID)
	} else if editorID.(float64) != float64(f.user.ID) {
		t.Errorf("expected editor_id %d, got %v", f.user.ID, editorID)
	}

	// Step 3: Editor A creates a zone (should succeed because lock is held)
	zoneParams := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	zoneBody := `{"polygons": [{"points": [{"x": 100, "y": 100}, {"x": 200, "y": 200}, {"x": 300, "y": 100}]}], "name": "zone-from-a"}`
	w3 := executeZoneRequest(t, f.zoneHandler.CreateZone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		zoneBody, f.user, f.token, zoneParams)

	if w3.Code != http.StatusCreated {
		t.Errorf("expected 201 for zone creation by lock owner, got %d. Body: %s", w3.Code, w3.Body.String())
	}
}

func TestEditLockHandler_ReleaseAndReacquire(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	user2, _, err := f.userRepo.FindOrCreateByDiscordID("lock-release-contender-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor B: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	// Step 1: Editor A acquires lock → 200
	w := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for editor A, got %d", w.Code)
	}

	// Step 2: Editor A releases lock → 200
	wRel := executeZoneRequest(t, f.lockHandler.ReleaseLock, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)
	if wRel.Code != http.StatusOK {
		t.Fatalf("expected 200 for release, got %d. Body: %s", wRel.Code, wRel.Body.String())
	}

	// Step 3: Editor B acquires lock → 200
	jwtServiceB := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	tokenB, _, err := jwtServiceB.GenerateAccessToken(user2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	wB := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", user2, tokenB, params)
	if wB.Code != http.StatusOK {
		t.Fatalf("expected 200 for editor B, got %d. Body: %s", wB.Code, wB.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(wB.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	lock, ok := resp["lock"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lock object, got %T", resp["lock"])
	}
	if lock["editor_id"].(float64) != float64(user2.ID) {
		t.Errorf("expected editor_id %d, got %v", user2.ID, lock["editor_id"])
	}
}

func TestEditLockHandler_LockTimeout_AnotherCanAcquire(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Step 1: Editor A acquires lock
	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Step 2: Manually expire the lock
	_, err = f.db.Exec(
		"UPDATE edit_locks SET expires_at = NOW() - INTERVAL '1 minute' WHERE team_id = $1 AND map_id = $2",
		f.team.ID, f.mapObj.ID,
	)
	if err != nil {
		t.Fatalf("failed to expire lock: %v", err)
	}

	// Step 3: Another editor can now acquire the lock
	user2, _, err := f.userRepo.FindOrCreateByDiscordID("lock-timeout-contender-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor B: %v", err)
	}

	jwtServiceB := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	tokenB, _, err := jwtServiceB.GenerateAccessToken(user2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", user2, tokenB, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for editor B (lock expired), got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	lock, ok := resp["lock"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lock object, got %T", resp["lock"])
	}
	if lock["editor_id"].(float64) != float64(user2.ID) {
		t.Errorf("expected editor_id %d, got %v", user2.ID, lock["editor_id"])
	}
}

func TestEditLockHandler_GetLockStatus_WithEditorNameAndRemainingSeconds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Set a display name on the user
	err := f.userRepo.UpdateDisplayName(f.user.ID, "EditorA")
	if err != nil {
		t.Fatalf("failed to set display name: %v", err)
	}

	// Acquire the lock
	_, err = f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeZoneRequest(t, f.lockHandler.GetLockStatus, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["locked"] != true {
		t.Errorf("expected locked=true, got %v", resp["locked"])
	}
	lock, ok := resp["lock"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lock object, got %T", resp["lock"])
	}
	if lock["editor_id"].(float64) != float64(f.user.ID) {
		t.Errorf("expected editor_id %d, got %v", f.user.ID, lock["editor_id"])
	}
	if lock["editor_name"] != "EditorA" {
		t.Errorf("expected editor_name 'EditorA', got '%v'", lock["editor_name"])
	}
	rem, ok := lock["remaining_seconds"].(float64)
	if !ok || rem <= 0 {
		t.Errorf("expected positive remaining_seconds, got %v", lock["remaining_seconds"])
	}
}

func TestEditLockHandler_RefreshLockResetsRemainingSeconds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Acquire the lock
	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Wait 1 second so we can detect a refresh
	time.Sleep(1 * time.Second)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	// Refresh the lock
	w := executeZoneRequest(t, f.lockHandler.RefreshLock, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	lock, ok := resp["lock"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lock object, got %T", resp["lock"])
	}

	rem, ok := lock["remaining_seconds"].(float64)
	if !ok || rem < 890 {
		t.Errorf("expected remaining_seconds ~900 after refresh, got %v", rem)
	}
}

func TestEditLockHandler_ModifyWithoutLock_Returns423(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Create a zone first
	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 200}, {X: 300, Y: 100}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0, nil); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	// Try to add a polygon without acquiring lock first
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
	}
	body := `{"points": [{"x": 400, "y": 400}, {"x": 500, "y": 500}, {"x": 600, "y": 400}]}`
	w := executeZoneRequest(t, f.zoneHandler.AddPolygon, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/polygons", f.team.ID, zone.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusLocked {
		t.Errorf("expected 423, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestEditLockHandler_AdminCleanupExpiredLocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Editor A acquires lock
	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Manually expire the lock
	_, err = f.db.Exec(
		"UPDATE edit_locks SET expires_at = NOW() - INTERVAL '1 minute' WHERE team_id = $1 AND map_id = $2",
		f.team.ID, f.mapObj.ID,
	)
	if err != nil {
		t.Fatalf("failed to expire lock: %v", err)
	}

	// Create an admin user
	admin, token := createAdminUserForTest(t, f.userRepo)

	// Cleanup expired locks through the admin handler
	w := executeAdminHandler(t, f.lockHandler.CleanupExpiredLocks, http.MethodPost,
		"/api/v1/admin/locks/cleanup", "", admin, token,
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	cleaned, ok := resp["cleaned"].(float64)
	if !ok || cleaned < 1 {
		t.Errorf("expected cleaned >= 1, got %v", resp["cleaned"])
	}

	// Verify lock is released
	lock, err := f.lockRepo.Get(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to get lock status: %v", err)
	}
	if lock != nil {
		t.Errorf("expected lock to be nil after cleanup, got %+v", lock)
	}
}

func TestEditLockHandler_AdminForceClearAllLocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Create a second map for this test
	map2, err := f.mapRepo.Create("force-clear-second-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create second map: %v", err)
	}

	// Acquire lock on first map
	_, err = f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock on map 1: %v", err)
	}

	// Acquire lock on second map
	_, err = f.lockRepo.Acquire(f.team.ID, map2.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock on map 2: %v", err)
	}

	// Create an admin user
	admin, token := createAdminUserForTest(t, f.userRepo)

	// Force clear all locks
	w := executeAdminHandler(t, f.lockHandler.ForceClearLocks, http.MethodDelete,
		"/api/v1/admin/locks", "", admin, token,
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	cnt, ok := resp["count"].(float64)
	if !ok || cnt != 2 {
		t.Errorf("expected count=2, got %v", resp["count"])
	}

	// Verify both locks are released
	l1, _ := f.lockRepo.Get(f.team.ID, f.mapObj.ID)
	if l1 != nil {
		t.Errorf("expected lock 1 to be cleared, got %+v", l1)
	}
	l2, _ := f.lockRepo.Get(f.team.ID, map2.ID)
	if l2 != nil {
		t.Errorf("expected lock 2 to be cleared, got %+v", l2)
	}
}

func TestEditLockHandler_ReleaseLockForEditor_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	// Acquire the lock
	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	// Release lock via ReleaseLockForEditor
	w := executeZoneRequest(t, f.lockHandler.ReleaseLockForEditor, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock/release", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify lock is released
	lock, err := f.lockRepo.Get(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to get lock: %v", err)
	}
	if lock != nil {
		t.Errorf("expected lock to be nil after release, got %+v", lock)
	}
}

func TestEditLockHandler_ReleaseLockForEditor_NoLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupZoneTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	// Release when no lock exists should still return 200
	w := executeZoneRequest(t, f.lockHandler.ReleaseLockForEditor, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock/release", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
