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

type lofTestFixtures struct {
	db             *sql.DB
	handler        *LineOfFireHandler
	userRepo       models.UserRepository
	mapRepo        models.MapRepository
	teamRepo       models.TeamRepository
	teamMemberRepo models.TeamMemberRepository
	lockRepo       models.EditLockRepository
	lineOfFireRepo models.LineOfFireRepository
	jwtService     *services.JWTService
	team           *models.Team
	mapObj         *models.Map
	editor         *models.User
	editorToken    string
}

func setupLineOfFireHandlerTest(t *testing.T) *lofTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	lockRepo := models.NewEditLockRepository(db)
	lineOfFireRepo := models.NewLineOfFireRepository(db)
	lofHistoryRepo := models.NewLineOfFireEditHistoryRepository(db)
	versionRepo := models.NewLineOfFireVersionRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	logger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	captain, _, err := userRepo.FindOrCreateByDiscordID("lof-handler-captain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	editor, _, err := userRepo.FindOrCreateByDiscordID("lof-handler-editor-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	team, err := teamRepo.Create("LOF Handler Team "+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, editor.ID, models.RoleEditor); err != nil {
		t.Fatalf("failed to add editor: %v", err)
	}

	m, err := mapRepo.Create("lof-handler-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	editorToken, _, err := jwtService.GenerateAccessToken(editor.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	handler := NewLineOfFireHandler(lineOfFireRepo, mapRepo, lockRepo, lofHistoryRepo, teamMemberRepo, userRepo, nil, versionRepo, logger)

	return &lofTestFixtures{
		db:             db,
		handler:        handler,
		userRepo:       userRepo,
		mapRepo:        mapRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		lockRepo:       lockRepo,
		lineOfFireRepo: lineOfFireRepo,
		jwtService:     jwtService,
		team:           team,
		mapObj:         m,
		editor:         editor,
		editorToken:    editorToken,
	}
}

func lofRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body []byte, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	c.Set("user", user)
	c.Params = params

	handler(c)
	return w
}

func acquireLock(t *testing.T, lockRepo models.EditLockRepository, teamID, mapID, userID int) {
	t.Helper()
	if _, err := lockRepo.Acquire(teamID, mapID, userID); err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
}

func TestListLinesOfFire(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	_, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Line A", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := lofRequest(t, f.handler.ListLinesOfFire, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lines-of-fire", f.team.ID, f.mapObj.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Lines []*models.LineOfFire `json:"lines_of_fire"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Lines) == 0 {
		t.Error("expected at least 1 line of fire")
	}
}

func TestCreateLineOfFire(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "New Line",
		"polygon_data": map[string]interface{}{
			"cones": []map[string]interface{}{
				{
					"origin":    map[string]float64{"x": 1, "y": 2},
					"left_end":  map[string]float64{"x": 3, "y": 4},
					"right_end": map[string]float64{"x": 5, "y": 6},
				},
			},
		},
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := lofRequest(t, f.handler.CreateLineOfFire, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lines-of-fire", f.team.ID, f.mapObj.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestCreateLineOfFire_NoLock(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name":         "No Lock Line",
		"polygon_data": map[string]interface{}{"cones": []interface{}{}},
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := lofRequest(t, f.handler.CreateLineOfFire, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lines-of-fire", f.team.ID, f.mapObj.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusLocked {
		t.Fatalf("expected 423, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestCreateLineOfFire_ObsoleteMap(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	replacement, err := f.mapRepo.Create("lof-obsolete-replacement-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Create replacement failed: %v", err)
	}
	f.mapRepo.MarkObsolete(f.mapObj.ID, replacement.ID)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Obsolete Line",
		"polygon_data": map[string]interface{}{"cones": []map[string]interface{}{
			{"origin": map[string]float64{"x": 1, "y": 2}, "left_end": map[string]float64{"x": 3, "y": 4}, "right_end": map[string]float64{"x": 5, "y": 6}},
		}},
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := lofRequest(t, f.handler.CreateLineOfFire, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lines-of-fire", f.team.ID, f.mapObj.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestGetLineOfFire(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)

	cone := models.Cone{Origin: models.Point{X: 100, Y: 200}, LeftEnd: models.Point{X: 300, Y: 400}, RightEnd: models.Point{X: 500, Y: 600}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Get Me", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
	}
	w := lofRequest(t, f.handler.GetLineOfFire, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d", f.team.ID, l.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Line models.LineOfFire `json:"line_of_fire"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Line.Name != "Get Me" {
		t.Errorf("expected name 'Get Me', got '%s'", resp.Line.Name)
	}
}

func TestGetLineOfFire_NotFound(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: "99999"},
	}
	w := lofRequest(t, f.handler.GetLineOfFire, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/99999", f.team.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestDeleteLineOfFire(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 1, Y: 2}, LeftEnd: models.Point{X: 3, Y: 4}, RightEnd: models.Point{X: 5, Y: 6}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Delete Me", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
	}
	w := lofRequest(t, f.handler.DeleteLineOfFire, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d", f.team.ID, l.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestSetLineOfFireName(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Old", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"name": "Renamed"})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
	}
	w := lofRequest(t, f.handler.SetLineOfFireName, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/name", f.team.ID, l.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	reloaded, _ := f.lineOfFireRepo.FindByID(l.ID)
	if reloaded.Name != "Renamed" {
		t.Errorf("expected name 'Renamed', got '%s'", reloaded.Name)
	}
}

func TestAddLineOfFireCone(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Cone Add", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"origin":    map[string]float64{"x": 100, "y": 200},
		"left_end":  map[string]float64{"x": 300, "y": 400},
		"right_end": map[string]float64{"x": 500, "y": 600},
	})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
	}
	w := lofRequest(t, f.handler.AddLineOfFireCone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/cones", f.team.ID, l.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	reloaded, _ := f.lineOfFireRepo.FindByID(l.ID)
	if len(reloaded.PolygonData.Cones) != 2 {
		t.Errorf("expected 2 cones, got %d", len(reloaded.PolygonData.Cones))
	}
}

func TestUpdateLineOfFireCone(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Cone Update", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"origin":    map[string]float64{"x": 999, "y": 888},
		"left_end":  map[string]float64{"x": 777, "y": 666},
		"right_end": map[string]float64{"x": 555, "y": 444},
	})
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
		{Key: "coneIndex", Value: "0"},
	}
	w := lofRequest(t, f.handler.UpdateLineOfFireCone, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/cones/0", f.team.ID, l.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	reloaded, _ := f.lineOfFireRepo.FindByID(l.ID)
	if reloaded.PolygonData.Cones[0].Origin.X != 999 {
		t.Errorf("expected origin X 999, got %f", reloaded.PolygonData.Cones[0].Origin.X)
	}
}

func TestRemoveLineOfFireCone(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Cone Remove", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"origin":    map[string]float64{"x": 100, "y": 200},
		"left_end":  map[string]float64{"x": 300, "y": 400},
		"right_end": map[string]float64{"x": 500, "y": 600},
	})
	paramsAdd := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
	}
	lofRequest(t, f.handler.AddLineOfFireCone, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/cones", f.team.ID, l.ID),
		body, f.editor, f.editorToken, paramsAdd)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
		{Key: "coneIndex", Value: "1"},
	}
	w := lofRequest(t, f.handler.RemoveLineOfFireCone, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/cones/1", f.team.ID, l.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRemoveLastCone_DeletesLine(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	l, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Last Cone", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "lineOfFireId", Value: fmt.Sprintf("%d", l.ID)},
		{Key: "coneIndex", Value: "0"},
	}
	w := lofRequest(t, f.handler.RemoveLineOfFireCone, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/lines-of-fire/%d/cones/0", f.team.ID, l.ID),
		nil, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	_, err = f.lineOfFireRepo.FindByID(l.ID)
	if err != models.ErrLineOfFireNotFound {
		t.Errorf("expected line to be deleted, found still exists: %v", err)
	}
}

func TestLineOfFireBatch_CreateLine(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	acquireLock(t, f.lockRepo, f.team.ID, f.mapObj.ID, f.editor.ID)

	body, _ := json.Marshal(map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"type":    "create_line_of_fire",
				"temp_id": "tmp1",
				"name":    "Batch Line",
				"polygon_data": map[string]interface{}{
					"cones": []map[string]interface{}{
						{"origin": map[string]float64{"x": 1, "y": 2}, "left_end": map[string]float64{"x": 3, "y": 4}, "right_end": map[string]float64{"x": 5, "y": 6}},
					},
				},
			},
		},
	})

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}
	w := lofRequest(t, f.handler.BatchActions, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/lines-of-fire/batch", f.team.ID, f.mapObj.ID),
		body, f.editor, f.editorToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRevisionItems_FilterLines(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	revisionHandler := NewRevisionHandler(nil, f.mapRepo, f.lineOfFireRepo, models.NewTemplateRepository(f.db), services.NewLogger(services.LogLevelError, services.LogFormatText))

	cone := models.Cone{Origin: models.Point{X: 10, Y: 20}, LeftEnd: models.Point{X: 30, Y: 40}, RightEnd: models.Point{X: 50, Y: 60}}
	_, err := f.lineOfFireRepo.Create(f.team.ID, f.mapObj.ID, "Revision Line", models.ConeData{Cones: []models.Cone{cone}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	body, _ := json.Marshal(map[string]bool{"zones": false, "lines": true})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.team.ID, f.mapObj.ID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.editorToken)
	c.Request = req
	c.Set("user", f.editor)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	revisionHandler.ListRevisionItems(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Lines []RevisionLineOfFire `json:"lines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Lines) == 0 {
		t.Error("expected at least 1 line of fire in revision")
	}
}

func TestRevisionItems_FilterBoth(t *testing.T) {
	f := setupLineOfFireHandlerTest(t)
	zoneRepo := models.NewZoneRepository(f.db)
	logger := services.NewLogger(services.LogLevelError, services.LogFormatText)
	revisionHandler := NewRevisionHandler(zoneRepo, f.mapRepo, f.lineOfFireRepo, models.NewTemplateRepository(f.db), logger)

	body, _ := json.Marshal(map[string]bool{"zones": true, "lines": true})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/maps/%d/revision-items", f.team.ID, f.mapObj.ID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.editorToken)
	c.Request = req
	c.Set("user", f.editor)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "mapId", Value: fmt.Sprintf("%d", f.mapObj.ID)},
	}

	revisionHandler.ListRevisionItems(c)

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
	if resp.Zones == nil {
		t.Error("expected zones key in response")
	}
	if resp.Lines == nil {
		t.Error("expected lines key in response")
	}
}
