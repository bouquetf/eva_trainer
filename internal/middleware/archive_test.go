package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
)

type mockGlobalRoleRepo struct {
	hasRole bool
	err     error
	users   []*models.UserGlobalRole
}

func (m *mockGlobalRoleRepo) Add(userID int, role string) error {
	return nil
}

func (m *mockGlobalRoleRepo) Remove(userID int, role string) error {
	return nil
}

func (m *mockGlobalRoleRepo) UserHasRole(userID int, role string) (bool, error) {
	return m.hasRole, m.err
}

func (m *mockGlobalRoleRepo) FindUsersByRole(role string) ([]*models.UserGlobalRole, error) {
	return m.users, nil
}

type mockTeamRoleRepo struct {
	hasRole         bool
	err             error
	userHasTeamRole bool
	teams           []*models.TeamWithRole
}

func (m *mockTeamRoleRepo) Add(teamID int, role string) error {
	return nil
}

func (m *mockTeamRoleRepo) Remove(teamID int, role string) error {
	return nil
}

func (m *mockTeamRoleRepo) TeamHasRole(teamID int, role string) (bool, error) {
	return m.hasRole, m.err
}

func (m *mockTeamRoleRepo) FindTeamsByRole(role string) ([]*models.TeamWithRole, error) {
	return m.teams, nil
}

func (m *mockTeamRoleRepo) UserIsMemberOfTeamWithRole(userID int, role string) (bool, error) {
	return m.userHasTeamRole, m.err
}

func TestCheckArchivePermissions_AdminBypass(t *testing.T) {
	user := &models.User{ID: 1, DiscordID: "admin", IsAdmin: true}
	globalRepo := &mockGlobalRoleRepo{}
	teamRepo := &mockTeamRoleRepo{}

	perms, err := CheckArchivePermissions(user, globalRepo, teamRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !perms.CanRead {
		t.Errorf("expected CanRead=true for admin")
	}
	if !perms.CanWrite {
		t.Errorf("expected CanWrite=true for admin")
	}
}

func TestCheckArchivePermissions_Archivist(t *testing.T) {
	user := &models.User{ID: 2, DiscordID: "archivist", IsAdmin: false}
	globalRepo := &mockGlobalRoleRepo{hasRole: true}
	teamRepo := &mockTeamRoleRepo{}

	perms, err := CheckArchivePermissions(user, globalRepo, teamRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !perms.CanRead {
		t.Errorf("expected CanRead=true for archivist")
	}
	if !perms.CanWrite {
		t.Errorf("expected CanWrite=true for archivist")
	}
}

func TestCheckArchivePermissions_ReaderTeamMember(t *testing.T) {
	user := &models.User{ID: 3, DiscordID: "reader", IsAdmin: false}
	globalRepo := &mockGlobalRoleRepo{hasRole: false}
	teamRepo := &mockTeamRoleRepo{userHasTeamRole: true}

	perms, err := CheckArchivePermissions(user, globalRepo, teamRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !perms.CanRead {
		t.Errorf("expected CanRead=true for reader team member")
	}
	if perms.CanWrite {
		t.Errorf("expected CanWrite=false for reader team member")
	}
}

func TestCheckArchivePermissions_NoAccess(t *testing.T) {
	user := &models.User{ID: 4, DiscordID: "nobody", IsAdmin: false}
	globalRepo := &mockGlobalRoleRepo{hasRole: false}
	teamRepo := &mockTeamRoleRepo{userHasTeamRole: false}

	perms, err := CheckArchivePermissions(user, globalRepo, teamRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perms.CanRead {
		t.Errorf("expected CanRead=false for unauthorized user")
	}
	if perms.CanWrite {
		t.Errorf("expected CanWrite=false for unauthorized user")
	}
}

func TestCheckArchivePermissions_GlobalRoleError(t *testing.T) {
	user := &models.User{ID: 5, DiscordID: "error", IsAdmin: false}
	globalRepo := &mockGlobalRoleRepo{err: models.ErrRoleNotFound}
	teamRepo := &mockTeamRoleRepo{}

	_, err := CheckArchivePermissions(user, globalRepo, teamRepo)
	if err == nil {
		t.Fatal("expected error from global role check")
	}
}

func setupRequireArchiveAccessTest(t *testing.T, isAdmin bool, hasGlobalRole bool, hasTeamRole bool, accessType string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	globalRepo := &mockGlobalRoleRepo{hasRole: hasGlobalRole}
	teamRepo := &mockTeamRoleRepo{userHasTeamRole: hasTeamRole}

	r := gin.New()
	r.GET("/protected",
		func(c *gin.Context) {
			user := &models.User{
				ID:        42,
				DiscordID: "test",
				IsAdmin:   isAdmin,
			}
			c.Set("user", user)
			c.Next()
		},
		RequireArchiveAccess(services.NewLogger(services.LogLevelError, services.LogFormatText), globalRepo, teamRepo, accessType),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	return w
}

func TestRequireArchiveAccess_Admin_ReadAccess(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, true, false, false, "read")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin read, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_Admin_WriteAccess(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, true, false, false, "write")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin write, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_Archivist_ReadAccess(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, true, false, "read")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for archivist read, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_Archivist_WriteAccess(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, true, false, "write")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for archivist write, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_ReaderTeamMember_ReadAccess(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, false, true, "read")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for reader team member read, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_ReaderTeamMember_WriteAccess_Forbidden(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, false, true, "write")
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for reader team member write, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%s'", resp["error"])
	}
}

func TestRequireArchiveAccess_NoAccess_Read_Forbidden(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, false, false, "read")
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unauthorized read, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%s'", resp["error"])
	}
}

func TestRequireArchiveAccess_NoAccess_Write_Forbidden(t *testing.T) {
	w := setupRequireArchiveAccessTest(t, false, false, false, "write")
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unauthorized write, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequireArchiveAccess_NoUser_ReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	globalRepo := &mockGlobalRoleRepo{}
	teamRepo := &mockTeamRoleRepo{}

	r := gin.New()
	r.GET("/protected",
		func(c *gin.Context) {
			c.Set("user", nil)
			c.Next()
		},
		RequireArchiveAccess(services.NewLogger(services.LogLevelError, services.LogFormatText), globalRepo, teamRepo, "read"),
		func(c *gin.Context) {
			t.Error("handler should not be reached when no user")
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d. Body: %s", w.Code, w.Body.String())
	}
}
