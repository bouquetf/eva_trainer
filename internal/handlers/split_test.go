//go:build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"eva_trainer/db/seed"
	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupSplitHandlerTest(t *testing.T) (*SplitHandler, *MapHandler, models.UserRepository, *services.JWTService, *sql.DB, models.PermissionRepository, models.RolePermissionRepository) {
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
		"000005_seed_permissions.up.sql",
		"000006_drop_current_version_id.up.sql",
		"000008_add_split_permission.up.sql",
		"000010_drop_split_dates.up.sql",
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

	if err := seed.SeedRolePermissions(db); err != nil {
		t.Fatalf("failed to seed role permissions: %v", err)
	}

	splitRepo := models.NewSplitRepository(db)
	mapRepo := models.NewMapRepository(db)
	userRepo := models.NewUserRepository(db)
	permRepo := models.NewPermissionRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	splitHandler := NewSplitHandler(splitRepo)
	mapHandler := NewMapHandler(mapRepo)

	return splitHandler, mapHandler, userRepo, jwtService, db, permRepo, rolePermRepo
}

func createSplitForTest(t *testing.T, handler *SplitHandler, token string) int {
	t.Helper()

	body := `{"name":"Test Split","map_names":["TestMap1"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateSplit(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("createSplitForTest: expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("createSplitForTest: failed to parse response: %v", err)
	}
	split := resp["split"].(map[string]interface{})
	return int(split["id"].(float64))
}

func uploadMapForSplitTest(t *testing.T, mapHandler *MapHandler, name string, token string) int {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("name", name); err != nil {
		t.Fatalf("failed to write name field: %v", err)
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s.webp"`, name))
	h.Set("Content-Type", "image/webp")
	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte("fake-webp-data")); err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/maps", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	mapHandler.UploadMap(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("uploadMapForSplitTest: expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	mapData := resp["map"].(map[string]interface{})
	return int(mapData["id"].(float64))
}

// --- CreateSplit ---

func TestSplitHandler_CreateSplit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	body := `{"name":"Season 1","map_names":["Atlantis","Polaris"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	splitHandler.CreateSplit(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	split, ok := resp["split"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'split' key in response")
	}
	if split["name"] != "Season 1" {
		t.Errorf("expected name 'Season 1', got '%v'", split["name"])
	}
	if split["id"].(float64) <= 0 {
		t.Errorf("expected valid id, got %v", split["id"])
	}
}

func TestSplitHandler_CreateSplit_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		body      string
		wantCode  int
		wantError string
	}{
		{
			name:      "empty name",
			body:      `{"name":"","map_names":["Atlantis"]}`,
			wantCode:  http.StatusBadRequest,
			wantError: "invalid_request_body",
		},
		{
			name:      "missing name",
			body:      `{"map_names":["Atlantis"]}`,
			wantCode:  http.StatusBadRequest,
			wantError: "invalid_request_body",
		},
		{
			name:      "empty map_names",
			body:      `{"name":"Season 1","map_names":[]}`,
			wantCode:  http.StatusBadRequest,
			wantError: "split_map_names_required",
		},
		{
			name:      "missing map_names",
			body:      `{"name":"Season 1"}`,
			wantCode:  http.StatusBadRequest,
			wantError: "invalid_request_body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
			_, token := createAdminUser(t, userRepo, jwtService)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			splitHandler.CreateSplit(c)

			if w.Code != tt.wantCode {
				t.Errorf("expected %d, got %d. Body: %s", tt.wantCode, w.Code, w.Body.String())
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != tt.wantError {
				t.Errorf("expected error '%s', got '%v'", tt.wantError, resp["error"])
			}
		})
	}
}

func TestSplitHandler_CreateSplit_ForbiddenForNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createNonAdminUser(t, userRepo, jwtService)

	body := `{"name":"Season 1","map_names":["Atlantis"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	authMw := middleware.RequireAuth(userRepo, jwtService)
	adminMw := middleware.RequireAdmin()

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		adminMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		splitHandler.CreateSplit(c)
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%v'", resp["error"])
	}
}

func TestSplitHandler_CreateSplit_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)

	body := `{"name":"Season 1","map_names":["Atlantis"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	authMw := middleware.RequireAuth(userRepo, jwtService)
	authMw(c)
	if !c.IsAborted() {
		c.Next()
		splitHandler.CreateSplit(c)
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "missing_authorization" {
		t.Errorf("expected error 'missing_authorization', got '%v'", resp["error"])
	}
}

// --- ListSplits (admin) ---

func TestSplitHandler_ListSplits_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	createSplitForTest(t, splitHandler, token)
	createSplitForTest(t, splitHandler, token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/splits", nil)

	splitHandler.ListSplits(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	splits, ok := resp["splits"].([]interface{})
	if !ok {
		t.Fatal("expected 'splits' array")
	}
	if len(splits) < 2 {
		t.Errorf("expected at least 2 splits, got %d", len(splits))
	}

	for _, s := range splits {
		entry := s.(map[string]interface{})
		if entry["id"].(float64) <= 0 {
			t.Errorf("expected valid id, got %v", entry["id"])
		}
	}
}

func TestSplitHandler_ListSplits_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/splits", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	splitHandler.ListSplits(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	splits, ok := resp["splits"].([]interface{})
	if !ok {
		t.Fatal("expected 'splits' array")
	}
	if len(splits) != 0 {
		t.Errorf("expected empty splits, got %d", len(splits))
	}
}

// --- GetSplit (admin) ---

func TestSplitHandler_GetSplit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	splitID := createSplitForTest(t, splitHandler, token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.GetSplit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	split, ok := resp["split"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'split' key")
	}
	if split["id"].(float64) != float64(splitID) {
		t.Errorf("expected id %d, got %v", splitID, split["id"])
	}
	if split["name"] != "Test Split" {
		t.Errorf("expected name 'Test Split', got '%v'", split["name"])
	}
}

func TestSplitHandler_GetSplit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/splits/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}
	c.Request.Header.Set("Authorization", "Bearer "+token)

	splitHandler.GetSplit(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "split_not_found" {
		t.Errorf("expected error 'split_not_found', got '%v'", resp["error"])
	}
}

// --- UpdateSplit (admin) ---

func TestSplitHandler_UpdateSplit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	splitID := createSplitForTest(t, splitHandler, token)

	body := `{"name":"Updated Split","map_names":["NewMap"]}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.UpdateSplit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	split, ok := resp["split"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'split' key")
	}
	if split["name"] != "Updated Split" {
		t.Errorf("expected name 'Updated Split', got '%v'", split["name"])
	}
}

func TestSplitHandler_UpdateSplit_InvalidData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	splitID := createSplitForTest(t, splitHandler, token)

	body := `{"name":"","map_names":["NewMap"]}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.UpdateSplit(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestSplitHandler_UpdateSplit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	body := `{"name":"Ghost","map_names":["Nowhere"]}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/splits/99999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	splitHandler.UpdateSplit(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "split_not_found" {
		t.Errorf("expected error 'split_not_found', got '%v'", resp["error"])
	}
}

// --- DeleteSplit (admin) ---

func TestSplitHandler_DeleteSplit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	splitID := createSplitForTest(t, splitHandler, token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.DeleteSplit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "split deleted" {
		t.Errorf("expected 'split deleted', got '%v'", resp["message"])
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), nil)
	c2.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.GetSplit(c2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestSplitHandler_DeleteSplit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/splits/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}
	c.Request.Header.Set("Authorization", "Bearer "+token)

	splitHandler.DeleteSplit(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "split_not_found" {
		t.Errorf("expected error 'split_not_found', got '%v'", resp["error"])
	}
}

// --- GetSplitMaps (admin) ---

func TestSplitHandler_GetSplitMaps_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, mapHandler, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	mapID1 := uploadMapForSplitTest(t, mapHandler, "AlphaMap", token)
	mapID2 := uploadMapForSplitTest(t, mapHandler, "BetaMap", token)

	body := `{"name":"Maps Split","map_names":["AlphaMap","BetaMap"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	splitHandler.CreateSplit(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("create split expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	splitID := int(createResp["split"].(map[string]interface{})["id"].(float64))

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/splits/%d/maps", splitID), nil)
	c2.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.GetSplitMaps(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)

	mapsList, ok := resp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array")
	}
	if len(mapsList) != 2 {
		t.Fatalf("expected 2 maps, got %d", len(mapsList))
	}

	foundIDs := make(map[int]bool)
	for _, m := range mapsList {
		entry := m.(map[string]interface{})
		foundIDs[int(entry["id"].(float64))] = true
	}
	if !foundIDs[mapID1] {
		t.Errorf("expected map id %d in results", mapID1)
	}
	if !foundIDs[mapID2] {
		t.Errorf("expected map id %d in results", mapID2)
	}
}

func TestSplitHandler_GetSplitMaps_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/splits/99999/maps", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}
	c.Request.Header.Set("Authorization", "Bearer "+token)

	splitHandler.GetSplitMaps(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "split_not_found" {
		t.Errorf("expected error 'split_not_found', got '%v'", resp["error"])
	}
}

// --- GetSplitMapsPublic ---

func TestSplitHandler_GetSplitMapsPublic_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, mapHandler, userRepo, jwtService, _, permRepo, rolePermRepo := setupSplitHandlerTest(t)

	_, adminToken := createAdminUser(t, userRepo, jwtService)
	splitID := createSplitForTest(t, splitHandler, adminToken)

	uploadMapForSplitTest(t, mapHandler, "TestMap1", adminToken)

	body := fmt.Sprintf(`{"name":"Test Split","map_names":["TestMap1"]}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/admin/splits/%d", splitID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	wUpd := httptest.NewRecorder()
	cUpd, _ := gin.CreateTestContext(wUpd)
	cUpd.Request = req
	cUpd.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.UpdateSplit(cUpd)

	if wUpd.Code != http.StatusOK {
		t.Fatalf("update split expected 200, got %d. Body: %s", wUpd.Code, wUpd.Body.String())
	}

	_, nonAdminToken := createNonAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/splits/%d/maps", splitID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}
	c.Request.Header.Set("Authorization", "Bearer "+nonAdminToken)

	authMw := middleware.RequireAuth(userRepo, jwtService)
	permMw := middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view")

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		permMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		splitHandler.GetSplitMapsPublic(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	mapsList, ok := resp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array")
	}
	if len(mapsList) != 1 {
		t.Errorf("expected 1 map, got %d", len(mapsList))
	}
}

func TestSplitHandler_GetSplitMapsPublic_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, permRepo, rolePermRepo := setupSplitHandlerTest(t)
	_, token := createNonAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/splits/99999/maps", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}
	c.Request.Header.Set("Authorization", "Bearer "+token)

	authMw := middleware.RequireAuth(userRepo, jwtService)
	permMw := middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view")

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		permMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		splitHandler.GetSplitMapsPublic(c)
	}

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "split_not_found" {
		t.Errorf("expected error 'split_not_found', got '%v'", resp["error"])
	}
}

// --- ListSplitsPublic ---

func TestSplitHandler_ListSplitsPublic_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, permRepo, rolePermRepo := setupSplitHandlerTest(t)
	_, adminToken := createAdminUser(t, userRepo, jwtService)

	createSplitForTest(t, splitHandler, adminToken)
	createSplitForTest(t, splitHandler, adminToken)

	_, nonAdminToken := createNonAdminUser(t, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/splits", nil)
	c.Request.Header.Set("Authorization", "Bearer "+nonAdminToken)

	authMw := middleware.RequireAuth(userRepo, jwtService)
	permMw := middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view")

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		permMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		splitHandler.ListSplitsPublic(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	splits, ok := resp["splits"].([]interface{})
	if !ok {
		t.Fatal("expected 'splits' array")
	}
	if len(splits) != 2 {
		t.Errorf("expected 2 splits, got %d", len(splits))
	}
}

func TestSplitHandler_ListSplitsPublic_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/splits", nil)

	authMw := middleware.RequireAuth(userRepo, jwtService)
	authMw(c)
	if !c.IsAborted() {
		c.Next()
		splitHandler.ListSplitsPublic(c)
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "missing_authorization" {
		t.Errorf("expected error 'missing_authorization', got '%v'", resp["error"])
	}
}

func TestSplitHandler_ListSplitsPublic_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, _, userRepo, jwtService, db, permRepo, rolePermRepo := setupSplitHandlerTest(t)
	_, token := createNonAdminUser(t, userRepo, jwtService)

	// Remove split:view permission from member role to trigger 403
	_, err := db.Exec(`DELETE FROM role_permissions WHERE role = 'member' AND permission_id = (SELECT id FROM permissions WHERE feature = 'split:view')`)
	if err != nil {
		t.Fatalf("failed to remove split:view from member: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/splits", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	authMw := middleware.RequireAuth(userRepo, jwtService)
	permMw := middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view")

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		permMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		splitHandler.ListSplitsPublic(c)
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "insufficient_permission" {
		t.Errorf("expected error 'insufficient_permission', got '%v'", resp["error"])
	}
}

// --- MapHandler ListMaps with name filter ---

func TestMapHandler_ListMaps_WithNameFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mapHandler, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	uploadMapForSplitTest(t, mapHandler, "Atlantis", token)
	uploadMapForSplitTest(t, mapHandler, "Polaris", token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/maps?name=Atlantis", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	mapHandler.ListMaps(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	mapsList, ok := resp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array")
	}
	if len(mapsList) != 1 {
		t.Fatalf("expected 1 map, got %d", len(mapsList))
	}

	entry := mapsList[0].(map[string]interface{})
	if entry["name"] != "Atlantis" {
		t.Errorf("expected name 'Atlantis', got '%v'", entry["name"])
	}
}

func TestMapHandler_ListMaps_WithNameFilterNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mapHandler, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	uploadMapForSplitTest(t, mapHandler, "Atlantis", token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/maps?name=Inexistant", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	mapHandler.ListMaps(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	mapsList, ok := resp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array")
	}
	if len(mapsList) != 0 {
		t.Errorf("expected 0 maps, got %d", len(mapsList))
	}
}

func TestMapHandler_ListMaps_WithoutNameFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mapHandler, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	uploadMapForSplitTest(t, mapHandler, "Atlantis", token)
	uploadMapForSplitTest(t, mapHandler, "Polaris", token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/maps", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	mapHandler.ListMaps(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	mapsList, ok := resp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array")
	}
	if len(mapsList) != 2 {
		t.Errorf("expected 2 maps, got %d", len(mapsList))
	}
}

// --- E2E: CreateSplit + ResolveMaps ---

func TestSplitHandler_E2E_CreateSplitAndResolveMaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	splitHandler, mapHandler, userRepo, jwtService, _, _, _ := setupSplitHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	// Step 1: Create 2 maps
	mapID1 := uploadMapForSplitTest(t, mapHandler, "Atlantis", token)
	mapID2 := uploadMapForSplitTest(t, mapHandler, "Polaris", token)

	// Step 2: Create split with those map names
	body := `{"name":"Season 1","map_names":["Atlantis","Polaris"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	splitHandler.CreateSplit(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("step 2: expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	splitID := int(createResp["split"].(map[string]interface{})["id"].(float64))

	// Step 3: Resolve maps for split
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/splits/%d/maps", splitID), nil)
	c3.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.GetSplitMaps(c3)

	if w3.Code != http.StatusOK {
		t.Fatalf("step 3: expected 200, got %d. Body: %s", w3.Code, w3.Body.String())
	}

	var resp3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &resp3)

	mapsList, ok := resp3["maps"].([]interface{})
	if !ok {
		t.Fatal("step 3: expected 'maps' array")
	}
	if len(mapsList) != 2 {
		t.Fatalf("step 3: expected 2 maps, got %d", len(mapsList))
	}

	foundIDs := map[int]bool{}
	for _, m := range mapsList {
		entry := m.(map[string]interface{})
		foundIDs[int(entry["id"].(float64))] = true
	}
	if !foundIDs[mapID1] {
		t.Errorf("step 3: expected map id %d in results", mapID1)
	}
	if !foundIDs[mapID2] {
		t.Errorf("step 3: expected map id %d in results", mapID2)
	}

	// Step 5: Create a 3rd map "Atlantis" (v2, makes v1 obsolete)
	mapID3 := uploadMapForSplitTest(t, mapHandler, "Atlantis", token)

	// Step 6: Resolve maps again
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/splits/%d/maps", splitID), nil)
	c6.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", splitID)}}

	splitHandler.GetSplitMaps(c6)

	if w6.Code != http.StatusOK {
		t.Fatalf("step 6: expected 200, got %d. Body: %s", w6.Code, w6.Body.String())
	}

	var resp6 map[string]interface{}
	json.Unmarshal(w6.Body.Bytes(), &resp6)

	mapsList2, ok := resp6["maps"].([]interface{})
	if !ok {
		t.Fatal("step 6: expected 'maps' array")
	}
	// Should return all 3 maps (including the obsolete v1 "Atlantis")
	if len(mapsList2) != 3 {
		t.Fatalf("step 7: expected 3 maps (including obsolete), got %d", len(mapsList2))
	}

	allIDs := map[int]bool{}
	for _, m := range mapsList2 {
		entry := m.(map[string]interface{})
		allIDs[int(entry["id"].(float64))] = true
	}
	if !allIDs[mapID1] {
		t.Errorf("step 7: expected obsolete map id %d in results", mapID1)
	}
	if !allIDs[mapID2] {
		t.Errorf("step 7: expected map id %d in results", mapID2)
	}
	if !allIDs[mapID3] {
		t.Errorf("step 7: expected map id %d in results", mapID3)
	}
}
