package models

import (
	"database/sql"
	"fmt"
	"time"
)

type EditLock struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	MapID     int       `json:"map_id"`
	EditorID  int       `json:"editor_id"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type editLockRepo struct {
	db *sql.DB
}

func NewEditLockRepository(db *sql.DB) EditLockRepository {
	return &editLockRepo{db: db}
}

func scanEditLock(scanner interface {
	Scan(dest ...interface{}) error
}) (*EditLock, error) {
	l := &EditLock{}
	err := scanner.Scan(&l.ID, &l.TeamID, &l.MapID, &l.EditorID, &l.LockedAt, &l.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *editLockRepo) Acquire(teamID, mapID, editorID int) (*EditLock, error) {
	l, err := scanEditLock(r.db.QueryRow(
		`INSERT INTO edit_locks (team_id, map_id, editor_id, locked_at, expires_at)
		 VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '15 minutes')
		 RETURNING id, team_id, map_id, editor_id, locked_at, expires_at`,
		teamID, mapID, editorID,
	))
	if err != nil {
		if isUniqueViolation(err) {
			existing, getErr := r.Get(teamID, mapID)
			if getErr != nil {
				return nil, fmt.Errorf("failed to check existing lock: %w", getErr)
			}
			if existing == nil {
				return nil, fmt.Errorf("failed to acquire lock: %w", err)
			}

			if existing.EditorID == editorID {
				return r.Refresh(teamID, mapID)
			}

			if existing.ExpiresAt.Before(time.Now()) {
				_, releaseErr := r.db.Exec(
					`DELETE FROM edit_locks WHERE team_id = $1 AND map_id = $2`,
					teamID, mapID,
				)
				if releaseErr != nil {
					return nil, fmt.Errorf("failed to release expired lock: %w", releaseErr)
				}

				return scanEditLock(r.db.QueryRow(
					`INSERT INTO edit_locks (team_id, map_id, editor_id, locked_at, expires_at)
					 VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '15 minutes')
					 RETURNING id, team_id, map_id, editor_id, locked_at, expires_at`,
					teamID, mapID, editorID,
				))
			}

			return nil, ErrLockAlreadyAcquired
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return l, nil
}

func (r *editLockRepo) Release(teamID, mapID int) error {
	result, err := r.db.Exec(
		`DELETE FROM edit_locks WHERE team_id = $1 AND map_id = $2`,
		teamID, mapID,
	)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLockNotFound
	}
	return nil
}

func (r *editLockRepo) ReleaseByEditor(teamID, mapID, editorID int) error {
	result, err := r.db.Exec(
		`DELETE FROM edit_locks WHERE team_id = $1 AND map_id = $2 AND editor_id = $3`,
		teamID, mapID, editorID,
	)
	if err != nil {
		return fmt.Errorf("failed to release lock by editor: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLockNotFound
	}
	return nil
}

func (r *editLockRepo) Get(teamID, mapID int) (*EditLock, error) {
	l, err := scanEditLock(r.db.QueryRow(
		`SELECT id, team_id, map_id, editor_id, locked_at, expires_at
		 FROM edit_locks WHERE team_id = $1 AND map_id = $2`,
		teamID, mapID,
	))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *editLockRepo) Refresh(teamID, mapID int) (*EditLock, error) {
	l, err := scanEditLock(r.db.QueryRow(
		`UPDATE edit_locks
		 SET expires_at = NOW() + INTERVAL '15 minutes'
		 WHERE team_id = $1 AND map_id = $2
		 RETURNING id, team_id, map_id, editor_id, locked_at, expires_at`,
		teamID, mapID,
	))
	if err == sql.ErrNoRows {
		return nil, ErrLockNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to refresh lock: %w", err)
	}
	return l, nil
}

func (r *editLockRepo) IsLocked(teamID, mapID int) (bool, error) {
	l, err := r.Get(teamID, mapID)
	if err != nil {
		return false, err
	}
	if l == nil {
		return false, nil
	}
	if l.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

func (r *editLockRepo) IsLockedByEditor(teamID, mapID, editorID int) (bool, error) {
	l, err := r.Get(teamID, mapID)
	if err != nil {
		return false, err
	}
	if l == nil {
		return false, nil
	}
	if l.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	return l.EditorID == editorID, nil
}

func (r *editLockRepo) CleanupExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM edit_locks WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired locks: %w", err)
	}
	return result.RowsAffected()
}
