//go:build integration

package handlers

import (
	"bytes"
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
		"000007_add_team_invitations_and_join_requests.up.sql",
		"000011_add_zone_polygon_constraints.up.sql",
		"000013_add_zone_versions.up.sql",
		"000015_add_template_is_public.up.sql",
		"000017_add_polygon_floor.up.sql",
		"000020_create_lines_of_fire.up.sql",
		"000021_create_template_lines_of_fire.up.sql",
		"000022_create_line_of_fire_edit_history.up.sql",
		"000023_create_line_of_fire_versions.up.sql",
		"000024_add_line_of_fire_permissions.up.sql",
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
	lineOfFireRepo  models.LineOfFireRepository
	templateRepo    models.TemplateRepository
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
	logger          services.Logger
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
	lineOfFireRepo := models.NewLineOfFireRepository(db)
	templateRepo := models.NewTemplateRepository(db)
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

	revisionHandler := NewRevisionHandler(zoneRepo, mapRepo, lineOfFireRepo, templateRepo, testLogger)
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
		lineOfFireRepo:  lineOfFireRepo,
		templateRepo:    templateRepo,
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
		logger:          testLogger,
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

func executeRevisionPostRequest(t *testing.T, handler gin.HandlerFunc, path string, body []byte, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user", user)
	c.Params = params

	handler(c)

	return w
}

func createRevisionLineOfFire(t *testing.T, repo models.LineOfFireRepository, teamID, mapID int, name string, cone models.Cone) int {
	t.Helper()
	l, err := repo.Create(teamID, mapID, name, models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("failed to create line of fire: %v", err)
	}
	return l.ID
}

func seedRevisionTemplate(t *testing.T, db *sql.DB, mapID int) int {
	t.Helper()
	var templateID int
	err := db.QueryRow(
		`INSERT INTO templates (original_team_id, map_id) VALUES (NULL, $1) RETURNING id`,
		mapID,
	).Scan(&templateID)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	coneJSON := `{"cones":[{"origin":{"x":1,"y":2},"left_end":{"x":3,"y":4},"right_end":{"x":5,"y":6}}]}`
	var tlID int
	err = db.QueryRow(
		`INSERT INTO template_lines_of_fire (template_id, polygon_data, name, order_index) VALUES ($1, $2, 'Template LOF', 0) RETURNING id`,
		templateID, coneJSON,
	).Scan(&tlID)
	if err != nil {
		t.Fatalf("failed to create template line of fire: %v", err)
	}
	return tlID
}

func createRevisionZone(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, points []models.Point) int {
	t.Helper()
	zone, err := zoneRepo.Create(teamID, mapID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	if _, err := zoneRepo.AddPolygon(zone.ID, points, 0, nil); err != nil {
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

func TestListRevisionZones_ObsoleteMapZones(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}}

	v1, err := f.mapRepo.Create("obsolete-test-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create v1 map: %v", err)
	}

	createRevisionZone(t, f.zoneRepo, f.teamA.ID, v1.ID, "Zone V1", pts)

	v2, err := f.mapRepo.Create("obsolete-test-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create v2 map: %v", err)
	}

	createRevisionZone(t, f.zoneRepo, f.teamA.ID, v2.ID, "Zone V2", pts)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", v2.ID)},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, v2.ID),
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

	if len(resp.Zones) != 2 {
		t.Fatalf("expected 2 zones (v1 + v2), got %d", len(resp.Zones))
	}

	for _, z := range resp.Zones {
		if z.Name == "Zone V1" {
			if z.MapID != v1.ID {
				t.Errorf("v1 zone map_id: expected %d, got %d", v1.ID, z.MapID)
			}
		} else if z.Name == "Zone V2" {
			if z.MapID != v2.ID {
				t.Errorf("v2 zone map_id: expected %d, got %d", v2.ID, z.MapID)
			}
		} else {
			t.Errorf("unexpected zone name: %s", z.Name)
		}
	}
}

func TestListRevisionZones_DeletedObsoleteMapExcluded(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}}

	v1, err := f.mapRepo.Create("deleted-test-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create v1 map: %v", err)
	}

	createRevisionZone(t, f.zoneRepo, f.teamA.ID, v1.ID, "Zone V1", pts)

	v2, err := f.mapRepo.Create("deleted-test-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create v2 map: %v", err)
	}

	v2ZoneID := createRevisionZone(t, f.zoneRepo, f.teamA.ID, v2.ID, "Zone V2", pts)

	if err := f.mapRepo.Delete(v1.ID); err != nil {
		t.Fatalf("failed to delete v1 map: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", v2.ID)},
	}
	w := executeRevisionRequest(t, f.revisionHandler.ListRevisionZones, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-zones", f.teamA.ID, v2.ID),
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
		t.Fatalf("expected 1 zone (only v2), got %d", len(resp.Zones))
	}

	if resp.Zones[0].ID != v2ZoneID {
		t.Errorf("expected v2 zone, got zone id %d", resp.Zones[0].ID)
	}
	if resp.Zones[0].MapID != v2.ID {
		t.Errorf("expected map_id %d, got %d", v2.ID, resp.Zones[0].MapID)
	}
}

func TestListRevisionItems_ZonesOnly(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Zone Only", pts)

	cone := models.Cone{Origin: models.Point{X: 100, Y: 200}, LeftEnd: models.Point{X: 300, Y: 400}, RightEnd: models.Point{X: 500, Y: 600}}
	createRevisionLineOfFire(t, f.lineOfFireRepo, f.teamA.ID, f.mapObj.ID, "LOF Should Not Appear", cone)

	body, _ := json.Marshal(map[string]bool{"zones": true, "lines": false})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone       `json:"zones"`
		Lines []RevisionLineOfFire `json:"lines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(resp.Zones))
	}
	if resp.Zones[0].Name != "Zone Only" {
		t.Errorf("expected 'Zone Only', got '%s'", resp.Zones[0].Name)
	}
	if resp.Lines != nil {
		t.Errorf("expected lines to be nil/absent, got %d lines", len(resp.Lines))
	}
}

func TestListRevisionItems_LinesOnly(t *testing.T) {
	f := setupRevisionTest(t)

	cone := models.Cone{Origin: models.Point{X: 100, Y: 200}, LeftEnd: models.Point{X: 300, Y: 400}, RightEnd: models.Point{X: 500, Y: 600}}
	createRevisionLineOfFire(t, f.lineOfFireRepo, f.teamA.ID, f.mapObj.ID, "Line Alpha", cone)

	body, _ := json.Marshal(map[string]bool{"zones": false, "lines": true})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone       `json:"zones"`
		Lines []RevisionLineOfFire `json:"lines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(resp.Lines))
	}
	if resp.Lines[0].Name != "Line Alpha" {
		t.Errorf("expected 'Line Alpha', got '%s'", resp.Lines[0].Name)
	}
	if resp.Lines[0].Source != "team" {
		t.Errorf("expected source 'team', got '%s'", resp.Lines[0].Source)
	}
	if resp.Zones != nil {
		t.Errorf("expected zones to be nil/absent, got %d zones", len(resp.Zones))
	}
}

func TestListRevisionItems_Both(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Zone A", pts)

	cone := models.Cone{Origin: models.Point{X: 100, Y: 200}, LeftEnd: models.Point{X: 300, Y: 400}, RightEnd: models.Point{X: 500, Y: 600}}
	createRevisionLineOfFire(t, f.lineOfFireRepo, f.teamA.ID, f.mapObj.ID, "Line A", cone)

	body, _ := json.Marshal(map[string]bool{"zones": true, "lines": true})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Zones []RevisionZone       `json:"zones"`
		Lines []RevisionLineOfFire `json:"lines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Zones) != 1 {
		t.Errorf("expected 1 zone, got %d", len(resp.Zones))
	}
	if len(resp.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(resp.Lines))
	}
}

func TestListRevisionItems_BothFalse_Returns400(t *testing.T) {
	f := setupRevisionTest(t)

	body, _ := json.Marshal(map[string]bool{"zones": false, "lines": false})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "revision_filter_empty" {
		t.Errorf("expected error 'revision_filter_empty', got '%v'", resp["error"])
	}
}

func TestListRevisionItems_IncludesTemplateLines(t *testing.T) {
	f := setupRevisionTest(t)

	seedRevisionTemplate(t, f.db, f.mapObj.ID)

	body, _ := json.Marshal(map[string]bool{"zones": false, "lines": true})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Lines []RevisionLineOfFire `json:"lines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	var foundTemplate bool
	for _, l := range resp.Lines {
		if l.Source == "template" {
			foundTemplate = true
			if l.Name != "Template LOF" {
				t.Errorf("expected name 'Template LOF', got '%s'", l.Name)
			}
		}
	}
	if !foundTemplate {
		t.Error("expected at least one template-sourced line of fire, none found")
	}
}

func TestListRevisionItems_DefaultsToZonesOnly(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "Default Zone", pts)

	body, _ := json.Marshal(map[string]interface{}{})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.member, f.memberToken, params)

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
		t.Errorf("expected 1 zone (default), got %d", len(resp.Zones))
	}
}

func TestListRevisionItems_Unauthenticated(t *testing.T) {
	f := setupRevisionTest(t)

	body, _ := json.Marshal(map[string]bool{"zones": true, "lines": true})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	f.revisionHandler.ListRevisionItems(c)

	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for unauthenticated request, got %d", w.Code)
	}
}

func TestListRevisionItems_TeamScoped(t *testing.T) {
	f := setupRevisionTest(t)

	pts := []models.Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}}
	createRevisionZone(t, f.zoneRepo, f.teamA.ID, f.mapObj.ID, "TeamA Zone", pts)
	createRevisionZone(t, f.zoneRepo, f.teamB.ID, f.mapObj.ID, "TeamB Zone", pts)

	body, _ := json.Marshal(map[string]bool{"zones": true, "lines": false})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.teamA.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeRevisionPostRequest(t, f.revisionHandler.ListRevisionItems,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.teamA.ID, f.mapObj.ID),
		body, f.captainA, f.captainAToken, params)

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
		t.Fatalf("expected 1 zone (TeamA only), got %d", len(resp.Zones))
	}
	if resp.Zones[0].Name != "TeamA Zone" {
		t.Errorf("expected 'TeamA Zone', got '%s'", resp.Zones[0].Name)
	}
}
