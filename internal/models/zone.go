package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Zone struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	MapID     int       `json:"map_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ZonePolygon struct {
	ID         int       `json:"id"`
	ZoneID     int       `json:"zone_id"`
	Points     []Point   `json:"points"`
	OrderIndex int       `json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
}

type ZoneName struct {
	ID        int       `json:"id"`
	ZoneID    int       `json:"zone_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ZoneWithDetails struct {
	ID        int           `json:"id"`
	TeamID    int           `json:"team_id"`
	MapID     int           `json:"map_id"`
	Polygons  []ZonePolygon `json:"polygons"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type zoneRepo struct {
	db *sql.DB
}

func NewZoneRepository(db *sql.DB) ZoneRepository {
	return &zoneRepo{db: db}
}

func scanZone(scanner interface {
	Scan(dest ...interface{}) error
}) (*Zone, error) {
	z := &Zone{}
	err := scanner.Scan(&z.ID, &z.TeamID, &z.MapID, &z.CreatedAt, &z.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return z, nil
}

func scanZoneWithDetails(scanner interface {
	Scan(dest ...interface{}) error
}) (*ZoneWithDetails, error) {
	z := &ZoneWithDetails{}
	err := scanner.Scan(&z.ID, &z.TeamID, &z.MapID, &z.CreatedAt, &z.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return z, nil
}

func scanPolygonData(data []byte) ([]Point, error) {
	if data == nil {
		return []Point{}, nil
	}
	var points []Point
	if err := json.Unmarshal(data, &points); err != nil {
		return nil, fmt.Errorf("failed to unmarshal polygon data: %w", err)
	}
	if points == nil {
		return []Point{}, nil
	}
	return points, nil
}

func marshalPoints(points []Point) ([]byte, error) {
	return json.Marshal(points)
}

func (r *zoneRepo) Create(teamID, mapID int) (*Zone, error) {
	z, err := scanZone(r.db.QueryRow(
		`INSERT INTO zones (team_id, map_id) VALUES ($1, $2)
		 RETURNING id, team_id, map_id, created_at, updated_at`,
		teamID, mapID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}
	return z, nil
}

func (r *zoneRepo) FindByID(zoneID int) (*Zone, error) {
	z, err := scanZone(r.db.QueryRow(
		`SELECT id, team_id, map_id, created_at, updated_at FROM zones WHERE id = $1`,
		zoneID,
	))
	if err == sql.ErrNoRows {
		return nil, ErrZoneNotFound
	}
	if err != nil {
		return nil, err
	}
	return z, nil
}

func (r *zoneRepo) FindByTeamAndMap(teamID, mapID int) ([]*Zone, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, map_id, created_at, updated_at FROM zones
		 WHERE team_id = $1 AND map_id = $2 ORDER BY id`,
		teamID, mapID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*Zone
	for rows.Next() {
		z, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		zones = append(zones, z)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if zones == nil {
		return []*Zone{}, nil
	}
	return zones, nil
}

func (r *zoneRepo) FindWithDetailsByTeamAndMap(teamID, mapID int) ([]*ZoneWithDetails, error) {
	zones, err := r.FindByTeamAndMap(teamID, mapID)
	if err != nil {
		return nil, err
	}

	result := make([]*ZoneWithDetails, 0, len(zones))
	for _, z := range zones {
		polygons, err := r.GetPolygons(z.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get polygons for zone %d: %w", z.ID, err)
		}

		zoneName := ""
		zn, err := r.GetName(z.ID)
		if err == nil && zn != nil {
			zoneName = zn.Name
		} else if err != ErrZoneNotFound && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get name for zone %d: %w", z.ID, err)
		}

		result = append(result, &ZoneWithDetails{
			ID:        z.ID,
			TeamID:    z.TeamID,
			MapID:     z.MapID,
			Polygons:  polygons,
			Name:      zoneName,
			CreatedAt: z.CreatedAt,
			UpdatedAt: z.UpdatedAt,
		})
	}

	return result, nil
}

func (r *zoneRepo) Delete(zoneID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM zone_versions WHERE zone_id = $1`, zoneID); err != nil {
		return fmt.Errorf("failed to delete zone versions: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM zones WHERE id = $1`, zoneID)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrZoneNotFound
	}

	return tx.Commit()
}

func (r *zoneRepo) AddPolygon(zoneID int, points []Point, orderIndex int) (*ZonePolygon, error) {
	if len(points) < 3 {
		return nil, ErrPolygonHasTooFewPoints
	}

	data, err := marshalPoints(points)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal polygon points: %w", err)
	}

	p := &ZonePolygon{}
	err = r.db.QueryRow(
		`INSERT INTO zone_polygons (zone_id, polygon_data, order_index)
		 VALUES ($1, $2, $3)
		 RETURNING id, zone_id, order_index, created_at`,
		zoneID, data, orderIndex,
	).Scan(&p.ID, &p.ZoneID, &p.OrderIndex, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add polygon: %w", err)
	}

	p.Points = points
	return p, nil
}

func (r *zoneRepo) GetPolygons(zoneID int) ([]ZonePolygon, error) {
	rows, err := r.db.Query(
		`SELECT id, zone_id, polygon_data, order_index, created_at
		 FROM zone_polygons WHERE zone_id = $1 ORDER BY order_index`,
		zoneID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polygons []ZonePolygon
	for rows.Next() {
		var p ZonePolygon
		var data []byte
		if err := rows.Scan(&p.ID, &p.ZoneID, &data, &p.OrderIndex, &p.CreatedAt); err != nil {
			return nil, err
		}
		points, err := scanPolygonData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to scan polygon data for polygon %d: %w", p.ID, err)
		}
		p.Points = points
		polygons = append(polygons, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if polygons == nil {
		return []ZonePolygon{}, nil
	}
	return polygons, nil
}

func (r *zoneRepo) UpdatePolygonPoints(polygonID int, points []Point) error {
	if len(points) < 3 {
		return ErrPolygonHasTooFewPoints
	}

	data, err := marshalPoints(points)
	if err != nil {
		return fmt.Errorf("failed to marshal polygon points: %w", err)
	}

	result, err := r.db.Exec(
		`UPDATE zone_polygons SET polygon_data = $1 WHERE id = $2`,
		data, polygonID,
	)
	if err != nil {
		return fmt.Errorf("failed to update polygon points: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPolygonNotFound
	}
	return nil
}

func (r *zoneRepo) RemovePolygon(polygonID int) error {
	result, err := r.db.Exec(`DELETE FROM zone_polygons WHERE id = $1`, polygonID)
	if err != nil {
		return fmt.Errorf("failed to remove polygon: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPolygonNotFound
	}
	return nil
}

func (r *zoneRepo) PolygonCount(zoneID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM zone_polygons WHERE zone_id = $1`,
		zoneID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *zoneRepo) SetName(zoneID int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrZoneNameRequired
	}
	if len(name) > 250 {
		return ErrZoneNameTooLong
	}

	zone, err := r.FindByID(zoneID)
	if err != nil {
		return err
	}

	taken, err := r.isNameTakenExcluding(zone.TeamID, zone.MapID, name, zoneID)
	if err != nil {
		return fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if taken {
		return ErrZoneNameTaken
	}

	result, err := r.db.Exec(
		`INSERT INTO zone_names (zone_id, name) VALUES ($1, $2)
		 ON CONFLICT (zone_id) DO UPDATE SET name = $2`,
		zoneID, name,
	)
	if err != nil {
		return fmt.Errorf("failed to set zone name: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		_, err = r.db.Exec(`UPDATE zone_names SET name = $1 WHERE zone_id = $2`, name, zoneID)
		if err != nil {
			return fmt.Errorf("failed to update zone name: %w", err)
		}
	}
	return nil
}

func (r *zoneRepo) GetName(zoneID int) (*ZoneName, error) {
	zn := &ZoneName{}
	err := r.db.QueryRow(
		`SELECT id, zone_id, name, created_at FROM zone_names WHERE zone_id = $1`,
		zoneID,
	).Scan(&zn.ID, &zn.ZoneID, &zn.Name, &zn.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrZoneNotFound
	}
	if err != nil {
		return nil, err
	}
	return zn, nil
}

func (r *zoneRepo) IsNameTaken(teamID, mapID int, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		 FROM zone_names zn
		 JOIN zones z ON z.id = zn.zone_id
		 WHERE z.team_id = $1 AND z.map_id = $2 AND zn.name = $3`,
		teamID, mapID, name,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *zoneRepo) isNameTakenExcluding(teamID, mapID int, name string, excludeZoneID int) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		 FROM zone_names zn
		 JOIN zones z ON z.id = zn.zone_id
		 WHERE z.team_id = $1 AND z.map_id = $2 AND zn.name = $3 AND z.id != $4`,
		teamID, mapID, name, excludeZoneID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
