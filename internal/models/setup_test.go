package models

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

const migrationsDir = "../../db/migrations"

var migrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000002_add_refresh_tokens.up.sql",
	"000003_add_display_name.up.sql",
	"000004_add_discord_username.up.sql",
	"000007_add_team_invitations_and_join_requests.up.sql",
	"000011_add_zone_polygon_constraints.up.sql",
	"000013_add_zone_versions.up.sql",
	"000015_add_template_is_public.up.sql",
	"000016_add_user_settings.up.sql",
	"000017_add_polygon_floor.up.sql",
	"000018_rename_template_is_public.up.sql",
	"000019_template_original_team_cascade.up.sql",
	"000020_create_lines_of_fire.up.sql",
	"000021_create_template_lines_of_fire.up.sql",
	"000022_create_line_of_fire_edit_history.up.sql",
	"000023_create_line_of_fire_versions.up.sql",
	"000024_add_line_of_fire_permissions.up.sql",
	"000025_add_archive_roles.up.sql",
	"000026_add_archive_leagues.up.sql",
	"000027_fix_ligue_locale_table.up.sql",
	"000028_add_ligue_locale_emplacement_unique.up.sql",
	"000029_add_ligue_locale_junction.up.sql",
	"000030_add_eva_teams.up.sql",
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
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
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to ping database: %v", err)
	}

	for _, mf := range migrationFiles {
		path := filepath.Join(migrationsDir, mf)
		content, err := os.ReadFile(path)
		if err != nil {
			db.Close()
			container.Terminate(ctx)
			t.Fatalf("failed to read migration %s: %v", mf, err)
		}

		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				db.Close()
				container.Terminate(ctx)
				t.Fatalf("failed to execute migration %s: %v\nSQL: %s", mf, err, stmt)
			}
		}
	}

	cleanup := func() {
		db.Close()
		container.Terminate(context.Background())
	}

	return db, cleanup
}
