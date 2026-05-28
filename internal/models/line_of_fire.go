package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Cone struct {
	Origin   Point    `json:"origin"`
	LeftEnd  Point    `json:"left_end"`
	RightEnd Point    `json:"right_end"`
	ArcBulge *float64 `json:"arc_bulge,omitempty"`
}

type ConeData struct {
	Cones []Cone `json:"cones"`
}

type LineOfFire struct {
	ID          int       `json:"id"`
	TeamID      int       `json:"team_id"`
	MapID       int       `json:"map_id"`
	Name        string    `json:"name"`
	PolygonData ConeData  `json:"polygon_data"`
	OrderIndex  int       `json:"order_index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TemplateLineOfFire struct {
	ID          int      `json:"id"`
	TemplateID  int      `json:"template_id"`
	PolygonData ConeData `json:"polygon_data"`
	Name        string   `json:"name"`
	OrderIndex  int      `json:"order_index"`
}

type LineOfFireVersion struct {
	ID           int             `json:"id"`
	LineOfFireID int             `json:"line_of_fire_id"`
	EditorID     *int            `json:"editor_id,omitempty"`
	VersionData  json.RawMessage `json:"version_data"`
	CreatedAt    time.Time       `json:"created_at"`
}

type lineOfFireRepo struct {
	db *sql.DB
}

func NewLineOfFireRepository(db *sql.DB) LineOfFireRepository {
	return &lineOfFireRepo{db: db}
}

type lineOfFireVersionRepo struct {
	db *sql.DB
}

func NewLineOfFireVersionRepository(db *sql.DB) LineOfFireVersionRepository {
	return &lineOfFireVersionRepo{db: db}
}

func marshalConeData(data ConeData) ([]byte, error) {
	return json.Marshal(data)
}

func unmarshalConeData(data []byte) (*ConeData, error) {
	if data == nil {
		return &ConeData{Cones: []Cone{}}, nil
	}
	var cd ConeData
	if err := json.Unmarshal(data, &cd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cone data: %w", err)
	}
	if cd.Cones == nil {
		cd.Cones = []Cone{}
	}
	return &cd, nil
}

func scanLineOfFire(scanner interface {
	Scan(dest ...interface{}) error
}) (*LineOfFire, error) {
	l := &LineOfFire{}
	var data []byte
	err := scanner.Scan(&l.ID, &l.TeamID, &l.MapID, &l.Name, &data, &l.OrderIndex, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	cd, err := unmarshalConeData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cone data for line_of_fire %d: %w", l.ID, err)
	}
	l.PolygonData = *cd
	return l, nil
}

func (r *lineOfFireRepo) Create(teamID, mapID int, name string, data ConeData) (*LineOfFire, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrLineOfFireNameRequired
	}
	if len(name) > 250 {
		return nil, ErrLineOfFireNameTooLong
	}
	if len(data.Cones) == 0 {
		return nil, ErrLineOfFireNoCones
	}
	for i, c := range data.Cones {
		if c.Origin.X == 0 && c.Origin.Y == 0 {
			return nil, fmt.Errorf("%w: cone %d has no origin", ErrConeOriginRequired, i)
		}
	}

	taken, err := r.IsNameTaken(teamID, mapID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if taken {
		return nil, ErrLineOfFireNameTaken
	}

	raw, err := marshalConeData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cone data: %w", err)
	}

	l, err := scanLineOfFire(r.db.QueryRow(
		`INSERT INTO lines_of_fire (team_id, map_id, name, polygon_data)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, team_id, map_id, name, polygon_data, order_index, created_at, updated_at`,
		teamID, mapID, name, raw,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create line of fire: %w", err)
	}
	return l, nil
}

func (r *lineOfFireRepo) FindByID(id int) (*LineOfFire, error) {
	l, err := scanLineOfFire(r.db.QueryRow(
		`SELECT id, team_id, map_id, name, polygon_data, order_index, created_at, updated_at
		 FROM lines_of_fire WHERE id = $1`,
		id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrLineOfFireNotFound
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *lineOfFireRepo) FindByTeamAndMap(teamID, mapID int) ([]*LineOfFire, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, map_id, name, polygon_data, order_index, created_at, updated_at
		 FROM lines_of_fire WHERE team_id = $1 AND map_id = $2 ORDER BY order_index`,
		teamID, mapID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []*LineOfFire
	for rows.Next() {
		l, err := scanLineOfFire(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if lines == nil {
		return []*LineOfFire{}, nil
	}
	return lines, nil
}

func (r *lineOfFireRepo) UpdateName(id int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrLineOfFireNameRequired
	}
	if len(name) > 250 {
		return ErrLineOfFireNameTooLong
	}

	l, err := r.FindByID(id)
	if err != nil {
		return err
	}

	taken, err := r.isNameTakenExcluding(l.TeamID, l.MapID, name, id)
	if err != nil {
		return fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if taken {
		return ErrLineOfFireNameTaken
	}

	result, err := r.db.Exec(
		`UPDATE lines_of_fire SET name = $1, updated_at = NOW() WHERE id = $2`,
		name, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update line of fire name: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrLineOfFireNotFound
	}
	return nil
}

func (r *lineOfFireRepo) UpdateOrder(id int, orderIndex int) error {
	result, err := r.db.Exec(
		`UPDATE lines_of_fire SET order_index = $1, updated_at = NOW() WHERE id = $2`,
		orderIndex, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update line of fire order: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrLineOfFireNotFound
	}
	return nil
}

func (r *lineOfFireRepo) UpdatePolygonData(id int, data ConeData) error {
	if len(data.Cones) == 0 {
		return ErrLineOfFireNoCones
	}
	for i, c := range data.Cones {
		if c.Origin.X == 0 && c.Origin.Y == 0 {
			return fmt.Errorf("%w: cone %d has no origin", ErrConeOriginRequired, i)
		}
	}

	raw, err := marshalConeData(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cone data: %w", err)
	}

	result, err := r.db.Exec(
		`UPDATE lines_of_fire SET polygon_data = $1, updated_at = NOW() WHERE id = $2`,
		raw, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update polygon data: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrLineOfFireNotFound
	}
	return nil
}

func (r *lineOfFireRepo) Delete(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM line_of_fire_versions WHERE line_of_fire_id = $1`, id); err != nil {
		return fmt.Errorf("failed to delete versions: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM line_of_fire_edit_history WHERE line_of_fire_id = $1`, id); err != nil {
		return fmt.Errorf("failed to delete edit history: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM lines_of_fire WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete line of fire: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLineOfFireNotFound
	}

	return tx.Commit()
}

func (r *lineOfFireRepo) IsNameTaken(teamID, mapID int, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM lines_of_fire
		 WHERE team_id = $1 AND map_id = $2 AND name = $3`,
		teamID, mapID, name,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *lineOfFireRepo) isNameTakenExcluding(teamID, mapID int, name string, excludeID int) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM lines_of_fire
		 WHERE team_id = $1 AND map_id = $2 AND name = $3 AND id != $4`,
		teamID, mapID, name, excludeID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *lineOfFireVersionRepo) Save(lineOfFireID int, editorID *int, snapshot json.RawMessage) (*LineOfFireVersion, error) {
	v := &LineOfFireVersion{}
	err := r.db.QueryRow(
		`INSERT INTO line_of_fire_versions (line_of_fire_id, editor_id, version_data)
		 VALUES ($1, $2, $3)
		 RETURNING id, line_of_fire_id, editor_id, version_data, created_at`,
		lineOfFireID, editorID, snapshot,
	).Scan(&v.ID, &v.LineOfFireID, &v.EditorID, &v.VersionData, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save line of fire version: %w", err)
	}
	return v, nil
}

func (r *lineOfFireVersionRepo) FindByLineOfFire(lineOfFireID int) ([]*LineOfFireVersion, error) {
	rows, err := r.db.Query(
		`SELECT id, line_of_fire_id, editor_id, version_data, created_at
		 FROM line_of_fire_versions WHERE line_of_fire_id = $1 ORDER BY created_at DESC`,
		lineOfFireID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*LineOfFireVersion
	for rows.Next() {
		v := &LineOfFireVersion{}
		if err := rows.Scan(&v.ID, &v.LineOfFireID, &v.EditorID, &v.VersionData, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if versions == nil {
		return []*LineOfFireVersion{}, nil
	}
	return versions, nil
}

func (r *lineOfFireVersionRepo) GetLatestForLineOfFire(lineOfFireID int) (*LineOfFireVersion, error) {
	v := &LineOfFireVersion{}
	err := r.db.QueryRow(
		`SELECT id, line_of_fire_id, editor_id, version_data, created_at
		 FROM line_of_fire_versions WHERE line_of_fire_id = $1 ORDER BY created_at DESC LIMIT 1`,
		lineOfFireID,
	).Scan(&v.ID, &v.LineOfFireID, &v.EditorID, &v.VersionData, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

type lofEditHistoryRepo struct {
	db *sql.DB
}

func NewLineOfFireEditHistoryRepository(db *sql.DB) LineOfFireEditHistoryRepository {
	return &lofEditHistoryRepo{db: db}
}

func (r *lofEditHistoryRepo) Log(lineOfFireID int, editorID *int, action string, oldData, newData json.RawMessage) error {
	_, err := r.db.Exec(
		`INSERT INTO line_of_fire_edit_history (line_of_fire_id, editor_id, action, old_data, new_data) VALUES ($1, $2, $3, $4, $5)`,
		lineOfFireID, editorID, action, oldData, newData,
	)
	if err != nil {
		return fmt.Errorf("failed to log line of fire edit history: %w", err)
	}
	return nil
}

func (r *lineOfFireVersionRepo) DeleteNewerThan(lineOfFireID int, olderThan time.Time) error {
	_, err := r.db.Exec(
		`DELETE FROM line_of_fire_versions WHERE line_of_fire_id = $1 AND created_at > $2`,
		lineOfFireID, olderThan,
	)
	if err != nil {
		return fmt.Errorf("failed to delete versions newer than: %w", err)
	}
	return nil
}
