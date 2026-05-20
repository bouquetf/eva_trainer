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

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTemplateTestDB(t *testing.T) *sql.DB {
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
		"000017_add_polygon_floor.up.sql",
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

type templateTestFixtures struct {
	db              *sql.DB
	templateHandler *TemplateHandler
	zoneHandler     *ZoneHandler
	userRepo        models.UserRepository
	templateRepo    models.TemplateRepository
	zoneRepo        models.ZoneRepository
	mapRepo         models.MapRepository
	teamRepo        models.TeamRepository
	teamMemberRepo  models.TeamMemberRepository
	lockRepo        models.EditLockRepository
	historyRepo     models.EditHistoryRepository
	captain         *models.User
	editor          *models.User
	member          *models.User
	team            *models.Team
	mapObj          *models.Map
	captainToken    string
	editorToken     string
	memberToken     string
}

func setupTemplateTest(t *testing.T) *templateTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupTemplateTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	historyRepo := models.NewEditHistoryRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	versionRepo := models.NewZoneVersionRepository(db)
	templateRepo := models.NewTemplateRepository(db)

	captain, _, err := userRepo.FindOrCreateByDiscordID("template-test-captain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	editor, _, err := userRepo.FindOrCreateByDiscordID("template-test-editor-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("template-test-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("Template Test Team "+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, editor.ID, models.RoleEditor); err != nil {
		t.Fatalf("failed to add editor: %v", err)
	}
	if err := teamMemberRepo.AddMember(team.ID, member.ID, models.RoleMember); err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	m, err := mapRepo.Create("template-test-map-"+t.Name(), nil)
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

	templateHandler := NewTemplateHandler(templateRepo, zoneRepo, teamMemberRepo)
	testLogger := services.NewLogger(services.LogLevelError, services.LogFormatText)
	zoneHandler := NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, nil, versionRepo, testLogger)

	return &templateTestFixtures{
		db:              db,
		templateHandler: templateHandler,
		zoneHandler:     zoneHandler,
		userRepo:        userRepo,
		templateRepo:    templateRepo,
		zoneRepo:        zoneRepo,
		mapRepo:         mapRepo,
		teamRepo:        teamRepo,
		teamMemberRepo:  teamMemberRepo,
		lockRepo:        lockRepo,
		historyRepo:     historyRepo,
		captain:         captain,
		editor:          editor,
		member:          member,
		team:            team,
		mapObj:          m,
		captainToken:    captainToken,
		editorToken:     editorToken,
		memberToken:     memberToken,
	}
}

func executeTemplateRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
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

func createZoneWithPolygon(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, points []models.Point) int {
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

func TestTemplateHandler_PublishMap_Success(t *testing.T) {
	f := setupTemplateTest(t)

	pts1 := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	pts2 := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 150, Y: 200}}
	createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone A", pts1)
	createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone B", pts2)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	tmplObj, ok := resp["template"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected template object, got %T", resp["template"])
	}

	if tmplObj["map_id"] == nil || tmplObj["map_id"].(float64) != float64(f.mapObj.ID) {
		t.Errorf("expected map_id %d, got %v", f.mapObj.ID, tmplObj["map_id"])
	}
	if tmplObj["is_public"] != true {
		t.Errorf("expected is_public true, got %v", tmplObj["is_public"])
	}

	templateID := tmplObj["id"].(float64)

	tmpl, err := f.templateRepo.FindByMapID(f.mapObj.ID)
	if err != nil {
		t.Fatalf("FindByMapID failed: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected template to exist")
	}
	if tmpl.ID != int(templateID) {
		t.Errorf("expected template id %d, got %d", int(templateID), tmpl.ID)
	}
}

func TestTemplateHandler_PublishMap_EmptyMap(t *testing.T) {
	f := setupTemplateTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	tmplObj, ok := resp["template"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected template object, got %T", resp["template"])
	}

	if tmplObj["is_public"] != true {
		t.Errorf("expected is_public true, got %v", tmplObj["is_public"])
	}
}

func TestTemplateHandler_PublishMap_AlreadyPublished(t *testing.T) {
	f := setupTemplateTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone", pts)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	w := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)
	if w.Code != http.StatusCreated {
		t.Fatalf("first publish: expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	w2 := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "already_published" {
		t.Errorf("expected error 'already_published', got %v", resp["error"])
	}
}

func TestTemplateHandler_PublishMap_NonCaptain(t *testing.T) {
	f := setupTemplateTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone", pts)

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	authMw := middleware.RequireAuth(nil, f.userRepo, jwtService)
	roleMw := middleware.RequireTeamRole(nil, f.teamMemberRepo, "captain", "cocaptain")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID), nil)
	req.Header.Set("Authorization", "Bearer "+f.editorToken)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		roleMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		f.templateHandler.PublishMap(c)
	}

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTemplateHandler_GetPublishStatus_Published(t *testing.T) {
	f := setupTemplateTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone", pts)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	pubResp := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)
	if pubResp.Code != http.StatusCreated {
		t.Fatalf("publish failed: %d", pubResp.Code)
	}

	w := executeTemplateRequest(t, f.templateHandler.GetPublishStatus, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish-status", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["published"] != true {
		t.Errorf("expected published true, got %v", resp["published"])
	}
	if resp["template"] == nil {
		t.Error("expected template to be present")
	}
}

func TestTemplateHandler_GetPublishStatus_NotPublished(t *testing.T) {
	f := setupTemplateTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	w := executeTemplateRequest(t, f.templateHandler.GetPublishStatus, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish-status", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["published"] != false {
		t.Errorf("expected published false, got %v", resp["published"])
	}
}

func TestTemplateHandler_PublishMap_IncludesFloorInPolygonData(t *testing.T) {
	f := setupTemplateTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	zoneID := createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Floor Zone", pts)
	polygons, err := f.zoneRepo.GetPolygons(zoneID)
	if err != nil || len(polygons) == 0 {
		t.Fatalf("failed to get polygons: %v", err)
	}
	if err := f.zoneRepo.SetPolygonFloor(polygons[0].ID, ptr("haut")); err != nil {
		t.Fatalf("failed to set floor: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var polygonDataStr string
	err = f.db.QueryRow(
		`SELECT polygon_data::text FROM template_zones WHERE template_id = (SELECT id FROM templates WHERE map_id = $1) ORDER BY order_index LIMIT 1`,
		f.mapObj.ID,
	).Scan(&polygonDataStr)
	if err != nil {
		t.Fatalf("failed to query polygon_data: %v", err)
	}
	if !strings.Contains(polygonDataStr, `"floor"`) {
		t.Errorf("expected polygon_data to contain 'floor' field, got: %s", polygonDataStr)
	}
	if !strings.Contains(polygonDataStr, `"haut"`) {
		t.Errorf("expected polygon_data to contain 'haut', got: %s", polygonDataStr)
	}
}

func TestTemplateHandler_DuplicateTemplate_PreservesFloor(t *testing.T) {
	f := setupTemplateTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	zoneID := createZoneWithPolygon(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Source Zone", pts)
	polygons, err := f.zoneRepo.GetPolygons(zoneID)
	if err != nil || len(polygons) == 0 {
		t.Fatalf("failed to get polygons: %v", err)
	}
	if err := f.zoneRepo.SetPolygonFloor(polygons[0].ID, ptr("bas")); err != nil {
		t.Fatalf("failed to set floor: %v", err)
	}

	publishParams := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	pubResp := executeTemplateRequest(t, f.templateHandler.PublishMap, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/publish", f.team.ID, f.mapObj.ID),
		"", f.captain, f.captainToken, publishParams)
	if pubResp.Code != http.StatusCreated {
		t.Fatalf("publish failed: %d. Body: %s", pubResp.Code, pubResp.Body.String())
	}

	var pubRespMap map[string]interface{}
	if err := json.Unmarshal(pubResp.Body.Bytes(), &pubRespMap); err != nil {
		t.Fatalf("failed to parse publish response: %v", err)
	}
	tmplObj := pubRespMap["template"].(map[string]interface{})
	templateID := int(tmplObj["id"].(float64))

	targetTeam, err := f.teamRepo.Create("Target Team "+t.Name(), f.captain.ID)
	if err != nil {
		t.Fatalf("failed to create target team: %v", err)
	}
	if err := f.teamMemberRepo.AddMember(targetTeam.ID, f.captain.ID, models.RoleCaptain); err != nil {
		t.Fatalf("failed to add captain to target team: %v", err)
	}

	dupParams := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", targetTeam.ID)},
		{Key: "templateId", Value: fmt.Sprintf("%d", templateID)},
	}
	dupResp := executeTemplateRequest(t, f.templateHandler.DuplicateTemplate, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/templates/%d/duplicate", targetTeam.ID, templateID),
		"", f.captain, f.captainToken, dupParams)
	if dupResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", dupResp.Code, dupResp.Body.String())
	}

	var dupRespMap map[string]interface{}
	if err := json.Unmarshal(dupResp.Body.Bytes(), &dupRespMap); err != nil {
		t.Fatalf("failed to parse duplicate response: %v", err)
	}
	zonesArr, ok := dupRespMap["zones"].([]interface{})
	if !ok || len(zonesArr) == 0 {
		t.Fatalf("expected zones array, got %T", dupRespMap["zones"])
	}
	firstZone := zonesArr[0].(map[string]interface{})
	polysArr, ok := firstZone["polygons"].([]interface{})
	if !ok || len(polysArr) == 0 {
		t.Fatalf("expected polygons array, got %T", firstZone["polygons"])
	}
	firstPoly := polysArr[0].(map[string]interface{})
	floorVal, hasFloor := firstPoly["floor"]
	if !hasFloor {
		t.Fatal("expected polygon to have 'floor' field after duplication")
	}
	if floorVal != "bas" {
		t.Errorf("expected floor 'bas', got %v", floorVal)
	}
}
