package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ZoneVersion struct {
	ID          int             `json:"id"`
	ZoneID      int             `json:"zone_id"`
	EditorID    *int            `json:"editor_id,omitempty"`
	VersionData VersionSnapshot `json:"version_data"`
	CreatedAt   time.Time       `json:"created_at"`
}

type VersionSnapshot struct {
	Polygons []SnapshotPolygon `json:"polygons"`
	Name     string            `json:"name"`
}

type SnapshotPolygon struct {
	Points     []Point `json:"points"`
	OrderIndex int     `json:"order_index"`
	Floor      *string `json:"floor,omitempty"`
}

type zoneVersionRepo struct {
	db *sql.DB
}

func NewZoneVersionRepository(db *sql.DB) ZoneVersionRepository {
	return &zoneVersionRepo{db: db}
}

func (r *zoneVersionRepo) Save(zoneID int, editorID *int, snapshot VersionSnapshot) (*ZoneVersion, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version snapshot: %w", err)
	}

	zv := &ZoneVersion{}
	var editorIDDB interface{}
	if editorID != nil {
		editorIDDB = *editorID
	}
	var versionDataDB []byte

	err = r.db.QueryRow(
		`INSERT INTO zone_versions (zone_id, editor_id, version_data)
		 VALUES ($1, $2, $3)
		 RETURNING id, zone_id, editor_id, version_data, created_at`,
		zoneID, editorIDDB, data,
	).Scan(&zv.ID, &zv.ZoneID, &zv.EditorID, &versionDataDB, &zv.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrZoneNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save zone version: %w", err)
	}

	if len(versionDataDB) > 0 {
		if err := json.Unmarshal(versionDataDB, &zv.VersionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal version data: %w", err)
		}
	}

	return zv, nil
}

func (r *zoneVersionRepo) FindByMap(teamID, mapID int, page, limit int) ([]*ZoneVersion, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM zone_versions zv
		 JOIN zones z ON z.id = zv.zone_id
		 WHERE z.team_id = $1 AND z.map_id = $2`,
		teamID, mapID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count zone versions: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT zv.id, zv.zone_id, zv.editor_id, zv.version_data, zv.created_at
		 FROM zone_versions zv
		 JOIN zones z ON z.id = zv.zone_id
		 WHERE z.team_id = $1 AND z.map_id = $2
		 ORDER BY zv.created_at DESC
		 LIMIT $3 OFFSET $4`,
		teamID, mapID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find zone versions by map: %w", err)
	}
	defer rows.Close()

	versions, err := scanVersionRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return versions, total, nil
}

func (r *zoneVersionRepo) FindByZone(zoneID int, page, limit int) ([]*ZoneVersion, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM zone_versions WHERE zone_id = $1`,
		zoneID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count zone versions: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT id, zone_id, editor_id, version_data, created_at
		 FROM zone_versions WHERE zone_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		zoneID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find zone versions by zone: %w", err)
	}
	defer rows.Close()

	versions, err := scanVersionRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return versions, total, nil
}

func (r *zoneVersionRepo) FindByID(versionID int) (*ZoneVersion, error) {
	zv := &ZoneVersion{}
	var versionDataDB []byte
	err := r.db.QueryRow(
		`SELECT id, zone_id, editor_id, version_data, created_at
		 FROM zone_versions WHERE id = $1`,
		versionID,
	).Scan(&zv.ID, &zv.ZoneID, &zv.EditorID, &versionDataDB, &zv.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find zone version by ID: %w", err)
	}

	if len(versionDataDB) > 0 {
		if err := json.Unmarshal(versionDataDB, &zv.VersionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal version data: %w", err)
		}
	}

	return zv, nil
}

func (r *zoneVersionRepo) GetLatestForZone(zoneID int) (*ZoneVersion, error) {
	zv := &ZoneVersion{}
	var versionDataDB []byte
	err := r.db.QueryRow(
		`SELECT id, zone_id, editor_id, version_data, created_at
		 FROM zone_versions WHERE zone_id = $1
		 ORDER BY created_at DESC LIMIT 1`,
		zoneID,
	).Scan(&zv.ID, &zv.ZoneID, &zv.EditorID, &versionDataDB, &zv.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest zone version: %w", err)
	}

	if len(versionDataDB) > 0 {
		if err := json.Unmarshal(versionDataDB, &zv.VersionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal version data: %w", err)
		}
	}

	return zv, nil
}

func (r *zoneVersionRepo) FindFirstOtherEditorVersion(zoneID, editorID int) (*int, error) {
	var versionID int
	err := r.db.QueryRow(
		`SELECT id FROM zone_versions
		 WHERE zone_id = $1 AND editor_id IS NOT NULL AND editor_id != $2
		 ORDER BY created_at ASC
		 LIMIT 1`,
		zoneID, editorID,
	).Scan(&versionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find first other editor version: %w", err)
	}
	return &versionID, nil
}

func (r *zoneVersionRepo) DeleteNewerThan(zoneID int, olderThan time.Time) error {
	_, err := r.db.Exec(
		`DELETE FROM zone_versions WHERE zone_id = $1 AND created_at > $2`,
		zoneID, olderThan,
	)
	if err != nil {
		return fmt.Errorf("failed to delete newer zone versions: %w", err)
	}
	return nil
}

func (r *zoneVersionRepo) CountByZone(zoneID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM zone_versions WHERE zone_id = $1`,
		zoneID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count zone versions: %w", err)
	}
	return count, nil
}

func scanVersionRows(rows *sql.Rows) ([]*ZoneVersion, error) {
	var versions []*ZoneVersion
	for rows.Next() {
		zv := &ZoneVersion{}
		var versionDataDB []byte
		if err := rows.Scan(&zv.ID, &zv.ZoneID, &zv.EditorID, &versionDataDB, &zv.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan zone version row: %w", err)
		}
		if len(versionDataDB) > 0 {
			if err := json.Unmarshal(versionDataDB, &zv.VersionData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal version data: %w", err)
			}
		}
		versions = append(versions, zv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if versions == nil {
		return []*ZoneVersion{}, nil
	}
	return versions, nil
}
