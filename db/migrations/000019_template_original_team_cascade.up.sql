ALTER TABLE templates DROP CONSTRAINT IF EXISTS templates_original_team_id_fkey;
ALTER TABLE templates ADD CONSTRAINT templates_original_team_id_fkey
    FOREIGN KEY (original_team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE templates ALTER COLUMN original_team_id SET NOT NULL;
