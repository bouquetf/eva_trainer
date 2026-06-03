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

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupArchivePermissionsTest(t *testing.T) (*ArchivePermissionsHandler, *models.User, string, models.UserRepository, models.GlobalRoleRepository, models.TeamRoleRepository, func()) {
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

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to ping database: %v", err)
	}

	migrationsDir := "../../db/migrations"
	migrationFiles := []string{
		"000001_initial_schema.up.sql",
		"000002_add_refresh_tokens.up.sql",
		"000003_add_display_name.up.sql",
		"000004_add_discord_username.up.sql",
		"000025_add_archive_roles.up.sql",
	}

	for _, mf := range migrationFiles {
		path := filepath.Join(migrationsDir, mf)
		content, err := os.ReadFile(path)
		if err != nil {
			db.Close()
			container.Terminate(ctx)
			t.Fatalf("failed to read migration %s: %v", mf, err)
		}
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				db.Close()
				container.Terminate(ctx)
				t.Fatalf("failed to execute migration %s: %v\nSQL: %s", mf, err, stmt)
			}
		}
	}

	userRepo := models.NewUserRepository(db)
	globalRoleRepo := models.NewGlobalRoleRepository(db)
	teamRoleRepo := models.NewTeamRoleRepository(db)

	handler := NewArchivePermissionsHandler(globalRoleRepo, teamRoleRepo, services.NewLogger(services.LogLevelError, services.LogFormatText))

	jwtService := services.NewJWTService(testSecret, 24*60*60, 7*24*60*60)

	adminUser, _, err := userRepo.FindOrCreateByDiscordID("archive-perm-admin-" + t.Name())
	if err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to create admin user: %v", err)
	}
	if err := userRepo.SetAdmin(adminUser.ID); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to set admin: %v", err)
	}
	adminUser, err = userRepo.FindByID(adminUser.ID)
	if err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to re-fetch admin user: %v", err)
	}
	adminToken, _, err := jwtService.GenerateAccessToken(adminUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	cleanup := func() {
		db.Close()
		container.Terminate(context.Background())
	}

	return handler, adminUser, adminToken, userRepo, globalRoleRepo, teamRoleRepo, cleanup
}

func executeArchivePermissionsHandler(t *testing.T, handler gin.HandlerFunc, user *models.User, token string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/v1/archive/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	c.Set("user", user)

	handler(c)

	return w
}

func TestArchivePermissionsHandler_GetPermissions_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, adminUser, adminToken, _, _, _, cleanup := setupArchivePermissionsTest(t)
	defer cleanup()

	w := executeArchivePermissionsHandler(t, handler.GetPermissions, adminUser, adminToken)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp middleware.ArchivePermissions
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.CanRead {
		t.Errorf("expected CanRead=true for admin")
	}
	if !resp.CanWrite {
		t.Errorf("expected CanWrite=true for admin")
	}
}

func TestArchivePermissionsHandler_GetPermissions_Archivist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, userRepo, globalRoleRepo, _, cleanup := setupArchivePermissionsTest(t)
	defer cleanup()

	archivistUser, _, err := userRepo.FindOrCreateByDiscordID("archivist-only-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create archivist user: %v", err)
	}

	if err := globalRoleRepo.Add(archivistUser.ID, models.RoleArchivist); err != nil {
		t.Fatalf("failed to add archivist role: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*60*60, 7*24*60*60)
	token, _, err := jwtService.GenerateAccessToken(archivistUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := executeArchivePermissionsHandler(t, handler.GetPermissions, archivistUser, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp middleware.ArchivePermissions
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.CanRead {
		t.Errorf("expected CanRead=true for archivist")
	}
	if !resp.CanWrite {
		t.Errorf("expected CanWrite=true for archivist")
	}
}

func TestArchivePermissionsHandler_GetPermissions_NoAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, userRepo, _, _, cleanup := setupArchivePermissionsTest(t)
	defer cleanup()

	regularUser, _, err := userRepo.FindOrCreateByDiscordID("regular-no-access-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create regular user: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*60*60, 7*24*60*60)
	token, _, err := jwtService.GenerateAccessToken(regularUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := executeArchivePermissionsHandler(t, handler.GetPermissions, regularUser, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp middleware.ArchivePermissions
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.CanRead {
		t.Errorf("expected CanRead=false for regular user")
	}
	if resp.CanWrite {
		t.Errorf("expected CanWrite=false for regular user")
	}
}
