CREATE TABLE line_of_fire_versions (
    id SERIAL PRIMARY KEY,
    line_of_fire_id INTEGER REFERENCES lines_of_fire(id) ON DELETE CASCADE NOT NULL,
    editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_lof_versions_line ON line_of_fire_versions(line_of_fire_id);
