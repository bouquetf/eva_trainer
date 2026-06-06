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

	"eva_trainer/db/seed"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const leaguesMigration = "000026_add_archive_leagues.up.sql"

func setupArchiveHandlerTest(t *testing.T) (*ArchiveHandler, *services.JWTService, models.UserRepository, *sql.DB) {
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
		"000025_add_archive_roles.up.sql",
		leaguesMigration,
		"000027_fix_ligue_locale_table.up.sql",
		"000028_add_ligue_locale_emplacement_unique.up.sql",
		"000029_add_ligue_locale_junction.up.sql",
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

	ligueLocaleRepo := models.NewLigueLocaleRepository(db)
	ligueRepo := models.NewLigueRepository(db)
	splitRepoModels := models.NewSplitRepository(db)
	userRepo := models.NewUserRepository(db)
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	archiveHandler := NewArchiveHandler(ligueLocaleRepo, ligueRepo, splitRepoModels, services.NewLogger(services.LogLevelError, services.LogFormatText))

	return archiveHandler, jwtService, userRepo, db
}

// --- ListLiguesLocales ---

func TestArchiveHandler_ListLiguesLocales_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/leagues-locales", nil)

	handler.ListLiguesLocales(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locales, ok := resp["ligues_locales"].([]interface{})
	if !ok {
		t.Fatal("expected 'leagues_locales' array")
	}
	if len(locales) != 0 {
		t.Errorf("expected empty locales, got %d", len(locales))
	}
}

func TestArchiveHandler_ListLiguesLocales_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	localeRepo.Create("Grenoble", "Grenoble")
	localeRepo.Create("Lyon", "Lyon")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/leagues-locales", nil)

	handler.ListLiguesLocales(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locales, ok := resp["ligues_locales"].([]interface{})
	if !ok {
		t.Fatal("expected 'leagues_locales' array")
	}
	if len(locales) != 2 {
		t.Errorf("expected 2 locales, got %d", len(locales))
	}
}

func TestArchiveHandler_ListLiguesLocales_FilterByEmplacement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	localeRepo.Create("Grenoble", "Grenoble")
	localeRepo.Create("Lyon", "Lyon")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/leagues-locales?emplacement=Grenoble", nil)

	handler.ListLiguesLocales(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locales, ok := resp["ligues_locales"].([]interface{})
	if !ok {
		t.Fatal("expected 'leagues_locales' array")
	}
	if len(locales) != 1 {
		t.Errorf("expected 1 locale, got %d", len(locales))
	}
	entry := locales[0].(map[string]interface{})
	if entry["emplacement"] != "Grenoble" {
		t.Errorf("expected emplacement 'Grenoble', got '%v'", entry["emplacement"])
	}
}

// --- GetLigueLocale ---

func TestArchiveHandler_GetLigueLocale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	created, err := localeRepo.Create("Grenoble", "Grenoble")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/archive/leagues-locales/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.GetLigueLocale(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locale, ok := resp["ligue_locale"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'league_locale' key")
	}
	if locale["nom"] != "Grenoble" {
		t.Errorf("expected nom 'Grenoble', got '%v'", locale["nom"])
	}
	if locale["emplacement"] != "Grenoble" {
		t.Errorf("expected emplacement 'Grenoble', got '%v'", locale["emplacement"])
	}
}

func TestArchiveHandler_GetLigueLocale_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/leagues-locales/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.GetLigueLocale(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "ligue_locale_not_found" {
		t.Errorf("expected error 'league_locale_not_found', got '%v'", resp["error"])
	}
}

func TestArchiveHandler_GetLigueLocale_WithLigue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)
	ligueRepo := models.NewLigueRepository(db)
	splitRepo := models.NewSplitRepository(db)

	split, err := splitRepo.Create("Split 1", []string{"cp", "mistral"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}

	now := time.Now().Truncate(24 * time.Hour)
	ligue, err := ligueRepo.Create(now, now.Add(30*24*time.Hour), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}

	created, err := localeRepo.Create("Grenoble", "Grenoble")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id) VALUES ($1, $2)`,
		ligue.ID, created.ID)
	if err != nil {
		t.Fatalf("failed to associate locale with ligue: %v", err)
	}

	locale2, err := localeRepo.Create("Lyon", "Lyon")
	if err != nil {
		t.Fatalf("failed to create second locale: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id) VALUES ($1, $2)`,
		ligue.ID, locale2.ID)
	if err != nil {
		t.Fatalf("failed to associate second locale: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/archive/leagues-locales/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.GetLigueLocale(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locale, ok := resp["ligue_locale"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'ligue_locale' key")
	}
	if locale["nom"] != "Grenoble" {
		t.Errorf("expected nom 'Grenoble', got '%v'", locale["nom"])
	}

	liguesResp, ok := resp["ligues"].([]interface{})
	if !ok {
		t.Fatal("expected 'ligues' key as array")
	}
	if len(liguesResp) != 1 {
		t.Fatalf("expected 1 ligue, got %d", len(liguesResp))
	}
	firstLigue := liguesResp[0].(map[string]interface{})
	if firstLigue["split_name"] != "Split 1" {
		t.Errorf("expected split_name 'Split 1', got '%v'", firstLigue["split_name"])
	}
}

func TestArchiveHandler_GetLigueLocale_WithoutLigue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	created, err := localeRepo.Create("Standalone", "Grenoble")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/archive/leagues-locales/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.GetLigueLocale(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	ligues, ok := resp["ligues"].([]interface{})
	if !ok {
		t.Fatal("expected 'ligues' key as array")
	}
	if len(ligues) != 0 {
		t.Errorf("expected empty ligues, got %d", len(ligues))
	}
}

// --- CreateLigueLocale ---

func TestArchiveHandler_CreateLigueLocale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"nom":"New Locale","emplacement":"Grenoble"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/archive/leagues-locales", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateLigueLocale(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locale, ok := resp["ligue_locale"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'league_locale' key")
	}
	if locale["nom"] != "New Locale" {
		t.Errorf("expected nom 'New Locale', got '%v'", locale["nom"])
	}
	if locale["id"].(float64) <= 0 {
		t.Errorf("expected valid id, got %v", locale["id"])
	}
}

func TestArchiveHandler_CreateLigueLocale_NomVide(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"nom":"","emplacement":"Grenoble"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/archive/leagues-locales", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateLigueLocale(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_request_body" {
		t.Errorf("expected error 'invalid_request_body', got '%v'", resp["error"])
	}
}

// --- UpdateLigueLocale ---

func TestArchiveHandler_UpdateLigueLocale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	created, err := localeRepo.Create("Original", "Salle A")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	body := `{"nom":"Updated","emplacement":"Salle B"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/archive/leagues-locales/%d", created.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.UpdateLigueLocale(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	locale, ok := resp["ligue_locale"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'league_locale' key")
	}
	if locale["nom"] != "Updated" {
		t.Errorf("expected nom 'Updated', got '%v'", locale["nom"])
	}
	if locale["emplacement"] != "Salle B" {
		t.Errorf("expected emplacement 'Salle B', got '%v'", locale["emplacement"])
	}
}

func TestArchiveHandler_UpdateLigueLocale_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"nom":"Ghost","emplacement":"Nowhere"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/archive/leagues-locales/99999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.UpdateLigueLocale(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "ligue_locale_not_found" {
		t.Errorf("expected error 'league_locale_not_found', got '%v'", resp["error"])
	}
}

// --- DeleteLigueLocale ---

func TestArchiveHandler_DeleteLigueLocale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	localeRepo := models.NewLigueLocaleRepository(db)

	created, err := localeRepo.Create("To Delete", "Grenoble")
	if err != nil {
		t.Fatalf("failed to create ligue locale: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/archive/leagues-locales/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.DeleteLigueLocale(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "ligue locale deleted" {
		t.Errorf("expected 'league_locale deleted', got '%v'", resp["message"])
	}
}

func TestArchiveHandler_DeleteLigueLocale_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/archive/leagues-locales/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.DeleteLigueLocale(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "ligue_locale_not_found" {
		t.Errorf("expected error 'league_locale_not_found', got '%v'", resp["error"])
	}
}
