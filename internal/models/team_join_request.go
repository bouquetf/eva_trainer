package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TeamJoinRequest struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	UserID    int       `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type teamJoinRequestRepo struct {
	db             *sql.DB
	teamMemberRepo TeamMemberRepository
}

func NewTeamJoinRequestRepository(db *sql.DB, teamMemberRepo TeamMemberRepository) TeamJoinRequestRepository {
	return &teamJoinRequestRepo{db: db, teamMemberRepo: teamMemberRepo}
}

func (r *teamJoinRequestRepo) Create(teamID, userID int) (*TeamJoinRequest, error) {
	_, err := r.teamMemberRepo.FindByUserAndTeam(userID, teamID)
	if err == nil {
		return nil, ErrJoinRequestAlreadyMember
	}
	if err != ErrMemberNotFound {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	jr := &TeamJoinRequest{}
	err = r.db.QueryRow(
		`INSERT INTO team_join_requests (team_id, user_id, status)
		 VALUES ($1, $2, 'pending')
		 RETURNING id, team_id, user_id, status, created_at`,
		teamID, userID,
	).Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrJoinRequestAlreadyExists
		}
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}

	return jr, nil
}

func (r *teamJoinRequestRepo) FindByID(id int) (*TeamJoinRequest, error) {
	jr := &TeamJoinRequest{}
	err := r.db.QueryRow(
		`SELECT id, team_id, user_id, status, created_at
		 FROM team_join_requests WHERE id = $1`,
		id,
	).Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrJoinRequestNotFound
	}
	if err != nil {
		return nil, err
	}

	return jr, nil
}

func (r *teamJoinRequestRepo) FindPendingByTeam(teamID int) ([]*TeamJoinRequest, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, user_id, status, created_at
		 FROM team_join_requests
		 WHERE team_id = $1 AND status = 'pending'
		 ORDER BY created_at ASC`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*TeamJoinRequest
	for rows.Next() {
		jr := &TeamJoinRequest{}
		if err := rows.Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, jr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if reqs == nil {
		return []*TeamJoinRequest{}, nil
	}
	return reqs, nil
}

func (r *teamJoinRequestRepo) FindPendingByUser(userID int) ([]*TeamJoinRequest, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, user_id, status, created_at
		 FROM team_join_requests
		 WHERE user_id = $1 AND status = 'pending'
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*TeamJoinRequest
	for rows.Next() {
		jr := &TeamJoinRequest{}
		if err := rows.Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, jr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if reqs == nil {
		return []*TeamJoinRequest{}, nil
	}
	return reqs, nil
}

func (r *teamJoinRequestRepo) FindByUserAndTeam(userID, teamID int) (*TeamJoinRequest, error) {
	jr := &TeamJoinRequest{}
	err := r.db.QueryRow(
		`SELECT id, team_id, user_id, status, created_at
		 FROM team_join_requests WHERE team_id = $1 AND user_id = $2`,
		teamID, userID,
	).Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrJoinRequestNotFound
	}
	if err != nil {
		return nil, err
	}

	return jr, nil
}

func (r *teamJoinRequestRepo) AcceptRequest(id int) error {
	jr, err := r.FindByID(id)
	if err != nil {
		return err
	}

	if jr.Status != "pending" {
		return fmt.Errorf("request already processed: %w", ErrJoinRequestNotFound)
	}

	_, err = r.teamMemberRepo.FindByUserAndTeam(jr.UserID, jr.TeamID)
	if err == nil {
		return ErrJoinRequestAlreadyMember
	}
	if err != ErrMemberNotFound {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`UPDATE team_join_requests SET status = 'accepted' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update join request: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrJoinRequestNotFound
	}

	_, err = tx.Exec(
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`,
		jr.TeamID, jr.UserID, RoleMember,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateTeamMember
		}
		return fmt.Errorf("failed to add team member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *teamJoinRequestRepo) DeclineRequest(id int) error {
	result, err := r.db.Exec(
		`UPDATE team_join_requests SET status = 'declined' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to decline join request: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrJoinRequestNotFound
	}
	return nil
}

func (r *teamJoinRequestRepo) CancelRequest(id int) error {
	result, err := r.db.Exec(
		`UPDATE team_join_requests SET status = 'cancelled' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel join request: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrJoinRequestNotFound
	}
	return nil
}
