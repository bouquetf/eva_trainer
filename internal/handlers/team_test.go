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

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTeamHandler(t *testing.T) (*TeamHandler, models.UserRepository, models.TeamRepository, models.TeamMemberRepository, *sql.DB) {
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
	teamMemberRepo := models.NewTeamMemberRepository(db)
	handler := NewTeamHandler(teamRepo, teamMemberRepo)

	return handler, userRepo, teamRepo, teamMemberRepo, db
}

func createAuthUser(t *testing.T, userRepo models.UserRepository) (*models.User, string) {
	t.Helper()

	user, _, err := userRepo.FindOrCreateByDiscordID("team-handler-user-" + t.Name())
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

func executeWithAuth(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string) *httptest.ResponseRecorder {
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
	pathStr := strings.TrimPrefix(path, "/api/v1/teams/")
	parts := strings.SplitN(pathStr, "/", 2)
	c.Params = gin.Params{{Key: "id", Value: parts[0]}}

	handler(c)

	return w
}

func TestTeamHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	w := executeWithAuth(t, handler.Create, http.MethodPost, "/api/v1/teams",
		`{"name": "New Team"}`, user, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["name"] != "New Team" {
		t.Errorf("expected name 'New Team', got %v", resp["name"])
	}
	if resp["id"] == nil || resp["id"].(float64) <= 0 {
		t.Errorf("expected positive id, got %v", resp["id"])
	}
}

func TestTeamHandler_Create_NoName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	w := executeWithAuth(t, handler.Create, http.MethodPost, "/api/v1/teams",
		`{}`, user, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTeamHandler_ListMyTeams_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	_, err := teamRepo.Create("Team A", user.ID)
	if err != nil {
		t.Fatalf("Create Team A failed: %v", err)
	}
	_, err = teamRepo.Create("Team B", user.ID)
	if err != nil {
		t.Fatalf("Create Team B failed: %v", err)
	}

	w := executeWithAuth(t, handler.ListMyTeams, http.MethodGet, "/api/v1/teams",
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(items))
	}
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Detail Team", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.GetTeam, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d", team.ID), "", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	teamData, ok := resp["team"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'team' field in response")
	}
	if teamData["name"] != "Detail Team" {
		t.Errorf("expected name 'Detail Team', got %v", teamData["name"])
	}

	members, ok := resp["members"].([]interface{})
	if !ok {
		t.Fatal("expected 'members' array in response")
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}

func TestTeamHandler_GetTeam_NonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	otherUser, _, err := userRepo.FindOrCreateByDiscordID("other-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create other user: %v", err)
	}

	team, err := teamRepo.Create("Private Team", otherUser.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.GetTeam, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d", team.ID), "", user, token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTeamHandler_GetTeam_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	w := executeWithAuth(t, handler.GetTeam, http.MethodGet,
		"/api/v1/teams/99999", "", user, token)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTeamHandler_UpdateTeam_ByCaptain_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Original", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.UpdateTeam, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{"name": "Updated Name"}`, user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["name"] != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %v", resp["name"])
	}
}

func TestTeamHandler_UpdateTeam_ByNonCaptain_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _ := setupTeamHandler(t)

	captain, _, err := userRepo.FindOrCreateByDiscordID("update-captain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}
	member, _, err := userRepo.FindOrCreateByDiscordID("update-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("Captain Only", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, member.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	memberToken, _, err := jwtService.GenerateAccessToken(member.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := executeWithAuth(t, handler.UpdateTeam, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{"name": "Hacked Name"}`, member, memberToken)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTeamHandler_DeleteTeam_ByCaptain_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Ma Team", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.DeleteTeam, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{"confirm_with_name": "Ma Team"}`, user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] != "team deleted" {
		t.Errorf("expected message 'team deleted', got '%v'", resp["message"])
	}
}

func TestTeamHandler_DeleteTeam_WrongName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Ma Team", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.DeleteTeam, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{"confirm_with_name": "Wrong Name"}`, user, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "confirmation_name_mismatch" {
		t.Errorf("expected error 'confirmation_name_mismatch', got '%v'", resp["error"])
	}
}

func TestTeamHandler_DeleteTeam_NoName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Ma Team", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.DeleteTeam, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{}`, user, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTeamHandler_LeaveTeam_Member_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _ := setupTeamHandler(t)

	captain, _, err := userRepo.FindOrCreateByDiscordID("leave-captain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}
	memberUser, _, err := userRepo.FindOrCreateByDiscordID("leave-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("Leave Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, memberUser.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	memberToken, _, err := jwtService.GenerateAccessToken(memberUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := executeWithAuth(t, handler.LeaveTeam, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/leave", team.ID),
		"", memberUser, memberToken)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] != "left team" {
		t.Errorf("expected message 'left team', got '%v'", resp["message"])
	}

	members, err := memberRepo.FindByTeamID(team.ID)
	if err != nil {
		t.Fatalf("FindByTeamID failed: %v", err)
	}
	for _, m := range members {
		if m.UserID == memberUser.ID {
			t.Errorf("expected member to be removed from team")
		}
	}
}

func TestTeamHandler_LeaveTeam_Captain_DeletesTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Captain Leave", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.LeaveTeam, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/leave", team.ID),
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] != "team deleted (captain left without transfer)" {
		t.Errorf("expected 'team deleted (captain left without transfer)', got '%v'", resp["message"])
	}

	_, err = teamRepo.FindByID(team.ID)
	if err != models.ErrTeamNotFound {
		t.Errorf("expected team to be deleted, got %v", err)
	}
}

func TestTeamHandler_DeleteTeam_ByNonCaptain_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, memberRepo, _ := setupTeamHandler(t)

	captain, _, err := userRepo.FindOrCreateByDiscordID("delcaptain-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}
	memberUser, _, err := userRepo.FindOrCreateByDiscordID("delmember-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("Delete By NonCaptain", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, memberUser.ID, models.RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	memberToken, _, err := jwtService.GenerateAccessToken(memberUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := executeWithAuth(t, handler.DeleteTeam, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{"confirm_with_name": "Delete By NonCaptain"}`, memberUser, memberToken)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTeamHandler_ListMyTeams_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, _, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	w := executeWithAuth(t, handler.ListMyTeams, http.MethodGet, "/api/v1/teams",
		"", user, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

func TestTeamHandler_UpdateTeam_NoName_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, userRepo, teamRepo, _, _ := setupTeamHandler(t)
	user, token := createAuthUser(t, userRepo)

	team, err := teamRepo.Create("Update No Name", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w := executeWithAuth(t, handler.UpdateTeam, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d", team.ID),
		`{}`, user, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
