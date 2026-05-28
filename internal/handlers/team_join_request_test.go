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

func setupTeamJoinRequestTestDB(t *testing.T) *sql.DB {
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
		"000007_add_team_invitations_and_join_requests.up.sql",
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

type teamJoinRequestTestFixtures struct {
	db                 *sql.DB
	joinRequestHandler *TeamJoinRequestHandler
	userRepo           models.UserRepository
	teamRepo           models.TeamRepository
	teamMemberRepo     models.TeamMemberRepository
	joinRequestRepo    models.TeamJoinRequestRepository
	captain            *models.User
	member             *models.User
	requester          *models.User
	team               *models.Team
	captainToken       string
	memberToken        string
	requesterToken     string
}

func setupTeamJoinRequestTest(t *testing.T) *teamJoinRequestTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupTeamJoinRequestTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	joinRequestRepo := models.NewTeamJoinRequestRepository(db, teamMemberRepo)
	testLogger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	captain, _, err := userRepo.FindOrCreateByDiscordID("jreq-cap-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("jreq-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("jreq-req-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("JoinReq Test Team "+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, member.ID, models.RoleMember); err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	captainToken, _, err := jwtService.GenerateAccessToken(captain.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	memberToken, _, err := jwtService.GenerateAccessToken(member.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	requesterToken, _, err := jwtService.GenerateAccessToken(requester.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	joinRequestHandler := NewTeamJoinRequestHandler(joinRequestRepo, teamMemberRepo, teamRepo, userRepo, testLogger)

	return &teamJoinRequestTestFixtures{
		db:                 db,
		joinRequestHandler: joinRequestHandler,
		userRepo:           userRepo,
		teamRepo:           teamRepo,
		teamMemberRepo:     teamMemberRepo,
		joinRequestRepo:    joinRequestRepo,
		captain:            captain,
		member:             member,
		requester:          requester,
		team:               team,
		captainToken:       captainToken,
		memberToken:        memberToken,
		requesterToken:     requesterToken,
	}
}

func executeJoinRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
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
	c.Params = params

	handler(c)

	return w
}

func TestTeamJoinRequestHandler_RequestJoin_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.RequestJoin, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/join-requests", f.team.ID),
		"", f.requester, f.requesterToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] == nil {
		t.Error("expected id in response")
	}
	if resp["status"] != "pending" {
		t.Errorf("expected status 'pending', got '%v'", resp["status"])
	}
}

func TestTeamJoinRequestHandler_RequestJoin_AlreadyMember(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.RequestJoin, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/join-requests", f.team.ID),
		"", f.member, f.memberToken, params)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for already member, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamJoinRequestHandler_RequestJoin_Duplicate(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	_, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create initial request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.RequestJoin, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/join-requests", f.team.ID),
		"", f.requester, f.requesterToken, params)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for duplicate, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamJoinRequestHandler_ListPendingRequests_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	_, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create join request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.ListPendingRequests, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/pending", f.team.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	requests, ok := resp["requests"].([]interface{})
	if !ok {
		t.Fatal("expected 'requests' array")
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestTeamJoinRequestHandler_AcceptRequest_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	jr, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create join request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "requestId", Value: fmt.Sprintf("%d", jr.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.AcceptRequest, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/%d/accept", f.team.ID, jr.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	member, err := f.teamMemberRepo.FindByUserAndTeam(f.requester.ID, f.team.ID)
	if err != nil {
		t.Fatalf("requester should now be a member: %v", err)
	}
	if member == nil {
		t.Fatal("expected non-nil membership after accept")
	}
}

func TestTeamJoinRequestHandler_DeclineRequest_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	jr, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create join request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "requestId", Value: fmt.Sprintf("%d", jr.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.DeclineRequest, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/%d/decline", f.team.ID, jr.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamJoinRequestHandler_GetMyRequest_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	_, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create join request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.GetMyRequest, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/mine", f.team.ID),
		"", f.requester, f.requesterToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["status"] != "pending" {
		t.Errorf("expected status 'pending', got '%v'", resp["status"])
	}
}

func TestTeamJoinRequestHandler_GetMyRequest_NotFound(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.GetMyRequest, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/mine", f.team.ID),
		"", f.requester, f.requesterToken, params)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamJoinRequestHandler_CancelMyRequest_Success(t *testing.T) {
	f := setupTeamJoinRequestTest(t)

	_, err := f.joinRequestRepo.Create(f.team.ID, f.requester.ID)
	if err != nil {
		t.Fatalf("failed to create join request: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeJoinRequest(t, f.joinRequestHandler.CancelMyRequest, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/join-requests/mine", f.team.ID),
		"", f.requester, f.requesterToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
