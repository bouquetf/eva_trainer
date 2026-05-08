package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"eva_trainer/pkg/config"
	"eva_trainer/pkg/database"

	"github.com/gin-gonic/gin"
)

func setupTestConfig() *config.AppConfig {
	return &config.AppConfig{
		Server: config.ServerConfig{
			Port:        "8080",
			GinMode:     "debug",
			FrontendURL: "http://localhost:5173",
		},
		Discord: config.DiscordConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/api/v1/auth/discord/callback",
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-that-is-at-least-32-chars-long-for-testing",
			Expiry:             24 * time.Hour,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
		},
		CORS: config.CORSConfig{
			Origins: []string{"*"},
		},
		Database: database.Config{
			DatabaseURL: "postgres://localhost:5432/test?sslmode=disable",
			Timeout:     5 * time.Second,
		},
	}
}

func setupTestFS(t *testing.T) (http.FileSystem, http.FileSystem) {
	t.Helper()

	rootFS := fstest.MapFS{
		"index.html":     {Data: []byte("<html></html>")},
		"assets/test.js": {Data: []byte("console.log")},
	}

	assetsSub, err := fs.Sub(rootFS, "assets")
	if err != nil {
		t.Fatal(err)
	}

	return http.FS(rootFS), http.FS(assetsSub)
}

func TestCORS_DevMode_AllowsAllOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	cfg.CORS.Origins = []string{"*"}
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_ProdMode_Restricted(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	cfg := setupTestConfig()
	cfg.CORS.Origins = []string{"http://localhost:8080"}
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://malicious.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin for non-matching origin in prod, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	cfg.CORS.Origins = []string{"*"}
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header on preflight")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header on preflight")
	}
}

func TestTeamRoute_List_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTeamRoute_Create_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/teams", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTeamRoute_GetByID_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTeamRoute_Members_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams/1/members", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTeamRoute_UpdateMemberRole_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/teams/1/members/2/role", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTeamRoute_TransferCaptaincy_WithoutAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/teams/1/transfer-captaincy", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCORS_RouteAPIIntact(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp["status"])
	}
}

func TestProd_ServesIndexHTML(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	cfg := setupTestConfig()
	frontendFS, assetsFS := setupTestFS(t)
	r := gin.New()
	SetupRoutes(r, cfg, frontendFS, assetsFS)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "<html></html>" {
		t.Errorf("expected index.html content, got %s", w.Body.String())
	}
}

func TestProd_ServesStaticAsset(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	cfg := setupTestConfig()
	frontendFS, assetsFS := setupTestFS(t)
	r := gin.New()
	SetupRoutes(r, cfg, frontendFS, assetsFS)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/test.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "console.log" {
		t.Errorf("expected asset content, got %s", w.Body.String())
	}
}

func TestProd_ServesSPAFallback(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	cfg := setupTestConfig()
	frontendFS, assetsFS := setupTestFS(t)
	r := gin.New()
	SetupRoutes(r, cfg, frontendFS, assetsFS)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/unknown-route", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "<html></html>" {
		t.Errorf("expected index.html content for SPA fallback, got %s", w.Body.String())
	}
}

func TestDev_Returns404ForRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.DebugMode)

	cfg := setupTestConfig()
	r := gin.New()
	SetupRoutes(r, cfg, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 in dev mode, got %d", w.Code)
	}
}
