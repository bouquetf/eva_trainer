package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type EvaTeam struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Tag              string    `json:"tag"`
	EvaTrainerTeamID *int      `json:"eva_trainer_team_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type evaTeamRepo struct {
	db *sql.DB
}

func NewEvaTeamRepository(db *sql.DB) EvaTeamRepository {
	return &evaTeamRepo{db: db}
}

func scanEvaTeam(scanner interface {
	Scan(dest ...interface{}) error
}) (*EvaTeam, error) {
	t := &EvaTeam{}
	err := scanner.Scan(&t.ID, &t.Name, &t.Tag, &t.EvaTrainerTeamID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func scanEvaTeams(rows *sql.Rows) ([]*EvaTeam, error) {
	var teams []*EvaTeam
	for rows.Next() {
		t, err := scanEvaTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if teams == nil {
		return []*EvaTeam{}, nil
	}
	return teams, nil
}

func ValidateEvaTeamName(name string) error {
	if name == "" {
		return ErrEvaTeamNameRequired
	}
	return nil
}

func ValidateEvaTeamTag(tag string) error {
	if tag == "" {
		return ErrEvaTeamTagRequired
	}
	if len(tag) > 50 {
		return ErrEvaTeamTagTooLong
	}
	return nil
}

func (r *evaTeamRepo) Create(name, tag string, evaTrainerTeamID *int) (*EvaTeam, error) {
	if err := ValidateEvaTeamName(name); err != nil {
		return nil, err
	}
	if err := ValidateEvaTeamTag(tag); err != nil {
		return nil, err
	}

	t, err := scanEvaTeam(r.db.QueryRow(
		`INSERT INTO eva_teams (name, tag, eva_trainer_team_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, tag, eva_trainer_team_id, created_at, updated_at`,
		name, tag, evaTrainerTeamID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create eva team: %w", err)
	}
	return t, nil
}

func (r *evaTeamRepo) FindByID(id int) (*EvaTeam, error) {
	t, err := scanEvaTeam(r.db.QueryRow(
		`SELECT id, name, tag, eva_trainer_team_id, created_at, updated_at
		 FROM eva_teams WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEvaTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find eva team by id: %w", err)
	}
	return t, nil
}

func (r *evaTeamRepo) FindAll() ([]*EvaTeam, error) {
	rows, err := r.db.Query(
		`SELECT id, name, tag, eva_trainer_team_id, created_at, updated_at
		 FROM eva_teams ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list eva teams: %w", err)
	}
	defer rows.Close()
	return scanEvaTeams(rows)
}

func (r *evaTeamRepo) Update(id int, name, tag string, evaTrainerTeamID *int) (*EvaTeam, error) {
	if err := ValidateEvaTeamName(name); err != nil {
		return nil, err
	}
	if err := ValidateEvaTeamTag(tag); err != nil {
		return nil, err
	}

	t, err := scanEvaTeam(r.db.QueryRow(
		`UPDATE eva_teams SET name = $1, tag = $2, eva_trainer_team_id = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, name, tag, eva_trainer_team_id, created_at, updated_at`,
		name, tag, evaTrainerTeamID, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEvaTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update eva team: %w", err)
	}
	return t, nil
}

func (r *evaTeamRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM eva_teams WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete eva team: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrEvaTeamNotFound
	}
	return nil
}

func (r *evaTeamRepo) HasReferences(id int) (bool, error) {
	rows, err := r.db.Query(`
		SELECT tc.table_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_name = 'eva_teams'
			AND ccu.column_name = 'id'`,
	)
	if err != nil {
		return false, fmt.Errorf("failed to query referencing tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, columnName string
		if err := rows.Scan(&tableName, &columnName); err != nil {
			return false, fmt.Errorf("failed to scan referencing table: %w", err)
		}

		var count int
		q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", tableName, columnName)
		if err := r.db.QueryRow(q, id).Scan(&count); err != nil {
			continue
		}
		if count > 0 {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("failed to iterate referencing tables: %w", err)
	}

	return false, nil
}
