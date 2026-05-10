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

const zoneMigrationsDir = "../../db/migrations"

var zoneMigrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000002_add_refresh_tokens.up.sql",
	"000003_add_display_name.up.sql",
	"000004_add_discord_username.up.sql",
	"000011_add_zone_polygon_constraints.up.sql",
}

func setupZoneTestDB(t *testing.T) *sql.DB {
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

	for _, mf := range zoneMigrationFiles {
		path := filepath.Join(zoneMigrationsDir, mf)
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

	return db
}

type zoneTestFixtures struct {
	db             *sql.DB
	zoneHandler    *ZoneHandler
	lockHandler    *EditLockHandler
	userRepo       models.UserRepository
	zoneRepo       models.ZoneRepository
	lockRepo       models.EditLockRepository
	mapRepo        models.MapRepository
	historyRepo    models.EditHistoryRepository
	teamMemberRepo models.TeamMemberRepository
	user           *models.User
	team           *models.Team
	mapObj         *models.Map
	token          string
}

func setupZoneTest(t *testing.T) *zoneTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupZoneTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)

	user, _, err := userRepo.FindOrCreateByDiscordID("zone-handler-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("Zone Handler Test Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("zone-handler-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	zoneHandler := NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo)
	lockHandler := NewEditLockHandler(lockRepo, userRepo)

	return &zoneTestFixtures{
		db:             db,
		zoneHandler:    zoneHandler,
		lockHandler:    lockHandler,
		userRepo:       userRepo,
		zoneRepo:       zoneRepo,
		lockRepo:       lockRepo,
		mapRepo:        mapRepo,
		historyRepo:    historyRepo,
		teamMemberRepo: teamMemberRepo,
		user:           user,
		team:           team,
		mapObj:         m,
		token:          token,
	}
}

func executeZoneRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
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
	c.Params = params

	handler(c)

	return w
}

func acquireLockForTest(t *testing.T, f *zoneTestFixtures) {
	t.Helper()
	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.user.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
}

func TestZoneHandler_ListZones_Success(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}
	if err := f.zoneRepo.SetName(zone.ID, "Zone A"); err != nil {
		t.Fatalf("failed to set name: %v", err)
	}

	zone2, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone 2: %v", err)
	}
	pts2 := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 150, Y: 200}}
	if _, err := f.zoneRepo.AddPolygon(zone2.ID, pts2, 0); err != nil {
		t.Fatalf("failed to add polygon 2: %v", err)
	}
	if err := f.zoneRepo.SetName(zone2.ID, "Zone B"); err != nil {
		t.Fatalf("failed to set name: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeZoneRequest(t, f.zoneHandler.ListZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	zones, ok := resp["zones"].([]interface{})
	if !ok {
		t.Fatalf("expected zones array, got %T", resp["zones"])
	}
	if len(zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(zones))
	}
	first := zones[0].(map[string]interface{})
	if first["name"] != "Zone A" {
		t.Errorf("expected name 'Zone A', got %v", first["name"])
	}
	polygons, ok := first["polygons"].([]interface{})
	if !ok || len(polygons) != 1 {
		t.Errorf("expected 1 polygon, got %d", len(polygons))
	}
}

func TestZoneHandler_CreateZone_Success(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	body := `{"polygons": [{"points": [{"x": 100, "y": 100}, {"x": 200, "y": 200}, {"x": 300, "y": 100}]}], "name": "top"}`
	w := executeZoneRequest(t, f.zoneHandler.CreateZone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	zone, ok := resp["zone"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected zone object, got %T", resp["zone"])
	}
	if zone["id"] == nil || zone["id"].(float64) <= 0 {
		t.Errorf("expected positive id, got %v", zone["id"])
	}
	if zone["name"] != "top" {
		t.Errorf("expected name 'top', got %v", zone["name"])
	}
}

func TestZoneHandler_CreateZone_NoName(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	body := `{"polygons": [{"points": [{"x": 100, "y": 100}, {"x": 200, "y": 200}, {"x": 300, "y": 100}]}]}`
	w := executeZoneRequest(t, f.zoneHandler.CreateZone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	zone, ok := resp["zone"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected zone object, got %T", resp["zone"])
	}
	name, _ := zone["name"].(string)
	if name != "" {
		t.Errorf("expected empty name, got '%s'", name)
	}
}

func TestZoneHandler_CreateZone_DuplicateName(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	body := `{"polygons": [{"points": [{"x": 100, "y": 100}, {"x": 200, "y": 200}, {"x": 300, "y": 100}]}], "name": "dupname"}`
	w := executeZoneRequest(t, f.zoneHandler.CreateZone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for first zone, got %d. Body: %s", w.Code, w.Body.String())
	}

	w2 := executeZoneRequest(t, f.zoneHandler.CreateZone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/zones", f.team.ID, f.mapObj.ID),
		body, f.user, f.token, params)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate name, got %d. Body: %s", w2.Code, w2.Body.String())
	}
}

func TestZoneHandler_DeleteZone_Success(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
	}
	w := executeZoneRequest(t, f.zoneHandler.DeleteZone, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d", f.team.ID, zone.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	_, err = f.zoneRepo.FindByID(zone.ID)
	if err != models.ErrZoneNotFound {
		t.Errorf("expected zone to be deleted, got %v", err)
	}
}

func TestZoneHandler_DeleteZone_NotFound(t *testing.T) {
	f := setupZoneTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: "99999"},
	}
	w := executeZoneRequest(t, f.zoneHandler.DeleteZone, http.MethodDelete,
		"/api/v1/teams/99999/zones/99999",
		"", f.user, f.token, params)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestZoneHandler_AddPolygon_Success(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}

	pts := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 200}, {X: 300, Y: 100}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
	}

	body := `{"points": [{"x": 400, "y": 400}, {"x": 500, "y": 500}, {"x": 600, "y": 400}]}`
	w := executeZoneRequest(t, f.zoneHandler.AddPolygon, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/polygons", f.team.ID, zone.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	polygon, ok := resp["polygon"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected polygon object, got %T", resp["polygon"])
	}
	if polygon["id"] == nil || polygon["id"].(float64) <= 0 {
		t.Errorf("expected positive polygon id, got %v", polygon["id"])
	}
}

func TestZoneHandler_RemovePolygon_LastDeletesZone(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	polygon, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0)
	if err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
		{Key: "polygonId", Value: fmt.Sprintf("%d", polygon.ID)},
	}
	w := executeZoneRequest(t, f.zoneHandler.RemovePolygon, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/polygons/%d", f.team.ID, zone.ID, polygon.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["zone_deleted"] != true {
		t.Errorf("expected zone_deleted true, got %v", resp["zone_deleted"])
	}
}

func TestZoneHandler_SetZoneName_Success(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
	}
	body := `{"name": "secret"}`
	w := executeZoneRequest(t, f.zoneHandler.SetZoneName, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/name", f.team.ID, zone.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	nameObj, ok := resp["name"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected name object, got %T", resp["name"])
	}
	if nameObj["name"] != "secret" {
		t.Errorf("expected name 'secret', got %v", nameObj["name"])
	}
}

func TestZoneHandler_AddPolygonPoint_Success(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	polygon, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0)
	if err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
		{Key: "polygonId", Value: fmt.Sprintf("%d", polygon.ID)},
	}
	body := `{"point": {"x": 150, "y": 150}, "index_after": 0}`
	w := executeZoneRequest(t, f.zoneHandler.AddPolygonPoint, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/polygons/%d/points", f.team.ID, zone.ID, polygon.ID),
		body, f.user, f.token, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	resultPolygon, ok := resp["polygon"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected polygon object, got %T", resp["polygon"])
	}
	points, ok := resultPolygon["points"].([]interface{})
	if !ok {
		t.Fatalf("expected points array, got %T", resultPolygon["points"])
	}
	if len(points) != 4 {
		t.Errorf("expected 4 points, got %d", len(points))
	}
}

func TestZoneHandler_RemovePolygonPoint_FewerThan3Returns400(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	polygon, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0)
	if err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
		{Key: "polygonId", Value: fmt.Sprintf("%d", polygon.ID)},
		{Key: "pointIndex", Value: "0"},
	}
	w := executeZoneRequest(t, f.zoneHandler.RemovePolygonPoint, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/polygons/%d/points/0", f.team.ID, zone.ID, polygon.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for removing point when only 3 exist, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestZoneHandler_AcquireLock_Success(t *testing.T) {
	f := setupZoneTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
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
	if lock["editor_id"].(float64) != float64(f.user.ID) {
		t.Errorf("expected editor_id %d, got %v", f.user.ID, lock["editor_id"])
	}
}

func TestZoneHandler_AcquireLock_AlreadyTaken(t *testing.T) {
	f := setupZoneTest(t)

	user2, _, err := f.userRepo.FindOrCreateByDiscordID("zone-lock-contender-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user 2: %v", err)
	}

	_, err = f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock for user 2: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeZoneRequest(t, f.lockHandler.AcquireLock, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lock", f.team.ID, f.mapObj.ID),
		"", f.user, f.token, params)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestZoneHandler_RefreshLock_Success(t *testing.T) {
	f := setupZoneTest(t)
	acquireLockForTest(t, f)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
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
	expiresAt, ok := lock["expires_at"].(string)
	if !ok || expiresAt == "" {
		t.Errorf("expected expires_at to be set, got %v", lock["expires_at"])
	}
}

func TestZoneHandler_ModifyWithoutLock_Returns423(t *testing.T) {
	f := setupZoneTest(t)

	zone, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}

	pts := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 200}, {X: 300, Y: 100}}
	if _, err := f.zoneRepo.AddPolygon(zone.ID, pts, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}

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
