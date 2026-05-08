package models

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var teamNameRegex = regexp.MustCompile(`^[\p{L}\p{M}0-9\-_' .]+$`)

type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CaptainID int       `json:"captain_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type teamRepo struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) TeamRepository {
	return &teamRepo{db: db}
}

func ValidateTeamName(name string) error {
	if name == "" {
		return ErrTeamNameRequired
	}
	if len(name) > 250 {
		return ErrTeamNameTooLong
	}
	if !teamNameRegex.MatchString(name) {
		return ErrInvalidTeamName
	}
	return nil
}

func (r *teamRepo) Create(name string, captainID int) (*Team, error) {
	if err := ValidateTeamName(name); err != nil {
		return nil, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	team := &Team{}
	err = tx.QueryRow(
		`INSERT INTO teams (name, captain_id) VALUES ($1, $2)
		 RETURNING id, name, captain_id, created_at, updated_at`,
		name, captainID,
	).Scan(&team.ID, &team.Name, &team.CaptainID, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("team name already exists: %w", err)
		}
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'captain')`,
		team.ID, captainID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add captain to team_members: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return team, nil
}

func (r *teamRepo) FindByID(id int) (*Team, error) {
	team := &Team{}
	err := r.db.QueryRow(
		`SELECT id, name, captain_id, created_at, updated_at FROM teams WHERE id = $1`,
		id,
	).Scan(&team.ID, &team.Name, &team.CaptainID, &team.CreatedAt, &team.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *teamRepo) FindByUserID(userID int) ([]*Team, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.name, t.captain_id, t.created_at, t.updated_at
		 FROM teams t
		 JOIN team_members tm ON t.id = tm.team_id
		 WHERE tm.user_id = $1
		 ORDER BY t.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		t := &Team{}
		if err := rows.Scan(&t.ID, &t.Name, &t.CaptainID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if teams == nil {
		return []*Team{}, nil
	}

	return teams, nil
}

func (r *teamRepo) UpdateName(id int, name string) error {
	if err := ValidateTeamName(name); err != nil {
		return err
	}

	result, err := r.db.Exec(
		`UPDATE teams SET name = $1, updated_at = NOW() WHERE id = $2`,
		name, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("team name already exists: %w", err)
		}
		return fmt.Errorf("failed to update team name: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTeamNotFound
	}

	return nil
}

func (r *teamRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM teams WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTeamNotFound
	}

	return nil
}
