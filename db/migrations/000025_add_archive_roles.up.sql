CREATE TABLE IF NOT EXISTS global_roles (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (user_id, role)
);

CREATE TABLE IF NOT EXISTS team_roles (
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (team_id, role)
);

INSERT INTO permissions (feature, description) VALUES
    ('archive:read', 'Consulter les archives'),
    ('archive:write', 'Gérer les archives (création, modification, suppression)')
ON CONFLICT (feature) DO NOTHING;
