package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DivNote struct {
	DivID     int       `json:"div_id"`
	UserID    int       `json:"user_id"`
	ContentMD string    `json:"content_md"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type divNoteRepo struct {
	db *sql.DB
}

func NewDivNoteRepository(db *sql.DB) DivNoteRepository {
	return &divNoteRepo{db: db}
}

func scanDivNote(scanner interface {
	Scan(dest ...interface{}) error
}) (*DivNote, error) {
	n := &DivNote{}
	err := scanner.Scan(&n.DivID, &n.UserID, &n.ContentMD, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *divNoteRepo) Upsert(divID, userID int, contentMD string) (*DivNote, error) {
	n, err := scanDivNote(r.db.QueryRow(
		`INSERT INTO div_notes (div_id, user_id, content_md, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (div_id, user_id) DO UPDATE SET content_md = $3, updated_at = NOW()
		 RETURNING div_id, user_id, content_md, created_at, updated_at`,
		divID, userID, contentMD,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to upsert div note: %w", err)
	}
	return n, nil
}

func (r *divNoteRepo) FindByDivAndUser(divID, userID int) (*DivNote, error) {
	n, err := scanDivNote(r.db.QueryRow(
		`SELECT div_id, user_id, content_md, created_at, updated_at
		 FROM div_notes
		 WHERE div_id = $1 AND user_id = $2`,
		divID, userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDivNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find div note: %w", err)
	}
	return n, nil
}

func (r *divNoteRepo) DeleteByDivAndUser(divID, userID int) error {
	result, err := r.db.Exec(
		`DELETE FROM div_notes WHERE div_id = $1 AND user_id = $2`,
		divID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete div note: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDivNoteNotFound
	}
	return nil
}
