DELETE FROM ligue_locale a USING ligue_locale b
WHERE a.id > b.id AND a.emplacement = b.emplacement;

CREATE UNIQUE INDEX IF NOT EXISTS ligue_locale_emplacement_key ON ligue_locale (emplacement);
