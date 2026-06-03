package models

import (
	"database/sql"
	"fmt"
	"time"
)

type TeamMemberWithDisplayName struct {
	TeamID      int       `json:"team_id"`
	UserID      int       `json:"user_id"`
	Role        string    `json:"role"`
	DisplayName string    `json:"display_name"`
	JoinedAt    time.Time `json:"joined_at"`
}

type TeamMember struct {
	TeamID   int       `json:"team_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

const (
	RoleCaptain   = "captain"
	RoleCocaptain = "cocaptain"
	RoleEditor    = "editor"
	RoleMember    = "member"
	RoleArchivist = "archivist"
	RoleReader    = "reader"
)

var ValidRoles = map[string]bool{
	RoleCaptain:   true,
	RoleCocaptain: true,
	RoleEditor:    true,
	RoleMember:    true,
}

type teamMemberRepo struct {
	db *sql.DB
}

func NewTeamMemberRepository(db *sql.DB) TeamMemberRepository {
	return &teamMemberRepo{db: db}
}

func (r *teamMemberRepo) AddMember(teamID, userID int, role string) error {
	if !ValidRoles[role] {
		return ErrInvalidRole
	}

	_, err := r.db.Exec(
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`,
		teamID, userID, role,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("user is already a member: %w", ErrDuplicateTeamMember)
		}
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (r *teamMemberRepo) RemoveMember(teamID, userID int) error {
	result, err := r.db.Exec(
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *teamMemberRepo) UpdateRole(teamID, userID int, role string) error {
	if !ValidRoles[role] {
		return ErrInvalidRole
	}

	if role == RoleCocaptain {
		count, err := r.CountByRole(teamID, RoleCocaptain)
		if err != nil {
			return fmt.Errorf("failed to check cocaptain count: %w", err)
		}
		if count > 0 {
			return ErrTeamAlreadyHasCocaptain
		}
	}

	result, err := r.db.Exec(
		`UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3`,
		role, teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *teamMemberRepo) FindByTeamIDWithDisplayNames(teamID int) ([]*TeamMemberWithDisplayName, error) {
	rows, err := r.db.Query(
		`SELECT tm.team_id, tm.user_id, tm.role, COALESCE(u.display_name, '') as display_name, tm.joined_at
		 FROM team_members tm
		 JOIN users u ON u.id = tm.user_id
		 WHERE tm.team_id = $1
		 ORDER BY tm.joined_at`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*TeamMemberWithDisplayName
	for rows.Next() {
		m := &TeamMemberWithDisplayName{}
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role, &m.DisplayName, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if members == nil {
		return []*TeamMemberWithDisplayName{}, nil
	}

	return members, nil
}

func (r *teamMemberRepo) FindByTeamID(teamID int) ([]*TeamMember, error) {
	rows, err := r.db.Query(
		`SELECT team_id, user_id, role, joined_at FROM team_members WHERE team_id = $1 ORDER BY joined_at`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*TeamMember
	for rows.Next() {
		m := &TeamMember{}
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if members == nil {
		return []*TeamMember{}, nil
	}

	return members, nil
}

func (r *teamMemberRepo) FindByUserID(userID int) ([]*TeamMember, error) {
	rows, err := r.db.Query(
		`SELECT team_id, user_id, role, joined_at FROM team_members WHERE user_id = $1 ORDER BY joined_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*TeamMember
	for rows.Next() {
		m := &TeamMember{}
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if members == nil {
		return []*TeamMember{}, nil
	}
	return members, nil
}

func (r *teamMemberRepo) FindByUserAndTeam(userID, teamID int) (*TeamMember, error) {
	m := &TeamMember{}
	err := r.db.QueryRow(
		`SELECT team_id, user_id, role, joined_at FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID,
	).Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt)

	if err == sql.ErrNoRows {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *teamMemberRepo) GetCaptainID(teamID int) (int, error) {
	var userID int
	err := r.db.QueryRow(
		`SELECT user_id FROM team_members WHERE team_id = $1 AND role = 'captain'`,
		teamID,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		return 0, ErrTeamNotFound
	}
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *teamMemberRepo) CountByRole(teamID int, role string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND role = $2`,
		teamID, role,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (isPostgresUniqueViolation(err.Error()))
}

func isPostgresUniqueViolation(msg string) bool {
	return contains(msg, "duplicate key") || contains(msg, "unique constraint")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
