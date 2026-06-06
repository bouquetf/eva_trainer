package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Ligue struct {
	ID        int       `json:"id"`
	DateDebut time.Time `json:"date_debut"`
	DateFin   time.Time `json:"date_fin"`
	SplitID   int       `json:"split_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ligueRepo struct {
	db *sql.DB
}

func NewLigueRepository(db *sql.DB) LigueRepository {
	return &ligueRepo{db: db}
}

func scanLigue(scanner interface {
	Scan(dest ...interface{}) error
}) (*Ligue, error) {
	l := &Ligue{}
	err := scanner.Scan(&l.ID, &l.DateDebut, &l.DateFin, &l.SplitID, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func scanLigues(rows *sql.Rows) ([]*Ligue, error) {
	var ligues []*Ligue
	for rows.Next() {
		l, err := scanLigue(rows)
		if err != nil {
			return nil, err
		}
		ligues = append(ligues, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if ligues == nil {
		return []*Ligue{}, nil
	}
	return ligues, nil
}

func (r *ligueRepo) Create(dateDebut, dateFin time.Time, splitID int) (*Ligue, error) {
	if dateDebut.After(dateFin) {
		return nil, ErrLigueInvalidDates
	}

	l, err := scanLigue(r.db.QueryRow(
		`INSERT INTO ligue (date_debut, date_fin, split_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, date_debut, date_fin, split_id, created_at, updated_at`,
		dateDebut, dateFin, splitID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create ligue: %w", err)
	}
	return l, nil
}

func (r *ligueRepo) FindByID(id int) (*Ligue, error) {
	l, err := scanLigue(r.db.QueryRow(
		`SELECT id, date_debut, date_fin, split_id, created_at, updated_at
		 FROM ligue WHERE id = $1`,
		id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrLigueNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ligue by id: %w", err)
	}
	return l, nil
}

func (r *ligueRepo) FindAll() ([]*Ligue, error) {
	rows, err := r.db.Query(
		`SELECT id, date_debut, date_fin, split_id, created_at, updated_at
		 FROM ligue ORDER BY date_debut DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list ligues: %w", err)
	}
	defer rows.Close()
	return scanLigues(rows)
}

func (r *ligueRepo) FindAllPaginated(page, perPage int) ([]*Ligue, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 5
	}

	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM ligue`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ligues: %w", err)
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT id, date_debut, date_fin, split_id, created_at, updated_at
		 FROM ligue ORDER BY date_debut DESC LIMIT $1 OFFSET $2`,
		perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list ligues paginated: %w", err)
	}
	defer rows.Close()

	ligues, err := scanLigues(rows)
	if err != nil {
		return nil, 0, err
	}
	return ligues, total, nil
}

func (r *ligueRepo) Update(id int, dateDebut, dateFin time.Time, splitID int) (*Ligue, error) {
	if dateDebut.After(dateFin) {
		return nil, ErrLigueInvalidDates
	}

	l, err := scanLigue(r.db.QueryRow(
		`UPDATE ligue SET date_debut = $1, date_fin = $2, split_id = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, date_debut, date_fin, split_id, created_at, updated_at`,
		dateDebut, dateFin, splitID, id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrLigueNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update ligue: %w", err)
	}
	return l, nil
}

func (r *ligueRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM ligue WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ligue: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLigueNotFound
	}
	return nil
}

func (r *ligueRepo) AssociateAllLocales(ligueID int) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id)
		 SELECT $1, id FROM ligue_locale
		 ON CONFLICT (ligue_id, ligue_locale_id) DO NOTHING`,
		ligueID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to associate locales: %w", err)
	}
	return result.RowsAffected()
}
