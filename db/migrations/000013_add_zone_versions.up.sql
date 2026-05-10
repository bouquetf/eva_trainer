CREATE TABLE IF NOT EXISTS zone_versions (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zone_versions_zone ON zone_versions(zone_id);
CREATE INDEX IF NOT EXISTS idx_zone_versions_editor ON zone_versions(editor_id);
CREATE INDEX IF NOT EXISTS idx_zone_versions_created ON zone_versions(created_at);
