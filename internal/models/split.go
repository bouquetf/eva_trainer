package models

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Split struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	MapNames []string `json:"map_names"`
}

type splitRepo struct {
	db *sql.DB
}

func NewSplitRepository(db *sql.DB) SplitRepository {
	return &splitRepo{db: db}
}

func scanSplit(scanner interface {
	Scan(dest ...interface{}) error
}) (*Split, error) {
	s := &Split{}
	var mapNames []string

	err := scanner.Scan(&s.ID, &s.Name, pq.Array(&mapNames))
	if err != nil {
		return nil, err
	}

	s.MapNames = mapNames

	return s, nil
}

func (r *splitRepo) Create(name string, mapNames []string) (*Split, error) {
	if name == "" {
		return nil, ErrSplitNameRequired
	}
	if len(mapNames) == 0 {
		return nil, ErrSplitMapNamesRequired
	}

	m, err := scanSplit(r.db.QueryRow(
		`INSERT INTO splits (name, map_names)
		 VALUES ($1, $2)
		 RETURNING id, name, map_names`,
		name, pq.Array(mapNames),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create split: %w", err)
	}
	return m, nil
}

func (r *splitRepo) FindByID(id int) (*Split, error) {
	s, err := scanSplit(r.db.QueryRow(
		`SELECT id, name, map_names FROM splits WHERE id = $1`,
		id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrSplitNotFound
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *splitRepo) FindAll() ([]*Split, error) {
	rows, err := r.db.Query(
		`SELECT id, name, map_names FROM splits ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []*Split
	for rows.Next() {
		s, err := scanSplit(rows)
		if err != nil {
			return nil, err
		}
		splits = append(splits, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if splits == nil {
		return []*Split{}, nil
	}
	return splits, nil
}

func (r *splitRepo) Update(id int, name string, mapNames []string) (*Split, error) {
	if name == "" {
		return nil, ErrSplitNameRequired
	}
	if len(mapNames) == 0 {
		return nil, ErrSplitMapNamesRequired
	}

	s, err := scanSplit(r.db.QueryRow(
		`UPDATE splits SET name = $1, map_names = $2
		 WHERE id = $3
		 RETURNING id, name, map_names`,
		name, pq.Array(mapNames), id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrSplitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update split: %w", err)
	}
	return s, nil
}

func (r *splitRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM splits WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete split: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSplitNotFound
	}
	return nil
}

func (r *splitRepo) FindMapsBySplitID(splitID int) ([]*Map, error) {
	split, err := r.FindByID(splitID)
	if err != nil {
		return nil, err
	}

	if len(split.MapNames) == 0 {
		return []*Map{}, nil
	}

	rows, err := r.db.Query(
		`SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps WHERE name = ANY($1)`,
		pq.Array(split.MapNames),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find maps by split id: %w", err)
	}
	defer rows.Close()

	var maps []*Map
	for rows.Next() {
		m, err := scanMap(rows)
		if err != nil {
			return nil, err
		}
		maps = append(maps, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if maps == nil {
		return []*Map{}, nil
	}
	return maps, nil
}
