CREATE TABLE IF NOT EXISTS ligue_locale_association (
    id SERIAL PRIMARY KEY,
    ligue_id INTEGER NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    ligue_locale_id INTEGER NOT NULL REFERENCES ligue_locale(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(ligue_id, ligue_locale_id)
);

INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id)
SELECT ligue_id, id FROM ligue_locale WHERE ligue_id IS NOT NULL;

ALTER TABLE ligue_locale DROP COLUMN ligue_id;
