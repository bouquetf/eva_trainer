package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
)

const testSecret = "middleware-test-secret-that-is-long-enough-32b!!"

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) FindOrCreateByDiscordID(discordID string) (*models.User, bool, error) {
	return nil, false, nil
}

func (m *mockUserRepo) FindByID(userID int) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepo) SetAdmin(userID int) error {
	return nil
}

func (m *mockUserRepo) IsFirstUser() (bool, error) {
	return false, nil
}

func setupTest(t *testing.T) (*services.JWTService, *mockUserRepo, *gin.Engine) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	svc := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	repo := &mockUserRepo{
		user: &models.User{
			ID:        42,
			DiscordID: "123456789",
			IsAdmin:   false,
		},
	}

	r := gin.New()
	r.GET("/protected", RequireAuth(repo, svc), func(c *gin.Context) {
		user, _ := GetUser(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id":    user.ID,
			"discord_id": user.DiscordID,
		})
	})

	return svc, repo, r
}

func TestRequireAuth_ValidToken(t *testing.T) {
	svc, _, r := setupTest(t)

	token, _, err := svc.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["user_id"] != float64(42) {
		t.Errorf("expected user_id 42, got %v", resp["user_id"])
	}
	if resp["discord_id"] != "123456789" {
		t.Errorf("expected discord_id 123456789, got %v", resp["discord_id"])
	}
}

func TestRequireAuth_MissingAuthorization(t *testing.T) {
	_, _, r := setupTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "missing_authorization" {
		t.Errorf("expected error 'missing_authorization', got '%s'", resp["error"])
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	_, _, r := setupTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "invalid_authorization" {
		t.Errorf("expected error 'invalid_authorization', got '%s'", resp["error"])
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	svc := services.NewJWTService(testSecret, -1*time.Hour, 7*24*time.Hour)
	repo := &mockUserRepo{
		user: &models.User{ID: 42, DiscordID: "123456789"},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", RequireAuth(repo, svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, _, err := svc.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "token_expired" {
		t.Errorf("expected error 'token_expired', got '%s'", resp["error"])
	}
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	svc := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
	repo := &mockUserRepo{
		err: models.ErrUserNotFound,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", RequireAuth(repo, svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, _, err := svc.GenerateAccessToken(999)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetUser_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	user, ok := GetUser(c)
	if ok {
		t.Errorf("expected ok=false when user not set")
	}
	if user != nil {
		t.Errorf("expected nil user when not set")
	}
}

func TestGetUser_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	expected := &models.User{ID: 1, DiscordID: "123"}
	c.Set("user", expected)

	user, ok := GetUser(c)
	if !ok {
		t.Errorf("expected ok=true when user is set")
	}
	if user != expected {
		t.Errorf("expected same user pointer")
	}
}

func TestMustGetUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	expected := &models.User{ID: 1, DiscordID: "123"}
	c.Set("user", expected)

	user, err := MustGetUser(c)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user != expected {
		t.Errorf("expected same user pointer")
	}
}

func TestMustGetUser_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := MustGetUser(c)
	if err == nil {
		t.Fatal("expected error when user not set")
	}
}
