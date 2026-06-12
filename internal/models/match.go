package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Match struct {
	ID              int       `json:"id"`
	BoID            int       `json:"bo_id"`
	MapID           int       `json:"map_id"`
	ScoreEquipe1    int       `json:"score_equipe1"`
	ScoreEquipe2    int       `json:"score_equipe2"`
	Forfait         int       `json:"forfait"`
	VainqueurID     *int      `json:"vainqueur_id,omitempty"`
	Commentaires    *string   `json:"commentaires,omitempty"`
	HorsCompetition bool      `json:"hors_competition"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type matchRepo struct {
	db *sql.DB
}

func NewMatchRepository(db *sql.DB) MatchRepository {
	return &matchRepo{db: db}
}

func scanMatch(scanner interface {
	Scan(dest ...interface{}) error
}) (*Match, error) {
	m := &Match{}
	err := scanner.Scan(&m.ID, &m.BoID, &m.MapID, &m.ScoreEquipe1, &m.ScoreEquipe2, &m.Forfait, &m.VainqueurID, &m.Commentaires, &m.HorsCompetition, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func scanMatchs(rows *sql.Rows) ([]*Match, error) {
	var matchs []*Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		matchs = append(matchs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if matchs == nil {
		return []*Match{}, nil
	}
	return matchs, nil
}

func (r *matchRepo) Create(boID, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error) {
	var vainqueurID *int
	if forfait == 1 {
		var eq2ID int
		err := r.db.QueryRow(`SELECT equipe2_id FROM bos WHERE id = $1`, boID).Scan(&eq2ID)
		if err == nil {
			vainqueurID = &eq2ID
		}
	} else if forfait == 2 {
		var eq1ID int
		err := r.db.QueryRow(`SELECT equipe1_id FROM bos WHERE id = $1`, boID).Scan(&eq1ID)
		if err == nil {
			vainqueurID = &eq1ID
		}
	} else if scoreEquipe1 > scoreEquipe2 {
		var eq1ID int
		err := r.db.QueryRow(`SELECT equipe1_id FROM bos WHERE id = $1`, boID).Scan(&eq1ID)
		if err == nil {
			vainqueurID = &eq1ID
		}
	} else if scoreEquipe2 > scoreEquipe1 {
		var eq2ID int
		err := r.db.QueryRow(`SELECT equipe2_id FROM bos WHERE id = $1`, boID).Scan(&eq2ID)
		if err == nil {
			vainqueurID = &eq2ID
		}
	}

	m, err := scanMatch(r.db.QueryRow(
		`INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at`,
		boID, mapID, scoreEquipe1, scoreEquipe2, forfait, vainqueurID, commentaires, horsCompetition,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}
	return m, nil
}

func (r *matchRepo) FindByID(id int) (*Match, error) {
	m, err := scanMatch(r.db.QueryRow(
		`SELECT id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at
		 FROM matchs WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find match by id: %w", err)
	}
	return m, nil
}

func (r *matchRepo) FindByBOID(boID int) ([]*Match, error) {
	rows, err := r.db.Query(
		`SELECT id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at
		 FROM matchs WHERE bo_id = $1
		 ORDER BY id ASC`,
		boID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find matchs by bo id: %w", err)
	}
	defer rows.Close()
	return scanMatchs(rows)
}

func (r *matchRepo) Update(id, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error) {
	var vainqueurID *int
	if forfait == 1 {
		err := r.db.QueryRow(
			`SELECT b.equipe2_id FROM matchs m JOIN bos b ON b.id = m.bo_id WHERE m.id = $1`,
			id,
		).Scan(&vainqueurID)
		if err != nil {
			vainqueurID = nil
		}
	} else if forfait == 2 {
		err := r.db.QueryRow(
			`SELECT b.equipe1_id FROM matchs m JOIN bos b ON b.id = m.bo_id WHERE m.id = $1`,
			id,
		).Scan(&vainqueurID)
		if err != nil {
			vainqueurID = nil
		}
	} else if scoreEquipe1 > scoreEquipe2 {
		err := r.db.QueryRow(
			`SELECT b.equipe1_id FROM matchs m JOIN bos b ON b.id = m.bo_id WHERE m.id = $1`,
			id,
		).Scan(&vainqueurID)
		if err != nil {
			vainqueurID = nil
		}
	} else if scoreEquipe2 > scoreEquipe1 {
		err := r.db.QueryRow(
			`SELECT b.equipe2_id FROM matchs m JOIN bos b ON b.id = m.bo_id WHERE m.id = $1`,
			id,
		).Scan(&vainqueurID)
		if err != nil {
			vainqueurID = nil
		}
	}

	m, err := scanMatch(r.db.QueryRow(
		`UPDATE matchs SET map_id = $1, score_equipe1 = $2, score_equipe2 = $3, forfait = $4, vainqueur_id = $5, commentaires = $6, hors_competition = $7, updated_at = NOW()
		 WHERE id = $8
		 RETURNING id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at`,
		mapID, scoreEquipe1, scoreEquipe2, forfait, vainqueurID, commentaires, horsCompetition, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update match: %w", err)
	}
	return m, nil
}

func (r *matchRepo) DeleteByBOID(boID int) error {
	_, err := r.db.Exec(`DELETE FROM matchs WHERE bo_id = $1`, boID)
	if err != nil {
		return fmt.Errorf("failed to delete matchs by bo id: %w", err)
	}
	return nil
}

func (r *matchRepo) SetHorsCompetition(id int, horsCompetition bool) (*Match, error) {
	m, err := scanMatch(r.db.QueryRow(
		`UPDATE matchs SET hors_competition = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at`,
		horsCompetition, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to set hors_competition: %w", err)
	}
	return m, nil
}

func (r *matchRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM matchs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete match: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMatchNotFound
	}
	return nil
}
