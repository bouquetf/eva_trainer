//go:build integration

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eva_trainer/db/seed"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/internal/testutil"

	"github.com/gin-gonic/gin"
)

const leaguesMigration = "000026_add_archive_leagues.up.sql"

var archiveMigrationFiles = []string{
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
	"000030_add_eva_teams.up.sql",
	"000031_add_archive_divs_and_bos.up.sql",
}

func setupArchiveHandlerTest(t *testing.T) (*ArchiveHandler, *services.JWTService, models.UserRepository, *sql.DB) {
	t.Helper()

	db, cleanup := testutil.SetupTestDB(t, archiveMigrationFiles)
	t.Cleanup(cleanup)

	if err := seed.SeedRolePermissions(db); err != nil {
		t.Fatalf("failed to seed role permissions: %v", err)
	}

	ligueLocaleRepo := models.NewLigueLocaleRepository(db)
	ligueRepo := models.NewLigueRepository(db)
	splitRepoModels := models.NewSplitRepository(db)
	evaTeamRepo := models.NewEvaTeamRepository(db)
	teamRepo := models.NewTeamRepository(db)
	divRepo := models.NewDivRepository(db)
	userRepo := models.NewUserRepository(db)
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	archiveHandler := NewArchiveHandler(ligueLocaleRepo, ligueRepo, splitRepoModels, evaTeamRepo, teamRepo, divRepo, services.NewLogger(services.LogLevelError, services.LogFormatText))

	return archiveHandler, jwtService, userRepo, db
}

func createTestLigue(t *testing.T, db *sql.DB) int {
	t.Helper()
	splitRepo := models.NewSplitRepository(db)
	split, err := splitRepo.Create("Test Split", []string{"cp", "mistral"})
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	ligueRepo := models.NewLigueRepository(db)
	now := time.Now().Truncate(24 * time.Hour)
	ligue, err := ligueRepo.Create(now, now.Add(30*24*time.Hour), split.ID)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}
	return ligue.ID
}

func createTestEvaTeam(t *testing.T, db *sql.DB, name, tag string) int {
	t.Helper()
	evaRepo := models.NewEvaTeamRepository(db)
	team, err := evaRepo.Create(name, tag, nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}
	return team.ID
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

// --- Eva Teams ---

func TestArchiveHandler_ListEvaTeams_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/teams", nil)

	handler.ListEvaTeams(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	teams, ok := resp["teams"].([]interface{})
	if !ok {
		t.Fatal("expected 'teams' array")
	}
	if len(teams) != 0 {
		t.Errorf("expected empty teams, got %d", len(teams))
	}
}

func TestArchiveHandler_ListEvaTeams_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	evaRepo := models.NewEvaTeamRepository(db)

	evaRepo.Create("Alpha", "AL", nil)
	evaRepo.Create("Beta", "BE", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/teams", nil)

	handler.ListEvaTeams(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	teams, ok := resp["teams"].([]interface{})
	if !ok {
		t.Fatal("expected 'teams' array")
	}
	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}
}

func TestArchiveHandler_GetEvaTeam_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	evaRepo := models.NewEvaTeamRepository(db)
	created, err := evaRepo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/archive/teams/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.GetEvaTeam(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	team, ok := resp["team"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'team' object")
	}
	if team["name"] != "Team A" {
		t.Errorf("expected name 'Team A', got '%v'", team["name"])
	}
}

func TestArchiveHandler_GetEvaTeam_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/archive/teams/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.GetEvaTeam(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_CreateEvaTeam_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	body := `{"name":"New Team","tag":"NT"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/archive/teams", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateEvaTeam(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	team, ok := resp["team"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'team' object")
	}
	if team["name"] != "New Team" {
		t.Errorf("expected name 'New Team', got '%v'", team["name"])
	}
	if team["tag"] != "NT" {
		t.Errorf("expected tag 'NT', got '%v'", team["tag"])
	}

	// Verify it was persisted
	evaRepo := models.NewEvaTeamRepository(db)
	found, err := evaRepo.FindByID(int(team["id"].(float64)))
	if err != nil {
		t.Fatalf("failed to find created team: %v", err)
	}
	if found.Name != "New Team" {
		t.Errorf("expected persisted name 'New Team', got %q", found.Name)
	}
}

func TestArchiveHandler_CreateEvaTeam_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"name":"","tag":"NT"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/archive/teams", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateEvaTeam(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_CreateEvaTeam_EmptyTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"name":"Team A","tag":""}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/archive/teams", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateEvaTeam(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_UpdateEvaTeam_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	evaRepo := models.NewEvaTeamRepository(db)
	created, err := evaRepo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	body := `{"name":"Updated Team","tag":"UT"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/archive/teams/%d", created.ID), strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.UpdateEvaTeam(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	team, ok := resp["team"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'team' object")
	}
	if team["name"] != "Updated Team" {
		t.Errorf("expected name 'Updated Team', got '%v'", team["name"])
	}
}

func TestArchiveHandler_UpdateEvaTeam_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	body := `{"name":"Updated","tag":"UP"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/archive/teams/99999", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.UpdateEvaTeam(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_DeleteEvaTeam_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	evaRepo := models.NewEvaTeamRepository(db)
	created, err := evaRepo.Create("Team A", "TA", nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/archive/teams/%d", created.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", created.ID)}}

	handler.DeleteEvaTeam(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "Équipe eva supprimée" {
		t.Errorf("expected 'Équipe eva supprimée', got '%v'", resp["message"])
	}
}

func TestArchiveHandler_DeleteEvaTeam_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/archive/teams/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.DeleteEvaTeam(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_CreateEvaTeam_WithAssociation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)

	userRepo := models.NewUserRepository(db)
	user, _, err := userRepo.FindOrCreateByDiscordID("discord_handler_assoc")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	teamRepo := models.NewTeamRepository(db)
	trainerTeam, err := teamRepo.Create("Handler Trainer Team", user.ID)
	if err != nil {
		t.Fatalf("failed to create trainer team: %v", err)
	}

	body := fmt.Sprintf(`{"name":"Eva Assoc Team","tag":"EAT","eva_trainer_team_id":%d}`, trainerTeam.ID)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/archive/teams", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateEvaTeam(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	team, ok := resp["team"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'team' object")
	}
	if team["eva_trainer_team_id"] == nil {
		t.Fatal("expected non-nil eva_trainer_team_id")
	}
	if team["eva_trainer_team_id"].(float64) != float64(trainerTeam.ID) {
		t.Errorf("expected eva_trainer_team_id %d, got %v", trainerTeam.ID, team["eva_trainer_team_id"])
	}
}

// --- Divs ---

func TestArchiveHandler_ListDivs_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/archive/divisions?ligue_id=%d", ligueID), nil)

	handler.ListDivs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	divs, ok := resp["divs"].([]interface{})
	if !ok {
		t.Fatal("expected 'divs' array")
	}
	if len(divs) != 0 {
		t.Errorf("expected empty divs, got %d", len(divs))
	}
}

func TestArchiveHandler_ListDivs_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	divRepo := models.NewDivRepository(db)

	divRepo.Create(ligueID, "Groupe A")
	divRepo.Create(ligueID, "Groupe B")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/archive/divisions?ligue_id=%d", ligueID), nil)

	handler.ListDivs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	divs, ok := resp["divs"].([]interface{})
	if !ok {
		t.Fatal("expected 'divs' array")
	}
	if len(divs) != 2 {
		t.Errorf("expected 2 divs, got %d", len(divs))
	}
}

func TestArchiveHandler_ListDivs_MissingLigueID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/archive/divisions", nil)

	handler.ListDivs(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "missing_ligue_id" {
		t.Errorf("expected error 'missing_ligue_id', got '%v'", resp["error"])
	}
}

func TestArchiveHandler_CreateDiv_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)

	body := fmt.Sprintf(`{"ligue_id":%d,"nom":"Groupe A"}`, ligueID)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/archive/divisions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDiv(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	div, ok := resp["div"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'div' object")
	}
	if div["nom"] != "Groupe A" {
		t.Errorf("expected nom 'Groupe A', got '%v'", div["nom"])
	}
	if div["id"].(float64) <= 0 {
		t.Errorf("expected valid id, got %v", div["id"])
	}
}

func TestArchiveHandler_CreateDiv_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	divRepo := models.NewDivRepository(db)
	_, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create initial div: %v", err)
	}

	body := fmt.Sprintf(`{"ligue_id":%d,"nom":"Groupe A"}`, ligueID)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/archive/divisions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDiv(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestArchiveHandler_GetDiv_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := setupArchiveHandlerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/archive/divisions/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.GetDiv(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "div_not_found" {
		t.Errorf("expected error 'div_not_found', got '%v'", resp["error"])
	}
}

func TestArchiveHandler_UpdateDiv_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	body := `{"nom":"Poule Haute"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/archive/divisions/%d", div.ID), strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}

	handler.UpdateDiv(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	updated, ok := resp["div"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'div' object")
	}
	if updated["nom"] != "Poule Haute" {
		t.Errorf("expected nom 'Poule Haute', got '%v'", updated["nom"])
	}
}

func TestArchiveHandler_DeleteDiv_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/archive/divisions/%d", div.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}

	handler.DeleteDiv(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify it was deleted
	_, err = divRepo.FindByID(div.ID)
	if err != models.ErrDivNotFound {
		t.Fatal("expected div to be deleted")
	}
}

func TestArchiveHandler_AssignTeamToDiv_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	teamID := createTestEvaTeam(t, db, "Team A", "TA")
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	body := fmt.Sprintf(`{"equipe_eva_id":%d}`, teamID)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/archive/divisions/%d/assign", div.ID), strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}

	handler.AssignTeamToDiv(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify via GetDiv
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/archive/divisions/%d", div.ID), nil)
	c2.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}
	handler.GetDiv(c2)

	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	teams, _ := resp2["teams"].([]interface{})
	if len(teams) != 1 {
		t.Errorf("expected 1 team in div, got %d", len(teams))
	}
}

func TestArchiveHandler_RemoveTeamFromDiv(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	teamID := createTestEvaTeam(t, db, "Team A", "TA")
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}

	err = divRepo.AssignTeam(div.ID, teamID)
	if err != nil {
		t.Fatalf("failed to assign team: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/archive/divisions/%d/teams/%d", div.ID, teamID), nil)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", div.ID)},
		{Key: "teamId", Value: fmt.Sprintf("%d", teamID)},
	}

	handler.RemoveTeamFromDiv(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify teams list is empty
	teams, err := divRepo.FindTeamsByDivID(div.ID)
	if err != nil {
		t.Fatalf("failed to list teams: %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("expected empty teams, got %d", len(teams))
	}
}

func TestArchiveHandler_GetDivStandings_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	teamID := createTestEvaTeam(t, db, "Team A", "TA")
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}
	err = divRepo.AssignTeam(div.ID, teamID)
	if err != nil {
		t.Fatalf("failed to assign team: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/archive/divisions/%d/standings", div.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}

	handler.GetDivStandings(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	standings, ok := resp["standings"].([]interface{})
	if !ok {
		t.Fatal("expected 'standings' array")
	}
	if len(standings) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(standings))
	}
	entry := standings[0].(map[string]interface{})
	if entry["bos_played"].(float64) != 0 {
		t.Errorf("expected bos_played 0, got %v", entry["bos_played"])
	}
}

func TestArchiveHandler_GetDivStandings_WithBOs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, db := setupArchiveHandlerTest(t)
	ligueID := createTestLigue(t, db)
	teamAID := createTestEvaTeam(t, db, "Team A", "TA")
	teamBID := createTestEvaTeam(t, db, "Team B", "TB")
	divRepo := models.NewDivRepository(db)
	div, err := divRepo.Create(ligueID, "Groupe A")
	if err != nil {
		t.Fatalf("failed to create div: %v", err)
	}
	err = divRepo.AssignTeam(div.ID, teamAID)
	if err != nil {
		t.Fatalf("failed to assign team A: %v", err)
	}
	err = divRepo.AssignTeam(div.ID, teamBID)
	if err != nil {
		t.Fatalf("failed to assign team B: %v", err)
	}

	// Create a BO: Team A beats Team B
	_, err = db.Exec(
		`INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, vainqueur_id, format)
		 VALUES ($1, $2, $3, $4, 'BO3')`,
		ligueID, teamAID, teamBID, teamAID,
	)
	if err != nil {
		t.Fatalf("failed to create bo: %v", err)
	}

	// Seed a map for matchs (maps are from earlier migrations)
	var mapID int
	err = db.QueryRow(`INSERT INTO maps (name)
		VALUES ('Test Map') RETURNING id`).Scan(&mapID)
	if err != nil {
		t.Fatalf("failed to seed map: %v", err)
	}

	// Create a match in that BO: Team A wins 16-10
	_, err = db.Exec(
		`INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id)
		 VALUES (1, $1, 16, 10, $2)`,
		mapID, teamAID,
	)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/archive/divisions/%d/standings", div.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", div.ID)}}

	handler.GetDivStandings(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	standings, ok := resp["standings"].([]interface{})
	if !ok {
		t.Fatal("expected 'standings' array")
	}
	if len(standings) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(standings))
	}

	entryA := standings[0].(map[string]interface{}) // Team A should be first (won the BO)
	if entryA["team_name"] != "Team A" {
		t.Errorf("expected Team A first, got '%v'", entryA["team_name"])
	}
	if entryA["bo_won"].(float64) != 1 {
		t.Errorf("expected bo_won 1, got %v", entryA["bo_won"])
	}
	if entryA["bo_points"].(float64) != 1 {
		t.Errorf("expected bo_points 1, got %v", entryA["bo_points"])
	}
}
