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

func setupHistoryTestDB(t *testing.T) *sql.DB {
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
		"000016_add_edit_history.up.sql",
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

type historyTestFixtures struct {
	db             *sql.DB
	historyHandler *HistoryHandler
	userRepo       models.UserRepository
	zoneRepo       models.ZoneRepository
	mapRepo        models.MapRepository
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	lockRepo       models.EditLockRepository
	versionRepo    models.ZoneVersionRepository
	historyRepo    models.EditHistoryRepository
	captain        *models.User
	editor         *models.User
	member         *models.User
	outsider       *models.User
	team           *models.Team
	mapObj         *models.Map
	captainToken   string
	editorToken    string
	memberToken    string
	outsiderToken  string
}

func setupHistoryTest(t *testing.T) *historyTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupHistoryTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	versionRepo := models.NewZoneVersionRepository(db)
	testLogger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	captain, _, err := userRepo.FindOrCreateByDiscordID("history-captain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	editor, _, err := userRepo.FindOrCreateByDiscordID("history-editor-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("history-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	outsider, _, err := userRepo.FindOrCreateByDiscordID("history-outsider-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create outsider: %v", err)
	}

	team, err := teamRepo.Create("History Test Team "+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, editor.ID, models.RoleEditor); err != nil {
		t.Fatalf("failed to add editor: %v", err)
	}
	if err := teamMemberRepo.AddMember(team.ID, member.ID, models.RoleMember); err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	m, err := mapRepo.Create("history-test-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	captainToken, _, err := jwtService.GenerateAccessToken(captain.ID)
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
	outsiderToken, _, err := jwtService.GenerateAccessToken(outsider.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	historyHandler := NewHistoryHandler(versionRepo, zoneRepo, lockRepo, teamMemberRepo, userRepo, historyRepo, testLogger)

	return &historyTestFixtures{
		db:             db,
		historyHandler: historyHandler,
		userRepo:       userRepo,
		zoneRepo:       zoneRepo,
		mapRepo:        mapRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		lockRepo:       lockRepo,
		versionRepo:    versionRepo,
		historyRepo:    historyRepo,
		captain:        captain,
		editor:         editor,
		member:         member,
		outsider:       outsider,
		team:           team,
		mapObj:         m,
		captainToken:   captainToken,
		editorToken:    editorToken,
		memberToken:    memberToken,
		outsiderToken:  outsiderToken,
	}
}

func executeHistoryRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
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

func createTestZone(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, points []models.Point) *models.Zone {
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
	return zone
}

func saveVersionForTest(t *testing.T, versionRepo models.ZoneVersionRepository, zoneID int, editorID *int, snapshot models.VersionSnapshot) *models.ZoneVersion {
	t.Helper()
	v, err := versionRepo.Save(zoneID, editorID, snapshot)
	if err != nil {
		t.Fatalf("failed to save version: %v", err)
	}
	return v
}

func TestHistoryHandler_ListVersions_Success(t *testing.T) {
	f := setupHistoryTest(t)

	pts := []models.Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}}
	zone := createTestZone(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone Alpha", pts)

	saveVersionForTest(t, f.versionRepo, zone.ID, &f.editor.ID, models.VersionSnapshot{
		Polygons: []models.SnapshotPolygon{{Points: pts, OrderIndex: 0}},
		Name:     "Zone Alpha",
	})
	saveVersionForTest(t, f.versionRepo, zone.ID, &f.captain.ID, models.VersionSnapshot{
		Polygons: []models.SnapshotPolygon{{Points: pts, OrderIndex: 0}},
		Name:     "Zone Alpha v2",
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeHistoryRequest(t, f.historyHandler.ListVersions, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/versions", f.team.ID, f.mapObj.ID),
		"", f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp listVersionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Total != 2 {
		t.Fatalf("expected 2 versions, got %d", resp.Total)
	}
	if len(resp.Versions) != 2 {
		t.Fatalf("expected 2 version items, got %d", len(resp.Versions))
	}
}

func TestHistoryHandler_ListVersions_Unauthenticated(t *testing.T) {
	f := setupHistoryTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/versions", f.team.ID, f.mapObj.ID), nil)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	f.historyHandler.ListVersions(c)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Errorf("expected 401 or 403 for unauthenticated request, got %d", w.Code)
	}
}

func TestHistoryHandler_ListVersions_NonMember(t *testing.T) {
	f := setupHistoryTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeHistoryRequest(t, f.historyHandler.ListVersions, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/versions", f.team.ID, f.mapObj.ID),
		"", f.outsider, f.outsiderToken, params)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-member, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_ListZoneVersions_Success(t *testing.T) {
	f := setupHistoryTest(t)

	pts := []models.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	zone := createTestZone(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone List", pts)

	saveVersionForTest(t, f.versionRepo, zone.ID, &f.editor.ID, models.VersionSnapshot{
		Polygons: []models.SnapshotPolygon{{Points: pts, OrderIndex: 0}},
		Name:     "Zone List",
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
	}
	w := executeHistoryRequest(t, f.historyHandler.ListZoneVersions, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/versions", f.team.ID, zone.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp listVersionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected 1 version, got %d", resp.Total)
	}
}

func TestHistoryHandler_ListZoneVersions_ZoneNotFound(t *testing.T) {
	f := setupHistoryTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: "999999"},
	}
	w := executeHistoryRequest(t, f.historyHandler.ListZoneVersions, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/versions", f.team.ID, 999999),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_RestoreVersion_Success(t *testing.T) {
	f := setupHistoryTest(t)

	pts := []models.Point{{X: 100, Y: 200}, {X: 300, Y: 400}, {X: 500, Y: 600}}
	zone := createTestZone(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Restore Zone", pts)

	_, err := f.lockRepo.Acquire(f.team.ID, f.mapObj.ID, f.editor.ID)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	v := saveVersionForTest(t, f.versionRepo, zone.ID, &f.editor.ID, models.VersionSnapshot{
		Polygons: []models.SnapshotPolygon{{Points: pts, OrderIndex: 0}},
		Name:     "Restore Zone",
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
		{Key: "versionId", Value: fmt.Sprintf("%d", v.ID)},
	}
	w := executeHistoryRequest(t, f.historyHandler.RestoreVersion, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/versions/%d/restore", f.team.ID, zone.ID, v.ID),
		"", f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_RestoreVersion_NoLock(t *testing.T) {
	f := setupHistoryTest(t)

	pts := []models.Point{{X: 1, Y: 1}, {X: 2, Y: 2}}
	zone := createTestZone(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "No Lock Zone", pts)

	v := saveVersionForTest(t, f.versionRepo, zone.ID, &f.editor.ID, models.VersionSnapshot{
		Polygons: []models.SnapshotPolygon{{Points: pts, OrderIndex: 0}},
		Name:     "No Lock Zone",
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "zoneId", Value: fmt.Sprintf("%d", zone.ID)},
		{Key: "versionId", Value: fmt.Sprintf("%d", v.ID)},
	}
	w := executeHistoryRequest(t, f.historyHandler.RestoreVersion, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/zones/%d/versions/%d/restore", f.team.ID, zone.ID, v.ID),
		"", f.editor, f.editorToken, params)

	if w.Code != http.StatusLocked {
		t.Errorf("expected 423, got %d. Body: %s", w.Code, w.Body.String())
	}
}
