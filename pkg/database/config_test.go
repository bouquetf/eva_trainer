package database

import (
	"os"
	"testing"
)

func TestLoadConfigFromEnv_UsesDatabaseURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db")
	defer os.Unsetenv("DATABASE_URL")

	cfg := LoadConfigFromEnv()

	if cfg.DatabaseURL != "postgres://user:pass@host:5432/db" {
		t.Errorf("expected DatabaseURL 'postgres://user:pass@host:5432/db', got '%s'", cfg.DatabaseURL)
	}
}

func TestLoadConfigFromEnv_UsesIndividualVars(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_TIMEOUT", "15")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_TIMEOUT")
	}()

	cfg := LoadConfigFromEnv()

	if cfg.Host != "myhost" {
		t.Errorf("expected Host 'myhost', got '%s'", cfg.Host)
	}
	if cfg.Port != "5433" {
		t.Errorf("expected Port '5433', got '%s'", cfg.Port)
	}
	if cfg.User != "admin" {
		t.Errorf("expected User 'admin', got '%s'", cfg.User)
	}
	if cfg.Password != "secret" {
		t.Errorf("expected Password 'secret', got '%s'", cfg.Password)
	}
	if cfg.DBName != "mydb" {
		t.Errorf("expected DBName 'mydb', got '%s'", cfg.DBName)
	}
	if cfg.SSLMode != "require" {
		t.Errorf("expected SSLMode 'require', got '%s'", cfg.SSLMode)
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("DB_TIMEOUT")

	cfg := LoadConfigFromEnv()

	if cfg.Host != "localhost" {
		t.Errorf("expected default Host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != "5432" {
		t.Errorf("expected default Port '5432', got '%s'", cfg.Port)
	}
	if cfg.DBName != "eva_trainer" {
		t.Errorf("expected default DBName 'eva_trainer', got '%s'", cfg.DBName)
	}
	if cfg.Timeout.String() != "10s" {
		t.Errorf("expected default Timeout '10s', got '%s'", cfg.Timeout)
	}
}
