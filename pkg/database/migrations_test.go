//go:build integration

package database

import (
	"os"
	"testing"
)

func TestRunMigrations_OnEmptyDB(t *testing.T) {
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

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations() returned error: %v", err)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
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

	if err := RunMigrations(db); err != nil {
		t.Fatalf("First RunMigrations() returned error: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("Second RunMigrations() returned error: %v", err)
	}
}

func TestRunMigrations_DownThenUp(t *testing.T) {
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

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations (up) returned error: %v", err)
	}

	if err := RunMigrationsDown(db); err != nil {
		t.Fatalf("RunMigrationsDown returned error: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations (up after down) returned error: %v", err)
	}
}
