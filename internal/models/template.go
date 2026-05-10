package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Template struct {
	ID             int       `json:"id"`
	OriginalTeamID int       `json:"original_team_id"`
	MapID          int       `json:"map_id"`
	IsPublic       bool      `json:"is_public"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TemplateZone struct {
	ID          int             `json:"id"`
	TemplateID  int             `json:"template_id"`
	PolygonData json.RawMessage `json:"polygon_data"`
	Name        string          `json:"name"`
	OrderIndex  int             `json:"order_index"`
}

type TemplateWithZones struct {
	Template Template       `json:"template"`
	Zones    []TemplateZone `json:"zones"`
}

type templateRepo struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) TemplateRepository {
	return &templateRepo{db: db}
}

func scanTemplate(scanner interface {
	Scan(dest ...interface{}) error
}) (*Template, error) {
	t := &Template{}
	err := scanner.Scan(&t.ID, &t.OriginalTeamID, &t.MapID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.IsPublic = true
	return t, nil
}

func (r *templateRepo) CreateFromMap(teamID, mapID int) (*Template, error) {
	t, err := scanTemplate(r.db.QueryRow(
		`INSERT INTO templates (original_team_id, map_id, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, original_team_id, map_id, created_at, updated_at`,
		teamID, mapID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT z.id FROM zones z WHERE z.team_id = $1 AND z.map_id = $2 ORDER BY z.id`,
		teamID, mapID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones for template: %w", err)
	}
	defer rows.Close()

	var zoneIDs []int
	for rows.Next() {
		var zoneID int
		if err := rows.Scan(&zoneID); err != nil {
			return nil, fmt.Errorf("failed to scan zone id: %w", err)
		}
		zoneIDs = append(zoneIDs, zoneID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, zoneID := range zoneIDs {
		polygonData, err := r.collectZonePolygons(zoneID)
		if err != nil {
			return nil, fmt.Errorf("failed to collect polygon data for zone %d: %w", zoneID, err)
		}

		zoneName, err := r.getZoneName(zoneID)
		if err != nil {
			return nil, fmt.Errorf("failed to get name for zone %d: %w", zoneID, err)
		}

		_, err = r.db.Exec(
			`INSERT INTO template_zones (template_id, polygon_data, name, order_index)
			 VALUES ($1, $2::jsonb, $3, $4)`,
			t.ID, polygonData, zoneName, i,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert template zone: %w", err)
		}
	}

	return t, nil
}

func (r *templateRepo) FindByMapID(mapID int) (*Template, error) {
	t, err := scanTemplate(r.db.QueryRow(
		`SELECT id, original_team_id, map_id, created_at, updated_at
		 FROM templates WHERE map_id = $1`,
		mapID,
	))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find template by map id: %w", err)
	}
	return t, nil
}

func (r *templateRepo) FindAllPublic() ([]*TemplateWithZones, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.original_team_id, t.map_id, t.created_at, t.updated_at,
		        tz.id, tz.polygon_data, tz.name, tz.order_index
		 FROM templates t
		 LEFT JOIN template_zones tz ON tz.template_id = t.id
		 ORDER BY t.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query all public templates: %w", err)
	}
	defer rows.Close()

	templateMap := make(map[int]*TemplateWithZones)
	var templateOrder []int

	for rows.Next() {
		var (
			t             Template
			tzID          sql.NullInt64
			tzPolygonData sql.NullString
			tzName        sql.NullString
			tzOrderIndex  sql.NullInt64
		)
		err := rows.Scan(
			&t.ID, &t.OriginalTeamID, &t.MapID, &t.CreatedAt, &t.UpdatedAt,
			&tzID, &tzPolygonData, &tzName, &tzOrderIndex,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		t.IsPublic = true

		entry, exists := templateMap[t.ID]
		if !exists {
			entry = &TemplateWithZones{
				Template: t,
				Zones:    []TemplateZone{},
			}
			templateMap[t.ID] = entry
			templateOrder = append(templateOrder, t.ID)
		}

		if tzID.Valid {
			zone := TemplateZone{
				ID:         int(tzID.Int64),
				TemplateID: t.ID,
				Name:       tzName.String,
				OrderIndex: int(tzOrderIndex.Int64),
			}
			if tzPolygonData.Valid {
				zone.PolygonData = json.RawMessage(tzPolygonData.String)
			}
			entry.Zones = append(entry.Zones, zone)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*TemplateWithZones, 0, len(templateOrder))
	for _, id := range templateOrder {
		result = append(result, templateMap[id])
	}
	if result == nil {
		return []*TemplateWithZones{}, nil
	}
	return result, nil
}

func (r *templateRepo) UpdateFromMap(templateID, teamID, mapID int) error {
	_, err := r.db.Exec(`DELETE FROM template_zones WHERE template_id = $1`, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete existing template zones: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT z.id FROM zones z WHERE z.team_id = $1 AND z.map_id = $2 ORDER BY z.id`,
		teamID, mapID,
	)
	if err != nil {
		return fmt.Errorf("failed to query zones for template update: %w", err)
	}
	defer rows.Close()

	var zoneIDs []int
	for rows.Next() {
		var zoneID int
		if err := rows.Scan(&zoneID); err != nil {
			return fmt.Errorf("failed to scan zone id: %w", err)
		}
		zoneIDs = append(zoneIDs, zoneID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i, zoneID := range zoneIDs {
		polygonData, err := r.collectZonePolygons(zoneID)
		if err != nil {
			return fmt.Errorf("failed to collect polygon data for zone %d: %w", zoneID, err)
		}

		zoneName, err := r.getZoneName(zoneID)
		if err != nil {
			return fmt.Errorf("failed to get name for zone %d: %w", zoneID, err)
		}

		_, err = r.db.Exec(
			`INSERT INTO template_zones (template_id, polygon_data, name, order_index)
			 VALUES ($1, $2::jsonb, $3, $4)`,
			templateID, polygonData, zoneName, i,
		)
		if err != nil {
			return fmt.Errorf("failed to insert template zone: %w", err)
		}
	}

	_, err = r.db.Exec(
		`UPDATE templates SET updated_at = NOW() WHERE id = $1`,
		templateID,
	)
	if err != nil {
		return fmt.Errorf("failed to update template timestamp: %w", err)
	}

	return nil
}

func (r *templateRepo) Delete(templateID int) error {
	result, err := r.db.Exec(`DELETE FROM templates WHERE id = $1`, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func (r *templateRepo) collectZonePolygons(zoneID int) ([]byte, error) {
	rows, err := r.db.Query(
		`SELECT polygon_data FROM zone_polygons WHERE zone_id = $1 ORDER BY order_index`,
		zoneID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polygons []map[string]interface{}
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var polyData map[string]interface{}
		if err := json.Unmarshal(data, &polyData); err != nil {
			pointsData := []map[string]interface{}{}
			if err := json.Unmarshal(data, &pointsData); err != nil {
				polyData = map[string]interface{}{"points": json.RawMessage(data)}
			} else {
				polyData = map[string]interface{}{"points": pointsData}
			}
		}
		polygons = append(polygons, polyData)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result, err := json.Marshal(polygons)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal polygon collection: %w", err)
	}
	return result, nil
}

func (r *templateRepo) getZoneName(zoneID int) (string, error) {
	var name string
	err := r.db.QueryRow(
		`SELECT name FROM zone_names WHERE zone_id = $1`,
		zoneID,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}
