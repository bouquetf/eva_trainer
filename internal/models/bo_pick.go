package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Pick struct {
	ID        int       `json:"id"`
	BoID      int       `json:"bo_id"`
	EquipeID  int       `json:"equipe_id"`
	MapID     int       `json:"map_id"`
	CreatedAt time.Time `json:"created_at"`
}

type pickRepo struct {
	db *sql.DB
}

func NewPickRepository(db *sql.DB) PickRepository {
	return &pickRepo{db: db}
}

func scanPick(scanner interface {
	Scan(dest ...interface{}) error
}) (*Pick, error) {
	p := &Pick{}
	err := scanner.Scan(&p.ID, &p.BoID, &p.EquipeID, &p.MapID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func scanPicks(rows *sql.Rows) ([]*Pick, error) {
	var picks []*Pick
	for rows.Next() {
		p, err := scanPick(rows)
		if err != nil {
			return nil, err
		}
		picks = append(picks, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if picks == nil {
		return []*Pick{}, nil
	}
	return picks, nil
}

func (r *pickRepo) Create(boID, equipeID, mapID int) (*Pick, error) {
	p, err := scanPick(r.db.QueryRow(
		`INSERT INTO bo_picks (bo_id, equipe_id, map_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, bo_id, equipe_id, map_id, created_at`,
		boID, equipeID, mapID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create pick: %w", err)
	}
	return p, nil
}

func (r *pickRepo) FindByID(id int) (*Pick, error) {
	p, err := scanPick(r.db.QueryRow(
		`SELECT id, bo_id, equipe_id, map_id, created_at
		 FROM bo_picks WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPickNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find pick by id: %w", err)
	}
	return p, nil
}

func (r *pickRepo) FindByBOID(boID int) ([]*Pick, error) {
	rows, err := r.db.Query(
		`SELECT id, bo_id, equipe_id, map_id, created_at
		 FROM bo_picks WHERE bo_id = $1
		 ORDER BY id ASC`,
		boID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find picks by bo id: %w", err)
	}
	defer rows.Close()
	return scanPicks(rows)
}

func (r *pickRepo) DeleteByBOID(boID int) error {
	_, err := r.db.Exec(`DELETE FROM bo_picks WHERE bo_id = $1`, boID)
	if err != nil {
		return fmt.Errorf("failed to delete picks by bo id: %w", err)
	}
	return nil
}

func (r *pickRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM bo_picks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete pick: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPickNotFound
	}
	return nil
}
