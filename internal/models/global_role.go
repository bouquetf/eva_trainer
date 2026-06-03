package models

import (
	"database/sql"
	"fmt"
)

type UserGlobalRole struct {
	UserID      int    `json:"user_id"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name,omitempty"`
}

type globalRoleRepo struct {
	db *sql.DB
}

func NewGlobalRoleRepository(db *sql.DB) GlobalRoleRepository {
	return &globalRoleRepo{db: db}
}

func (r *globalRoleRepo) Add(userID int, role string) error {
	_, err := r.db.Exec(
		`INSERT INTO global_roles (user_id, role) VALUES ($1, $2)`,
		userID, role,
	)
	if err != nil {
		return fmt.Errorf("failed to add global role: %w", err)
	}
	return nil
}

func (r *globalRoleRepo) Remove(userID int, role string) error {
	result, err := r.db.Exec(
		`DELETE FROM global_roles WHERE user_id = $1 AND role = $2`,
		userID, role,
	)
	if err != nil {
		return fmt.Errorf("failed to remove global role: %w", err)
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

func (r *globalRoleRepo) UserHasRole(userID int, role string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM global_roles WHERE user_id = $1 AND role = $2)`,
		userID, role,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check global role: %w", err)
	}
	return exists, nil
}

func (r *globalRoleRepo) FindUsersByRole(role string) ([]*UserGlobalRole, error) {
	rows, err := r.db.Query(
		`SELECT gr.user_id, gr.role, COALESCE(u.display_name, '') as display_name
		 FROM global_roles gr
		 JOIN users u ON u.id = gr.user_id
		 WHERE gr.role = $1
		 ORDER BY u.display_name`,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by role: %w", err)
	}
	defer rows.Close()

	var users []*UserGlobalRole
	for rows.Next() {
		u := &UserGlobalRole{}
		if err := rows.Scan(&u.UserID, &u.Role, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan user global role: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if users == nil {
		return []*UserGlobalRole{}, nil
	}

	return users, nil
}
