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

	"eva_trainer/db/seed"
	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTeamRolesHandler(t *testing.T) (*TeamRolesHandler, models.UserRepository, models.TeamRepository, models.TeamMemberRepository, models.PermissionRepository, models.RolePermissionRepository, *sql.DB) {
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

	// Run schema migrations
	allMigrationFiles := []string{
		"000001_initial_schema.up.sql",
		"000002_add_refresh_tokens.up.sql",
		"000003_add_display_name.up.sql",
		"000004_add_discord_username.up.sql",
		"000005_seed_permissions.up.sql",
		"000008_add_split_permission.up.sql",
	}

	for _, mf := range allMigrationFiles {
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

	// Seed role permissions
	if err := seed.SeedRolePermissions(db); err != nil {
		t.Fatalf("SeedRolePermissions failed: %v", err)
	}

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	rolePermRepo := models.NewRolePermissionRepository(db)
	permRepo := models.NewPermissionRepository(db)
	handler := NewTeamRolesHandler(teamRepo, teamMemberRepo, userRepo, rolePermRepo, permRepo, services.NewLogger(services.LogLevelError, services.LogFormatText))

	return handler, userRepo, teamRepo, teamMemberRepo, permRepo, rolePermRepo, db
}

func createRoleAuthUser(t *testing.T, userRepo models.UserRepository, discordID string) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID(discordID)
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

func executeRoleRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, memberRepo models.TeamMemberRepository) *httptest.ResponseRecorder {
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

	// Support both /:id/members and /:id/members/:userId/role and /:id/transfer-captaincy
	pathStr := strings.TrimPrefix(path, "/api/v1/teams/")
	parts := strings.SplitN(pathStr, "/", 2)
	teamIDStr := parts[0]
	teamID, _ := strconv.Atoi(teamIDStr)
	c.Params = gin.Params{{Key: "id", Value: teamIDStr}}
	if len(parts) > 1 {
		subParts := strings.Split(parts[1], "/")
		if len(subParts) >= 3 && subParts[0] == "members" {
			c.Params = append(c.Params, gin.Param{Key: "userId", Value: subParts[1]})
		}
	}

	if memberRepo != nil && teamID > 0 {
		member, err := memberRepo.FindByUserAndTeam(user.ID, teamID)
		if err == nil {
			c.Set("team_member", member)
		}
	}

	handler(c)

	return w
}

func TestTeamRolesHandler_ListMembers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "list-cap-"+t.Name())
	member1, _ := createRoleAuthUser(t, userRepo, "list-m1-"+t.Name())
	member2, _ := createRoleAuthUser(t, userRepo, "list-m2-"+t.Name())

	team, err := teamRepo.Create("List Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, member1.ID, models.RoleEditor); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	if err := memberRepo.AddMember(team.ID, member2.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.ListMembers, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/members", team.ID), "", captain, token, memberRepo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	members, ok := resp["members"].([]interface{})
	if !ok {
		t.Fatal("expected 'members' array")
	}
	if len(members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(members))
	}

	first := members[0].(map[string]interface{})
	if first["user_id"] == nil || first["role"] == nil || first["joined_at"] == nil {
		t.Errorf("member missing required fields: user_id, role, joined_at")
	}
}

func TestTeamRolesHandler_ListMembers_NonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, _ := createRoleAuthUser(t, userRepo, "list-nm-cap-"+t.Name())
	_, token := createRoleAuthUser(t, userRepo, "list-nm-out-"+t.Name())

	team, err := teamRepo.Create("Private", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	authMw := middleware.RequireAuth(nil, userRepo, jwtService)
	memberMw := middleware.RequireTeamRole(nil, memberRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/members", team.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", team.ID)},
	}

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		memberMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		handler.ListMembers(c)
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTeamRolesHandler_UpdateRole_CaptainToEditor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "ue-cap-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "ue-target-"+t.Name())

	team, err := teamRepo.Create("Role Update", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, target.ID),
		`{"role": "editor"}`, captain, token, memberRepo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["role"] != "editor" {
		t.Errorf("expected role 'editor', got '%v'", resp["role"])
	}

	member, _ := memberRepo.FindByUserAndTeam(target.ID, team.ID)
	if member.Role != models.RoleEditor {
		t.Errorf("expected member role to be 'editor', got '%s'", member.Role)
	}
}

func TestTeamRolesHandler_UpdateRole_PromoteCocaptain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "pc-cap-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "pc-target-"+t.Name())

	team, err := teamRepo.Create("Promote Co", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleEditor); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, target.ID),
		`{"role": "cocaptain"}`, captain, token, memberRepo)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	member, _ := memberRepo.FindByUserAndTeam(target.ID, team.ID)
	if member.Role != models.RoleCocaptain {
		t.Errorf("expected member role to be 'cocaptain', got '%s'", member.Role)
	}
}

func TestTeamRolesHandler_UpdateRole_AlreadyHasCocaptain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "ahc-cap-"+t.Name())
	cocaptain, _ := createRoleAuthUser(t, userRepo, "ahc-co-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "ahc-target-"+t.Name())

	team, err := teamRepo.Create("Already Co", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, cocaptain.ID, models.RoleCocaptain); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleEditor); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, target.ID),
		`{"role": "cocaptain"}`, captain, token, memberRepo)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "team_already_has_cocaptain" {
		t.Errorf("expected 'team_already_has_cocaptain', got '%s'", resp["error"])
	}
}

func TestTeamRolesHandler_UpdateRole_NonCaptain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, _ := createRoleAuthUser(t, userRepo, "nc-cap-"+t.Name())
	editor, editorToken := createRoleAuthUser(t, userRepo, "nc-editor-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "nc-target-"+t.Name())

	team, err := teamRepo.Create("Non Captain", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, editor.ID, models.RoleEditor); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Set up context with auth middleware manually
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	authMw := middleware.RequireAuth(nil, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, target.ID),
		strings.NewReader(`{"role": "editor"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+editorToken)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", team.ID)},
		{Key: "userId", Value: fmt.Sprintf("%d", target.ID)},
	}

	roleMw := middleware.RequireTeamRole(nil, memberRepo, "captain")

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		roleMw(c)
	}
	if !c.IsAborted() {
		c.Next()
		handler.UpdateRole(c)
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamRolesHandler_UpdateRole_DowngradeAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "da-cap-"+t.Name())
	adminUser, _ := createRoleAuthUser(t, userRepo, "da-admin-"+t.Name())

	team, err := teamRepo.Create("Admin Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, adminUser.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Make the target user an admin
	if err := userRepo.SetAdmin(adminUser.ID); err != nil {
		t.Fatalf("SetAdmin failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, adminUser.ID),
		`{"role": "editor"}`, captain, token, memberRepo)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "cannot_downgrade_admin" {
		t.Errorf("expected 'cannot_downgrade_admin', got '%s'", resp["error"])
	}
}

func TestTeamRolesHandler_UpdateRole_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "ir-cap-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "ir-target-"+t.Name())

	team, err := teamRepo.Create("Invalid Role", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/%d/role", team.ID, target.ID),
		`{"role": "superadmin"}`, captain, token, memberRepo)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_role" {
		t.Errorf("expected 'invalid_role', got '%s'", resp["error"])
	}
}

func TestTeamRolesHandler_UpdateRole_TargetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "tnf-cap-"+t.Name())

	team, err := teamRepo.Create("Target Not Found", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeRoleRequest(t, handler.UpdateRole, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/members/99999/role", team.ID),
		`{"role": "member"}`, captain, token, memberRepo)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "member_not_found" {
		t.Errorf("expected 'member_not_found', got '%s'", resp["error"])
	}
}

func TestTeamRolesHandler_TransferCaptaincy_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "tc-cap-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "tc-target-"+t.Name())

	team, err := teamRepo.Create("Transfer", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleEditor); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	w := executeRoleRequest(t, handler.TransferCaptaincy, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/transfer-captaincy", team.ID),
		fmt.Sprintf(`{"user_id": %d}`, target.ID), captain, token, memberRepo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["new_captain_id"] != float64(target.ID) {
		t.Errorf("expected new_captain_id %d, got %v", target.ID, resp["new_captain_id"])
	}

	// Verify old captain is now cocaptain
	oldMember, err := memberRepo.FindByUserAndTeam(captain.ID, team.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTeam failed: %v", err)
	}
	if oldMember.Role != models.RoleCocaptain {
		t.Errorf("expected old captain to be cocaptain, got '%s'", oldMember.Role)
	}

	// Verify new captain
	newMember, err := memberRepo.FindByUserAndTeam(target.ID, team.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTeam failed: %v", err)
	}
	if newMember.Role != models.RoleCaptain {
		t.Errorf("expected new member to be captain, got '%s'", newMember.Role)
	}

	// Verify team captain_id changed
	updatedTeam, err := teamRepo.FindByID(team.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if updatedTeam.CaptainID != target.ID {
		t.Errorf("expected team captain_id %d, got %d", target.ID, updatedTeam.CaptainID)
	}
}

func TestTeamRolesHandler_TransferCaptaincy_SameUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "tsu-cap-"+t.Name())

	team, err := teamRepo.Create("Same User", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeRoleRequest(t, handler.TransferCaptaincy, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/transfer-captaincy", team.ID),
		fmt.Sprintf(`{"user_id": %d}`, captain.ID), captain, token, memberRepo)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "same_user" {
		t.Errorf("expected 'same_user', got '%s'", resp["error"])
	}
}

func TestTeamRolesHandler_TransferCaptaincy_AdminTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, token := createRoleAuthUser(t, userRepo, "ta-cap-"+t.Name())
	adminUser, _ := createRoleAuthUser(t, userRepo, "ta-admin-"+t.Name())

	team, err := teamRepo.Create("Admin Target", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, adminUser.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	if err := userRepo.SetAdmin(adminUser.ID); err != nil {
		t.Fatalf("SetAdmin failed: %v", err)
	}

	w := executeRoleRequest(t, handler.TransferCaptaincy, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/transfer-captaincy", team.ID),
		fmt.Sprintf(`{"user_id": %d}`, adminUser.ID), captain, token, memberRepo)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamRolesHandler_TransferCaptaincy_NonCaptain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _, _, _ := setupTeamRolesHandler(t)

	captain, _ := createRoleAuthUser(t, userRepo, "tnc-cap-"+t.Name())
	cocaptain, coToken := createRoleAuthUser(t, userRepo, "tnc-co-"+t.Name())
	target, _ := createRoleAuthUser(t, userRepo, "tnc-target-"+t.Name())

	team, err := teamRepo.Create("Non Captain Transfer", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, cocaptain.ID, models.RoleCocaptain); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	if err := memberRepo.AddMember(team.ID, target.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	authMw := middleware.RequireAuth(nil, userRepo, jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/transfer-captaincy", team.ID),
		strings.NewReader(fmt.Sprintf(`{"user_id": %d}`, target.ID)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+coToken)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", team.ID)},
	}

	coMember, err := memberRepo.FindByUserAndTeam(cocaptain.ID, team.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTeam failed: %v", err)
	}
	c.Set("team_member", coMember)
	c.Set("team_id", team.ID)

	authMw(c)
	if !c.IsAborted() {
		c.Next()
		handler.TransferCaptaincy(c)
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}
