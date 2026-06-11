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

func setupTeamInvitationTestDB(t *testing.T) *sql.DB {
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

type teamInvitationTestFixtures struct {
	db                *sql.DB
	invitationHandler *TeamInvitationHandler
	userRepo          models.UserRepository
	teamRepo          models.TeamRepository
	teamMemberRepo    models.TeamMemberRepository
	invitationRepo    models.TeamInvitationRepository
	captain           *models.User
	member            *models.User
	otherUser         *models.User
	team              *models.Team
	captainToken      string
	memberToken       string
	otherToken        string
}

func setupTeamInvitationTest(t *testing.T) *teamInvitationTestFixtures {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupTeamInvitationTestDB(t)

	userRepo := models.NewUserRepository(db)
	teamRepo := models.NewTeamRepository(db)
	teamMemberRepo := models.NewTeamMemberRepository(db)
	invitationRepo := models.NewTeamInvitationRepository(db, teamMemberRepo)
	testLogger := services.NewLogger(services.LogLevelError, services.LogFormatText)

	captain, _, err := userRepo.FindOrCreateByDiscordID("invite-cap-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("invite-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	otherUser, _, err := userRepo.FindOrCreateByDiscordID("invite-other-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create other user: %v", err)
	}
	// Set a display name so invitation by display_name works
	_, err = db.Exec("UPDATE users SET display_name = $1 WHERE id = $2", "DisplayName-"+t.Name(), otherUser.ID)
	if err != nil {
		t.Fatalf("failed to set display name: %v", err)
	}
	otherUser.DisplayName = "DisplayName-" + t.Name()

	team, err := teamRepo.Create("Invite Test Team "+t.Name(), captain.ID)
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
	otherToken, _, err := jwtService.GenerateAccessToken(otherUser.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	invitationHandler := NewTeamInvitationHandler(invitationRepo, teamMemberRepo, userRepo, teamRepo, "http://localhost:3000", testLogger)

	return &teamInvitationTestFixtures{
		db:                db,
		invitationHandler: invitationHandler,
		userRepo:          userRepo,
		teamRepo:          teamRepo,
		teamMemberRepo:    teamMemberRepo,
		invitationRepo:    invitationRepo,
		captain:           captain,
		member:            member,
		otherUser:         otherUser,
		team:              team,
		captainToken:      captainToken,
		memberToken:       memberToken,
		otherToken:        otherToken,
	}
}

func executeInvitationRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body string, user *models.User, token string, params gin.Params) *httptest.ResponseRecorder {
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

func TestTeamInvitationHandler_InviteByDisplayName_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	body := fmt.Sprintf(`{"display_name":"%s"}`, f.otherUser.DisplayName)
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.InviteByDisplayName, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/invitations", f.team.ID),
		body, f.captain, f.captainToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["invited_user_id"] == nil {
		t.Error("expected invited_user_id in response")
	}
}

func TestTeamInvitationHandler_InviteByDisplayName_UserNotFound(t *testing.T) {
	f := setupTeamInvitationTest(t)

	body := `{"display_name":"nonexistent-user-xxxx"}`
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.InviteByDisplayName, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/invitations", f.team.ID),
		body, f.captain, f.captainToken, params)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_InviteByDisplayName_NonCaptain(t *testing.T) {
	f := setupTeamInvitationTest(t)

	body := fmt.Sprintf(`{"display_name":"%s"}`, f.otherUser.DisplayName)
	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.InviteByDisplayName, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/invitations", f.team.ID),
		body, f.member, f.memberToken, params)

	if w.Code == http.StatusCreated || w.Code == http.StatusOK {
		t.Errorf("expected non-2xx for non-captain, got %d", w.Code)
	}
}

func TestTeamInvitationHandler_GenerateInviteLink_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.GenerateInviteLink, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/invite-links", f.team.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token")
	}
	if resp["invite_url"] == nil {
		t.Error("expected invite_url in response")
	}
}

func TestTeamInvitationHandler_AcceptInvitation_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	otherUser2, _, err := f.userRepo.FindOrCreateByDiscordID("invite-accept-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	otherToken2, _, err := jwtService.GenerateAccessToken(otherUser2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, otherUser2.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.AcceptInvitation, http.MethodPost,
		fmt.Sprintf("/api/v1/invitations/%d/accept", inv.ID),
		"", otherUser2, otherToken2, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_AcceptInvitation_NotForYou(t *testing.T) {
	f := setupTeamInvitationTest(t)

	otherUser2, _, err := f.userRepo.FindOrCreateByDiscordID("invite-notforu-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, otherUser2.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.AcceptInvitation, http.MethodPost,
		fmt.Sprintf("/api/v1/invitations/%d/accept", inv.ID),
		"", f.otherUser, f.otherToken, params)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_DeclineInvitation_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	otherUser2, _, err := f.userRepo.FindOrCreateByDiscordID("invite-decline-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	otherToken2, _, err := jwtService.GenerateAccessToken(otherUser2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, otherUser2.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.DeclineInvitation, http.MethodPost,
		fmt.Sprintf("/api/v1/invitations/%d/decline", inv.ID),
		"", otherUser2, otherToken2, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_AcceptInviteLink_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	token, err := models.GenerateInviteToken()
	if err != nil {
		t.Fatalf("GenerateInviteToken failed: %v", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = f.invitationRepo.CreateLink(f.team.ID, f.captain.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateLink failed: %v", err)
	}

	params := gin.Params{
		{Key: "token", Value: token},
	}
	w := executeInvitationRequest(t, f.invitationHandler.AcceptInviteLink, http.MethodPost,
		fmt.Sprintf("/api/v1/invite-links/%s/accept", token),
		"", f.otherUser, f.otherToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_AcceptInviteLink_Expired(t *testing.T) {
	f := setupTeamInvitationTest(t)

	token, err := models.GenerateInviteToken()
	if err != nil {
		t.Fatalf("GenerateInviteToken failed: %v", err)
	}
	expiresAt := time.Now().Add(-1 * time.Hour)
	_, err = f.invitationRepo.CreateLink(f.team.ID, f.captain.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateLink failed: %v", err)
	}

	params := gin.Params{
		{Key: "token", Value: token},
	}
	w := executeInvitationRequest(t, f.invitationHandler.AcceptInviteLink, http.MethodPost,
		fmt.Sprintf("/api/v1/invite-links/%s/accept", token),
		"", f.otherUser, f.otherToken, params)

	if w.Code != http.StatusGone {
		t.Errorf("expected 410, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_ListPendingInvitations_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	otherUser2, _, err := f.userRepo.FindOrCreateByDiscordID("invite-pending-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err = f.invitationRepo.CreateByUser(f.team.ID, otherUser2.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.ListPendingByTeam, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/invitations/pending", f.team.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	invitations, ok := resp["invitations"].([]interface{})
	if !ok {
		t.Fatal("expected 'invitations' array")
	}
	if len(invitations) != 1 {
		t.Fatalf("expected 1 invitation, got %d", len(invitations))
	}
}

func TestTeamInvitationHandler_ListMyInvitations_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	otherUser2, _, err := f.userRepo.FindOrCreateByDiscordID("invite-my-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	otherToken2, _, err := jwtService.GenerateAccessToken(otherUser2.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	_, err = f.invitationRepo.CreateByUser(f.team.ID, otherUser2.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{}
	w := executeInvitationRequest(t, f.invitationHandler.ListMyInvitations, http.MethodGet,
		"/api/v1/invitations",
		"", otherUser2, otherToken2, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 invitation, got %d", len(resp))
	}
}

func TestTeamInvitationHandler_CancelInvitation_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, f.otherUser.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.CancelInvitation, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/invitations/%d", f.team.ID, inv.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] != "invitation cancelled" {
		t.Errorf("expected message 'invitation cancelled', got %v", resp["message"])
	}

	cancelled, err := f.invitationRepo.FindByID(inv.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %s", cancelled.Status)
	}
}

func TestTeamInvitationHandler_CancelInvitation_ByCocaptain_Success(t *testing.T) {
	f := setupTeamInvitationTest(t)

	cocaptain, _, err := f.userRepo.FindOrCreateByDiscordID("cancel-coca-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create cocaptain: %v", err)
	}
	if err := f.teamMemberRepo.AddMember(f.team.ID, cocaptain.ID, models.RoleCocaptain); err != nil {
		t.Fatalf("failed to add cocaptain: %v", err)
	}
	jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	cocaptainToken, _, err := jwtService.GenerateAccessToken(cocaptain.ID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, f.otherUser.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.CancelInvitation, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/invitations/%d", f.team.ID, inv.ID),
		"", cocaptain, cocaptainToken, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_CancelInvitation_ForbiddenForMember(t *testing.T) {
	f := setupTeamInvitationTest(t)

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, f.otherUser.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.CancelInvitation, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/invitations/%d", f.team.ID, inv.ID),
		"", f.member, f.memberToken, params)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_CancelInvitation_InvitationNotFound(t *testing.T) {
	f := setupTeamInvitationTest(t)

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", f.team.ID)},
		{Key: "invitationId", Value: "99999"},
	}
	w := executeInvitationRequest(t, f.invitationHandler.CancelInvitation, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/invitations/99999", f.team.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTeamInvitationHandler_CancelInvitation_WrongTeam(t *testing.T) {
	f := setupTeamInvitationTest(t)

	inv, err := f.invitationRepo.CreateByUser(f.team.ID, f.otherUser.ID, f.captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	otherTeam, err := f.teamRepo.Create("Other Team "+t.Name(), f.captain.ID)
	if err != nil {
		t.Fatalf("failed to create other team: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", otherTeam.ID)},
		{Key: "invitationId", Value: fmt.Sprintf("%d", inv.ID)},
	}
	w := executeInvitationRequest(t, f.invitationHandler.CancelInvitation, http.MethodDelete,
		fmt.Sprintf("/api/v1/teams/%d/invitations/%d", otherTeam.ID, inv.ID),
		"", f.captain, f.captainToken, params)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}
