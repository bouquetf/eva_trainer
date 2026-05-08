package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(key, value string) func() {
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if old == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, old)
		}
	}
}

func unsetEnv(key string) func() {
	old := os.Getenv(key)
	os.Unsetenv(key)
	return func() {
		if old != "" {
			os.Setenv(key, old)
		}
	}
}

func TestGetEnv(t *testing.T) {
	defer setEnv("MY_VAR", "value")()
	if got := getEnv("MY_VAR", "default"); got != "value" {
		t.Errorf("expected 'value', got '%s'", got)
	}
	os.Unsetenv("MY_VAR")
	if got := getEnv("MY_VAR", "default"); got != "default" {
		t.Errorf("expected 'default', got '%s'", got)
	}
}

func TestGetDuration(t *testing.T) {
	defer setEnv("DUR", "5m")()
	if got := getDuration("DUR", time.Hour); got != 5*time.Minute {
		t.Errorf("expected 5m, got %v", got)
	}
	os.Unsetenv("DUR")
	if got := getDuration("DUR", time.Hour); got != time.Hour {
		t.Errorf("expected default 1h, got %v", got)
	}
}

func TestGetOrigins(t *testing.T) {
	defer setEnv("ORIGINS", "http://a.com,http://b.com")()
	got := getOrigins("ORIGINS")
	if len(got) != 2 || got[0] != "http://a.com" || got[1] != "http://b.com" {
		t.Errorf("expected 2 origins, got %v", got)
	}
	os.Setenv("ORIGINS", "*")
	got = getOrigins("ORIGINS")
	if len(got) != 1 || got[0] != "*" {
		t.Errorf("expected ['*'], got %v", got)
	}
	os.Unsetenv("ORIGINS")
	got = getOrigins("ORIGINS")
	if len(got) != 1 || got[0] != "*" {
		t.Errorf("expected default ['*'], got %v", got)
	}
}

func TestValidate(t *testing.T) {
	cfg := &AppConfig{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty config")
	}
	for _, field := range []string{"JWT_SECRET", "DISCORD_CLIENT_ID", "DISCORD_CLIENT_SECRET"} {
		if !contains(err.Error(), field) {
			t.Errorf("error should mention %s, got: %s", field, err.Error())
		}
	}

	cfg = &AppConfig{
		JWT:     JWTConfig{Secret: "s"},
		Discord: DiscordConfig{ClientID: "i", ClientSecret: "s"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDefaults(t *testing.T) {
	defer unsetEnv("PORT")()
	defer unsetEnv("GIN_MODE")()
	defer unsetEnv("JWT_EXPIRY")()
	defer unsetEnv("REFRESH_TOKEN_EXPIRY")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected default Port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.GinMode != "debug" {
		t.Errorf("expected default GinMode 'debug', got %s", cfg.Server.GinMode)
	}
	if cfg.JWT.Expiry != 24*time.Hour {
		t.Errorf("expected JWT expiry 24h, got %v", cfg.JWT.Expiry)
	}
	if cfg.JWT.RefreshTokenExpiry != 7*24*time.Hour {
		t.Errorf("expected refresh token expiry 7d, got %v", cfg.JWT.RefreshTokenExpiry)
	}
}

func TestLoad_WithSystemEnv(t *testing.T) {
	defer setEnv("PORT", "5000")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != "5000" {
		t.Errorf("expected Port 5000, got %s", cfg.Server.Port)
	}
}

func TestLoad_SystemEnvOverridesDotEnv(t *testing.T) {
	defer setEnv("PORT", "4000")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	os.WriteFile(".env", []byte("PORT=3000"), 0644)
	defer os.Remove(".env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != "4000" {
		t.Errorf("expected Port 4000 (system override), got %s", cfg.Server.Port)
	}
}

func TestLoad_DotEnvWithoutSystem(t *testing.T) {
	defer unsetEnv("PORT")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	os.WriteFile(".env", []byte("PORT=3000"), 0644)
	defer os.Remove(".env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != "3000" {
		t.Errorf("expected Port 3000 from .env, got %s", cfg.Server.Port)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	defer unsetEnv("JWT_SECRET")()
	defer unsetEnv("DISCORD_CLIENT_ID")()
	defer unsetEnv("DISCORD_CLIENT_SECRET")()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestLoad_DatabaseTimeout(t *testing.T) {
	defer setEnv("DB_TIMEOUT", "15")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Database.Timeout != 15*time.Second {
		t.Errorf("expected DB timeout 15s, got %v", cfg.Database.Timeout)
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	defer setEnv("CORS_ORIGINS", "http://a.com,http://b.com")()
	defer setEnv("JWT_SECRET", "test")()
	defer setEnv("DISCORD_CLIENT_ID", "id")()
	defer setEnv("DISCORD_CLIENT_SECRET", "secret")()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if len(cfg.CORS.Origins) != 2 || cfg.CORS.Origins[0] != "http://a.com" || cfg.CORS.Origins[1] != "http://b.com" {
		t.Errorf("expected 2 origins, got %v", cfg.CORS.Origins)
	}
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
