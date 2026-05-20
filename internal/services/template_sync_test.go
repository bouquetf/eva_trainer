//go:build integration

package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"eva_trainer/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupSyncTestDB(t *testing.T) *sql.DB {
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
		WaitingFor:   wait.ForAll(wait.ForLog("database system is ready to accept connections"), wait.ForListeningPort("5432/tcp")),
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
		"000002_add_refresh_tokens.up.sql",
		"000003_add_display_name.up.sql",
		"000004_add_discord_username.up.sql",
		"000011_add_zone_polygon_constraints.up.sql",
		"000015_add_template_is_public.up.sql",
		"000017_add_polygon_floor.up.sql",
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

type syncTestFixtures struct {
	db           *sql.DB
	syncService  *TemplateSyncService
	templateRepo models.TemplateRepository
	zoneRepo     models.ZoneRepository
	mapRepo      models.MapRepository
	teamRepo     models.TeamRepository
	team         *models.Team
	mapObj       *models.Map
}

func countTemplateZones(t *testing.T, db *sql.DB, templateID int) int {
	t.Helper()
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM template_zones WHERE template_id = $1`, templateID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count template zones: %v", err)
	}
	return count
}

func setupSyncTest(t *testing.T) *syncTestFixtures {
	t.Helper()

	db := setupSyncTestDB(t)

	teamRepo := models.NewTeamRepository(db)
	mapRepo := models.NewMapRepository(db)
	zoneRepo := models.NewZoneRepository(db)
	templateRepo := models.NewTemplateRepository(db)

	user, _, err := models.NewUserRepository(db).FindOrCreateByDiscordID("sync-test-user-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("Sync Test Team "+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	m, err := mapRepo.Create("sync-test-map-"+t.Name(), nil)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}

	testLogger := NewLogger(LogLevelError, LogFormatText)
	syncService := NewTemplateSyncService(templateRepo, zoneRepo, testLogger)

	return &syncTestFixtures{
		db:           db,
		syncService:  syncService,
		templateRepo: templateRepo,
		zoneRepo:     zoneRepo,
		mapRepo:      mapRepo,
		teamRepo:     teamRepo,
		team:         team,
		mapObj:       m,
	}
}

func createZoneForSync(t *testing.T, zoneRepo models.ZoneRepository, teamID, mapID int, name string, points []models.Point) int {
	t.Helper()
	zone, err := zoneRepo.Create(teamID, mapID)
	if err != nil {
		t.Fatalf("failed to create zone: %v", err)
	}
	if _, err := zoneRepo.AddPolygon(zone.ID, points, 0, nil); err != nil {
		t.Fatalf("failed to add polygon: %v", err)
	}
	if name != "" {
		if err := zoneRepo.SetName(zone.ID, name); err != nil {
			t.Fatalf("failed to set name: %v", err)
		}
	}
	return zone.ID
}

func TestTemplateSyncService_CreateZone_UpdatesTemplate(t *testing.T) {
	f := setupSyncTest(t)

	pts1 := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	pts2 := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 150, Y: 200}}
	createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone A", pts1)
	createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone B", pts2)

	tmpl, err := f.templateRepo.CreateFromMap(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}

	pts3 := []models.Point{{X: 300, Y: 300}, {X: 400, Y: 300}, {X: 350, Y: 400}}
	zone3, err := f.zoneRepo.Create(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("failed to create zone 3: %v", err)
	}
	if _, err := f.zoneRepo.AddPolygon(zone3.ID, pts3, 0, nil); err != nil {
		t.Fatalf("failed to add polygon 3: %v", err)
	}

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("SyncOnZoneChange failed: %v", err)
	}

	zoneCount := countTemplateZones(t, f.db, tmpl.ID)
	if zoneCount != 3 {
		t.Fatalf("expected 3 zones in template, got %d", zoneCount)
	}
}

func TestTemplateSyncService_DeleteZone_UpdatesTemplate(t *testing.T) {
	f := setupSyncTest(t)

	pts1 := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	pts2 := []models.Point{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 150, Y: 200}}
	z1 := createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone A", pts1)
	createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone B", pts2)

	tmpl, err := f.templateRepo.CreateFromMap(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}

	initialCount := countTemplateZones(t, f.db, tmpl.ID)
	if initialCount != 2 {
		t.Fatalf("expected 2 zones initially, got %d", initialCount)
	}

	if err := f.zoneRepo.Delete(z1); err != nil {
		t.Fatalf("failed to delete zone: %v", err)
	}

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("SyncOnZoneChange after delete failed: %v", err)
	}

	zoneCount := countTemplateZones(t, f.db, tmpl.ID)
	if zoneCount != 1 {
		t.Fatalf("expected 1 zone in template after delete, got %d", zoneCount)
	}
}

func TestTemplateSyncService_NonPublishedMap_NoAction(t *testing.T) {
	f := setupSyncTest(t)

	tmpl, err := f.templateRepo.FindByMapID(f.mapObj.ID)
	if err != nil {
		t.Fatalf("FindByMapID failed: %v", err)
	}
	if tmpl != nil {
		t.Fatal("expected no template for non-published map")
	}

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone", pts)

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("SyncOnZoneChange failed: %v", err)
	}

	tmpl, err = f.templateRepo.FindByMapID(f.mapObj.ID)
	if err != nil {
		t.Fatalf("FindByMapID failed: %v", err)
	}
	if tmpl != nil {
		t.Fatal("expected no template to be created for non-published map")
	}
}

func TestTemplateSyncService_UpdatePolygon_UpdatesTemplate(t *testing.T) {
	f := setupSyncTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	zoneID := createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "Zone", pts)

	tmpl, err := f.templateRepo.CreateFromMap(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}

	polygons, err := f.zoneRepo.GetPolygons(zoneID)
	if err != nil || len(polygons) == 0 {
		t.Fatalf("failed to get polygons: %v", err)
	}
	polygonID := polygons[0].ID

	newPoints := []models.Point{{X: 50, Y: 50}, {X: 100, Y: 50}, {X: 75, Y: 100}}
	if err := f.zoneRepo.UpdatePolygonPoints(polygonID, newPoints); err != nil {
		t.Fatalf("failed to update polygon points: %v", err)
	}

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("SyncOnZoneChange failed: %v", err)
	}

	zoneCount := countTemplateZones(t, f.db, tmpl.ID)
	if zoneCount != 1 {
		t.Fatalf("expected 1 zone in template, got %d", zoneCount)
	}
}

func TestTemplateSyncService_RenameZone_UpdatesTemplate(t *testing.T) {
	f := setupSyncTest(t)

	pts := []models.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 10}}
	zoneID := createZoneForSync(t, f.zoneRepo, f.team.ID, f.mapObj.ID, "top", pts)

	tmpl, err := f.templateRepo.CreateFromMap(f.team.ID, f.mapObj.ID)
	if err != nil {
		t.Fatalf("CreateFromMap failed: %v", err)
	}

	if err := f.zoneRepo.SetName(zoneID, "secret"); err != nil {
		t.Fatalf("failed to rename zone: %v", err)
	}

	if err := f.syncService.SyncOnZoneChange(f.team.ID, f.mapObj.ID); err != nil {
		t.Fatalf("SyncOnZoneChange after rename failed: %v", err)
	}

	var nameAfter string
	err = f.db.QueryRow(`SELECT name FROM template_zones WHERE template_id = $1`, tmpl.ID).Scan(&nameAfter)
	if err != nil {
		t.Fatalf("failed to get template zone name after rename: %v", err)
	}
	if nameAfter != "secret" {
		t.Fatalf("expected name 'secret' after rename, got '%s'", nameAfter)
	}
}
