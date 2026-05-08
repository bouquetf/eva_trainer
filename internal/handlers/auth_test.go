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
	"golang.org/x/oauth2"
)

const testSecret = "handler-test-secret-that-is-32-bytes-long-for-test!!"

var testDB *sql.DB

const migrationsDir = "../../db/migrations"

var migrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000002_add_refresh_tokens.up.sql",
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
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
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

	handler := NewAuthHandler(userRepo, refreshRepo, oauthCfg, jwtService, "http://localhost:5173")

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

	middleware.RequireAuth(userRepo, jwtService)(c)
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

	middleware.RequireAuth(userRepo, jwtService)(c)
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

func TestRefresh_WithExpiredToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, refreshRepo, _ := setupAuthHandler(t)

	user, _, err := userRepo.FindOrCreateByDiscordID("expired-test-user")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	refreshPlain, err := refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(-1*time.Hour))
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
