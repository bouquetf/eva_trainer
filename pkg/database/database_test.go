//go:build integration

package database

import (
	"os"
	"testing"
)

func TestConnect_ValidURL(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	cfg := Config{DatabaseURL: url, Timeout: defaultTimeout}
	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connect() returned error: %v", err)
	}
	defer db.Close()

	if err := HealthCheck(db); err != nil {
		t.Errorf("HealthCheck() returned error: %v", err)
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	cfg := Config{
		DatabaseURL: "postgres://invalid:invalid@localhost:9999/nonexistent",
		Timeout:     defaultTimeout,
	}

	_, err := Connect(cfg)
	if err == nil {
		t.Fatal("Connect() expected error, got nil")
	}
}
