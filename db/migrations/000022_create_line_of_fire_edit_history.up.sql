CREATE TABLE line_of_fire_edit_history (
    id SERIAL PRIMARY KEY,
    line_of_fire_id INTEGER REFERENCES lines_of_fire(id) ON DELETE CASCADE NOT NULL,
    editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_lof_edit_history_line ON line_of_fire_edit_history(line_of_fire_id);
