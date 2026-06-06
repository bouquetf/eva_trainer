ALTER TABLE ligue_locale ADD COLUMN ligue_id INTEGER REFERENCES ligue(id);

UPDATE ligue_locale
SET ligue_id = sub.ligue_id
FROM (
    SELECT DISTINCT ON (ligue_locale_id) ligue_id, ligue_locale_id
    FROM ligue_locale_association
    ORDER BY ligue_locale_id, id DESC
) sub
WHERE ligue_locale.id = sub.ligue_locale_id;

DROP TABLE IF EXISTS ligue_locale_association;
