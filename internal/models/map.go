package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Map struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	WebpBlob   []byte    `json:"-"`
	AddedAt    time.Time `json:"added_at"`
	ObsoleteOf *int      `json:"obsolete_of,omitempty"`
	IsObsolete bool      `json:"is_obsolete"`
}

type mapRepo struct {
	db *sql.DB
}

func NewMapRepository(db *sql.DB) MapRepository {
	return &mapRepo{db: db}
}

func scanMap(scanner interface {
	Scan(dest ...interface{}) error
}) (*Map, error) {
	m := &Map{}
	var obsoleteOf sql.NullInt32

	err := scanner.Scan(&m.ID, &m.Name, &m.AddedAt, &obsoleteOf, &m.IsObsolete)
	if err != nil {
		return nil, err
	}

	if obsoleteOf.Valid {
		v := int(obsoleteOf.Int32)
		m.ObsoleteOf = &v
	}

	return m, nil
}

func (r *mapRepo) Create(name string, blobData []byte) (*Map, error) {
	if name == "" {
		return nil, ErrMapNameRequired
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int
	hasExisting := false
	err = tx.QueryRow(
		`SELECT id FROM maps WHERE name = $1 AND is_obsolete = FALSE LIMIT 1`,
		name,
	).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing map: %w", err)
	}
	if err == nil {
		hasExisting = true
	}

	m, err := scanMap(tx.QueryRow(
		`INSERT INTO maps (name, webp_blob) VALUES ($1, $2)
		 RETURNING id, name, added_at, obsolete_of, is_obsolete`,
		name, blobData,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create map: %w", err)
	}

	if hasExisting {
		_, err = tx.Exec(
			`UPDATE maps SET is_obsolete = TRUE, obsolete_of = $1 WHERE id = $2`,
			m.ID, existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to mark existing map as obsolete: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return m, nil
}

func (r *mapRepo) FindByID(id int) (*Map, error) {
	m, err := scanMap(r.db.QueryRow(
		`SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps WHERE id = $1`,
		id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrMapNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *mapRepo) FindAll() ([]*Map, error) {
	rows, err := r.db.Query(
		`SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps ORDER BY added_at DESC`,
	)
	if err != nil {
		return nil, err
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

func (r *mapRepo) FindByName(name string) ([]*Map, error) {
	rows, err := r.db.Query(
		`SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps WHERE name = $1 ORDER BY added_at DESC`,
		name,
	)
	if err != nil {
		return nil, err
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

func (r *mapRepo) FindByNameLatest(name string) (*Map, error) {
	m, err := scanMap(r.db.QueryRow(
		`SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps
		 WHERE name = $1
		 ORDER BY is_obsolete ASC, added_at DESC
		 LIMIT 1`,
		name,
	))
	if err == sql.ErrNoRows {
		return nil, ErrMapNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *mapRepo) MarkObsolete(mapID int, replacedByMapID int) error {
	result, err := r.db.Exec(
		`UPDATE maps SET is_obsolete = TRUE, obsolete_of = $1 WHERE id = $2`,
		replacedByMapID, mapID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark map as obsolete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMapNotFound
	}
	return nil
}

func (r *mapRepo) IsMapObsolete(mapID int) (bool, error) {
	var isObsolete bool
	err := r.db.QueryRow(
		`SELECT is_obsolete FROM maps WHERE id = $1`,
		mapID,
	).Scan(&isObsolete)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isObsolete, nil
}

func (r *mapRepo) FindObsoleteWithoutZones() ([]*Map, error) {
	rows, err := r.db.Query(
		`SELECT m.id, m.name, m.added_at, m.obsolete_of, m.is_obsolete
		 FROM maps m
		 LEFT JOIN zones z ON m.id = z.map_id
		 WHERE m.is_obsolete = TRUE AND z.id IS NULL`,
	)
	if err != nil {
		return nil, err
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

func (r *mapRepo) Delete(mapID int) error {
	result, err := r.db.Exec(`DELETE FROM maps WHERE id = $1`, mapID)
	if err != nil {
		return fmt.Errorf("failed to delete map: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMapNotFound
	}
	return nil
}

func (r *mapRepo) GetBlob(mapID int) ([]byte, error) {
	var blob []byte
	err := r.db.QueryRow(
		`SELECT webp_blob FROM maps WHERE id = $1`,
		mapID,
	).Scan(&blob)
	if err == sql.ErrNoRows {
		return nil, ErrMapNotFound
	}
	if err != nil {
		return nil, err
	}
	return blob, nil
}
