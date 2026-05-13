package seed

import (
	"database/sql"
	"fmt"
)

func SeedRolePermissions(db *sql.DB) error {
	permMap := make(map[string]int)
	rows, err := db.Query(`SELECT id, feature FROM permissions`)
	if err != nil {
		return fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var feature string
		if err := rows.Scan(&id, &feature); err != nil {
			return fmt.Errorf("failed to scan permission: %w", err)
		}
		permMap[feature] = id
	}

	if err := rows.Err(); err != nil {
		return err
	}

	type assignment struct {
		role    string
		feature string
	}

	assignments := []assignment{
		{role: "captain", feature: "team:edit"},
		{role: "captain", feature: "team:delete"},
		{role: "captain", feature: "team:leave"},
		{role: "captain", feature: "team:invite"},
		{role: "captain", feature: "team:requests:view"},
		{role: "captain", feature: "team:requests:manage"},
		{role: "captain", feature: "member:view"},
		{role: "captain", feature: "member:kick"},
		{role: "captain", feature: "member:role:assign"},
		{role: "captain", feature: "member:role:promote-cocaptain"},
		{role: "captain", feature: "member:role:transfer-captain"},
		{role: "captain", feature: "zone:view"},
		{role: "captain", feature: "zone:create"},
		{role: "captain", feature: "zone:edit"},
		{role: "captain", feature: "zone:delete"},
		{role: "captain", feature: "zone:publish"},
		{role: "captain", feature: "zone:lock"},
		{role: "captain", feature: "review:start"},
		{role: "captain", feature: "template:duplicate"},
		{role: "captain", feature: "edit_history:view"},
		{role: "captain", feature: "edit_history:rollback"},

		{role: "cocaptain", feature: "team:edit"},
		{role: "cocaptain", feature: "team:leave"},
		{role: "cocaptain", feature: "team:invite"},
		{role: "cocaptain", feature: "team:requests:view"},
		{role: "cocaptain", feature: "team:requests:manage"},
		{role: "cocaptain", feature: "member:view"},
		{role: "cocaptain", feature: "member:role:assign"},
		{role: "cocaptain", feature: "zone:view"},
		{role: "cocaptain", feature: "zone:create"},
		{role: "cocaptain", feature: "zone:edit"},
		{role: "cocaptain", feature: "zone:delete"},
		{role: "cocaptain", feature: "zone:publish"},
		{role: "cocaptain", feature: "zone:lock"},
		{role: "cocaptain", feature: "review:start"},
		{role: "cocaptain", feature: "template:duplicate"},
		{role: "cocaptain", feature: "edit_history:view"},

		{role: "editor", feature: "team:leave"},
		{role: "editor", feature: "member:view"},
		{role: "editor", feature: "zone:view"},
		{role: "editor", feature: "zone:create"},
		{role: "editor", feature: "zone:edit"},
		{role: "editor", feature: "zone:delete"},
		{role: "editor", feature: "zone:lock"},
		{role: "editor", feature: "review:start"},
		{role: "editor", feature: "template:duplicate"},

		{role: "member", feature: "team:create"},
		{role: "member", feature: "team:leave"},
		{role: "member", feature: "member:view"},
		{role: "member", feature: "zone:view"},
		{role: "member", feature: "review:start"},
		{role: "member", feature: "template:duplicate"},
		{role: "member", feature: "edit_history:view"},

		{role: "captain", feature: "split:view"},
		{role: "cocaptain", feature: "split:view"},
		{role: "editor", feature: "split:view"},
		{role: "member", feature: "split:view"},
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, a := range assignments {
		permID, ok := permMap[a.feature]
		if !ok {
			return fmt.Errorf("permission not found for feature %q", a.feature)
		}

		_, err := tx.Exec(
			`INSERT INTO role_permissions (role, permission_id) VALUES ($1, $2)
			 ON CONFLICT (role, permission_id) DO NOTHING`,
			a.role, permID,
		)
		if err != nil {
			return fmt.Errorf("failed to assign %q to role %q: %w", a.feature, a.role, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
