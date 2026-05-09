package models

import (
	"database/sql"
	"fmt"
	"time"
)

type MapCatalogEntry struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type mapCatalogRepo struct {
	db *sql.DB
}

func NewMapCatalogRepository(db *sql.DB) *mapCatalogRepo {
	return &mapCatalogRepo{db: db}
}

func (r *mapCatalogRepo) FindAll() ([]*MapCatalogEntry, error) {
	rows, err := r.db.Query(`SELECT id, name, created_at FROM map_catalog ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query map catalog: %w", err)
	}
	defer rows.Close()

	var entries []*MapCatalogEntry
	for rows.Next() {
		var e MapCatalogEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan map catalog entry: %w", err)
		}
		entries = append(entries, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []*MapCatalogEntry{}
	}
	return entries, nil
}

func (r *mapCatalogRepo) Create(name string) (*MapCatalogEntry, error) {
	var e MapCatalogEntry
	err := r.db.QueryRow(
		`INSERT INTO map_catalog (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&e.ID, &e.Name, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create map catalog entry: %w", err)
	}
	return &e, nil
}

func (r *mapCatalogRepo) Update(id int, name string) (*MapCatalogEntry, error) {
	var e MapCatalogEntry
	err := r.db.QueryRow(
		`UPDATE map_catalog SET name = $1 WHERE id = $2 RETURNING id, name, created_at`,
		name, id,
	).Scan(&e.ID, &e.Name, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("map catalog entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update map catalog entry: %w", err)
	}
	return &e, nil
}

func (r *mapCatalogRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM map_catalog WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete map catalog entry: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("map catalog entry not found")
	}
	return nil
}
