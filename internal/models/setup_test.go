package models

import (
	"database/sql"
	"testing"

	"eva_trainer/internal/testutil"
)

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
	"000031_add_archive_divs_and_bos.up.sql",
	"000032_add_divs_replay_url.up.sql",
	"000033_add_div_notes.up.sql",
	"000034_add_ligue_nom.up.sql",
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	return testutil.SetupTestDB(t, migrationFiles)
}
