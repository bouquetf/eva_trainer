CREATE TABLE lines_of_fire (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE NOT NULL,
    map_id INTEGER REFERENCES maps(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(250) NOT NULL,
    polygon_data JSONB NOT NULL DEFAULT '{"cones":[]}'::jsonb,
    order_index INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_lines_of_fire_team_map ON lines_of_fire(team_id, map_id);
