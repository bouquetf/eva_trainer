CREATE TABLE IF NOT EXISTS divs (
    id SERIAL PRIMARY KEY,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    nom VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_divs_ligue_nom ON divs(ligue_id, nom);

CREATE TABLE IF NOT EXISTS equipe_eva_div (
    equipe_eva_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE CASCADE,
    div_id INT NOT NULL REFERENCES divs(id) ON DELETE CASCADE,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    PRIMARY KEY (equipe_eva_id, div_id),
    UNIQUE (equipe_eva_id, ligue_id)
);

CREATE INDEX IF NOT EXISTS idx_equipe_eva_div_div_id ON equipe_eva_div(div_id);
CREATE INDEX IF NOT EXISTS idx_equipe_eva_div_ligue_id ON equipe_eva_div(ligue_id);

CREATE TABLE IF NOT EXISTS bos (
    id SERIAL PRIMARY KEY,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    equipe1_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE RESTRICT,
    equipe2_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE RESTRICT,
    vainqueur_id INT NULL REFERENCES eva_teams(id) ON DELETE SET NULL,
    format VARCHAR(3) NOT NULL DEFAULT 'BO3' CHECK (format IN ('BO3', 'BO5')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bos_ligue_id ON bos(ligue_id);

CREATE TABLE IF NOT EXISTS matchs (
    id SERIAL PRIMARY KEY,
    bo_id INT NOT NULL REFERENCES bos(id) ON DELETE CASCADE,
    map_id INT NOT NULL REFERENCES maps(id) ON DELETE RESTRICT,
    score_equipe1 INT NOT NULL CHECK (score_equipe1 >= 0 AND score_equipe1 <= 100),
    score_equipe2 INT NOT NULL CHECK (score_equipe2 >= 0 AND score_equipe2 <= 100),
    vainqueur_id INT NULL REFERENCES eva_teams(id) ON DELETE SET NULL,
    commentaires TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matchs_bo_id ON matchs(bo_id);
