# Tâche #001 - Story #035 : Migration BDD + Modèle LigueLocale

## Objectif
Créer les tables `ligue_locale` et `ligue_locale_association` en base de données, et les modèles Go associés (struct, repository, tests unitaires).

## Contexte
- Story #035 : `docs/stories/story-035.md`
- Dépend de : Story #034 (rôles archiviste/lecteur déjà en place)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Deux entités liées par une table de jonction many-to-many :

- **LigueLocale** (table `ligue_locale`) : lieu de compétition — `id`, `nom`, `emplacement` (unique), `created_at`, `updated_at`
- **Ligue** (table `ligue`) : période de compétition — `id`, `date_debut`, `date_fin`, `split_id` → `splits`, `created_at`, `updated_at`
- **LigueLocaleAssociation** (table `ligue_locale_association`) : association many-to-many — `ligue_id`, `ligue_locale_id` (UNIQUE, FK avec ON DELETE CASCADE)

**Cas nominaux :**
- Création d'une ligue locale avec nom et emplacement
- Lecture d'une ligue locale par ID
- Liste de toutes les ligues locales, avec possibilité de filtrer par emplacement
- Mise à jour d'une ligue locale (nom, emplacement)
- Suppression d'une ligue locale (cascade supprime la jonction)
- Associer une ligue locale à une ligue (INSERT dans la jonction)
- Dissocier une ligue locale d'une ligue (DELETE dans la jonction)
- Trouver toutes les ligues associées à une ligue locale
- Trouver toutes les ligues locales associées à une ligue
- Création d'une ligue (période) avec date début, date fin, split
- Liste des ligues avec pagination
- Association en masse de toutes les ligues locales existantes à une ligue (`AssociateAllLocales`)

**Cas limites :**
- Nom de ligue locale vide → rejeté
- Emplacement doit être unique (contrainte unique en base)
- Filtre par emplacement : si l'emplacement n'existe pas → liste vide (pas d'erreur)
- `AssociateAllLocales` utilise `ON CONFLICT DO NOTHING` — ne crée pas de doublons
- Supprimer une ligue ne supprime pas les ligues locales (seule la jonction est supprimée par FK cascade)

**Gestion d'erreurs :**
- `ErrLeagueNameRequired` si nom vide
- `ErrLeagueNotFound` si ID inconnu
- `ErrLeagueAlreadyExists` si contrainte d'unicité violée (emplacement)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000026_add_archive_leagues.up.sql` | Créer | Migration haute — table `ligue_locale` |
| `db/migrations/000026_add_archive_leagues.down.sql` | Créer | Migration basse |
| `db/migrations/000027_fix_ligue_locale_table.up.sql` | Créer | Fix avec `IF NOT EXISTS` |
| `db/migrations/000027_fix_ligue_locale_table.down.sql` | Créer | Migration basse |
| `db/migrations/000028_add_ligue_locale_emplacement_unique.up.sql` | Créer | Unicité sur `emplacement` + déduplication |
| `db/migrations/000028_add_ligue_locale_emplacement_unique.down.sql` | Créer | Migration basse |
| `db/migrations/000029_add_ligue_locale_junction.up.sql` | Créer | Table `ligue_locale_association`, migre données, drop `ligue_id` |
| `db/migrations/000029_add_ligue_locale_junction.down.sql` | Créer | Migration basse |
| `internal/models/ligue_locale.go` | Créer | Struct LigueLocale + repository |
| `internal/models/ligue.go` | Créer | Struct Ligue + repository |
| `internal/models/repository.go` | Modifier | Ajouter interfaces + méthodes d'association |
| `internal/models/ligue_locale_test.go` | Créer | Tests intégration ligue locale |
| `internal/models/ligue_test.go` | Créer | Tests intégration ligue |
| `internal/models/setup_test.go` | Modifier | Ajouter migrations 000026-000029 |

### Modèle de données

```sql
CREATE TABLE ligue_locale (
    id SERIAL PRIMARY KEY,
    nom VARCHAR(255) NOT NULL,
    emplacement VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE ligue (
    id SERIAL PRIMARY KEY,
    date_debut DATE NOT NULL,
    date_fin DATE NOT NULL,
    split_id INTEGER NOT NULL REFERENCES splits(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE ligue_locale_association (
    ligue_id INTEGER NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    ligue_locale_id INTEGER NOT NULL REFERENCES ligue_locale(id) ON DELETE CASCADE,
    UNIQUE(ligue_id, ligue_locale_id)
);
```

### Signatures

```go
// internal/models/ligue_locale.go

type LigueLocale struct {
    ID          int       `json:"id"`
    Nom         string    `json:"nom"`
    Emplacement string    `json:"emplacement"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type ligueLocaleRepo struct {
    db *sql.DB
}

func NewLigueLocaleRepository(db *sql.DB) LigueLocaleRepository

// LigueLocaleRepository (dans repository.go)
type LigueLocaleRepository interface {
    Create(nom string, emplacement string) (*LigueLocale, error)
    FindByID(id int) (*LigueLocale, error)
    FindAll(emplacement string) ([]*LigueLocale, error)
    Update(id int, nom string, emplacement string) (*LigueLocale, error)
    Delete(id int) error
    AssociateLigue(localeID, ligueID int) error
    DisassociateLigue(localeID, ligueID int) error
    FindLiguesByLocaleID(localeID int) ([]*Ligue, error)
    FindLocalesByLigueID(ligueID int) ([]*LigueLocale, error)
}

// internal/models/ligue.go

type Ligue struct {
    ID        int       `json:"id"`
    DateDebut time.Time `json:"date_debut"`
    DateFin   time.Time `json:"date_fin"`
    SplitID   int       `json:"split_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type ligueRepo struct {
    db *sql.DB
}

func NewLigueRepository(db *sql.DB) LigueRepository

type LigueRepository interface {
    Create(dateDebut, dateFin time.Time, splitID int) (*Ligue, error)
    FindByID(id int) (*Ligue, error)
    FindAll(page, perPage int) ([]*Ligue, int, error)
    Update(id int, dateDebut, dateFin time.Time, splitID int) (*Ligue, error)
    Delete(id int) error
    AssociateAllLocales(ligueID int) (int, error)
    HasBOs(ligueID int) (bool, error)
}
```

### Contraintes techniques
- **Framework** : Go standard library + database/sql + pgx
- **Pattern** : Suivre exactement le pattern de `internal/models/split.go` (constructeur `NewXRepository`, méthodes avec `scanX`, requêtes SQL avec `$1`, `$2`, etc.)
- **Dates** : Stocker en `DATE` PostgreSQL, utiliser `time.Time` en Go
- **Migration** : Utiliser `CREATE TABLE IF NOT EXISTS` pour l'idempotence (000027)
- **Tests** : Les tests doivent utiliser `setupTestDB` pour créer un conteneur PostgreSQL
- **AssociateAllLocales** : `INSERT INTO ligue_locale_association (ligue_id, ligue_locale_id) SELECT $1, id FROM ligue_locale ON CONFLICT DO NOTHING` — retourne le nombre de lignes insérées
- **HasBOs** : Vérifie l'existence de BOs associés à la ligue (table `bos` créée dans Story #038). Retourne `false` si la table n'existe pas encore (requête avec vérification de l'existence de la table)

### Tests à implémenter

#### Tests LigueLocale (intégration)
- **Fichier** : `internal/models/ligue_locale_test.go`

1. **Create_Success** : Créer une ligue locale → vérifier tous les champs + ID > 0
2. **Create_NomVide** : `repo.Create("", "Salle A")` → `ErrLeagueNameRequired`
3. **FindByID_Success** : Créer puis FindByID → champs identiques
4. **FindByID_NotFound** : ID 99999 → `ErrLeagueNotFound`
5. **FindAll_NoFilter** : Créer 2 ligues locales → FindAll("") → 2 résultats
6. **FindAll_FilterByEmplacement** : Créer avec emplacements différents → filter → 1 résultat
7. **FindAll_FilterNoMatch** : FindAll("Inexistant") → slice vide (pas d'erreur)
8. **Update_Success** : Créer puis update → champs mis à jour
9. **Delete_Success** : Créer puis delete → FindByID retourne `ErrLeagueNotFound`
10. **AssociateLigue** : Créer locale + ligue → AssociateLigue → FindLiguesByLocaleID → 1 résultat
11. **DisassociateLigue** : Associer puis dissocier → FindLiguesByLocaleID → 0 résultats

#### Tests Ligue (intégration)
- **Fichier** : `internal/models/ligue_test.go`

1. **Create_Success** : Créer une ligue → vérifier tous les champs
2. **FindByID_Success** : Créer puis FindByID → champs identiques
3. **FindAll_WithPagination** : Créer plusieurs ligues → pagination correcte
4. **Update_Success** : Créer puis update → champs mis à jour
5. **Delete_Success** : Créer puis delete → FindByID retourne `ErrLeagueNotFound`
6. **AssociateAllLocales** : Créer ligue + 2 locales → AssociateAllLocales → 2 associées
7. **AssociateAllLocales_Idempotent** : Appeler 2 fois → même nombre d'associations
8. **HasBOs_NoBOs** : Créer ligue sans BO → `false`
9. **FindLocalesByLigueID** : Créer ligue + 2 locales associées → 2 résultats
