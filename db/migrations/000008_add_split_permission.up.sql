INSERT INTO permissions (feature, description) VALUES
    ('split:view', 'Voir les splits et leurs cartes')
ON CONFLICT (feature) DO NOTHING;
