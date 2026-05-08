package models

import (
	"database/sql"
	"fmt"
)

type RolePermission struct {
	Role         string `json:"role"`
	PermissionID int    `json:"permission_id"`
}

type rolePermissionRepo struct {
	db *sql.DB
}

func NewRolePermissionRepository(db *sql.DB) RolePermissionRepository {
	return &rolePermissionRepo{db: db}
}

func (r *rolePermissionRepo) FindByRole(role string) ([]*Permission, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.feature, p.description
		 FROM permissions p
		 JOIN role_permissions rp ON p.id = rp.permission_id
		 WHERE rp.role = $1
		 ORDER BY p.feature`,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions for role: %w", err)
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

func (r *rolePermissionRepo) FindByPermissionID(permissionID int) ([]*RolePermission, error) {
	rows, err := r.db.Query(
		`SELECT role, permission_id FROM role_permissions WHERE permission_id = $1`,
		permissionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query role_permissions: %w", err)
	}
	defer rows.Close()

	var rolePerms []*RolePermission
	for rows.Next() {
		rp := &RolePermission{}
		if err := rows.Scan(&rp.Role, &rp.PermissionID); err != nil {
			return nil, fmt.Errorf("failed to scan role_permission: %w", err)
		}
		rolePerms = append(rolePerms, rp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if rolePerms == nil {
		return []*RolePermission{}, nil
	}

	return rolePerms, nil
}

func (r *rolePermissionRepo) FindAllRoles() ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT role FROM role_permissions ORDER BY role`)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *rolePermissionRepo) AssignPermissionToRole(role string, permissionID int) error {
	_, err := r.db.Exec(
		`INSERT INTO role_permissions (role, permission_id) VALUES ($1, $2)
		 ON CONFLICT (role, permission_id) DO NOTHING`,
		role, permissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

func (r *rolePermissionRepo) RemovePermissionFromRole(role string, permissionID int) error {
	result, err := r.db.Exec(
		`DELETE FROM role_permissions WHERE role = $1 AND permission_id = $2`,
		role, permissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPermissionNotFound
	}

	return nil
}

func (r *rolePermissionRepo) RoleHasPermission(role string, permissionID int) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM role_permissions WHERE role = $1 AND permission_id = $2`,
		role, permissionID,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}

	return count > 0, nil
}
