CREATE TABLE IF NOT EXISTS ligue (
    id SERIAL PRIMARY KEY,
    date_debut DATE NOT NULL,
    date_fin DATE NOT NULL,
    split_id INTEGER NOT NULL REFERENCES splits(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ligue_locale (
    id SERIAL PRIMARY KEY,
    nom VARCHAR(255) NOT NULL,
    emplacement VARCHAR(255) NOT NULL,
    ligue_id INTEGER REFERENCES ligue(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
