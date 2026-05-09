package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type EditHistory struct {
	ID        int              `json:"id"`
	ZoneID    int              `json:"zone_id"`
	EditorID  *int             `json:"editor_id,omitempty"`
	Action    string           `json:"action"`
	OldData   *json.RawMessage `json:"old_data,omitempty"`
	NewData   *json.RawMessage `json:"new_data,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

type editHistoryRepo struct {
	db *sql.DB
}

func NewEditHistoryRepository(db *sql.DB) EditHistoryRepository {
	return &editHistoryRepo{db: db}
}

func (r *editHistoryRepo) Log(zoneID int, editorID *int, action string, oldData, newData json.RawMessage) (*EditHistory, error) {
	h := &EditHistory{
		ZoneID:   zoneID,
		EditorID: editorID,
		Action:   action,
	}

	var oldDataDB, newDataDB interface{}
	if oldData != nil {
		h.OldData = &oldData
		oldDataDB = oldData
	}
	if newData != nil {
		h.NewData = &newData
		newDataDB = newData
	}

	err := r.db.QueryRow(
		`INSERT INTO edit_history (zone_id, editor_id, action, old_data, new_data)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, zone_id, editor_id, action, old_data, new_data, timestamp`,
		zoneID, editorID, action, oldDataDB, newDataDB,
	).Scan(&h.ID, &h.ZoneID, &h.EditorID, &h.Action, &h.OldData, &h.NewData, &h.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to log edit history: %w", err)
	}

	return h, nil
}

func (r *editHistoryRepo) FindByZone(zoneID int) ([]*EditHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, zone_id, editor_id, action, old_data, new_data, timestamp
		 FROM edit_history WHERE zone_id = $1 ORDER BY timestamp DESC`,
		zoneID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

func (r *editHistoryRepo) FindByTeamAndMap(teamID, mapID int) ([]*EditHistory, error) {
	rows, err := r.db.Query(
		`SELECT eh.id, eh.zone_id, eh.editor_id, eh.action, eh.old_data, eh.new_data, eh.timestamp
		 FROM edit_history eh
		 JOIN zones z ON z.id = eh.zone_id
		 WHERE z.team_id = $1 AND z.map_id = $2
		 ORDER BY eh.timestamp DESC`,
		teamID, mapID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

func scanHistoryRows(rows *sql.Rows) ([]*EditHistory, error) {
	var history []*EditHistory
	for rows.Next() {
		h := &EditHistory{}
		var oldDataDB, newDataDB sql.NullString
		err := rows.Scan(&h.ID, &h.ZoneID, &h.EditorID, &h.Action, &oldDataDB, &newDataDB, &h.Timestamp)
		if err != nil {
			return nil, err
		}

		if oldDataDB.Valid {
			raw := json.RawMessage(oldDataDB.String)
			h.OldData = &raw
		}
		if newDataDB.Valid {
			raw := json.RawMessage(newDataDB.String)
			h.NewData = &raw
		}

		history = append(history, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if history == nil {
		return []*EditHistory{}, nil
	}
	return history, nil
}
