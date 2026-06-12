# Tâche #001 - Story #045 : Migration BDD + Modèle Go + Repository

## Objectif

Ajouter le champ `nom` à l'entité `Ligue` en base de données, dans le modèle Go, le repository, et l'interface repository. Mettre à jour les tests du repository.

## Contexte

- Story #045 : `docs/stories/story-045.md`
- Dépend de : Rien
- Nécessaire pour : Tâche #002 (Handlers Go)

## Critères d'acceptation concernés
- `[ ]` L'archiviste peut saisir un nom (texte libre, obligatoire) lors de la création d'une ligue (ligne 18)
- `[ ]` L'archiviste peut modifier le nom d'une ligue existante (ligne 19)
- `[ ]` L'API expose les champs `nom` et `split_id` sur les endpoints ligue (ligne 23, partie modèle)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter un champ `nom` (texte, obligatoire) à l'entité `Ligue` (table `ligue`). Le nom est une chaîne libre saisie par l'archiviste pour identifier une période de compétition.

**Cas nominaux :**
- Créer une ligue avec un nom valide (non vide)
- Lire une ligue retourne le nom
- Mettre à jour une ligue avec un nouveau nom
- Lister les ligues retourne le nom

**Cas limites :**
- Nom vide → erreur `ErrLigueNameRequired`
- Les ligues existantes sans nom reçoivent une chaîne vide par défaut (migration avec `DEFAULT ''`)

**Gestion d'erreurs :**
- Nom vide lors de la création → `ErrLigueNameRequired`
- Nom vide lors de la mise à jour → `ErrLigueNameRequired`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000034_add_ligue_nom.up.sql` | Créer | Migration pour ajouter `nom` à `ligue` |
| `internal/models/ligue.go` | Modifier | Ajouter `Nom` au struct, mettre à jour toutes les requêtes SQL |
| `internal/models/repository.go` | Modifier | Ajouter `nom` aux signatures `Create` et `Update` de `LigueRepository` |
| `internal/models/ligue_test.go` | Modifier | Ajouter tests pour le champ `nom` |

### Signatures

```go
// Repository interface (dans repository.go)
type LigueRepository interface {
    Create(dateDebut, dateFin time.Time, splitID int, nom string) (*Ligue, error)
    FindByID(id int) (*Ligue, error)
    FindAll() ([]*Ligue, error)
    FindAllPaginated(page, perPage int) ([]*Ligue, int, error)
    Update(id int, dateDebut, dateFin time.Time, splitID int, nom string) (*Ligue, error)
    Delete(id int) error
    AssociateAllLocales(ligueID int) (int64, error)
}

// Dans ligue.go
type Ligue struct {
    ID        int       `json:"id"`
    Nom       string    `json:"nom"`
    DateDebut time.Time `json:"date_debut"`
    DateFin   time.Time `json:"date_fin"`
    SplitID   int       `json:"split_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Contraintes techniques

- **Migration** : Le numéro de migration est `000034`. La colonne `nom` est `VARCHAR(255) NOT NULL DEFAULT ''`. Utiliser `ALTER TABLE ligue ADD COLUMN nom VARCHAR(255) NOT NULL DEFAULT '';`
- **Modèle** : Ajouter `Nom string` avec le tag JSON `json:"nom"`
- **scanLigue** : Ajouter `&l.Nom` dans l'appel `Scan()` à la position correspondant à la colonne `nom` dans la clause `SELECT` (entre `id` et `date_debut`, ou à la fin selon l'ordre SQL)
- **Toutes les requêtes SQL** (`INSERT`, `SELECT`, `UPDATE`) dans `ligueRepo` doivent inclure la colonne `nom`
- **Validation** : Dans `Create()` et `Update()`, si `nom` est vide après trimming, retourner `ErrLigueNameRequired`
- **Style** : Suivre les conventions du fichier existant (nommage, gestion d'erreurs, `fmt.Errorf` avec wrapping)

#### Détail des requêtes SQL à modifier

**Create** :
```sql
INSERT INTO ligue (date_debut, date_fin, split_id, nom)
VALUES ($1, $2, $3, $4)
RETURNING id, nom, date_debut, date_fin, split_id, created_at, updated_at
```

**FindByID** :
```sql
SELECT id, nom, date_debut, date_fin, split_id, created_at, updated_at
FROM ligue WHERE id = $1
```

**FindAll** :
```sql
SELECT id, nom, date_debut, date_fin, split_id, created_at, updated_at
FROM ligue ORDER BY date_debut DESC
```

**FindAllPaginated** (deux occurrences : COUNT et SELECT) :
```sql
SELECT id, nom, date_debut, date_fin, split_id, created_at, updated_at
FROM ligue ORDER BY date_debut DESC LIMIT $1 OFFSET $2
```

**Update** :
```sql
UPDATE ligue SET date_debut = $1, date_fin = $2, split_id = $3, nom = $4, updated_at = NOW()
WHERE id = $5
RETURNING id, nom, date_debut, date_fin, split_id, created_at, updated_at
```

### Tests à implémenter

**Fichier** : `internal/models/ligue_test.go`

#### Modifications des tests existants

Chaque appel à `repo.Create()` et `repo.Update()` doit maintenant passer un `nom` (ex: `"Saison 2026"`).

#### Nouveaux tests

1. **CreateLigue avec nom vide**
   - Appeler `repo.Create(..., "")` 
   - Vérifier que l'erreur retournée est `ErrLigueNameRequired`

2. **CreateLigue avec nom valide**
   - Créer une ligue avec nom "Saison 2026"
   - Vérifier que `ligue.Nom == "Saison 2026"`

3. **UpdateLigue avec nouveau nom**
   - Créer une ligue avec nom "Saison 2026"
   - Update avec nom "Saison 2027"
   - Vérifier que `updated.Nom == "Saison 2027"`

4. **FindByID retourne le nom**
   - Créer une ligue avec un nom
   - FindByID → vérifier que le nom est présent

### Exécution des tests

```bash
go test -tags=integration ./internal/models/ -run TestLigueRepository -v
```
