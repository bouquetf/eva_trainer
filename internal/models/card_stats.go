package models

import (
	"database/sql"
	"fmt"
	"sort"
)

type OpponentCardStats struct {
	TeamID      int     `json:"team_id"`
	TeamName    string  `json:"team_name"`
	AvgScore    float64 `json:"avg_score"`
	MedianScore float64 `json:"median_score"`
}

type CardStats struct {
	MapID       int                 `json:"map_id"`
	MapName     string              `json:"map_name"`
	Wins        int                 `json:"wins"`
	Losses      int                 `json:"losses"`
	Winrate     float64             `json:"winrate"`
	AvgScore    float64             `json:"avg_score"`
	MedianScore float64             `json:"median_score"`
	PickRate    float64             `json:"pick_rate"`
	BanRate     float64             `json:"ban_rate"`
	Opponents   []OpponentCardStats `json:"opponents"`
}

type cardStatsRepo struct {
	db *sql.DB
}

func NewCardStatsRepository(db *sql.DB) CardStatsRepository {
	return &cardStatsRepo{db: db}
}

func ComputeMedian(scores []int) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	sorted := make([]int, len(scores))
	copy(sorted, scores)
	sort.Ints(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return float64(sorted[n/2-1]+sorted[n/2]) / 2.0
	}
	return float64(sorted[n/2])
}

func ComputeWinrate(wins, losses int) float64 {
	total := wins + losses
	if total == 0 {
		return 0.0
	}
	return float64(wins) / float64(total)
}

type rawMatchRow struct {
	mapID        int
	mapName      string
	teamScore    int
	opponentID   int
	opponentName string
	isWin        bool
	isDraw       bool
	pickCount    int
	banCount     int
}

func (r *cardStatsRepo) Compute(teamID int) ([]*CardStats, error) {
	var totalPicks, totalBans int
	r.db.QueryRow(`SELECT COUNT(*) FROM bo_picks WHERE equipe_id = $1`, teamID).Scan(&totalPicks)
	r.db.QueryRow(`SELECT COUNT(*) FROM bo_bans WHERE equipe_id = $1`, teamID).Scan(&totalBans)

	rows, err := r.db.Query(`
		SELECT m.map_id, mp.name AS map_name,
		       m.score_equipe1, m.score_equipe2,
		       m.vainqueur_id,
		       b.equipe1_id, b.equipe2_id,
		       e1.name AS equipe1_name, e2.name AS equipe2_name,
		       COALESCE(picks.pick_count, 0) AS pick_count,
		       COALESCE(bans.ban_count, 0) AS ban_count
		FROM matchs m
		JOIN bos b ON b.id = m.bo_id
		JOIN maps mp ON mp.id = m.map_id
		JOIN eva_teams e1 ON e1.id = b.equipe1_id
		JOIN eva_teams e2 ON e2.id = b.equipe2_id
		LEFT JOIN (
			SELECT bp.map_id, COUNT(*) AS pick_count
			FROM bo_picks bp
			WHERE bp.equipe_id = $1
			GROUP BY bp.map_id
		) picks ON picks.map_id = mp.id
		LEFT JOIN (
			SELECT bb.map_id, COUNT(*) AS ban_count
			FROM bo_bans bb
			WHERE bb.equipe_id = $1
			GROUP BY bb.map_id
		) bans ON bans.map_id = mp.id
		WHERE (b.equipe1_id = $1 OR b.equipe2_id = $1)
		  AND m.forfait = 0
		ORDER BY mp.name ASC, m.id ASC`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query matchs for card stats: %w", err)
	}
	defer rows.Close()

	var rawRows []rawMatchRow
	for rows.Next() {
		var mapID int
		var mapName string
		var score1, score2 int
		var vainqueurID sql.NullInt64
		var equipe1ID, equipe2ID int
		var equipe1Name, equipe2Name string
		var pickCount, banCount int

		if err := rows.Scan(&mapID, &mapName, &score1, &score2, &vainqueurID, &equipe1ID, &equipe2ID, &equipe1Name, &equipe2Name, &pickCount, &banCount); err != nil {
			return nil, fmt.Errorf("failed to scan match row: %w", err)
		}

		var teamScore int
		var opponentID int
		var opponentName string
		if equipe1ID == teamID {
			teamScore = score1
			opponentID = equipe2ID
			opponentName = equipe2Name
		} else {
			teamScore = score2
			opponentID = equipe1ID
			opponentName = equipe1Name
		}

		row := rawMatchRow{
			mapID:        mapID,
			mapName:      mapName,
			teamScore:    teamScore,
			opponentID:   opponentID,
			opponentName: opponentName,
			pickCount:    pickCount,
			banCount:     banCount,
		}

		if vainqueurID.Valid {
			row.isWin = int(vainqueurID.Int64) == teamID
		} else {
			row.isDraw = true
		}

		rawRows = append(rawRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate match rows: %w", err)
	}

	type mapAgg struct {
		mapName string
		wins    int
		losses  int
		scores  []int
		picks   int
		bans    int
	}
	type opponentAgg struct {
		name   string
		scores []int
	}

	mapData := make(map[int]*mapAgg)
	opponentData := make(map[int]map[int]*opponentAgg)

	for _, r := range rawRows {
		ma, ok := mapData[r.mapID]
		if !ok {
			ma = &mapAgg{mapName: r.mapName, picks: r.pickCount, bans: r.banCount}
			mapData[r.mapID] = ma
		}
		if !r.isDraw {
			if r.isWin {
				ma.wins++
			} else {
				ma.losses++
			}
		}
		ma.scores = append(ma.scores, r.teamScore)

		if opponentData[r.mapID] == nil {
			opponentData[r.mapID] = make(map[int]*opponentAgg)
		}
		oa, ok := opponentData[r.mapID][r.opponentID]
		if !ok {
			oa = &opponentAgg{name: r.opponentName}
			opponentData[r.mapID][r.opponentID] = oa
		}
		oa.scores = append(oa.scores, r.teamScore)
	}

	allMapRows, err := r.db.Query(`SELECT id, name FROM maps ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all maps: %w", err)
	}
	for allMapRows.Next() {
		var mid int
		var mname string
		if err := allMapRows.Scan(&mid, &mname); err != nil {
			allMapRows.Close()
			return nil, fmt.Errorf("failed to scan all-maps row: %w", err)
		}
		if _, ok := mapData[mid]; !ok {
			mapData[mid] = &mapAgg{mapName: mname}
		}
	}
	allMapRows.Close()
	if err := allMapRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate all-maps rows: %w", err)
	}

	// Fetch pick counts per map for the team (also covers maps with picks but no matches)
	pickRows, err := r.db.Query(`
		SELECT bp.map_id, COUNT(*) AS pick_count
		FROM bo_picks bp
		JOIN bos b ON b.id = bp.bo_id
		WHERE bp.equipe_id = $1
		  AND (b.equipe1_id = $1 OR b.equipe2_id = $1)
		GROUP BY bp.map_id
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pick counts: %w", err)
	}
	for pickRows.Next() {
		var mapID int
		var count int
		if err := pickRows.Scan(&mapID, &count); err != nil {
			pickRows.Close()
			return nil, fmt.Errorf("failed to scan pick count row: %w", err)
		}
		if ma, ok := mapData[mapID]; ok {
			ma.picks = count
		}
	}
	pickRows.Close()

	// Fetch ban counts per map for the team (also covers maps with bans but no matches)
	banRows, err := r.db.Query(`
		SELECT bb.map_id, COUNT(*) AS ban_count
		FROM bo_bans bb
		JOIN bos b ON b.id = bb.bo_id
		WHERE bb.equipe_id = $1
		  AND (b.equipe1_id = $1 OR b.equipe2_id = $1)
		GROUP BY bb.map_id
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ban counts: %w", err)
	}
	for banRows.Next() {
		var mapID int
		var count int
		if err := banRows.Scan(&mapID, &count); err != nil {
			banRows.Close()
			return nil, fmt.Errorf("failed to scan ban count row: %w", err)
		}
		if ma, ok := mapData[mapID]; ok {
			ma.bans = count
		}
	}
	banRows.Close()

	result := make([]*CardStats, 0, len(mapData))
	for mapID, ma := range mapData {
		cs := &CardStats{
			MapID:       mapID,
			MapName:     ma.mapName,
			Wins:        ma.wins,
			Losses:      ma.losses,
			Winrate:     ComputeWinrate(ma.wins, ma.losses),
			AvgScore:    average(ma.scores),
			MedianScore: ComputeMedian(ma.scores),
			PickRate:    computeRate(ma.picks, totalPicks),
			BanRate:     computeRate(ma.bans, totalBans),
			Opponents:   make([]OpponentCardStats, 0),
		}

		opponents := opponentData[mapID]
		if len(opponents) > 0 {
			cs.Opponents = make([]OpponentCardStats, 0, len(opponents))
			for oppID, oa := range opponents {
				cs.Opponents = append(cs.Opponents, OpponentCardStats{
					TeamID:      oppID,
					TeamName:    oa.name,
					AvgScore:    average(oa.scores),
					MedianScore: ComputeMedian(oa.scores),
				})
			}
			sort.Slice(cs.Opponents, func(i, j int) bool {
				return cs.Opponents[i].TeamName < cs.Opponents[j].TeamName
			})
		}

		result = append(result, cs)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MapName < result[j].MapName
	})

	return result, nil
}

func computeRate(value, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(value) / float64(total)
}

func average(scores []int) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return float64(sum) / float64(len(scores))
}
