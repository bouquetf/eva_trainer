//go:build integration

package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func seedRolePermissions(db *sql.DB) error {
	permMap := make(map[string]int)
	rows, err := db.Query(`SELECT id, feature FROM permissions`)
	if err != nil {
		return fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var feature string
		if err := rows.Scan(&id, &feature); err != nil {
			return fmt.Errorf("failed to scan permission: %w", err)
		}
		permMap[feature] = id
	}

	if err := rows.Err(); err != nil {
		return err
	}

	type assignment struct {
		role    string
		feature string
	}

	assignments := []assignment{
		{role: "captain", feature: "team:edit"},
		{role: "captain", feature: "team:delete"},
		{role: "captain", feature: "team:leave"},
		{role: "captain", feature: "team:invite"},
		{role: "captain", feature: "team:requests:view"},
		{role: "captain", feature: "team:requests:manage"},
		{role: "captain", feature: "member:view"},
		{role: "captain", feature: "member:kick"},
		{role: "captain", feature: "member:role:assign"},
		{role: "captain", feature: "member:role:promote-cocaptain"},
		{role: "captain", feature: "member:role:transfer-captain"},
		{role: "captain", feature: "zone:view"},
		{role: "captain", feature: "zone:create"},
		{role: "captain", feature: "zone:edit"},
		{role: "captain", feature: "zone:delete"},
		{role: "captain", feature: "zone:publish"},
		{role: "captain", feature: "zone:lock"},
		{role: "captain", feature: "review:start"},
		{role: "captain", feature: "template:duplicate"},
		{role: "captain", feature: "edit_history:view"},
		{role: "captain", feature: "edit_history:rollback"},

		{role: "cocaptain", feature: "team:edit"},
		{role: "cocaptain", feature: "team:leave"},
		{role: "cocaptain", feature: "team:invite"},
		{role: "cocaptain", feature: "team:requests:view"},
		{role: "cocaptain", feature: "team:requests:manage"},
		{role: "cocaptain", feature: "member:view"},
		{role: "cocaptain", feature: "member:role:assign"},
		{role: "cocaptain", feature: "zone:view"},
		{role: "cocaptain", feature: "zone:create"},
		{role: "cocaptain", feature: "zone:edit"},
		{role: "cocaptain", feature: "zone:delete"},
		{role: "cocaptain", feature: "zone:publish"},
		{role: "cocaptain", feature: "zone:lock"},
		{role: "cocaptain", feature: "review:start"},
		{role: "cocaptain", feature: "template:duplicate"},
		{role: "cocaptain", feature: "edit_history:view"},

		{role: "editor", feature: "team:leave"},
		{role: "editor", feature: "member:view"},
		{role: "editor", feature: "zone:view"},
		{role: "editor", feature: "zone:create"},
		{role: "editor", feature: "zone:edit"},
		{role: "editor", feature: "zone:delete"},
		{role: "editor", feature: "review:start"},
		{role: "editor", feature: "template:duplicate"},

		{role: "member", feature: "team:create"},
		{role: "member", feature: "team:leave"},
		{role: "member", feature: "member:view"},
		{role: "member", feature: "zone:view"},
		{role: "member", feature: "review:start"},
		{role: "member", feature: "template:duplicate"},
		{role: "member", feature: "edit_history:view"},
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, a := range assignments {
		permID, ok := permMap[a.feature]
		if !ok {
			return fmt.Errorf("permission not found for feature %q", a.feature)
		}

		_, err := tx.Exec(
			`INSERT INTO role_permissions (role, permission_id) VALUES ($1, $2)
			 ON CONFLICT (role, permission_id) DO NOTHING`,
			a.role, permID,
		)
		if err != nil {
			return fmt.Errorf("failed to assign %q to role %q: %w", a.feature, a.role, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func setupPermissionDB(t *testing.T) (PermissionRepository, RolePermissionRepository, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)

	permRepo := NewPermissionRepository(db)
	rolePermRepo := NewRolePermissionRepository(db)

	// Run seed migration
	path := filepath.Join(migrationsDir, "000005_seed_permissions.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		cleanup()
		t.Fatalf("failed to read seed migration: %v", err)
	}

	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			cleanup()
			t.Fatalf("failed to execute seed migration: %v\nSQL: %s", err, stmt)
		}
	}

	if err := seedRolePermissions(db); err != nil {
		cleanup()
		t.Fatalf("SeedRolePermissions failed: %v", err)
	}

	return permRepo, rolePermRepo, cleanup
}

func TestPermissionRepository_Create(t *testing.T) {
	permRepo, _, cleanup := setupPermissionDB(t)
	defer cleanup()

	perm, err := permRepo.Create("zone:test", "Test permission")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if perm.Feature != "zone:test" {
		t.Errorf("expected feature 'zone:test', got '%s'", perm.Feature)
	}
	if perm.Description != "Test permission" {
		t.Errorf("expected description 'Test permission', got '%s'", perm.Description)
	}
	if perm.ID <= 0 {
		t.Errorf("expected positive ID, got %d", perm.ID)
	}
}

func TestPermissionRepository_FindByFeature_Success(t *testing.T) {
	permRepo, _, cleanup := setupPermissionDB(t)
	defer cleanup()

	perm, err := permRepo.FindByFeature("team:create")
	if err != nil {
		t.Fatalf("FindByFeature failed: %v", err)
	}
	if perm.Feature != "team:create" {
		t.Errorf("expected 'team:create', got '%s'", perm.Feature)
	}
	if perm.ID <= 0 {
		t.Errorf("expected positive ID, got %d", perm.ID)
	}
}

func TestPermissionRepository_FindByFeature_NotFound(t *testing.T) {
	permRepo, _, cleanup := setupPermissionDB(t)
	defer cleanup()

	_, err := permRepo.FindByFeature("nonexistent:feature")
	if err != ErrPermissionNotFound {
		t.Errorf("expected ErrPermissionNotFound, got %v", err)
	}
}

func TestPermissionRepository_FindByID_Success(t *testing.T) {
	permRepo, _, cleanup := setupPermissionDB(t)
	defer cleanup()

	created, err := permRepo.Create("zone:test", "Test permission")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := permRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Feature != "zone:test" {
		t.Errorf("expected 'zone:test', got '%s'", found.Feature)
	}
	if found.Description != "Test permission" {
		t.Errorf("expected 'Test permission', got '%s'", found.Description)
	}
}

func TestPermissionRepository_FindAll_ContainsSeeds(t *testing.T) {
	permRepo, _, cleanup := setupPermissionDB(t)
	defer cleanup()

	all, err := permRepo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	featureMap := make(map[string]bool)
	for _, p := range all {
		featureMap[p.Feature] = true
	}

	expected := []string{
		"team:create", "team:edit", "team:delete", "team:leave",
		"member:view", "zone:view", "zone:create", "zone:edit",
		"map:add",
	}

	for _, f := range expected {
		if !featureMap[f] {
			t.Errorf("expected feature '%s' to be present in FindAll", f)
		}
	}
}

func TestRolePermissionRepository_AssignAndCheck(t *testing.T) {
	permRepo, rolePermRepo, cleanup := setupPermissionDB(t)
	defer cleanup()

	perm, err := permRepo.Create("zone:other", "Other permission")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := rolePermRepo.AssignPermissionToRole("editor", perm.ID); err != nil {
		t.Fatalf("AssignPermissionToRole failed: %v", err)
	}

	has, err := rolePermRepo.RoleHasPermission("editor", perm.ID)
	if err != nil {
		t.Fatalf("RoleHasPermission failed: %v", err)
	}
	if !has {
		t.Errorf("expected true after assignment")
	}
}

func TestRolePermissionRepository_RoleHasPermission_Seed(t *testing.T) {
	permRepo, rolePermRepo, cleanup := setupPermissionDB(t)
	defer cleanup()

	deletePerm, err := permRepo.FindByFeature("team:delete")
	if err != nil {
		t.Fatalf("FindByFeature(team:delete) failed: %v", err)
	}

	has, err := rolePermRepo.RoleHasPermission("captain", deletePerm.ID)
	if err != nil {
		t.Fatalf("RoleHasPermission failed: %v", err)
	}
	if !has {
		t.Errorf("expected captain to have team:delete")
	}

	memberHas, err := rolePermRepo.RoleHasPermission("member", deletePerm.ID)
	if err != nil {
		t.Fatalf("RoleHasPermission failed: %v", err)
	}
	if memberHas {
		t.Errorf("expected member NOT to have team:delete")
	}
}

func TestRolePermissionRepository_FindByRole(t *testing.T) {
	_, rolePermRepo, cleanup := setupPermissionDB(t)
	defer cleanup()

	perms, err := rolePermRepo.FindByRole("captain")
	if err != nil {
		t.Fatalf("FindByRole failed: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected captain to have at least one permission")
	}

	featureMap := make(map[string]bool)
	for _, p := range perms {
		featureMap[p.Feature] = true
	}

	for _, f := range []string{"team:edit", "team:delete", "member:role:assign"} {
		if !featureMap[f] {
			t.Errorf("expected captain to have '%s'", f)
		}
	}
}

func TestRolePermissionRepository_RemovePermission(t *testing.T) {
	permRepo, rolePermRepo, cleanup := setupPermissionDB(t)
	defer cleanup()

	perm, err := permRepo.Create("zone:temp", "Temp")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := rolePermRepo.AssignPermissionToRole("editor", perm.ID); err != nil {
		t.Fatalf("Assign failed: %v", err)
	}

	if err := rolePermRepo.RemovePermissionFromRole("editor", perm.ID); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	has, err := rolePermRepo.RoleHasPermission("editor", perm.ID)
	if err != nil {
		t.Fatalf("RoleHasPermission failed: %v", err)
	}
	if has {
		t.Errorf("expected false after removal")
	}
}

func TestHasPermission_AdminBypass(t *testing.T) {
	has, err := HasPermission("admin", "map:add", nil, nil)
	if err != nil {
		t.Fatalf("HasPermission returned error: %v", err)
	}
	if !has {
		t.Errorf("expected admin to bypass permission check")
	}
}

func TestHasPermission_UnknownFeature(t *testing.T) {
	permRepo, rolePermRepo, cleanup := setupPermissionDB(t)
	defer cleanup()

	has, err := HasPermission("member", "does:not:exist", rolePermRepo, permRepo)
	if err != nil {
		t.Fatalf("HasPermission returned error: %v", err)
	}
	if has {
		t.Errorf("expected false for unknown feature")
	}
}

func TestSeedPermissions_IdempotentSQL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	path := filepath.Join(migrationsDir, "000005_seed_permissions.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	executeSQL := func() {
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("SQL failed: %v\nSQL: %s", err, stmt)
			}
		}
	}

	executeSQL()
	executeSQL()

	permRepo := NewPermissionRepository(db)
	all, err := permRepo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	featureCount := make(map[string]int)
	for _, p := range all {
		featureCount[p.Feature]++
	}

	for feature, count := range featureCount {
		if count > 1 {
			t.Errorf("duplicate feature '%s' (count=%d)", feature, count)
		}
	}

	if len(featureCount) < 23 {
		t.Errorf("expected at least 23 features, got %d", len(featureCount))
	}
}

func TestSeedRolePermissions_Idempotent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Run seed migration
	path := filepath.Join(migrationsDir, "000005_seed_permissions.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed migration failed: %v", err)
		}
	}

	// Run seedRolePermissions twice
	if err := seedRolePermissions(db); err != nil {
		t.Fatalf("first SeedRolePermissions failed: %v", err)
	}
	if err := seedRolePermissions(db); err != nil {
		t.Fatalf("second SeedRolePermissions failed: %v", err)
	}

	rolePermRepo := NewRolePermissionRepository(db)
	captainPerms, err := rolePermRepo.FindByRole("captain")
	if err != nil {
		t.Fatalf("FindByRole failed: %v", err)
	}

	if len(captainPerms) < 15 {
		t.Errorf("expected at least 15 captain permissions, got %d", len(captainPerms))
	}
}
