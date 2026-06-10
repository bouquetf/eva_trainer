# Tâche #001 - Story #042 : Migration, modèle et repository replay URL

## Objectif
Ajouter la colonne `replay_url` à la table `divs`, mettre à jour le modèle `Div`, ajouter les méthodes repository pour lire/modifier/supprimer l'URL, et implémenter la validation des formats YouTube et Twitch.

## Contexte
- Story #042 : `docs/stories/story-042.md`
- Dépend de : Story #037 (Divs, assignation d'équipes et classement) — déjà livrée
- Nécessaire pour : Tâche #002 (Handler API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Migration base de données

Créer un nouveau fichier de migration `db/migrations/000032_add_divs_replay_url.up.sql` :

```sql
ALTER TABLE divs ADD COLUMN replay_url VARCHAR(500) NOT NULL DEFAULT '';
```

Et le down migration correspondant `db/migrations/000032_add_divs_replay_url.down.sql` :

```sql
ALTER TABLE divs DROP COLUMN IF EXISTS replay_url;
```

**Note :** La colonne est NOT NULL avec une valeur par défaut de chaîne vide (plutôt que nullable) pour simplifier les requêtes et le scan. Une chaîne vide signifie "pas de lien".

#### 2. Mise à jour du modèle `Div`

Dans `internal/models/div.go`, ajouter le champ `ReplayURL` :

```go
type Div struct {
    ID        int    `json:"id"`
    LigueID   int    `json:"ligue_id"`
    Nom       string `json:"nom"`
    ReplayURL string `json:"replay_url"`
}
```

#### 3. Mise à jour de `scanDiv`

Mettre à jour la fonction `scanDiv` pour inclure `replay_url` :

```go
func scanDiv(scanner interface {
    Scan(dest ...interface{}) error
}) (*Div, error) {
    d := &Div{}
    err := scanner.Scan(&d.ID, &d.LigueID, &d.Nom, &d.ReplayURL)
    if err != nil {
        return nil, err
    }
    return d, nil
}
```

#### 4. Mise à jour de toutes les requêtes SQL

Ajouter `replay_url` dans tous les SELECT et RETURNING du fichier `internal/models/div.go` :

- `Create` : `INSERT ... RETURNING id, ligue_id, nom, replay_url`
- `FindByID` : `SELECT id, ligue_id, nom, replay_url FROM divs WHERE id = $1`
- `FindByLigueID` : `SELECT id, ligue_id, nom, replay_url FROM divs WHERE ligue_id = $1 ORDER BY nom ASC`
- `Update` : `UPDATE divs SET nom = $1 WHERE id = $2 RETURNING id, ligue_id, nom, replay_url`

#### 5. Nouvelles méthodes repository

Ajouter dans l'interface `DivRepository` (fichier `internal/models/repository.go`) :

```go
SetReplayURL(divID int, url string) (*Div, error)
```

**Note :** Pas de méthode `DeleteReplayURL` séparée — la suppression se fait en appelant `SetReplayURL` avec une chaîne vide, ce qui remet la valeur par défaut.

Ajouter l'erreur dans `internal/models/errors.go` :

```go
ErrDivInvalidReplayURL = errors.New("invalid replay URL: only YouTube and Twitch URLs are allowed")
```

#### 6. Validation d'URL

Créer une fonction de validation exportée et testable dans `internal/models/div.go` :

```go
// IsValidReplayURL vérifie que l'URL est un lien YouTube ou Twitch valide.
// Retourne true si valide, false sinon.
func IsValidReplayURL(url string) bool
```

**Règles de validation :**

YouTube (l'URL doit correspondre EXACTEMENT à un des patterns — pas de paramètres supplémentaires) :
- `https://www.youtube.com/watch?v=VIDEO_ID` (VIDEO_ID = caractères `\w-`)
- `https://youtu.be/VIDEO_ID`
- `https://www.youtube.com/embed/VIDEO_ID`
- Avec ou sans `https://` (accepter aussi `http://`)
- **Refuser** toute URL avec des paramètres supplémentaires (`&t=123`, `&list=...`, etc.)
- **Refuser** tout chemin après le video ID

Twitch (au moins un des patterns doit matcher) :
- `https://www.twitch.tv/videos/NUMERIC_ID`
- `https://clips.twitch.tv/CLIP_NAME`
- `https://www.twitch.tv/CHANNEL_NAME`
- `https://www.twitch.tv/CHANNEL_NAME/video/VIDEO_ID`
- Avec ou sans `https://` (accepter aussi `http://`)

**Tout autre format** est rejeté.

**Implémentation recommandée avec regex :**

```go
var youtubeRe = regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=[\w-]+|youtu\.be/[\w-]+|youtube\.com/embed/[\w-]+)$`)
var twitchRe = regexp.MustCompile(`^(https?://)?(www\.|clips\.)?twitch\.tv/(videos/\d+|[a-zA-Z0-9_]+(/video/\d+)?)$`)

func IsValidReplayURL(url string) bool {
    if url == "" {
        return true
    }
    return youtubeRe.MatchString(url) || twitchRe.MatchString(url)
}
```

**Note importante :** Les regex utilisent `$` en fin de pattern pour garantir qu'il n'y a pas de paramètres supplémentaires après l'URL. `youtubeRe.MatchString("https://www.youtube.com/watch?v=test&t=123")` retournera `false`.

**Chaîne vide :** `IsValidReplayURL("")` retourne `true` (une chaîne vide signifie "pas de lien" et est toujours acceptée). La validation stricte est faite dans le handler uniquement quand l'utilisateur fournit une URL non-vide.

#### 7. Implémentation de `SetReplayURL`

Dans `internal/models/div.go` :

```go
func (r *divRepo) SetReplayURL(divID int, url string) (*Div, error) {
    d, err := scanDiv(r.db.QueryRow(
        `UPDATE divs SET replay_url = $1 WHERE id = $2
         RETURNING id, ligue_id, nom, replay_url`,
        url, divID,
    ))
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrDivNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to set replay url: %w", err)
    }
    return d, nil
}
```

**Note :** La validation de l'URL est faite au niveau du handler, pas du repository. Le repository stocke la valeur telle quelle.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000032_add_divs_replay_url.up.sql` | Créer | Migration up |
| `db/migrations/000032_add_divs_replay_url.down.sql` | Créer | Migration down |
| `internal/models/div.go` | Modifier | Ajouter ReplayURL, mettre à jour scanDiv et requêtes, ajouter SetReplayURL, IsValidReplayURL |
| `internal/models/repository.go` | Modifier | Ajouter SetReplayURL à DivRepository |
| `internal/models/errors.go` | Modifier | Ajouter ErrDivInvalidReplayURL |
| `internal/models/div_test.go` | Créer | Tests unitaires validation URL + tests intégration |

### Signatures

```go
package models

// Dans errors.go
var ErrDivInvalidReplayURL = errors.New("invalid replay URL: only YouTube and Twitch URLs are allowed")

// Dans div.go (nouvelle méthode)
func (r *divRepo) SetReplayURL(divID int, url string) (*Div, error)

// Dans div.go (nouvelle fonction exportée)
func IsValidReplayURL(url string) bool

// Dans repository.go (ajout à l'interface DivRepository)
SetReplayURL(divID int, url string) (*Div, error)

// Div struct mis à jour
type Div struct {
    ID        int    `json:"id"`
    LigueID   int    `json:"ligue_id"`
    Nom       string `json:"nom"`
    ReplayURL string `json:"replay_url"`
}
```

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern des repositories existants (struct `divRepo`, constructeur déjà existant)
- **Tests** : Coverage minimum 80% pour les nouvelles fonctionnalités
- **Convention** : Utiliser `sql.DB` comme dépendance, pas de transaction pour une simple UPDATE
- **Migration** : Ajouter le numéro de migration dans `db/migrations/` (après la dernière migration existante)

### Tests à implémenter

#### Tests unitaires de validation (`internal/models/div_test.go`)

- **Scénario 1 : URL YouTube valide (watch)**
  - Données : `"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
  - Attendu : `true`

- **Scénario 2 : URL YouTube valide (short)**
  - Données : `"https://youtu.be/dQw4w9WgXcQ"`
  - Attendu : `true`

- **Scénario 3 : URL YouTube valide (embed)**
  - Données : `"https://www.youtube.com/embed/dQw4w9WgXcQ"`
  - Attendu : `true`

- **Scénario 4 : URL YouTube sans https**
  - Données : `"youtube.com/watch?v=dQw4w9WgXcQ"`
  - Attendu : `true`

- **Scénario 5 : URL Twitch valide (videos)**
  - Données : `"https://www.twitch.tv/videos/123456789"`
  - Attendu : `true`

- **Scénario 6 : URL Twitch valide (clips)**
  - Données : `"https://clips.twitch.tv/CoolClipName_123"`
  - Attendu : `true`

- **Scénario 7 : URL Twitch valide (channel)**
  - Données : `"https://www.twitch.tv/monchannel"`
  - Attendu : `true`

- **Scénario 8 : URL invalide (site inconnu)**
  - Données : `"https://vimeo.com/123456789"`
  - Attendu : `false`

- **Scénario 9 : URL invalide (texte aléatoire)**
  - Données : `"pas une url"`
  - Attendu : `false`

- **Scénario 10 : Chaîne vide**
  - Données : `""`
  - Attendu : `true`

- **Scénario 11 : URL YouTube sans video ID**
  - Données : `"https://www.youtube.com/watch"`
  - Attendu : `false`

- **Scénario 12 : URL YouTube avec paramètres supplémentaires (rejeté)**
  - Données : `"https://www.youtube.com/watch?v=test123&t=30"`
  - Attendu : `false`

- **Scénario 13 : URL YouTube avec playlist (rejeté)**
  - Données : `"https://www.youtube.com/watch?v=test123&list=PLabc"`
  - Attendu : `false`

- **Scénario 14 : URL Twitch valide avec http**
  - Données : `"http://www.twitch.tv/videos/987654321"`
  - Attendu : `true`

- **Scénario 15 : URL YouTube avec chemin supplémentaire (rejeté)**
  - Données : `"https://www.youtube.com/watch?v=test123/extra"`
  - Attendu : `false`

- **Scénario 16 : URL Twitch avec paramètres (rejeté)**
  - Données : `"https://www.twitch.tv/videos/123456?t=1h"`
  - Attendu : `false`

#### Tests d'intégration (même fichier, avec tag `integration`)

- **Scénario 13 : SetReplayURL sur une div existante (YouTube)**
  - Setup : créer une ligue, créer une div
  - Appel : `SetReplayURL(divID, "https://www.youtube.com/watch?v=test123")`
  - Vérifier : retourne Div avec replay_url = "https://www.youtube.com/watch?v=test123"

- **Scénario 14 : SetReplayURL avec chaîne vide (suppression)**
  - Setup : créer une ligue, créer une div avec un replay_url déjà défini
  - Appel : `SetReplayURL(divID, "")`
  - Vérifier : retourne Div avec replay_url = ""

- **Scénario 15 : SetReplayURL sur div inexistante**
  - Appel : `SetReplayURL(99999, "https://www.youtube.com/watch?v=test")`
  - Vérifier : `errors.Is(err, ErrDivNotFound)`
