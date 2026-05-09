package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eva_trainer/internal/models"

	"github.com/gin-gonic/gin"
)

func setupRequireAdminTest(isAdmin bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/v1/admin/test",
		func(c *gin.Context) {
			user := &models.User{
				ID:        1,
				DiscordID: "test",
				IsAdmin:   isAdmin,
			}
			c.Set("user", user)
			c.Next()
		},
		RequireAdmin(),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	return r
}

func TestRequireAdmin_AdminUser_Returns200(t *testing.T) {
	r := setupRequireAdminTest(true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
}

func TestRequireAdmin_NonAdminUser_Returns403(t *testing.T) {
	r := setupRequireAdminTest(false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got '%v'", resp["error"])
	}
	if resp["message"] != "admin access required" {
		t.Errorf("expected message 'admin access required', got '%v'", resp["message"])
	}
}

func TestRequireAdmin_NoUser_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/v1/admin/test", RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got '%v'", resp["error"])
	}
}
