# Tâche #001 - Story #043 : Migration, modèle et repository DivNote

## Objectif
Créer la table `div_notes` en base, le modèle Go `DivNote`, l'interface `DivNoteRepository` et son implémentation complète avec les opérations CRUD.

## Contexte
- Story #043 : `docs/stories/story-043.md`
- Dépend de : Story #037 (les divs existent)
- Nécessaire pour : Tâche #002 (handlers API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Une note personnelle est liée à un utilisateur et une div. La contrainte est : **un seul commentaire par personne et par div**. Un nouvel upsert remplace l'ancien.

**Cas nominaux :**
- Upsert : un utilisateur écrit une note sur une div → insertion si inexistante, mise à jour si existante
- Get : récupération de la note d'un utilisateur pour une div donnée
- Delete : suppression de la note d'un utilisateur pour une div donnée
- CASCADE : si la div est supprimée, les notes associées sont supprimées ; si l'utilisateur est supprimé, idem

**Cas limites :**
- Note vide → autorisé (l'utilisateur peut effacer son contenu, pas de validation de contenu non vide)
- Div inexistante → erreur (contrainte FK)
- Utilisateur inexistant → erreur (contrainte FK)
- Upsert sans changement de contenu → mise à jour de `updated_at` normale

**Gestion d'erreurs :**
- `ErrDivNoteNotFound` si get/delete sur une note inexistante
- Violation FK (div ou user invalide) → `ErrDivNotFound` ou `ErrUserNotFound` selon le cas

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000033_add_div_notes.up.sql` | Créer | Migration UP |
| `db/migrations/000033_add_div_notes.down.sql` | Créer | Migration DOWN |
| `internal/models/div_note.go` | Créer | Modèle DivNote + repository |
| `internal/models/repository.go` | Modifier | Ajouter interface DivNoteRepository |
| `internal/models/errors.go` | Modifier | Ajouter ErrDivNoteNotFound |

### Modèle de données

```sql
CREATE TABLE IF NOT EXISTS div_notes (
    div_id     INT       NOT NULL REFERENCES divs(id) ON DELETE CASCADE,
    user_id    INT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_md TEXT      NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (div_id, user_id)
);
```

### Signatures

```go
// Dans repository.go, ajouter dans la zone DivRepository (après ligne 273)
type DivNoteRepository interface {
    Upsert(divID, userID int, contentMD string) (*DivNote, error)
    FindByDivAndUser(divID, userID int) (*DivNote, error)
    DeleteByDivAndUser(divID, userID int) error
}

// Dans div_note.go
type DivNote struct {
    DivID     int       `json:"div_id"`
    UserID    int       `json:"user_id"`
    ContentMD string    `json:"content_md"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type divNoteRepo struct {
    db *sql.DB
}

func NewDivNoteRepository(db *sql.DB) DivNoteRepository
func scanDivNote(scanner interface{ Scan(dest ...interface{}) error }) (*DivNote, error)

// Repository methods
func (r *divNoteRepo) Upsert(divID, userID int, contentMD string) (*DivNote, error)  — INSERT ON CONFLICT (div_id, user_id) DO UPDATE SET content_md = $3, updated_at = NOW()
func (r *divNoteRepo) FindByDivAndUser(divID, userID int) (*DivNote, error)           — SELECT avec WHERE div_id=$1 AND user_id=$2
func (r *divNoteRepo) DeleteByDivAndUser(divID, userID int) error                      — DELETE avec WHERE div_id=$1 AND user_id=$2
```

### Contraintes techniques
- **Framework** : Go standard library + `database/sql` (comme les autres repositories)
- **Pattern** : Exactement le même pattern que `divRepo` dans `internal/models/div.go` (scanner, QueryRow, Exec, sql.ErrNoRows)
- **Style** : Respecter `gofmt`, les conventions d'erreurs avec `fmt.Errorf("failed to ...: %w", err)`
- **Migration** : Ajouter les numéros de migration dans les helpers E2E si nécessaire (vérifier `frontend/e2e/helpers/db.ts`)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/models/div_note_test.go`
- Scénario 1 : Upsert insère une nouvelle note
  - Données : divID=1, userID=1, contentMD="Hello"
  - Résultat attendu : note créée avec contentMD="Hello"
- Scénario 2 : Upsert met à jour une note existante
  - Données : même divID/userID, nouveau contentMD="World"
  - Résultat attendu : contentMD="World", updated_at changé
- Scénario 3 : FindByDivAndUser retourne la note existante
  - Résultat attendu : note avec les bons champs
- Scénario 4 : FindByDivAndUser retourne ErrDivNoteNotFound si inexistante
  - Résultat attendu : `errors.Is(err, ErrDivNoteNotFound)`
- Scénario 5 : DeleteByDivAndUser supprime la note
  - Résultat attendu : aucun erreur, puis Find retourne ErrDivNoteNotFound
- Scénario 6 : DeleteByDivAndUser retourne ErrDivNoteNotFound si inexistante

### Documentation
- Mettre à jour `docs/archives-notes.md` après la tâche #005
