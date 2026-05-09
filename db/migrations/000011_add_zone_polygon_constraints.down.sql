DROP INDEX IF EXISTS idx_edit_locks_expires_at;
ALTER TABLE zone_names DROP CONSTRAINT IF EXISTS zone_names_zone_id_key;
