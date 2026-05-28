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
	"strconv"
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

func setupAdminHandler(t *testing.T) (*AdminHandler, models.UserRepository, *sql.DB) {
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
		container.Terminate(ctx)
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
	permRepo := models.NewPermissionRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)

	handler := NewAdminHandler(userRepo, permRepo, rolePermRepo, teamRepo, teamMemberRepo, services.NewLogger(services.LogLevelError, services.LogFormatText))

	return handler, userRepo, db
}

func createAdminUserForTest(t *testing.T, userRepo models.UserRepository) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("admin-test-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := userRepo.SetAdmin(user.ID); err != nil {
		t.Fatalf("failed to set admin: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	return user, token
}

func createNonAdminUserForTest(t *testing.T, userRepo models.UserRepository) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("non-admin-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	token, _, err := jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	return user, token
}

func executeAdminHandler(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string) *httptest.ResponseRecorder {
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

	relative := strings.TrimPrefix(path, "/api/v1/admin/users/")
	parts := strings.SplitN(relative, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		idStr := strings.Split(parts[0], "?")[0]
		if _, err := strconv.Atoi(idStr); err == nil {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		}
		if len(parts) > 1 {
			if strings.HasPrefix(parts[1], "teams/") {
				teamParts := strings.Split(strings.TrimPrefix(parts[1], "teams/"), "/")
				if len(teamParts) > 0 {
					c.Params = append(c.Params, gin.Param{Key: "teamId", Value: teamParts[0]})
				}
			}
		}
	}

	handler(c)

	return w
}

func TestAdminHandler_ListUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, token := createAdminUserForTest(t, userRepo)

	_, _, err := userRepo.FindOrCreateByDiscordID("user1-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	_, _, err = userRepo.FindOrCreateByDiscordID("user2-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	w := executeAdminHandler(t, handler.ListUsers, http.MethodGet, "/api/v1/admin/users", "", admin, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Users []*models.User `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Users) < 3 {
		t.Errorf("expected at least 3 users, got %d", len(resp.Users))
	}
}

func TestAdminHandler_PromoteToAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	targetUser, _, err := userRepo.FindOrCreateByDiscordID("promote-target-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create target user: %v", err)
	}

	path := fmt.Sprintf("/api/v1/admin/users/%d/promote", targetUser.ID)
	w := executeAdminHandler(t, handler.PromoteToAdmin, http.MethodPost, path, "", admin, adminToken)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Message string `json:"message"`
		UserID  int    `json:"user_id"`
		IsAdmin bool   `json:"is_admin"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Message != "user promoted to admin" {
		t.Errorf("expected 'user promoted to admin', got '%s'", resp.Message)
	}
	if !resp.IsAdmin {
		t.Errorf("expected is_admin=true in response")
	}

	updated, err := userRepo.FindByID(targetUser.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if !updated.IsAdmin {
		t.Errorf("expected target user to be admin after promotion")
	}
}

func TestAdminHandler_PromoteToAdmin_AlreadyAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	path := fmt.Sprintf("/api/v1/admin/users/%d/promote", admin.ID)
	w := executeAdminHandler(t, handler.PromoteToAdmin, http.MethodPost, path, "", admin, adminToken)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != "already_admin" {
		t.Errorf("expected 'already_admin', got '%s'", resp.Error)
	}
}

func TestAdminHandler_PromoteToAdmin_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	path := "/api/v1/admin/users/99999/promote"
	w := executeAdminHandler(t, handler.PromoteToAdmin, http.MethodPost, path, "", admin, adminToken)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != "user_not_found" {
		t.Errorf("expected 'user_not_found', got '%s'", resp.Error)
	}
}

func TestAdminHandler_GetUserDetail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	path := fmt.Sprintf("/api/v1/admin/users/%d", admin.ID)
	w := executeAdminHandler(t, handler.GetUserDetail, http.MethodGet, path, "", admin, adminToken)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		User        *models.User `json:"user"`
		Memberships []any        `json:"memberships"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.User == nil {
		t.Errorf("expected user in response")
	}
	if resp.User.ID != admin.ID {
		t.Errorf("expected user id %d, got %d", admin.ID, resp.User.ID)
	}
}

func TestAdminHandler_GetUserDetail_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	path := "/api/v1/admin/users/99999"
	w := executeAdminHandler(t, handler.GetUserDetail, http.MethodGet, path, "", admin, adminToken)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAdminHandler_GetUserDetail_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	admin, adminToken := createAdminUserForTest(t, userRepo)

	w := executeAdminHandler(t, handler.GetUserDetail, http.MethodGet, "/api/v1/admin/users/abc", "", admin, adminToken)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAdminHandler_Middleware_NonAdmin_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	nonAdmin, nonAdminToken := createNonAdminUserForTest(t, userRepo)

	userRepo.FindAll()

	r := gin.New()
	r.GET("/api/v1/admin/users",
		func(c *gin.Context) {
			c.Set("user", nonAdmin)
			c.Next()
		},
		middleware.RequireAdmin(services.NewLogger(services.LogLevelError, services.LogFormatText)),
		handler.ListUsers,
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+nonAdminToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%v'", resp["error"])
	}
}

func TestAdminHandler_Middleware_NoAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _ := setupAdminHandler(t)

	r := gin.New()
	r.GET("/api/v1/admin/users",
		func(c *gin.Context) {
			c.Set("user", nil)
			c.Next()
		},
		middleware.RequireAdmin(services.NewLogger(services.LogLevelError, services.LogFormatText)),
		handler.ListUsers,
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAdminHandler_Middleware_NonAdminPromote_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _ := setupAdminHandler(t)
	nonAdmin, nonAdminToken := createNonAdminUserForTest(t, userRepo)

	targetUser, _, err := userRepo.FindOrCreateByDiscordID("promote-target-nonadmin-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create target user: %v", err)
	}

	path := fmt.Sprintf("/api/v1/admin/users/%d/promote", targetUser.ID)

	r := gin.New()
	r.POST("/api/v1/admin/users/:id/promote",
		func(c *gin.Context) {
			c.Set("user", nonAdmin)
			c.Params = append(c.Params, gin.Param{Key: "id", Value: strconv.Itoa(targetUser.ID)})
			c.Next()
		},
		middleware.RequireAdmin(services.NewLogger(services.LogLevelError, services.LogFormatText)),
		handler.PromoteToAdmin,
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, nil)
	req.Header.Set("Authorization", "Bearer "+nonAdminToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%v'", resp["error"])
	}
}
