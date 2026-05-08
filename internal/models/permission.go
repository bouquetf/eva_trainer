package models

import (
	"database/sql"
	"fmt"
)

type Permission struct {
	ID          int    `json:"id"`
	Feature     string `json:"feature"`
	Description string `json:"description"`
}

type permissionRepo struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) FindAll() ([]*Permission, error) {
	rows, err := r.db.Query(`SELECT id, feature, description FROM permissions ORDER BY feature`)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*Permission
	for rows.Next() {
		p := &Permission{}
		if err := rows.Scan(&p.ID, &p.Feature, &p.Description); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if permissions == nil {
		return []*Permission{}, nil
	}

	return permissions, nil
}

func (r *permissionRepo) FindByID(id int) (*Permission, error) {
	p := &Permission{}
	err := r.db.QueryRow(
		`SELECT id, feature, description FROM permissions WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Feature, &p.Description)

	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find permission by id: %w", err)
	}

	return p, nil
}

func (r *permissionRepo) FindByFeature(feature string) (*Permission, error) {
	p := &Permission{}
	err := r.db.QueryRow(
		`SELECT id, feature, description FROM permissions WHERE feature = $1`,
		feature,
	).Scan(&p.ID, &p.Feature, &p.Description)

	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find permission by feature: %w", err)
	}

	return p, nil
}

func (r *permissionRepo) Create(feature, description string) (*Permission, error) {
	p := &Permission{}
	err := r.db.QueryRow(
		`INSERT INTO permissions (feature, description) VALUES ($1, $2)
		 RETURNING id, feature, description`,
		feature, description,
	).Scan(&p.ID, &p.Feature, &p.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return p, nil
}

func HasPermission(
	role string,
	feature string,
	rolePermRepo RolePermissionRepository,
	permRepo PermissionRepository,
) (bool, error) {
	if role == "admin" {
		return true, nil
	}

	if feature == "" {
		return false, nil
	}

	perm, err := permRepo.FindByFeature(feature)
	if err == ErrPermissionNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to find permission: %w", err)
	}

	has, err := rolePermRepo.RoleHasPermission(role, perm.ID)
	if err != nil {
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}

	return has, nil
}
