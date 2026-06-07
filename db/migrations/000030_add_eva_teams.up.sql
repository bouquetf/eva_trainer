CREATE TABLE IF NOT EXISTS eva_teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tag VARCHAR(50) NOT NULL,
    eva_trainer_team_id INT REFERENCES teams(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(eva_trainer_team_id)
);

CREATE INDEX IF NOT EXISTS idx_eva_teams_name ON eva_teams(name);
