package models

import (
	"database/sql"
	"fmt"
	"sort"
)

type HeadToHeadTeamStats struct {
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Winrate     float64 `json:"winrate"`
	AvgScore    float64 `json:"avg_score"`
	MedianScore float64 `json:"median_score"`
}

type HeadToHeadMapStats struct {
	MapID      int                 `json:"map_id"`
	MapName    string              `json:"map_name"`
	Team1Stats HeadToHeadTeamStats `json:"team1_stats"`
	Team2Stats HeadToHeadTeamStats `json:"team2_stats"`
}

type headToHeadStatsRepo struct {
	db *sql.DB
}

func NewHeadToHeadStatsRepository(db *sql.DB) HeadToHeadStatsRepository {
	return &headToHeadStatsRepo{db: db}
}

type rawH2HRow struct {
	mapID      int
	mapName    string
	team1Score int
	team2Score int
	team1Win   bool
	team2Win   bool
	isDraw     bool
}

func (r *headToHeadStatsRepo) Compute(team1ID, team2ID int) ([]*HeadToHeadMapStats, error) {
	if team1ID == team2ID {
		return nil, ErrInvalidTeamPair
	}

	rows, err := r.db.Query(`
		SELECT m.map_id, mp.name AS map_name,
		       m.score_equipe1, m.score_equipe2,
		       m.vainqueur_id,
		       b.equipe1_id, b.equipe2_id
		FROM matchs m
		JOIN bos b ON b.id = m.bo_id
		JOIN maps mp ON mp.id = m.map_id
		WHERE ((b.equipe1_id = $1 AND b.equipe2_id = $2)
		   OR (b.equipe1_id = $2 AND b.equipe2_id = $1))
		  AND m.forfait = 0
		ORDER BY mp.name ASC, m.id ASC`, team1ID, team2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query head-to-head matchs: %w", err)
	}
	defer rows.Close()

	var rawRows []rawH2HRow
	for rows.Next() {
		var mapID int
		var mapName string
		var score1, score2 int
		var vainqueurID sql.NullInt64
		var equipe1ID, equipe2ID int

		if err := rows.Scan(&mapID, &mapName, &score1, &score2, &vainqueurID, &equipe1ID, &equipe2ID); err != nil {
			return nil, fmt.Errorf("failed to scan head-to-head row: %w", err)
		}

		var team1Score, team2Score int
		if equipe1ID == team1ID {
			team1Score = score1
			team2Score = score2
		} else {
			team1Score = score2
			team2Score = score1
		}

		row := rawH2HRow{
			mapID:      mapID,
			mapName:    mapName,
			team1Score: team1Score,
			team2Score: team2Score,
		}

		if vainqueurID.Valid {
			row.team1Win = int(vainqueurID.Int64) == team1ID
			row.team2Win = int(vainqueurID.Int64) == team2ID
		} else {
			row.isDraw = true
		}

		rawRows = append(rawRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate head-to-head rows: %w", err)
	}

	type mapAgg struct {
		mapName     string
		team1Scores []int
		team2Scores []int
		team1Wins   int
		team1Losses int
		team2Wins   int
		team2Losses int
	}

	mapData := make(map[int]*mapAgg)

	for _, r := range rawRows {
		ma, ok := mapData[r.mapID]
		if !ok {
			ma = &mapAgg{mapName: r.mapName}
			mapData[r.mapID] = ma
		}
		ma.team1Scores = append(ma.team1Scores, r.team1Score)
		ma.team2Scores = append(ma.team2Scores, r.team2Score)

		if !r.isDraw {
			if r.team1Win {
				ma.team1Wins++
				ma.team2Losses++
			} else {
				ma.team2Wins++
				ma.team1Losses++
			}
		}
	}

	result := make([]*HeadToHeadMapStats, 0, len(mapData))
	for mapID, ma := range mapData {
		result = append(result, &HeadToHeadMapStats{
			MapID:   mapID,
			MapName: ma.mapName,
			Team1Stats: HeadToHeadTeamStats{
				Wins:        ma.team1Wins,
				Losses:      ma.team1Losses,
				Winrate:     ComputeWinrate(ma.team1Wins, ma.team1Losses),
				AvgScore:    average(ma.team1Scores),
				MedianScore: ComputeMedian(ma.team1Scores),
			},
			Team2Stats: HeadToHeadTeamStats{
				Wins:        ma.team2Wins,
				Losses:      ma.team2Losses,
				Winrate:     ComputeWinrate(ma.team2Wins, ma.team2Losses),
				AvgScore:    average(ma.team2Scores),
				MedianScore: ComputeMedian(ma.team2Scores),
			},
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MapName < result[j].MapName
	})

	return result, nil
}
