-- Ensure one name per zone by removing duplicates, keeping the most recent
DELETE FROM zone_names
WHERE id NOT IN (
  SELECT MIN(id)
  FROM zone_names
  GROUP BY zone_id
);

-- Add UNIQUE constraint on zone_names.zone_id to enforce one name per zone
ALTER TABLE zone_names ADD CONSTRAINT zone_names_zone_id_key UNIQUE (zone_id);

-- Add index on edit_locks.expires_at for efficient cleanup of expired locks
CREATE INDEX IF NOT EXISTS idx_edit_locks_expires_at ON edit_locks(expires_at);
