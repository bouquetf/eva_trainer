//go:build integration

package seed

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupSeedTestDB(t *testing.T) *sql.DB {
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
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
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
		"000006_drop_current_version_id.up.sql",
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

func createTempMapDir(t *testing.T, files map[string]string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "mapseed-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	return dir
}

func TestSeedMaps_CreatesMaps(t *testing.T) {
	db := setupSeedTestDB(t)

	mapFiles := map[string]string{
		"atlantis-map.webp":      "atlantis-blob",
		"lunar-outpost-map.webp": "lunar-blob",
	}
	tmpDir := createTempMapDir(t, mapFiles)

	if err := SeedMaps(db, tmpDir); err != nil {
		t.Fatalf("SeedMaps returned error: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM maps`).Scan(&count); err != nil {
		t.Fatalf("failed to count maps: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 maps, got %d", count)
	}
}

func TestSeedMaps_Idempotent(t *testing.T) {
	db := setupSeedTestDB(t)

	mapFiles := map[string]string{
		"atlantis-map.webp": "atlantis-blob",
		"polaris-map.webp":  "polaris-blob",
	}
	tmpDir := createTempMapDir(t, mapFiles)

	if err := SeedMaps(db, tmpDir); err != nil {
		t.Fatalf("first SeedMaps failed: %v", err)
	}

	if err := SeedMaps(db, tmpDir); err != nil {
		t.Fatalf("second SeedMaps failed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM maps`).Scan(&count); err != nil {
		t.Fatalf("failed to count maps: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 maps (idempotent), got %d", count)
	}
}

func TestSeedMaps_EmptyDir(t *testing.T) {
	db := setupSeedTestDB(t)

	tmpDir := createTempMapDir(t, map[string]string{})

	if err := SeedMaps(db, tmpDir); err != nil {
		t.Fatalf("SeedMaps with empty dir returned error: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM maps`).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 maps for empty dir, got %d", count)
	}
}

func TestExtractMapName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"atlantis-map.webp", "Atlantis"},
		{"lunar-outpost-map.webp", "Lunar Outpost"},
		{"polaris-map.webp", "Polaris"},
		{"file.txt", ""},
		{"", ""},
		{"simple.webp", "Simple"},
		{"multi-word-name.webp", "Multi Word Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractMapName(tt.input)
			if result != tt.expected {
				t.Errorf("extractMapName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
