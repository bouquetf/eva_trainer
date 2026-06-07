package models

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgconn"
)

type Div struct {
	ID      int    `json:"id"`
	LigueID int    `json:"ligue_id"`
	Nom     string `json:"nom"`
}

type StandingEntry struct {
	TeamID        int    `json:"team_id"`
	TeamName      string `json:"team_name"`
	TeamTag       string `json:"team_tag"`
	BOsPlayed     int    `json:"bos_played"`
	BOWon         int    `json:"bo_won"`
	BODrawn       int    `json:"bo_drawn"`
	BOLost        int    `json:"bo_lost"`
	BOPoints      int    `json:"bo_points"`
	MatchesPlayed int    `json:"matches_played"`
	MatchWon      int    `json:"match_won"`
	MatchDrawn    int    `json:"match_drawn"`
	MatchLost     int    `json:"match_lost"`
	MatchPoints   int    `json:"match_points"`
}

type divRepo struct {
	db *sql.DB
}

func NewDivRepository(db *sql.DB) DivRepository {
	return &divRepo{db: db}
}

func scanDiv(scanner interface {
	Scan(dest ...interface{}) error
}) (*Div, error) {
	d := &Div{}
	err := scanner.Scan(&d.ID, &d.LigueID, &d.Nom)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func scanDivs(rows *sql.Rows) ([]*Div, error) {
	var divs []*Div
	for rows.Next() {
		d, err := scanDiv(rows)
		if err != nil {
			return nil, err
		}
		divs = append(divs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if divs == nil {
		return []*Div{}, nil
	}
	return divs, nil
}

func (r *divRepo) Create(ligueID int, nom string) (*Div, error) {
	if nom == "" {
		return nil, ErrDivNameRequired
	}

	d, err := scanDiv(r.db.QueryRow(
		`INSERT INTO divs (ligue_id, nom)
		 VALUES ($1, $2)
		 RETURNING id, ligue_id, nom`,
		ligueID, nom,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDivNameExists
		}
		return nil, fmt.Errorf("failed to create div: %w", err)
	}
	return d, nil
}

func (r *divRepo) FindByID(id int) (*Div, error) {
	d, err := scanDiv(r.db.QueryRow(
		`SELECT id, ligue_id, nom FROM divs WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDivNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find div by id: %w", err)
	}
	return d, nil
}

func (r *divRepo) FindByLigueID(ligueID int) ([]*Div, error) {
	rows, err := r.db.Query(
		`SELECT id, ligue_id, nom FROM divs WHERE ligue_id = $1 ORDER BY nom ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list divs by ligue: %w", err)
	}
	defer rows.Close()
	return scanDivs(rows)
}

func (r *divRepo) Update(id int, nom string) (*Div, error) {
	if nom == "" {
		return nil, ErrDivNameRequired
	}

	d, err := scanDiv(r.db.QueryRow(
		`UPDATE divs SET nom = $1 WHERE id = $2
		 RETURNING id, ligue_id, nom`,
		nom, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDivNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDivNameExists
		}
		return nil, fmt.Errorf("failed to update div: %w", err)
	}
	return d, nil
}

func (r *divRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM divs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete div: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDivNotFound
	}
	return nil
}

func (r *divRepo) AssignTeam(divID, equipeEvaID int) error {
	_, err := r.db.Exec(
		`INSERT INTO equipe_eva_div (equipe_eva_id, div_id, ligue_id)
		 VALUES ($1, $2, (SELECT ligue_id FROM divs WHERE id = $2))`,
		equipeEvaID, divID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var count int
			checkErr := r.db.QueryRow(
				`SELECT COUNT(*) FROM equipe_eva_div WHERE equipe_eva_id = $1 AND div_id = $2`,
				equipeEvaID, divID,
			).Scan(&count)
			if checkErr == nil && count > 0 {
				return nil
			}
			return ErrDivTeamAlreadyAssigned
		}
		return fmt.Errorf("failed to assign team to div: %w", err)
	}
	return nil
}

func (r *divRepo) RemoveTeam(divID, equipeEvaID int) error {
	result, err := r.db.Exec(
		`DELETE FROM equipe_eva_div WHERE equipe_eva_id = $1 AND div_id = $2`,
		equipeEvaID, divID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove team from div: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrEvaTeamNotFound
	}
	return nil
}

func (r *divRepo) FindTeamsByDivID(divID int) ([]*EvaTeam, error) {
	rows, err := r.db.Query(
		`SELECT e.id, e.name, e.tag, e.eva_trainer_team_id, e.created_at, e.updated_at
		 FROM eva_teams e
		 JOIN equipe_eva_div d ON d.equipe_eva_id = e.id
		 WHERE d.div_id = $1
		 ORDER BY e.name ASC`,
		divID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by div id: %w", err)
	}
	defer rows.Close()
	return scanEvaTeams(rows)
}

func (r *divRepo) FindTeamsByLigueID(ligueID int) (map[int][]*EvaTeam, error) {
	rows, err := r.db.Query(
		`SELECT eed.div_id, e.id, e.name, e.tag, e.eva_trainer_team_id, e.created_at, e.updated_at
		 FROM eva_teams e
		 JOIN equipe_eva_div eed ON eed.equipe_eva_id = e.id
		 WHERE eed.ligue_id = $1
		 ORDER BY eed.div_id, e.name ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by ligue id: %w", err)
	}
	defer rows.Close()

	result := make(map[int][]*EvaTeam)
	for rows.Next() {
		var divID int
		t := &EvaTeam{}
		err := rows.Scan(&divID, &t.ID, &t.Name, &t.Tag, &t.EvaTrainerTeamID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		result[divID] = append(result[divID], t)
	}
	return result, nil
}

func (r *divRepo) GetStandings(divID int) ([]*StandingEntry, error) {
	teams, err := r.FindTeamsByDivID(divID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return []*StandingEntry{}, nil
	}

	div, err := r.FindByID(divID)
	if err != nil {
		return nil, err
	}

	boRows, err := r.db.Query(
		`SELECT id, equipe1_id, equipe2_id, vainqueur_id FROM bos WHERE ligue_id = $1`,
		div.LigueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query bos: %w", err)
	}
	defer boRows.Close()

	type boData struct {
		ID          int
		Equipe1ID   int
		Equipe2ID   int
		VainqueurID *int
	}
	var allBos []boData
	for boRows.Next() {
		var b boData
		if err := boRows.Scan(&b.ID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID); err != nil {
			return nil, fmt.Errorf("failed to scan bo: %w", err)
		}
		allBos = append(allBos, b)
	}
	if err := boRows.Err(); err != nil {
		return nil, err
	}

	matchRows, err := r.db.Query(
		`SELECT m.bo_id, m.vainqueur_id
		 FROM matchs m
		 JOIN bos b ON b.id = m.bo_id
		 WHERE b.ligue_id = $1`,
		div.LigueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query matchs: %w", err)
	}
	defer matchRows.Close()

	matchCountByBO := make(map[int]int)
	matchWinsByTeamBO := make(map[int]map[int]int)
	for matchRows.Next() {
		var boID int
		var vainqueurID *int
		if err := matchRows.Scan(&boID, &vainqueurID); err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matchCountByBO[boID]++
		if vainqueurID != nil {
			if matchWinsByTeamBO[*vainqueurID] == nil {
				matchWinsByTeamBO[*vainqueurID] = make(map[int]int)
			}
			matchWinsByTeamBO[*vainqueurID][boID]++
		}
	}
	if err := matchRows.Err(); err != nil {
		return nil, err
	}

	entries := make([]*StandingEntry, 0, len(teams))
	for _, team := range teams {
		entry := &StandingEntry{
			TeamID:   team.ID,
			TeamName: team.Name,
			TeamTag:  team.Tag,
		}

		for _, b := range allBos {
			if b.Equipe1ID != team.ID && b.Equipe2ID != team.ID {
				continue
			}
			if b.VainqueurID == nil {
				continue
			}
			entry.BOsPlayed++
			if *b.VainqueurID == team.ID {
				entry.BOWon++
			} else {
				entry.BOLost++
			}

			totalMatchs := matchCountByBO[b.ID]
			entry.MatchesPlayed += totalMatchs
			wins := matchWinsByTeamBO[team.ID][b.ID]
			entry.MatchWon += wins
		}

		entry.BOPoints = entry.BOWon
		entry.MatchPoints = entry.MatchWon
		entry.MatchLost = entry.MatchesPlayed - entry.MatchWon - entry.MatchDrawn

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].BOPoints != entries[j].BOPoints {
			return entries[i].BOPoints > entries[j].BOPoints
		}
		if entries[i].MatchPoints != entries[j].MatchPoints {
			return entries[i].MatchPoints > entries[j].MatchPoints
		}
		return entries[i].TeamName < entries[j].TeamName
	})

	return entries, nil
}

func (r *divRepo) GenerateRoundRobin(divID int) ([]*BO, error) {
	div, err := r.FindByID(divID)
	if err != nil {
		return nil, fmt.Errorf("div not found: %w", err)
	}

	teams, err := r.FindTeamsByDivID(divID)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams: %w", err)
	}
	if len(teams) < 2 {
		return nil, ErrDivNotEnoughTeams
	}

	teamIDs := make([]int, len(teams))
	teamSet := make(map[int]bool, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
		teamSet[t.ID] = true
	}

	existingBOs, err := r.findBOsByTeamIDs(div.LigueID, teamSet)
	if err != nil {
		return nil, err
	}

	expected := generateRoundRobinPairs(teamIDs)

	existingPair := make(map[[2]int]*BO)
	for _, bo := range existingBOs {
		p := pair(bo.Equipe1ID, bo.Equipe2ID)
		existingPair[p] = bo
	}

	expectedSet := make(map[[2]int]bool, len(expected))
	for _, p := range expected {
		expectedSet[p] = true
	}

	for _, bo := range existingBOs {
		p := pair(bo.Equipe1ID, bo.Equipe2ID)
		if !expectedSet[p] {
			if err := r.deleteBO(bo.ID); err != nil {
				return nil, fmt.Errorf("failed to delete stale BO %d: %w", bo.ID, err)
			}
			delete(existingPair, p)
		}
	}

	for _, p := range expected {
		if _, exists := existingPair[p]; !exists {
			bo, err := r.createBO(div.LigueID, p[0], p[1])
			if err != nil {
				return nil, fmt.Errorf("failed to create BO for pair (%d,%d): %w", p[0], p[1], err)
			}
			existingPair[p] = bo
		}
	}

	result := make([]*BO, 0, len(existingPair))
	for _, bo := range existingPair {
		result = append(result, bo)
	}
	return result, nil
}

func pair(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

func generateRoundRobinPairs(teamIDs []int) [][2]int {
	n := len(teamIDs)
	hasBye := false
	if n%2 != 0 {
		n++
		hasBye = true
	}

	teams := make([]int, n)
	copy(teams, teamIDs)

	rounds := n - 1
	var pairs [][2]int
	seen := make(map[[2]int]bool)

	for round := 0; round < rounds; round++ {
		for i := 0; i < n/2; i++ {
			t1 := teams[i]
			t2 := teams[n-1-i]

			if hasBye && (t1 == 0 || t2 == 0) {
				continue
			}
			if t1 == 0 || t2 == 0 {
				continue
			}

			p := pair(t1, t2)
			if !seen[p] {
				seen[p] = true
				pairs = append(pairs, p)
			}
		}

		last := teams[n-1]
		for i := n - 1; i > 1; i-- {
			teams[i] = teams[i-1]
		}
		teams[1] = last
	}

	return pairs
}

func (r *divRepo) findBOsByTeamIDs(ligueID int, teamSet map[int]bool) ([]*BO, error) {
	rows, err := r.db.Query(
		`SELECT id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, created_at, updated_at
		 FROM bos WHERE ligue_id = $1
		 ORDER BY id ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find BOs by ligue: %w", err)
	}
	defer rows.Close()

	var result []*BO
	for rows.Next() {
		b := &BO{}
		err := rows.Scan(&b.ID, &b.LigueID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID, &b.Format, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan BO: %w", err)
		}
		if teamSet[b.Equipe1ID] || teamSet[b.Equipe2ID] {
			result = append(result, b)
		}
	}
	return result, rows.Err()
}

func (r *divRepo) createBO(ligueID, equipe1ID, equipe2ID int) (*BO, error) {
	b := &BO{}
	err := r.db.QueryRow(
		`INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, format, created_at, updated_at)
		 VALUES ($1, $2, $3, 'BO3', NOW(), NOW())
		 RETURNING id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, created_at, updated_at`,
		ligueID, equipe1ID, equipe2ID,
	).Scan(&b.ID, &b.LigueID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID, &b.Format, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create BO: %w", err)
	}
	return b, nil
}

func (r *divRepo) deleteBO(id int) error {
	result, err := r.db.Exec(`DELETE FROM bos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete BO: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBONotFound
	}
	return nil
}
