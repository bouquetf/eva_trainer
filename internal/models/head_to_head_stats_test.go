//go:build integration

package models

import (
	"database/sql"
	"testing"
	"time"

	"eva_trainer/internal/testutil"
)

var headToHeadMigrationFiles = []string{
	"000001_initial_schema.up.sql",
	"000005_seed_permissions.up.sql",
	"000025_add_archive_roles.up.sql",
	"000026_add_archive_leagues.up.sql",
	"000027_fix_ligue_locale_table.up.sql",
	"000028_add_ligue_locale_emplacement_unique.up.sql",
	"000029_add_ligue_locale_junction.up.sql",
	"000030_add_eva_teams.up.sql",
	"000031_add_archive_divs_and_bos.up.sql",
}

func setupH2HTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t, headToHeadMigrationFiles)
	t.Cleanup(cleanup)
	return db
}

func createH2HTestMap(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	var id int
	err := db.QueryRow(`INSERT INTO maps (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	return id
}

func createH2HTestSplit(t *testing.T, db *sql.DB, name string, mapNames []string) int {
	t.Helper()
	repo := NewSplitRepository(db)
	split, err := repo.Create(name, mapNames)
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	return split.ID
}

func createH2HTestLigue(t *testing.T, db *sql.DB, splitID int) int {
	t.Helper()
	now := time.Now().Truncate(24 * time.Hour)
	var id int
	err := db.QueryRow(
		`INSERT INTO ligue (date_debut, date_fin, split_id, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		now, now.Add(30*24*time.Hour), splitID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create ligue: %v", err)
	}
	return id
}

func createH2HTestEvaTeam(t *testing.T, db *sql.DB, name, tag string) int {
	t.Helper()
	repo := NewEvaTeamRepository(db)
	team, err := repo.Create(name, tag, nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}
	return team.ID
}

func createH2HTestBO(t *testing.T, db *sql.DB, ligueID, equipe1ID, equipe2ID int) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, format, created_at, updated_at) VALUES ($1, $2, $3, 'BO3', NOW(), NOW()) RETURNING id`,
		ligueID, equipe1ID, equipe2ID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create BO: %v", err)
	}
	return id
}

func createH2HTestMatch(t *testing.T, db *sql.DB, boID, mapID, score1, score2 int, vainqueurID *int, forfait int) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id, forfait, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		boID, mapID, score1, score2, vainqueurID, forfait,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}
	return id
}

func TestH2HCompute_Team1EqualsTeam2(t *testing.T) {
	db := setupH2HTestDB(t)
	teamID := createH2HTestEvaTeam(t, db, "Team", "T")

	repo := NewHeadToHeadStatsRepository(db)
	_, err := repo.Compute(teamID, teamID)
	if err != ErrInvalidTeamPair {
		t.Fatalf("expected ErrInvalidTeamPair, got %v", err)
	}
}

func TestH2HCompute_NoMatches(t *testing.T) {
	db := setupH2HTestDB(t)
	teamA := createH2HTestEvaTeam(t, db, "Team A", "TA")
	teamB := createH2HTestEvaTeam(t, db, "Team B", "TB")

	repo := NewHeadToHeadStatsRepository(db)
	stats, err := repo.Compute(teamA, teamB)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d entries", len(stats))
	}
}

func TestH2HCompute_SingleMapWithWinAndLoss(t *testing.T) {
	db := setupH2HTestDB(t)
	splitID := createH2HTestSplit(t, db, "H2H Test Split", []string{"cp_metal"})
	mapID := createH2HTestMap(t, db, "cp_metal")
	ligueID := createH2HTestLigue(t, db, splitID)
	teamA := createH2HTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createH2HTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createH2HTestBO(t, db, ligueID, teamA, teamB)
	createH2HTestMatch(t, db, boID, mapID, 80, 60, &teamA, 0)
	createH2HTestMatch(t, db, boID, mapID, 55, 70, &teamB, 0)

	repo := NewHeadToHeadStatsRepository(db)
	stats, err := repo.Compute(teamA, teamB)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}

	s := stats[0]
	if s.MapName != "cp_metal" {
		t.Errorf("expected map_name 'cp_metal', got '%s'", s.MapName)
	}

	// Team1 stats
	if s.Team1Stats.Wins != 1 {
		t.Errorf("team1: expected wins=1, got %d", s.Team1Stats.Wins)
	}
	if s.Team1Stats.Losses != 1 {
		t.Errorf("team1: expected losses=1, got %d", s.Team1Stats.Losses)
	}
	if s.Team1Stats.Winrate != 0.5 {
		t.Errorf("team1: expected winrate=0.5, got %.2f", s.Team1Stats.Winrate)
	}
	expectedTeam1Avg := (80.0 + 55.0) / 2.0
	if s.Team1Stats.AvgScore != expectedTeam1Avg {
		t.Errorf("team1: expected avg_score=%.1f, got %.1f", expectedTeam1Avg, s.Team1Stats.AvgScore)
	}
	expectedTeam1Median := (55.0 + 80.0) / 2.0
	if s.Team1Stats.MedianScore != expectedTeam1Median {
		t.Errorf("team1: expected median_score=%.1f, got %.1f", expectedTeam1Median, s.Team1Stats.MedianScore)
	}

	// Team2 stats
	if s.Team2Stats.Wins != 1 {
		t.Errorf("team2: expected wins=1, got %d", s.Team2Stats.Wins)
	}
	if s.Team2Stats.Losses != 1 {
		t.Errorf("team2: expected losses=1, got %d", s.Team2Stats.Losses)
	}
	if s.Team2Stats.Winrate != 0.5 {
		t.Errorf("team2: expected winrate=0.5, got %.2f", s.Team2Stats.Winrate)
	}
	expectedTeam2Avg := (60.0 + 70.0) / 2.0
	if s.Team2Stats.AvgScore != expectedTeam2Avg {
		t.Errorf("team2: expected avg_score=%.1f, got %.1f", expectedTeam2Avg, s.Team2Stats.AvgScore)
	}
	expectedTeam2Median := (60.0 + 70.0) / 2.0
	if s.Team2Stats.MedianScore != expectedTeam2Median {
		t.Errorf("team2: expected median_score=%.1f, got %.1f", expectedTeam2Median, s.Team2Stats.MedianScore)
	}
}

func TestH2HCompute_MultipleMaps(t *testing.T) {
	db := setupH2HTestDB(t)
	splitID := createH2HTestSplit(t, db, "H2H Multi Map", []string{"cp_metal", "mistral"})
	mapMetal := createH2HTestMap(t, db, "cp_metal")
	mapMistral := createH2HTestMap(t, db, "mistral")
	ligueID := createH2HTestLigue(t, db, splitID)
	teamA := createH2HTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createH2HTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createH2HTestBO(t, db, ligueID, teamA, teamB)
	createH2HTestMatch(t, db, boID, mapMetal, 80, 60, &teamA, 0)
	createH2HTestMatch(t, db, boID, mapMetal, 55, 70, &teamB, 0)
	createH2HTestMatch(t, db, boID, mapMistral, 90, 50, &teamA, 0)

	repo := NewHeadToHeadStatsRepository(db)
	stats, err := repo.Compute(teamA, teamB)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 map stats, got %d", len(stats))
	}

	// First map (alphabetical): cp_metal
	s1 := stats[0]
	if s1.MapName != "cp_metal" {
		t.Errorf("expected first map 'cp_metal', got '%s'", s1.MapName)
	}
	if s1.Team1Stats.Wins != 1 || s1.Team1Stats.Losses != 1 || s1.Team1Stats.Winrate != 0.5 {
		t.Errorf("cp_metal team1: expected W=1 L=1 WR=0.5, got W=%d L=%d WR=%.2f", s1.Team1Stats.Wins, s1.Team1Stats.Losses, s1.Team1Stats.Winrate)
	}
	if s1.Team2Stats.Wins != 1 || s1.Team2Stats.Losses != 1 || s1.Team2Stats.Winrate != 0.5 {
		t.Errorf("cp_metal team2: expected W=1 L=1 WR=0.5, got W=%d L=%d WR=%.2f", s1.Team2Stats.Wins, s1.Team2Stats.Losses, s1.Team2Stats.Winrate)
	}
	if s1.Team1Stats.AvgScore != 67.5 {
		t.Errorf("cp_metal team1: expected avg_score=67.5, got %.1f", s1.Team1Stats.AvgScore)
	}
	if s1.Team2Stats.AvgScore != 65.0 {
		t.Errorf("cp_metal team2: expected avg_score=65.0, got %.1f", s1.Team2Stats.AvgScore)
	}

	// Second map: mistral
	s2 := stats[1]
	if s2.MapName != "mistral" {
		t.Errorf("expected second map 'mistral', got '%s'", s2.MapName)
	}
	if s2.Team1Stats.Wins != 1 || s2.Team1Stats.Losses != 0 || s2.Team1Stats.Winrate != 1.0 {
		t.Errorf("mistral team1: expected W=1 L=0 WR=1.0, got W=%d L=%d WR=%.2f", s2.Team1Stats.Wins, s2.Team1Stats.Losses, s2.Team1Stats.Winrate)
	}
	if s2.Team2Stats.Wins != 0 || s2.Team2Stats.Losses != 1 || s2.Team2Stats.Winrate != 0.0 {
		t.Errorf("mistral team2: expected W=0 L=1 WR=0.0, got W=%d L=%d WR=%.2f", s2.Team2Stats.Wins, s2.Team2Stats.Losses, s2.Team2Stats.Winrate)
	}
	if s2.Team1Stats.AvgScore != 90.0 {
		t.Errorf("mistral team1: expected avg_score=90.0, got %.1f", s2.Team1Stats.AvgScore)
	}
	if s2.Team2Stats.AvgScore != 50.0 {
		t.Errorf("mistral team2: expected avg_score=50.0, got %.1f", s2.Team2Stats.AvgScore)
	}
	if s2.Team1Stats.MedianScore != 90.0 {
		t.Errorf("mistral team1: expected median_score=90.0, got %.1f", s2.Team1Stats.MedianScore)
	}
	if s2.Team2Stats.MedianScore != 50.0 {
		t.Errorf("mistral team2: expected median_score=50.0, got %.1f", s2.Team2Stats.MedianScore)
	}
}

func TestH2HCompute_Draw(t *testing.T) {
	db := setupH2HTestDB(t)
	splitID := createH2HTestSplit(t, db, "H2H Draw", []string{"cp_metal"})
	mapID := createH2HTestMap(t, db, "cp_metal")
	ligueID := createH2HTestLigue(t, db, splitID)
	teamA := createH2HTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createH2HTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createH2HTestBO(t, db, ligueID, teamA, teamB)
	createH2HTestMatch(t, db, boID, mapID, 60, 60, nil, 0)

	repo := NewHeadToHeadStatsRepository(db)
	stats, err := repo.Compute(teamA, teamB)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}

	s := stats[0]
	if s.Team1Stats.Wins != 0 || s.Team1Stats.Losses != 0 || s.Team1Stats.Winrate != 0.0 {
		t.Errorf("team1: expected W=0 L=0 WR=0.0, got W=%d L=%d WR=%.2f", s.Team1Stats.Wins, s.Team1Stats.Losses, s.Team1Stats.Winrate)
	}
	if s.Team2Stats.Wins != 0 || s.Team2Stats.Losses != 0 || s.Team2Stats.Winrate != 0.0 {
		t.Errorf("team2: expected W=0 L=0 WR=0.0, got W=%d L=%d WR=%.2f", s.Team2Stats.Wins, s.Team2Stats.Losses, s.Team2Stats.Winrate)
	}
	if s.Team1Stats.AvgScore != 60.0 {
		t.Errorf("team1: expected avg_score=60.0, got %.1f", s.Team1Stats.AvgScore)
	}
	if s.Team1Stats.MedianScore != 60.0 {
		t.Errorf("team1: expected median_score=60.0, got %.1f", s.Team1Stats.MedianScore)
	}
	if s.Team2Stats.AvgScore != 60.0 {
		t.Errorf("team2: expected avg_score=60.0, got %.1f", s.Team2Stats.AvgScore)
	}
	if s.Team2Stats.MedianScore != 60.0 {
		t.Errorf("team2: expected median_score=60.0, got %.1f", s.Team2Stats.MedianScore)
	}
}

func TestH2HCompute_ForfaitExcluded(t *testing.T) {
	db := setupH2HTestDB(t)
	splitID := createH2HTestSplit(t, db, "H2H Forfait", []string{"cp_metal"})
	mapID := createH2HTestMap(t, db, "cp_metal")
	ligueID := createH2HTestLigue(t, db, splitID)
	teamA := createH2HTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createH2HTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createH2HTestBO(t, db, ligueID, teamA, teamB)
	createH2HTestMatch(t, db, boID, mapID, 80, 60, &teamA, 0)
	createH2HTestMatch(t, db, boID, mapID, 0, 100, nil, 1)

	repo := NewHeadToHeadStatsRepository(db)
	stats, err := repo.Compute(teamA, teamB)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}

	s := stats[0]
	if s.Team1Stats.Wins != 1 || s.Team1Stats.Losses != 0 {
		t.Errorf("team1: expected W=1 L=0, got W=%d L=%d", s.Team1Stats.Wins, s.Team1Stats.Losses)
	}
	if s.Team2Stats.Wins != 0 || s.Team2Stats.Losses != 1 {
		t.Errorf("team2: expected W=0 L=1, got W=%d L=%d", s.Team2Stats.Wins, s.Team2Stats.Losses)
	}
	if s.Team1Stats.AvgScore != 80.0 {
		t.Errorf("team1: expected avg_score=80.0, got %.1f", s.Team1Stats.AvgScore)
	}
	if s.Team2Stats.AvgScore != 60.0 {
		t.Errorf("team2: expected avg_score=60.0, got %.1f", s.Team2Stats.AvgScore)
	}
}
