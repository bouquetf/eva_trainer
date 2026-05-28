CREATE TABLE template_lines_of_fire (
    id SERIAL PRIMARY KEY,
    template_id INTEGER REFERENCES templates(id) ON DELETE CASCADE NOT NULL,
    polygon_data JSONB NOT NULL DEFAULT '{"cones":[]}'::jsonb,
    name VARCHAR(250) NOT NULL,
    order_index INTEGER DEFAULT 0
);

CREATE INDEX idx_template_lines_of_fire_template ON template_lines_of_fire(template_id);
