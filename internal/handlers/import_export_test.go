//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const ieMigrationsDir = "../../db/migrations"

var ieMigrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000002_add_refresh_tokens.up.sql",
	"000003_add_display_name.up.sql",
	"000004_add_discord_username.up.sql",
	"000009_add_map_catalog.up.sql",
	"000011_add_zone_polygon_constraints.up.sql",
	"000013_add_zone_versions.up.sql",
}

func setupImportExportTestDB(t *testing.T) (*ImportExportHandler, *sql.DB, models.UserRepository, models.TeamRepository, models.MapRepository, models.ZoneRepository) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForAll(wait.ForLog("database system is ready to accept connections"), wait.ForListeningPort("5432/tcp")),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		container.Terminate(context.Background())
	})

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	for _, mf := range ieMigrationFiles {
		path := filepath.Join(ieMigrationsDir, mf)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", mf, err)
		}
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("failed to execute migration %s: %v\nSQL: %s", mf, err, stmt)
			}
		}
	}

	zoneRepo := models.NewZoneRepository(db)
	mapRepo := models.NewMapRepository(db)
	teamRepo := models.NewTeamRepository(db)
	userRepo := models.NewUserRepository(db)
	logger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	handler := NewImportExportHandler(zoneRepo, mapRepo, teamRepo, logger)
	return handler, db, userRepo, teamRepo, mapRepo, zoneRepo
}

func createTestTeam(t *testing.T, teamRepo models.TeamRepository, userRepo models.UserRepository) (*models.Team, *models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("import-export-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := userRepo.SetAdmin(user.ID); err != nil {
		t.Fatalf("failed to set admin: %v", err)
	}

	team, err := teamRepo.Create("test-team-"+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	return team, user, token
}

func createTestMap(t *testing.T, mapRepo models.MapRepository, name string) *models.Map {
	t.Helper()

	evaMap, err := mapRepo.Create(name, []byte("fake-webp-data"))
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	return evaMap
}

func createTestZone(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, pointCount int) *models.Zone {
	t.Helper()

	zone, err := zoneRepo.Create(teamID, mapID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}

	points := make([]models.Point, pointCount)
	for i := range points {
		points[i] = models.Point{X: float64(i * 10), Y: float64(i * 10)}
	}
	_, err = zoneRepo.AddPolygon(zone.ID, points, 0, nil)
	if err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	if err := zoneRepo.SetName(zone.ID, name); err != nil {
		t.Fatalf("failed to set zone name: %v", err)
	}

	return zone
}

func executeImportExportHandler(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, nil)
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	c.Set("user", user)

	// Parse path: /api/v1/admin/teams/:id/export or /api/v1/admin/teams/:id/export/:mapId or /api/v1/admin/teams/:id/import
	relative := strings.TrimPrefix(path, "/api/v1/admin/teams/")
	parts := strings.SplitN(relative, "/", 3)
	if len(parts) > 0 && parts[0] != "" {
		idStr := strings.Split(parts[0], "?")[0]
		if _, err := strconv.Atoi(idStr); err == nil {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		}
		if len(parts) > 1 {
			if parts[1] == "export" && len(parts) > 2 {
				c.Params = append(c.Params, gin.Param{Key: "mapId", Value: parts[2]})
			}
		}
	}

	handler(c)

	return w
}

func TestExportTeamMap_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, zoneRepo := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	evaMap := createTestMap(t, mapRepo, "Atlantis")
	createTestZone(t, zoneRepo, team.ID, evaMap.ID, "top", 4)
	createTestZone(t, zoneRepo, team.ID, evaMap.ID, "secret", 3)

	w := executeImportExportHandler(t, handler.ExportTeamMap, http.MethodGet,
		fmt.Sprintf("/api/v1/admin/teams/%d/export/%d", team.ID, evaMap.ID),
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["map_name"] != "Atlantis" {
		t.Errorf("expected map_name 'Atlantis', got %v", result["map_name"])
	}
	if result["team"] != team.Name {
		t.Errorf("expected team '%s', got %v", team.Name, result["team"])
	}
	if _, ok := result["export_date"]; !ok {
		t.Error("expected export_date field")
	}
	if result["data"] != nil {
		t.Error("response should not wrap in 'data' field")
	}

	zones, ok := result["zones"].([]interface{})
	if !ok {
		t.Fatalf("expected zones array, got %T", result["zones"])
	}
	if len(zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(zones))
	}
}

func TestExportTeamMap_EmptyZones(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	evaMap := createTestMap(t, mapRepo, "EmptyMap")

	w := executeImportExportHandler(t, handler.ExportTeamMap, http.MethodGet,
		fmt.Sprintf("/api/v1/admin/teams/%d/export/%d", team.ID, evaMap.ID),
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	zones := result["zones"].([]interface{})
	if len(zones) != 0 {
		t.Errorf("expected empty zones, got %d", len(zones))
	}
}

func TestExportTeamMap_TeamNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, _ := setupImportExportTestDB(t)
	_, user, token := createTestTeam(t, teamRepo, userRepo)
	evaMap := createTestMap(t, mapRepo, "OrphanMap")

	w := executeImportExportHandler(t, handler.ExportTeamMap, http.MethodGet,
		"/api/v1/admin/teams/99999/export/"+strconv.Itoa(evaMap.ID),
		"", user, token)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestExportTeamMap_MapNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, _, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)

	w := executeImportExportHandler(t, handler.ExportTeamMap, http.MethodGet,
		fmt.Sprintf("/api/v1/admin/teams/%d/export/99999", team.ID),
		"", user, token)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestExportTeamMaps_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, zoneRepo := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)

	map1 := createTestMap(t, mapRepo, "Atlantis")
	map2 := createTestMap(t, mapRepo, "Lunar")
	createTestZone(t, zoneRepo, team.ID, map1.ID, "top", 4)
	createTestZone(t, zoneRepo, team.ID, map1.ID, "secret", 3)
	createTestZone(t, zoneRepo, team.ID, map2.ID, "base", 3)

	w := executeImportExportHandler(t, handler.ExportTeamMaps, http.MethodGet,
		fmt.Sprintf("/api/v1/admin/teams/%d/export", team.ID),
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 map exports, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r["map_name"].(string)] = true
	}
	if !names["Atlantis"] || !names["Lunar"] {
		t.Errorf("expected both Atlantis and Lunar maps, got %v", names)
	}
}

func TestImportTeamZones_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	createTestMap(t, mapRepo, "Atlantis")

	importBody := fmt.Sprintf(`[
		{
			"map_name": "Atlantis",
			"zones": [
				{
					"name": "imported-top",
					"polygons": [
						{"points": [{"x": 0, "y": 0}, {"x": 100, "y": 0}, {"x": 100, "y": 100}, {"x": 0, "y": 100}]}
					]
				}
			]
		}
	]`)

	w := executeImportExportHandler(t, handler.ImportTeamZones, http.MethodPost,
		fmt.Sprintf("/api/v1/admin/teams/%d/import", team.ID),
		importBody, user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["imported_count"] != 1.0 {
		t.Errorf("expected imported_count 1, got %v", result["imported_count"])
	}
	if result["maps_processed"] != 1.0 {
		t.Errorf("expected maps_processed 1, got %v", result["maps_processed"])
	}
	if result["errors"] != nil {
		t.Errorf("expected no errors, got %v", result["errors"])
	}
}

func TestImportTeamZones_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, zoneRepo := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	evaMap := createTestMap(t, mapRepo, "Atlantis")
	createTestZone(t, zoneRepo, team.ID, evaMap.ID, "existing-zone", 4)

	importBody := fmt.Sprintf(`[
		{
			"map_name": "Atlantis",
			"zones": [
				{
					"name": "existing-zone",
					"polygons": [
						{"points": [{"x": 0, "y": 0}, {"x": 100, "y": 0}, {"x": 100, "y": 100}, {"x": 0, "y": 100}]}
					]
				}
			]
		}
	]`)

	w := executeImportExportHandler(t, handler.ImportTeamZones, http.MethodPost,
		fmt.Sprintf("/api/v1/admin/teams/%d/import", team.ID),
		importBody, user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["imported_count"] != 0.0 {
		t.Errorf("expected imported_count 0, got %v", result["imported_count"])
	}

	errors, ok := result["errors"].([]interface{})
	if !ok || len(errors) == 0 {
		t.Fatalf("expected errors array, got %v", result["errors"])
	}
}

func TestImportTeamZones_MapNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, _, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)

	importBody := `[
		{
			"map_name": "NonExistentMap",
			"zones": []
		}
	]`

	w := executeImportExportHandler(t, handler.ImportTeamZones, http.MethodPost,
		fmt.Sprintf("/api/v1/admin/teams/%d/import", team.ID),
		importBody, user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	errors, ok := result["errors"].([]interface{})
	if !ok || len(errors) == 0 {
		t.Fatalf("expected errors for missing map, got %v", result["errors"])
	}

	firstErr := errors[0].(map[string]interface{})
	if firstErr["error"] != "map not found" {
		t.Errorf("expected 'map not found' error, got %v", firstErr["error"])
	}
}

func TestImportTeamZones_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, _, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)

	w := executeImportExportHandler(t, handler.ImportTeamZones, http.MethodPost,
		fmt.Sprintf("/api/v1/admin/teams/%d/import", team.ID),
		"{invalid}", user, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestImportTeamZones_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	createTestMap(t, mapRepo, "Atlantis")

	w := executeImportExportHandler(t, handler.ImportTeamZones, http.MethodPost,
		fmt.Sprintf("/api/v1/admin/teams/%d/import", team.ID),
		"[]", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["imported_count"] != 0.0 {
		t.Errorf("expected imported_count 0 for empty body, got %v", result["imported_count"])
	}
}

func TestExportTeamMaps_NoZones(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, userRepo, teamRepo, mapRepo, _ := setupImportExportTestDB(t)
	team, user, token := createTestTeam(t, teamRepo, userRepo)
	createTestMap(t, mapRepo, "Atlantis")
	createTestMap(t, mapRepo, "Lunar")

	w := executeImportExportHandler(t, handler.ExportTeamMaps, http.MethodGet,
		fmt.Sprintf("/api/v1/admin/teams/%d/export", team.ID),
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var results []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &results)

	if len(results) != 0 {
		t.Errorf("expected no map exports (no zones), got %d", len(results))
	}
}
