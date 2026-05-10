package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

type mockTeamMemberRepoPerm struct {
	member *models.TeamMember
	err    error
}

func (m *mockTeamMemberRepoPerm) FindByUserAndTeam(userID, teamID int) (*models.TeamMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.member, nil
}

func (m *mockTeamMemberRepoPerm) AddMember(teamID, userID int, role string) error {
	return nil
}

func (m *mockTeamMemberRepoPerm) RemoveMember(teamID, userID int) error {
	return nil
}

func (m *mockTeamMemberRepoPerm) UpdateRole(teamID, userID int, role string) error {
	return nil
}

func (m *mockTeamMemberRepoPerm) FindByTeamID(teamID int) ([]*models.TeamMember, error) {
	return nil, nil
}

func (m *mockTeamMemberRepoPerm) FindByTeamIDWithDisplayNames(teamID int) ([]*models.TeamMemberWithDisplayName, error) {
	return nil, nil
}

func (m *mockTeamMemberRepoPerm) FindByUserID(userID int) ([]*models.TeamMember, error) {
	return nil, nil
}

func (m *mockTeamMemberRepoPerm) GetCaptainID(teamID int) (int, error) {
	return 0, nil
}

func (m *mockTeamMemberRepoPerm) CountByRole(teamID int, role string) (int, error) {
	return 0, nil
}

type mockRolePermRepoPerm struct {
	hasPermission bool
	err           error
	perms         []*models.Permission
}

func (m *mockRolePermRepoPerm) FindByRole(role string) ([]*models.Permission, error) {
	return m.perms, nil
}

func (m *mockRolePermRepoPerm) FindByPermissionID(permissionID int) ([]*models.RolePermission, error) {
	return nil, nil
}

func (m *mockRolePermRepoPerm) AssignPermissionToRole(role string, permissionID int) error {
	return nil
}

func (m *mockRolePermRepoPerm) RemovePermissionFromRole(role string, permissionID int) error {
	return nil
}

func (m *mockRolePermRepoPerm) FindAllRoles() ([]string, error) {
	return nil, nil
}

func (m *mockRolePermRepoPerm) RoleHasPermission(role string, permissionID int) (bool, error) {
	return m.hasPermission, m.err
}

type mockPermRepoPerm struct {
	perm *models.Permission
	err  error
}

func (m *mockPermRepoPerm) FindAll() ([]*models.Permission, error) {
	return nil, nil
}

func (m *mockPermRepoPerm) FindByID(id int) (*models.Permission, error) {
	return m.perm, m.err
}

func (m *mockPermRepoPerm) FindByFeature(feature string) (*models.Permission, error) {
	return m.perm, m.err
}

func (m *mockPermRepoPerm) Create(feature, description string) (*models.Permission, error) {
	return nil, nil
}

func setupPermissionMiddlewareTest(t *testing.T, role string, isAdmin bool, hasPermission bool) (*gin.Engine, *mockTeamMemberRepoPerm) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	member := &models.TeamMember{
		TeamID: 1,
		UserID: 42,
		Role:   role,
	}

	memberRepo := &mockTeamMemberRepoPerm{
		member: member,
	}

	rolePermRepo := &mockRolePermRepoPerm{
		hasPermission: hasPermission,
		perms: []*models.Permission{
			{ID: 1, Feature: "team:delete", Description: "Delete team"},
		},
	}

	permRepo := &mockPermRepoPerm{
		perm: &models.Permission{ID: 1, Feature: "team:delete"},
	}

	r := gin.New()
	r.GET("/api/v1/teams/:id/protected",
		func(c *gin.Context) {
			user := &models.User{
				ID:        42,
				DiscordID: "test",
				IsAdmin:   isAdmin,
			}
			c.Set("user", user)
			c.Set("team_member", member)
			c.Next()
		},
		RequirePermission(nil, memberRepo, rolePermRepo, permRepo, "team:delete"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	return r, memberRepo
}

func TestRequirePermission_AdminBypass(t *testing.T) {
	r, _ := setupPermissionMiddlewareTest(t, "member", true, false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true for admin bypass")
	}
}

func TestRequirePermission_MemberWithPermission(t *testing.T) {
	r, _ := setupPermissionMiddlewareTest(t, "captain", false, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_MemberWithoutPermission(t *testing.T) {
	r, _ := setupPermissionMiddlewareTest(t, "member", false, false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected 'forbidden', got '%s'", resp["error"])
	}
}

func TestRequirePermission_NonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memberRepo := &mockTeamMemberRepoPerm{
		member: nil,
		err:    models.ErrMemberNotFound,
	}

	rolePermRepo := &mockRolePermRepoPerm{}
	permRepo := &mockPermRepoPerm{}

	r := gin.New()
	r.GET("/api/v1/teams/:id/protected",
		func(c *gin.Context) {
			user := &models.User{
				ID:        99,
				DiscordID: "non-member",
				IsAdmin:   false,
			}
			c.Set("user", user)
			c.Next()
		},
		RequirePermission(nil, memberRepo, rolePermRepo, permRepo, "team:delete"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected 'forbidden', got '%s'", resp["error"])
	}
}

func TestRequireTeamRole_BlocksBeforePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	member := &models.TeamMember{
		TeamID: 1,
		UserID: 42,
		Role:   "editor",
	}

	memberRepo := &mockTeamMemberRepoPerm{
		member: member,
	}

	r := gin.New()
	r.GET("/api/v1/teams/:id/protected",
		func(c *gin.Context) {
			user := &models.User{
				ID:        42,
				DiscordID: "test",
				IsAdmin:   false,
			}
			c.Set("user", user)
			c.Set("team_member", member)
			c.Next()
		},
		RequireTeamRole(nil, memberRepo, "captain"),
		func(c *gin.Context) {
			t.Error("handler should not be reached")
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected 'forbidden', got '%s'", resp["error"])
	}
}
