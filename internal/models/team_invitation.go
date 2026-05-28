package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type TeamInvitation struct {
	ID              int       `json:"id"`
	TeamID          int       `json:"team_id"`
	InvitedUserID   *int      `json:"invited_user_id,omitempty"`
	InvitedByUserID int       `json:"invited_by_user_id"`
	Token           *string   `json:"token,omitempty"`
	Status          string    `json:"status"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type teamInvitationRepo struct {
	db             *sql.DB
	teamMemberRepo TeamMemberRepository
}

func NewTeamInvitationRepository(db *sql.DB, teamMemberRepo TeamMemberRepository) TeamInvitationRepository {
	return &teamInvitationRepo{db: db, teamMemberRepo: teamMemberRepo}
}

func GenerateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (r *teamInvitationRepo) CreateByUser(teamID, invitedUserID, invitedByUserID int) (*TeamInvitation, error) {
	if invitedUserID == invitedByUserID {
		return nil, ErrCannotInviteSelf
	}

	_, err := r.teamMemberRepo.FindByUserAndTeam(invitedUserID, teamID)
	if err == nil {
		return nil, ErrJoinRequestAlreadyMember
	}
	if !errors.Is(err, ErrMemberNotFound) {
		return nil, fmt.Errorf("failed to check team membership: %w", err)
	}

	inv := &TeamInvitation{}
	err = r.db.QueryRow(
		`INSERT INTO team_invitations (team_id, invited_user_id, invited_by_user_id, status, expires_at)
		 VALUES ($1, $2, $3, 'pending', NOW() + INTERVAL '1 day')
		 RETURNING id, team_id, invited_user_id, invited_by_user_id, token, status, expires_at, created_at`,
		teamID, invitedUserID, invitedByUserID,
	).Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return inv, nil
}

func (r *teamInvitationRepo) CreateLink(teamID, invitedByUserID int, token string, expiresAt time.Time) (*TeamInvitation, error) {
	inv := &TeamInvitation{}
	err := r.db.QueryRow(
		`INSERT INTO team_invitations (team_id, invited_by_user_id, token, status, expires_at)
		 VALUES ($1, $2, $3, 'pending', $4)
		 RETURNING id, team_id, invited_user_id, invited_by_user_id, token, status, expires_at, created_at`,
		teamID, invitedByUserID, token, expiresAt.UTC(),
	).Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("duplicate token, try again: %w", err)
		}
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	return inv, nil
}

func (r *teamInvitationRepo) FindByID(id int) (*TeamInvitation, error) {
	inv := &TeamInvitation{}
	err := r.db.QueryRow(
		`SELECT id, team_id, invited_user_id, invited_by_user_id, token, status, expires_at, created_at
		 FROM team_invitations WHERE id = $1`,
		id,
	).Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (r *teamInvitationRepo) FindByToken(token string) (*TeamInvitation, error) {
	inv := &TeamInvitation{}
	err := r.db.QueryRow(
		`SELECT id, team_id, invited_user_id, invited_by_user_id, token, status, expires_at, created_at
		 FROM team_invitations WHERE token = $1`,
		token,
	).Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (r *teamInvitationRepo) FindPendingByTeam(teamID int) ([]*TeamInvitation, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, invited_user_id, invited_by_user_id, token, status, expires_at, created_at
		 FROM team_invitations
		 WHERE team_id = $1 AND status = 'pending'
		 ORDER BY created_at DESC`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []*TeamInvitation
	for rows.Next() {
		inv := &TeamInvitation{}
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if invs == nil {
		return []*TeamInvitation{}, nil
	}
	return invs, nil
}

func (r *teamInvitationRepo) FindPendingByUser(userID int) ([]*TeamInvitation, error) {
	rows, err := r.db.Query(
		`SELECT i.id, i.team_id, i.invited_user_id, i.invited_by_user_id, i.token, i.status, i.expires_at, i.created_at
		 FROM team_invitations i
		 WHERE i.invited_user_id = $1 AND i.status = 'pending'
		 ORDER BY i.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []*TeamInvitation
	for rows.Next() {
		inv := &TeamInvitation{}
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.InvitedUserID, &inv.InvitedByUserID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if invs == nil {
		return []*TeamInvitation{}, nil
	}
	return invs, nil
}

func (r *teamInvitationRepo) AcceptInvitation(id int) error {
	inv, err := r.FindByID(id)
	if err != nil {
		return err
	}

	if time.Now().After(inv.ExpiresAt) {
		return ErrInvitationExpired
	}

	if inv.Status != "pending" {
		return ErrInvitationAlreadyUsed
	}

	if inv.InvitedUserID == nil {
		return fmt.Errorf("cannot accept a link invitation by id: %w", ErrInvitationNotFound)
	}

	_, err = r.teamMemberRepo.FindByUserAndTeam(*inv.InvitedUserID, inv.TeamID)
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
		`UPDATE team_invitations SET status = 'accepted' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationAlreadyUsed
	}

	_, err = tx.Exec(
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`,
		inv.TeamID, *inv.InvitedUserID, RoleMember,
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

func (r *teamInvitationRepo) AcceptByToken(token string, userID int) error {
	inv, err := r.FindByToken(token)
	if err != nil {
		return err
	}

	if time.Now().UTC().After(inv.ExpiresAt.UTC()) {
		return ErrInvitationExpired
	}

	if inv.Status != "pending" {
		return ErrInvitationAlreadyUsed
	}

	_, err = r.teamMemberRepo.FindByUserAndTeam(userID, inv.TeamID)
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
		`UPDATE team_invitations SET status = 'accepted', invited_user_id = $1 WHERE token = $2 AND status = 'pending'`,
		userID, token,
	)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationAlreadyUsed
	}

	_, err = tx.Exec(
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`,
		inv.TeamID, userID, RoleMember,
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

func (r *teamInvitationRepo) DeclineInvitation(id int) error {
	result, err := r.db.Exec(
		`UPDATE team_invitations SET status = 'declined' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *teamInvitationRepo) CancelInvitation(id int) error {
	result, err := r.db.Exec(
		`UPDATE team_invitations SET status = 'cancelled' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *teamInvitationRepo) ExpireOverdue() (int64, error) {
	result, err := r.db.Exec(
		`UPDATE team_invitations SET status = 'expired' WHERE status = 'pending' AND expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to expire overdue invitations: %w", err)
	}
	return result.RowsAffected()
}
