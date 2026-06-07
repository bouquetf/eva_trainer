package models

import (
	"database/sql"
	"fmt"
	"time"
)

type BO struct {
	ID          int       `json:"id"`
	LigueID     int       `json:"ligue_id"`
	Equipe1ID   int       `json:"equipe1_id"`
	Equipe2ID   int       `json:"equipe2_id"`
	VainqueurID *int      `json:"vainqueur_id,omitempty"`
	Format      string    `json:"format"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type boRepo struct {
	db *sql.DB
}

func NewBORepository(db *sql.DB) BORepository {
	return &boRepo{db: db}
}

func scanBO(scanner interface {
	Scan(dest ...interface{}) error
}) (*BO, error) {
	b := &BO{}
	err := scanner.Scan(&b.ID, &b.LigueID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID, &b.Format, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func scanBOs(rows *sql.Rows) ([]*BO, error) {
	var bos []*BO
	for rows.Next() {
		b, err := scanBO(rows)
		if err != nil {
			return nil, err
		}
		bos = append(bos, b)
	}
	return bos, rows.Err()
}

func (r *boRepo) Create(ligueID, equipe1ID, equipe2ID int) (*BO, error) {
	b, err := scanBO(r.db.QueryRow(
		`INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, format, created_at, updated_at)
		 VALUES ($1, $2, $3, 'BO3', NOW(), NOW())
		 RETURNING id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, created_at, updated_at`,
		ligueID, equipe1ID, equipe2ID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create BO: %w", err)
	}
	return b, nil
}

func (r *boRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM bos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete BO: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBONotFound
	}
	return nil
}

func (r *boRepo) FindByLigueID(ligueID int) ([]*BO, error) {
	rows, err := r.db.Query(
		`SELECT id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, created_at, updated_at
		 FROM bos WHERE ligue_id = $1
		 ORDER BY id ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find BOs by ligue: %w", err)
	}
	defer rows.Close()
	return scanBOs(rows)
}
