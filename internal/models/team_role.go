package models

import (
	"database/sql"
	"fmt"
)

type TeamWithRole struct {
	TeamID   int    `json:"team_id"`
	Role     string `json:"role"`
	TeamName string `json:"team_name,omitempty"`
}

type teamRoleRepo struct {
	db *sql.DB
}

func NewTeamRoleRepository(db *sql.DB) TeamRoleRepository {
	return &teamRoleRepo{db: db}
}

func (r *teamRoleRepo) Add(teamID int, role string) error {
	_, err := r.db.Exec(
		`INSERT INTO team_roles (team_id, role) VALUES ($1, $2)`,
		teamID, role,
	)
	if err != nil {
		return fmt.Errorf("failed to add team role: %w", err)
	}
	return nil
}

func (r *teamRoleRepo) Remove(teamID int, role string) error {
	result, err := r.db.Exec(
		`DELETE FROM team_roles WHERE team_id = $1 AND role = $2`,
		teamID, role,
	)
	if err != nil {
		return fmt.Errorf("failed to remove team role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRoleNotFound
	}

	return nil
}

func (r *teamRoleRepo) TeamHasRole(teamID int, role string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM team_roles WHERE team_id = $1 AND role = $2)`,
		teamID, role,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team role: %w", err)
	}
	return exists, nil
}

func (r *teamRoleRepo) FindTeamsByRole(role string) ([]*TeamWithRole, error) {
	rows, err := r.db.Query(
		`SELECT tr.team_id, tr.role, COALESCE(t.name, '') as team_name
		 FROM team_roles tr
		 JOIN teams t ON t.id = tr.team_id
		 WHERE tr.role = $1
		 ORDER BY t.name`,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by role: %w", err)
	}
	defer rows.Close()

	var teams []*TeamWithRole
	for rows.Next() {
		t := &TeamWithRole{}
		if err := rows.Scan(&t.TeamID, &t.Role, &t.TeamName); err != nil {
			return nil, fmt.Errorf("failed to scan team role: %w", err)
		}
		teams = append(teams, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if teams == nil {
		return []*TeamWithRole{}, nil
	}

	return teams, nil
}

func (r *teamRoleRepo) UserIsMemberOfTeamWithRole(userID int, role string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM team_roles tr
			JOIN team_members tm ON tm.team_id = tr.team_id
			WHERE tm.user_id = $1 AND tr.role = $2
		)`,
		userID, role,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user team role membership: %w", err)
	}
	return exists, nil
}
