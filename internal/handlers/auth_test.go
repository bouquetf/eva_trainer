//go:build integration

package handlers

import (
	"context"
	"crypto/sha256"
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
	"golang.org/x/oauth2"
)

const testSecret = "handler-test-secret-that-is-32-bytes-long-for-test!!"

var testDB *sql.DB

const migrationsDir = "../../db/migrations"

var migrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000002_add_refresh_tokens.up.sql",
	"000003_add_display_name.up.sql",
	"000004_add_discord_username.up.sql",
	"000016_add_user_settings.up.sql",
}

func setupTestDB(t *testing.T) *sql.DB {
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

func setupAuthHandler(t *testing.T) (*AuthHandler, models.UserRepository, models.RefreshTokenRepository, *services.JWTService) {
	t.Helper()

	db := setupTestDB(t)

	userRepo := models.NewUserRepository(db)
	refreshRepo := models.NewRefreshTokenRepository(db)
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	oauthCfg := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
		RedirectURL: "http://localhost:8080/api/v1/auth/discord/callback",
		Scopes:      []string{"identify"},
	}

	handler := NewAuthHandler(userRepo, refreshRepo, oauthCfg, jwtService, "http://localhost:5173", services.NewLogger(services.LogLevelError, services.LogFormatText))

	return handler, userRepo, refreshRepo, jwtService
}

func TestLoginRedirect_Returns302(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _ := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/discord/login", nil)

	handler.LoginRedirect(c)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 Found, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "https://discord.com/api/oauth2/authorize") {
		t.Errorf("expected redirect to Discord, got Location: %s", location)
	}
	if !strings.Contains(location, "client_id=test-client-id") {
		t.Errorf("expected client_id in redirect URL")
	}
	if !strings.Contains(location, "state=") {
		t.Errorf("expected state parameter in redirect URL")
	}
}

func TestMe_WithoutToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.Me(c)
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMe_WithValidToken_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("test-discord-id")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()

		handler.Me(c)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["discord_id"] != "test-discord-id" {
		t.Errorf("expected discord_id 'test-discord-id', got %v", resp["discord_id"])
	}
}

func TestRefresh_WithValidToken_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, refreshRepo, _ := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("refresh-test-user")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	refreshPlain, err := refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(7*24*time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	body := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshPlain)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Refresh(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Errorf("expected access_token in response")
	}
	if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
		t.Errorf("expected refresh_token in response")
	}
	if resp["refresh_token"] == refreshPlain {
		t.Errorf("expected new refresh token (rotation)")
	}
}

func TestRefresh_WithRevokedToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, refreshRepo, _ := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("revoked-test-user")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	refreshPlain, err := refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(7*24*time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshPlain)))
	if err := refreshRepo.RevokeRefreshToken(hash); err != nil {
		t.Fatalf("RevokeRefreshToken failed: %v", err)
	}

	body := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshPlain)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Refresh(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "refresh_token_revoked" {
		t.Errorf("expected error 'refresh_token_revoked', got '%s'", resp["error"])
	}
}

func TestGetSettings_NoSettings_ReturnsDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("settings-defaults")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/settings", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.GetSettings(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if v := resp["revision_correct_timer"]; v != float64(3) {
		t.Errorf("expected correct_timer 3, got %v", v)
	}
	if v := resp["revision_incorrect_timer"]; v != float64(5) {
		t.Errorf("expected incorrect_timer 5, got %v", v)
	}
	if v := resp["revision_question_count"]; v != float64(20) {
		t.Errorf("expected question_count 20, got %v", v)
	}
}

func TestGetSettings_WithoutToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/settings", nil)

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.GetSettings(c)
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetSettings_WithPartialSettings_ReturnsMerged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("settings-partial")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Stocker des settings où seuls certains champs sont > 0
	// GetUserSettings fusionne avec les défauts (champs à 0 → defaults)
	if err := userRepo.UpdateUserSettings(user.ID, &models.UserSettings{RevisionCorrectTimer: 2}); err != nil {
		t.Fatalf("failed to set partial settings: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/settings", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.GetSettings(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if v := resp["revision_correct_timer"]; v != float64(2) {
		t.Errorf("expected correct_timer 2, got %v", v)
	}
	if v := resp["revision_incorrect_timer"]; v != float64(5) {
		t.Errorf("expected incorrect_timer 5 (default), got %v", v)
	}
}

func TestUpdateSettings_StoresAndPersists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("settings-store")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	body := `{"revision_correct_timer": 4, "revision_incorrect_timer": 6, "revision_question_count": 15}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/auth/settings", strings.NewReader(body))
	c.Request.Header.Set("Authorization", "Bearer "+token)
	c.Request.Header.Set("Content-Type", "application/json")

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.UpdateSettings(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Vérifier la persistance via GET
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/settings", nil)
	c2.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth(nil, userRepo, jwtService)(c2)
	if !c2.IsAborted() {
		c2.Next()
		handler.GetSettings(c2)
	}

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET, got %d", w2.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if v := resp["revision_correct_timer"]; v != float64(4) {
		t.Errorf("expected correct_timer 4, got %v", v)
	}
	if v := resp["revision_incorrect_timer"]; v != float64(6) {
		t.Errorf("expected incorrect_timer 6, got %v", v)
	}
	if v := resp["revision_question_count"]; v != float64(15) {
		t.Errorf("expected question_count 15, got %v", v)
	}
}

func TestUpdateSettings_PartialUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("settings-partial-update")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Pré-remplir avec des valeurs connues
	initial := `{"revision_correct_timer": 10, "revision_incorrect_timer": 20, "revision_question_count": 30}`
	w0 := httptest.NewRecorder()
	c0, _ := gin.CreateTestContext(w0)
	c0.Request = httptest.NewRequest(http.MethodPut, "/api/v1/auth/settings", strings.NewReader(initial))
	c0.Request.Header.Set("Authorization", "Bearer "+token)
	c0.Request.Header.Set("Content-Type", "application/json")

	middleware.RequireAuth(nil, userRepo, jwtService)(c0)
	if !c0.IsAborted() {
		c0.Next()
		handler.UpdateSettings(c0)
	}
	if w0.Code != http.StatusOK {
		t.Fatalf("failed to set initial settings: %d", w0.Code)
	}

	// Mise à jour partielle : seulement correct_timer
	partial := `{"revision_correct_timer": 2}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/auth/settings", strings.NewReader(partial))
	c.Request.Header.Set("Authorization", "Bearer "+token)
	c.Request.Header.Set("Content-Type", "application/json")

	middleware.RequireAuth(nil, userRepo, jwtService)(c)
	if !c.IsAborted() {
		c.Next()
		handler.UpdateSettings(c)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if v := resp["revision_correct_timer"]; v != float64(2) {
		t.Errorf("expected correct_timer 2, got %v", v)
	}
	if v := resp["revision_incorrect_timer"]; v != float64(20) {
		t.Errorf("expected incorrect_timer 20 (unchanged), got %v", v)
	}
	if v := resp["revision_question_count"]; v != float64(30) {
		t.Errorf("expected question_count 30 (unchanged), got %v", v)
	}
}

func TestUpdateSettings_InvalidValues_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, jwtService := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("settings-invalid")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	tests := []struct {
		name string
		body string
	}{
		{"timer correct = 0", `{"revision_correct_timer": 0}`},
		{"question_count = 0", `{"revision_question_count": 0}`},
		{"body invalide", `not-json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/auth/settings", strings.NewReader(tt.body))
			c.Request.Header.Set("Authorization", "Bearer "+token)
			c.Request.Header.Set("Content-Type", "application/json")

			middleware.RequireAuth(nil, userRepo, jwtService)(c)
			if !c.IsAborted() {
				c.Next()
				handler.UpdateSettings(c)
			}

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestRefresh_WithExpiredToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, refreshRepo, _ := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("expired-test-user")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	refreshPlain, err := refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	body := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshPlain)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Refresh(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "refresh_token_expired" {
		t.Errorf("expected error 'refresh_token_expired', got '%s'", resp["error"])
	}
}
