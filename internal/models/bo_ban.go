package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Ban struct {
	ID        int       `json:"id"`
	BoID      int       `json:"bo_id"`
	EquipeID  int       `json:"equipe_id"`
	MapID     int       `json:"map_id"`
	CreatedAt time.Time `json:"created_at"`
}

type banRepo struct {
	db *sql.DB
}

func NewBanRepository(db *sql.DB) BanRepository {
	return &banRepo{db: db}
}

func scanBan(scanner interface {
	Scan(dest ...interface{}) error
}) (*Ban, error) {
	b := &Ban{}
	err := scanner.Scan(&b.ID, &b.BoID, &b.EquipeID, &b.MapID, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func scanBans(rows *sql.Rows) ([]*Ban, error) {
	var bans []*Ban
	for rows.Next() {
		b, err := scanBan(rows)
		if err != nil {
			return nil, err
		}
		bans = append(bans, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if bans == nil {
		return []*Ban{}, nil
	}
	return bans, nil
}

func (r *banRepo) Create(boID, equipeID, mapID int) (*Ban, error) {
	b, err := scanBan(r.db.QueryRow(
		`INSERT INTO bo_bans (bo_id, equipe_id, map_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, bo_id, equipe_id, map_id, created_at`,
		boID, equipeID, mapID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create ban: %w", err)
	}
	return b, nil
}

func (r *banRepo) FindByID(id int) (*Ban, error) {
	b, err := scanBan(r.db.QueryRow(
		`SELECT id, bo_id, equipe_id, map_id, created_at
		 FROM bo_bans WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ban by id: %w", err)
	}
	return b, nil
}

func (r *banRepo) FindByBOID(boID int) ([]*Ban, error) {
	rows, err := r.db.Query(
		`SELECT id, bo_id, equipe_id, map_id, created_at
		 FROM bo_bans WHERE bo_id = $1
		 ORDER BY id ASC`,
		boID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find bans by bo id: %w", err)
	}
	defer rows.Close()
	return scanBans(rows)
}

func (r *banRepo) DeleteByBOID(boID int) error {
	_, err := r.db.Exec(`DELETE FROM bo_bans WHERE bo_id = $1`, boID)
	if err != nil {
		return fmt.Errorf("failed to delete bans by bo id: %w", err)
	}
	return nil
}

func (r *banRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM bo_bans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ban: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBanNotFound
	}
	return nil
}
