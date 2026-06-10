//go:build integration

package models

import (
	"database/sql"
	"testing"
	"time"

	"eva_trainer/internal/testutil"
)

var cardStatsMigrationFiles = []string{
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

func setupCardStatsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t, cardStatsMigrationFiles)
	t.Cleanup(cleanup)
	return db
}

func createTestSplit(t *testing.T, db *sql.DB, name string, mapNames []string) int {
	t.Helper()
	repo := NewSplitRepository(db)
	split, err := repo.Create(name, mapNames)
	if err != nil {
		t.Fatalf("failed to create split: %v", err)
	}
	return split.ID
}

func createTestPick(t *testing.T, db *sql.DB, boID, equipeID, mapID int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO bo_picks (bo_id, equipe_id, map_id) VALUES ($1, $2, $3)`,
		boID, equipeID, mapID,
	)
	if err != nil {
		t.Fatalf("failed to create pick: %v", err)
	}
}

func createTestBan(t *testing.T, db *sql.DB, boID, equipeID, mapID int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO bo_bans (bo_id, equipe_id, map_id) VALUES ($1, $2, $3)`,
		boID, equipeID, mapID,
	)
	if err != nil {
		t.Fatalf("failed to create ban: %v", err)
	}
}

func createTestMap(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	var id int
	err := db.QueryRow(`INSERT INTO maps (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create map: %v", err)
	}
	return id
}

func createTestLigue(t *testing.T, db *sql.DB, splitID int) int {
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

func createTestEvaTeam(t *testing.T, db *sql.DB, name, tag string) int {
	t.Helper()
	repo := NewEvaTeamRepository(db)
	team, err := repo.Create(name, tag, nil)
	if err != nil {
		t.Fatalf("failed to create eva team: %v", err)
	}
	return team.ID
}

func createTestBO(t *testing.T, db *sql.DB, ligueID, equipe1ID, equipe2ID int) int {
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

func createTestMatch(t *testing.T, db *sql.DB, boID, mapID, score1, score2 int, vainqueurID *int) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id, forfait, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, 0, NOW(), NOW()) RETURNING id`,
		boID, mapID, score1, score2, vainqueurID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}
	return id
}

func TestComputeStats_NoMatchs(t *testing.T) {
	db := setupCardStatsTestDB(t)
	teamID := createTestEvaTeam(t, db, "Lonely Team", "LONE")

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamID)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d entries", len(stats))
	}
}

func TestComputeStats_TeamNotFound(t *testing.T) {
	db := setupCardStatsTestDB(t)
	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(99999)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats for unknown team, got %d entries", len(stats))
	}
}

func TestComputeStats_SingleMapWithWinAndLoss(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal", "mistral"})
	mapID := createTestMap(t, db, "cp_metal")
	createTestMap(t, db, "mistral")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapID, 80, 60, &teamA)
	createTestMatch(t, db, boID, mapID, 55, 70, &teamB)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
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
	if s.Wins != 1 {
		t.Errorf("expected wins=1, got %d", s.Wins)
	}
	if s.Losses != 1 {
		t.Errorf("expected losses=1, got %d", s.Losses)
	}
	expectedWinrate := 0.5
	if s.Winrate != expectedWinrate {
		t.Errorf("expected winrate=%.2f, got %.2f", expectedWinrate, s.Winrate)
	}
	expectedAvg := (80.0 + 55.0) / 2.0
	if s.AvgScore != expectedAvg {
		t.Errorf("expected avg_score=%.1f, got %.1f", expectedAvg, s.AvgScore)
	}
	expectedMedian := (55.0 + 80.0) / 2.0
	if s.MedianScore != expectedMedian {
		t.Errorf("expected median_score=%.1f, got %.1f", expectedMedian, s.MedianScore)
	}
	if len(s.Opponents) == 0 {
		t.Fatal("expected 1 opponent stat, got 0")
	}
	opp := s.Opponents[0]
	if opp.TeamID != teamB {
		t.Errorf("expected opponent team_id=%d, got %d", teamB, opp.TeamID)
	}
	if opp.TeamName != "Team Beta" {
		t.Errorf("expected opponent name 'Team Beta', got '%s'", opp.TeamName)
	}
}

func TestComputeStats_MultipleMaps(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal", "mistral"})
	mapMetal := createTestMap(t, db, "cp_metal")
	mapMistral := createTestMap(t, db, "mistral")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapMetal, 80, 60, &teamA)
	createTestMatch(t, db, boID, mapMistral, 90, 50, &teamA)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 map stats, got %d", len(stats))
	}

	s1 := stats[0]
	if s1.MapName != "cp_metal" {
		t.Errorf("expected first map 'cp_metal', got '%s'", s1.MapName)
	}
	if s1.Wins != 1 || s1.Losses != 0 || s1.Winrate != 1.0 {
		t.Errorf("cp_metal: expected W=1 L=0 WR=1.0, got W=%d L=%d WR=%.2f", s1.Wins, s1.Losses, s1.Winrate)
	}

	s2 := stats[1]
	if s2.MapName != "mistral" {
		t.Errorf("expected second map 'mistral', got '%s'", s2.MapName)
	}
	if s2.Wins != 1 || s2.Losses != 0 || s2.Winrate != 1.0 {
		t.Errorf("mistral: expected W=1 L=0 WR=1.0, got W=%d L=%d WR=%.2f", s2.Wins, s2.Losses, s2.Winrate)
	}
}

func TestComputeStats_DrawsIgnored(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal"})
	mapID := createTestMap(t, db, "cp_metal")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapID, 80, 60, &teamA)
	createTestMatch(t, db, boID, mapID, 50, 50, nil)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}
	s := stats[0]
	if s.Wins != 1 {
		t.Errorf("expected wins=1 (draw ignored), got %d", s.Wins)
	}
	if s.Losses != 0 {
		t.Errorf("expected losses=0 (draw ignored), got %d", s.Losses)
	}
	if s.AvgScore != 65.0 {
		t.Errorf("expected avg_score=65.0 (both matches counted), got %.1f", s.AvgScore)
	}
}

func TestComputeStats_ForfaitExcluded(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal"})
	mapID := createTestMap(t, db, "cp_metal")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)

	var matchID int
	err := db.QueryRow(
		`INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id, forfait, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		boID, mapID, 80, 60, &teamA, 1,
	).Scan(&matchID)
	if err != nil {
		t.Fatalf("failed to create forfait match: %v", err)
	}
	_ = matchID

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats (only forfait matches), got %d", len(stats))
	}
}

func TestComputeStats_WinrateAllLosses(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal"})
	mapID := createTestMap(t, db, "cp_metal")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapID, 40, 60, &teamB)
	createTestMatch(t, db, boID, mapID, 30, 70, &teamB)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}
	s := stats[0]
	if s.Wins != 0 {
		t.Errorf("expected wins=0, got %d", s.Wins)
	}
	if s.Losses != 2 {
		t.Errorf("expected losses=2, got %d", s.Losses)
	}
	if s.Winrate != 0.0 {
		t.Errorf("expected winrate=0.0, got %.2f", s.Winrate)
	}
}

func TestComputeStats_OpponentStats(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal"})
	mapID := createTestMap(t, db, "cp_metal")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")
	teamC := createTestEvaTeam(t, db, "Team Gamma", "GAMMA")

	boAB := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boAB, mapID, 80, 60, &teamA)
	createTestMatch(t, db, boAB, mapID, 55, 70, &teamB)

	boAC := createTestBO(t, db, ligueID, teamA, teamC)
	createTestMatch(t, db, boAC, mapID, 90, 50, &teamA)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}
	s := stats[0]
	if len(s.Opponents) != 2 {
		t.Fatalf("expected 2 opponents, got %d", len(s.Opponents))
	}

	oppB := findOpponent(s.Opponents, teamB)
	if oppB == nil {
		t.Fatal("expected opponent Beta")
	}
	if oppB.AvgScore != 67.5 {
		t.Errorf("expected avg_score=67.5 vs Beta, got %.1f", oppB.AvgScore)
	}
	if oppB.MedianScore != 67.5 {
		t.Errorf("expected median_score=67.5 vs Beta, got %.1f", oppB.MedianScore)
	}

	oppC := findOpponent(s.Opponents, teamC)
	if oppC == nil {
		t.Fatal("expected opponent Gamma")
	}
	if oppC.AvgScore != 90.0 {
		t.Errorf("expected avg_score=90.0 vs Gamma, got %.1f", oppC.AvgScore)
	}
	if oppC.MedianScore != 90.0 {
		t.Errorf("expected median_score=90.0 vs Gamma, got %.1f", oppC.MedianScore)
	}
}

func findOpponent(opponents []OpponentCardStats, teamID int) *OpponentCardStats {
	for i := range opponents {
		if opponents[i].TeamID == teamID {
			return &opponents[i]
		}
	}
	return nil
}

func TestComputeStats_ScoreExtremes(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal"})
	mapID := createTestMap(t, db, "cp_metal")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapID, 0, 100, &teamB)
	createTestMatch(t, db, boID, mapID, 100, 0, &teamA)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 map stat, got %d", len(stats))
	}
	s := stats[0]
	if s.Wins != 1 || s.Losses != 1 {
		t.Errorf("expected W=1 L=1, got W=%d L=%d", s.Wins, s.Losses)
	}
	if s.AvgScore != 50.0 {
		t.Errorf("expected avg_score=50.0, got %.1f", s.AvgScore)
	}
	if s.MedianScore != 50.0 {
		t.Errorf("expected median_score=50.0, got %.1f", s.MedianScore)
	}
}

func TestComputeStats_PicksAndBans(t *testing.T) {
	db := setupCardStatsTestDB(t)
	splitID := createTestSplit(t, db, "Test Split", []string{"cp_metal", "mistral"})
	mapMetal := createTestMap(t, db, "cp_metal")
	mapMistral := createTestMap(t, db, "mistral")
	ligueID := createTestLigue(t, db, splitID)
	teamA := createTestEvaTeam(t, db, "Team Alpha", "ALPHA")
	teamB := createTestEvaTeam(t, db, "Team Beta", "BETA")

	boID := createTestBO(t, db, ligueID, teamA, teamB)
	createTestMatch(t, db, boID, mapMetal, 80, 60, &teamA)
	createTestMatch(t, db, boID, mapMistral, 90, 50, &teamA)

	// Team A picks cp_metal twice, bans mistral once
	createTestPick(t, db, boID, teamA, mapMetal)
	createTestPick(t, db, boID, teamA, mapMetal)
	createTestBan(t, db, boID, teamA, mapMistral)

	repo := NewCardStatsRepository(db)
	stats, err := repo.Compute(teamA)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 map stats, got %d", len(stats))
	}

	s1 := stats[0]
	if s1.MapName != "cp_metal" {
		t.Errorf("expected first map 'cp_metal', got '%s'", s1.MapName)
	}
	if s1.PickRate != 1.0 {
		t.Errorf("cp_metal: expected pick_rate=1.0, got %.2f", s1.PickRate)
	}
	if s1.BanRate != 0.0 {
		t.Errorf("cp_metal: expected ban_rate=0.0, got %.2f", s1.BanRate)
	}

	s2 := stats[1]
	if s2.MapName != "mistral" {
		t.Errorf("expected second map 'mistral', got '%s'", s2.MapName)
	}
	if s2.PickRate != 0.0 {
		t.Errorf("mistral: expected pick_rate=0.0, got %.2f", s2.PickRate)
	}
	if s2.BanRate != 1.0 {
		t.Errorf("mistral: expected ban_rate=1.0, got %.2f", s2.BanRate)
	}
}
