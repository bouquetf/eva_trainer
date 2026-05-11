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

func setupRevisionTestDB(t *testing.T) *sql.DB {
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

	migrationsDir := "../../db/migrations"
	migrationFiles := []string{
		"000001_initial_schema.up.sql",
		"000002_add_refresh_tokens.up.sql",
		"000003_add_display_name.up.sql",
		"000004_add_discord_username.up.sql",
		"000011_add_zone_polygon_constraints.up.sql",
		"000013_add_zone_versions.up.sql",
		"000015_add_template_is_public.up.sql",
	}

	for _, mf := range migrationFiles {
		path := filepath.Join(migrationsDir, mf)
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

type revisionTestFixtures struct {
	db              *sql.DB
	revisionHandler *RevisionHandler
	zoneHandler     *ZoneHandler
	userRepo        models.UserRepository
	zoneRepo        models.ZoneRepository
	mapRepo         models.MapRepository
	teamRepo        models.TeamRepository
	teamMemberRepo  models.TeamMemberRepository
	lockRepo        models.EditLockRepository
	captainA        *models.User
	captainB        *models.User
	editor          *models.User
	member          *models.User
	teamA           *models.Team
	teamB           *models.Team
	mapObj          *models.Map
	captainAToken   string
	captainBToken   string
	editorToken     string
	memberToken     string
	jwtService      *services.JWTService
}

func setupRevisionTest(t *testing.T) *revisionTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupRevisionTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	versionRepo := models.NewZoneVersionRepository(db)
	testLogger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	captainA, _, err := userRepo.FindOrCreateByDiscordID("revision-test-captain-a-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain A: %v", err)
	}

	captainB, _, err := userRepo.FindOrCreateByDiscordID("revision-test-captain-b-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain B: %v", err)
	}

	editor, _, err := userRepo.FindOrCreateByDiscordID("revision-test-editor-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("revision-test-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	teamA, err := teamRepo.Create("Revision Test Team A "+t.Name(), captainA.ID)
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}

	teamB, err := teamRepo.Create("Revision Test Team B "+t.Name(), captainB.ID)
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}

	if err := teamMemberRepo.AddMember(teamA.ID, editor.ID, models.RoleEditor); err != nil {
		t.Fatalf("failed to add editor to team A: %v", err)
	}
	if err := teamMemberRepo.AddMember(teamA.ID, member.ID, models.RoleMember); err != nil {
		t.Fatalf("failed to add member to team A: %v", err)
	}

	m, err := mapRepo.Create("revision-test-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	captainAToken, _, err := jwtService.GenerateAccessToken(captainA.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	captainBToken, _, err := jwtService.GenerateAccessToken(captainB.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	editorToken, _, err := jwtService.GenerateAccessToken(editor.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	memberToken, _, err := jwtService.GenerateAccessToken(member.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	revisionHandler := NewRevisionHandler(zoneRepo, mapRepo, testLogger)
	zoneHandler := NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, nil, versionRepo, testLogger)

	return &revisionTestFixtures{
		db:              db,
		revisionHandler: revisionHandler,
		zoneHandler:     zoneHandler,
		userRepo:        userRepo,
		zoneRepo:        zoneRepo,
		mapRepo:         mapRepo,
		teamRepo:        teamRepo,
		teamMemberRepo:  teamMemberRepo,
		lockRepo:        lockRepo,
		captainA:        captainA,
		captainB:        captainB,
		editor:          editor,
		member:          member,
		teamA:           teamA,
		teamB:           teamB,
		mapObj:          m,
		captainAToken:   captainAToken,
		captainBToken:   captainBToken,
		editorToken:     editorToken,
		memberToken:     memberToken,
		jwtService:      jwtService,
	}
}

func executeRevisionRequest(t *testing.T, handler gin.HandlerFunc, method, path string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	c.Set("user", user)
	c.Params = params

	handler(c)

	return w
}

func createRevisionZone(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, points []models.Point) int {
	t.Helper()
	zone, err := zoneRepo.Create(teamID, mapID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	if _, err := zoneRepo.AddPolygon(zone.ID, points, 0); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}
	if name != "" {
		if err := zoneRepo.SetName(zone.ID, name); err != nil {
			t.Fatalf("failed to set name: %v", err)
		}
	}
	return zone.ID
}

func TestListRevisionZones_OnlyTeamZones(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Zone Alpha", pts)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, f.mapObj.ID),
		f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone `json:"zones"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(resp.Zones))
	}

	z := resp.Zones[0]
	if z.Source != "team" {
		t.Errorf("expected source 'team', got '%s'", z.Source)
	}
	if z.Polygons[0].ZoneID == -1 {
		t.Error("expected real zone_id, not template placeholder -1")
	}
	if z.Name != "Zone Alpha" {
		t.Errorf("expected name 'Zone Alpha', got '%s'", z.Name)
	}
	if z.MapID != f.mapObj.ID {
		t.Errorf("expected map_id %d, got %d", f.mapObj.ID, z.MapID)
	}
	if len(z.Polygons) != 1 {
		t.Fatalf("expected 1 polygon, got %d", len(z.Polygons))
	}
	if len(z.Polygons[0].Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(z.Polygons[0].Points))
	}
}

func TestListRevisionZones_NoZones(t *testing.T) {
	f := setupRevisionTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, f.mapObj.ID),
		f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone `json:"zones"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Zones) != 0 {
		t.Errorf("expected 0 zones, got %d", len(resp.Zones))
	}
}

func TestListRevisionZones_Unauthenticated(t *testing.T) {
	f := setupRevisionTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, f.mapObj.ID), nil)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	f.revisionHandler.ListRevisionZones(c)

	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for unauthenticated request, got %d", w.Code)
	}
}

func TestListRevisionZones_MultipleTeamZones(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Alpha", pts)
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Beta", pts)
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Gamma", pts)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, f.mapObj.ID),
		f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone `json:"zones"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Zones) != 3 {
		t.Fatalf("expected 3 zones, got %d", len(resp.Zones))
	}

	names := make(map[string]bool)
	for _, z := range resp.Zones {
		names[z.Name] = true
		if z.Source != "team" {
			t.Errorf("expected source 'team' for zone %s, got '%s'", z.Name, z.Source)
		}
	}
	for _, n := range []string{"Alpha", "Beta", "Gamma"} {
		if !names[n] {
			t.Errorf("missing zone: %s", n)
		}
	}
}

func TestListRevisionZones_InvalidMapID(t *testing.T) {
	f := setupRevisionTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: "999999"},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, 999999),
		f.member, f.memberToken, params)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent map, got %d. Body: %s", w.Code, w.Body.String())
	}
}
