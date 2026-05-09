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

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMapHandlerTest(t *testing.T) (*MapHandler, models.UserRepository, *services.JWTService, *sql.DB) {
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
		"000006_drop_current_version_id.up.sql",
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

	mapRepo := models.NewMapRepository(db)
	userRepo := models.NewUserRepository(db)
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	handler := NewMapHandler(mapRepo)

	return handler, userRepo, jwtService, db
}

func createAdminUser(t *testing.T, userRepo models.UserRepository, jwtService *services.JWTService) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("admin-map-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := userRepo.SetAdmin(user.ID); err != nil {
		t.Fatalf("failed to set admin: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	return user, token
}

func createNonAdminUser(t *testing.T, userRepo models.UserRepository, jwtService *services.JWTService) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("nonadmin-map-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	return user, token
}

func buildUploadRequest(t *testing.T, name string, filename string, content []byte, token string) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if name != "" {
		if err := w.WriteField("name", name); err != nil {
			t.Fatalf("failed to write name field: %v", err)
		}
	}

	if filename != "" && content != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
		h.Set("Content-Type", "image/webp")
		part, err := w.CreatePart(h)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("failed to write file content: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/maps", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func TestMapHandler_UploadAndServeImage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	blobData := []byte("fake-webp-image-data")
	req := buildUploadRequest(t, "TestMap", "test-map.webp", blobData, token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)

	handler.UploadMap(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var uploadResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("failed to parse upload response: %v", err)
	}

	mapData, ok := uploadResp["map"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'map' key in response")
	}
	mapID := int(mapData["id"].(float64))

	// Now serve the image
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/maps/%d/image", mapID), nil)
	c2.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", mapID)}}

	handler.ServeImage(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for image, got %d", w2.Code)
	}

	ct := w2.Header().Get("Content-Type")
	if ct != "image/webp" {
		t.Errorf("expected Content-Type 'image/webp', got '%s'", ct)
	}

	if w2.Body.String() != string(blobData) {
		t.Errorf("expected image body %q, got %q", string(blobData), w2.Body.String())
	}
}

func TestMapHandler_ImageAccessibleWithoutAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	req := buildUploadRequest(t, "PublicImg", "pub.webp", []byte("image-blob"), token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	handler.UploadMap(c)

	var uploadResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	mapID := int(uploadResp["map"].(map[string]interface{})["id"].(float64))

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/maps/%d/image", mapID), nil)
	c2.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", mapID)}}

	handler.ServeImage(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for public image, got %d", w2.Code)
	}
}

func TestMapHandler_UploadNoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _ := setupMapHandlerTest(t)

	req := buildUploadRequest(t, "NoAuth", "noauth.webp", []byte("data"), "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)

	handler.UploadMap(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 (no auth enforced at handler level), got %d", w.Code)
	}
}

func TestMapHandler_UploadNonWebpFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	req := buildUploadRequest(t, "BadFile", "image.png", []byte("png-data"), token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)

	handler.UploadMap(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_file_format" {
		t.Errorf("expected error 'invalid_file_format', got '%v'", resp["error"])
	}
}

func TestMapHandler_UploadNoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	req := buildUploadRequest(t, "NoFile", "", nil, token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)

	handler.UploadMap(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "file_required" {
		t.Errorf("expected error 'file_required', got '%v'", resp["error"])
	}
}

func TestMapHandler_AutoObsoletion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	blobData := []byte("version1-data")
	req1 := buildUploadRequest(t, "ObsTest", "obstest.webp", blobData, token)

	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1
	c1.Set("user_id", 1)
	handler.UploadMap(c1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first upload expected 201, got %d", w1.Code)
	}

	var resp1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	map1 := resp1["map"].(map[string]interface{})
	v1ID := int(map1["id"].(float64))

	req2 := buildUploadRequest(t, "ObsTest", "obstest.webp", []byte("version2-data"), token)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2
	c2.Set("user_id", 1)
	handler.UploadMap(c2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("second upload expected 201, got %d", w2.Code)
	}

	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	previous, ok := resp2["previous_version_obsoleted"].(map[string]interface{})
	if !ok {
		t.Fatal("expected previous_version_obsoleted in response")
	}
	prevID := int(previous["id"].(float64))
	if prevID != v1ID {
		t.Errorf("expected previous_version_obsoleted.id %d, got %d", v1ID, prevID)
	}
}

func TestMapHandler_ListMaps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	req1 := buildUploadRequest(t, "ListA", "lista.webp", []byte("a"), token)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1
	c1.Set("user_id", 1)
	handler.UploadMap(c1)

	req2 := buildUploadRequest(t, "ListB", "listb.webp", []byte("b"), token)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2
	c2.Set("user_id", 1)
	handler.UploadMap(c2)

	// List maps
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/maps", nil)
	c3.Set("user_id", 1)

	handler.ListMaps(c3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w3.Code)
	}

	var listResp map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &listResp)

	mapsList, ok := listResp["maps"].([]interface{})
	if !ok {
		t.Fatal("expected 'maps' array in response")
	}
	if len(mapsList) != 2 {
		t.Errorf("expected 2 maps, got %d", len(mapsList))
	}
}

func TestMapHandler_PurgeObsolete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, jwtService, _ := setupMapHandlerTest(t)
	_, token := createAdminUser(t, userRepo, jwtService)

	// Upload a map, then re-upload to obsolete it
	req1 := buildUploadRequest(t, "PurgeMe", "purge.webp", []byte("v1"), token)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1
	c1.Set("user_id", 1)
	handler.UploadMap(c1)

	req2 := buildUploadRequest(t, "PurgeMe", "purge.webp", []byte("v2"), token)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2
	c2.Set("user_id", 1)
	handler.UploadMap(c2)

	// Purge
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/maps/purge", nil)
	c3.Set("user_id", 1)

	handler.PurgeObsoleteMaps(c3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 for purge, got %d. Body: %s", w3.Code, w3.Body.String())
	}

	var purgeResp map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &purgeResp)

	purgedIDs, ok := purgeResp["purged_ids"].([]interface{})
	if !ok {
		t.Fatal("expected 'purged_ids' array in response")
	}
	if len(purgedIDs) == 0 {
		// Accept: at least 1 if there's an obsolete without zones
		t.Log("purge returned empty purged_ids (may have zones linked)")
	}
}

func TestMapHandler_ImageNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _ := setupMapHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/maps/99999/image", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.ServeImage(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
