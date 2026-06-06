package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type LigueLocale struct {
	ID          int       `json:"id"`
	Nom         string    `json:"nom"`
	Emplacement string    `json:"emplacement"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ligueLocaleRepo struct {
	db *sql.DB
}

func NewLigueLocaleRepository(db *sql.DB) LigueLocaleRepository {
	return &ligueLocaleRepo{db: db}
}

func scanLigueLocale(scanner interface {
	Scan(dest ...interface{}) error
}) (*LigueLocale, error) {
	l := &LigueLocale{}
	err := scanner.Scan(&l.ID, &l.Nom, &l.Emplacement, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func scanLigueLocales(rows *sql.Rows) ([]*LigueLocale, error) {
	var locales []*LigueLocale
	for rows.Next() {
		l, err := scanLigueLocale(rows)
		if err != nil {
			return nil, err
		}
		locales = append(locales, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if locales == nil {
		return []*LigueLocale{}, nil
	}
	return locales, nil
}

func (r *ligueLocaleRepo) Create(nom string, emplacement string) (*LigueLocale, error) {
	if nom == "" {
		return nil, ErrLigueLocaleNameRequired
	}

	l, err := scanLigueLocale(r.db.QueryRow(
		`INSERT INTO ligue_locale (nom, emplacement)
		 VALUES ($1, $2)
		 RETURNING id, nom, emplacement, created_at, updated_at`,
		nom, emplacement,
	))
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return nil, ErrLigueLocaleEmplacementExists
		}
		return nil, fmt.Errorf("failed to create ligue locale: %w", err)
	}
	return l, nil
}

func (r *ligueLocaleRepo) FindByID(id int) (*LigueLocale, error) {
	l, err := scanLigueLocale(r.db.QueryRow(
		`SELECT id, nom, emplacement, created_at, updated_at
		 FROM ligue_locale WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLigueLocaleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ligue locale by id: %w", err)
	}
	return l, nil
}

func (r *ligueLocaleRepo) FindAll(emplacement string) ([]*LigueLocale, error) {
	var rows *sql.Rows
	var err error

	if emplacement == "" {
		rows, err = r.db.Query(
			`SELECT id, nom, emplacement, created_at, updated_at
			 FROM ligue_locale ORDER BY nom ASC`,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, nom, emplacement, created_at, updated_at
			 FROM ligue_locale WHERE emplacement = $1 ORDER BY nom ASC`,
			emplacement,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list ligue locales: %w", err)
	}
	defer rows.Close()

	return scanLigueLocales(rows)
}

func (r *ligueLocaleRepo) FindByLigueID(ligueID int, page, perPage int) ([]*LigueLocale, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 5
	}

	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM ligue_locale_association WHERE ligue_id = $1`, ligueID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ligue locales by ligue id: %w", err)
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT ll.id, ll.nom, ll.emplacement, ll.created_at, ll.updated_at
		 FROM ligue_locale ll
		 INNER JOIN ligue_locale_association lla ON lla.ligue_locale_id = ll.id
		 WHERE lla.ligue_id = $1
		 ORDER BY ll.nom ASC LIMIT $2 OFFSET $3`,
		ligueID, perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find ligue locales by ligue id: %w", err)
	}
	defer rows.Close()

	locales, err := scanLigueLocales(rows)
	if err != nil {
		return nil, 0, err
	}
	return locales, total, nil
}

func (r *ligueLocaleRepo) Update(id int, nom string, emplacement string) (*LigueLocale, error) {
	if nom == "" {
		return nil, ErrLigueLocaleNameRequired
	}

	l, err := scanLigueLocale(r.db.QueryRow(
		`UPDATE ligue_locale SET nom = $1, emplacement = $2, updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, nom, emplacement, created_at, updated_at`,
		nom, emplacement, id,
	))
	if err == sql.ErrNoRows {
		return nil, ErrLigueLocaleNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrLigueLocaleEmplacementExists
		}
		return nil, fmt.Errorf("failed to update ligue locale: %w", err)
	}
	return l, nil
}

func (r *ligueLocaleRepo) AssociateLigue(localeID, ligueID int) error {
	_, err := r.db.Exec(
		`INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id)
		 VALUES ($1, $2)
		 ON CONFLICT (ligue_id, ligue_locale_id) DO NOTHING`,
		ligueID, localeID,
	)
	if err != nil {
		return fmt.Errorf("failed to associate ligue with locale: %w", err)
	}
	return nil
}

func (r *ligueLocaleRepo) DisassociateLigue(localeID, ligueID int) error {
	result, err := r.db.Exec(
		`DELETE FROM ligue_locale_association WHERE ligue_id = $1 AND ligue_locale_id = $2`,
		ligueID, localeID,
	)
	if err != nil {
		return fmt.Errorf("failed to disassociate ligue from locale: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLigueLocaleNotFound
	}
	return nil
}

func (r *ligueLocaleRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM ligue_locale WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ligue locale: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLigueLocaleNotFound
	}
	return nil
}

func (r *ligueLocaleRepo) FindLiguesByLocaleID(localeID int) ([]*Ligue, error) {
	rows, err := r.db.Query(
		`SELECT l.id, l.date_debut, l.date_fin, l.split_id, l.created_at, l.updated_at
		 FROM ligue l
		 INNER JOIN ligue_locale_association lla ON lla.ligue_id = l.id
		 WHERE lla.ligue_locale_id = $1
		 ORDER BY l.date_debut DESC`,
		localeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find ligues by locale id: %w", err)
	}
	defer rows.Close()
	return scanLigues(rows)
}
