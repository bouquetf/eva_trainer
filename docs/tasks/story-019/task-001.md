# Tâche #001 - Story #019 : User Settings — Couche Base de Données & Repository

## Objectif
Ajouter le support du stockage des préférences utilisateur dans la base de données : migration SQL, extension du modèle `User`, nouvelles méthodes sur `UserRepository`.

## Contexte
- Story #019 : [Paramètres de révision et timers configurables](../../stories/story-019-done.md)
- Dépend de : *aucune*
- Nécessaire pour : Tâche #002 (API), Tâche #003 (Frontend Settings UI), Tâche #004 (Revision timers)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter une colonne JSONB `settings` à la table `users` pour stocker les préférences utilisateur. Étendre le modèle `User` et le repository avec les méthodes nécessaires pour lire et écrire ces settings.

**Contenu du JSON settings :**
```json
{
  "revision_correct_timer": 3,
  "revision_incorrect_timer": 5,
  "revision_question_count": 20
}
```

**Cas nominaux :**
- Un utilisateur sans settings (NULL en base) retourne les valeurs par défaut (3, 5, 20)
- Un utilisateur avec settings partiels (ex: seulement `revision_correct_timer: 2`) fusionne avec les défauts
- Mise à jour des settings : remplacement complet du JSON

**Gestion d'erreurs :**
- UserID invalide (n'existe pas) → retourne `ErrUserNotFound`
- Settings JSON invalide côté Go → erreur de parsing (ne devrait pas arriver car validation Go)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000016_add_user_settings.up.sql` | Créer | Ajouter colonne `settings JSONB` à `users` |
| `db/migrations/000016_add_user_settings.down.sql` | Créer | DROP COLUMN settings |
| `internal/models/user.go` | Modifier | Ajouter champ `Settings`, constructeurs, méthode de merge |
| `internal/models/repository.go` | Modifier | Ajouter `GetUserSettings` et `UpdateUserSettings` à l'interface |
| `internal/models/user_test.go` | Créer | Tests unitaires des fonctions de settings |

### Signatures

```go
// internal/models/user.go

// UserSettings représente les préférences persistantes d'un utilisateur
type UserSettings struct {
    RevisionCorrectTimer   int `json:"revision_correct_timer"`
    RevisionIncorrectTimer int `json:"revision_incorrect_timer"`
    RevisionQuestionCount  int `json:"revision_question_count"`
}

// DefaultUserSettings retourne les valeurs par défaut
func DefaultUserSettings() UserSettings { ... }

// MergeUserSettings fusionne des settings partiels avec les défauts
// (les champs à 0 sont remplacés par les défauts)
func MergeUserSettings(overrides *UserSettings) UserSettings { ... }

// User étendu avec un champ Settings
type User struct {
    // ... champs existants ...
    Settings UserSettings `json:"settings"`
}
```

```go
// internal/models/repository.go — à ajouter à UserRepository

GetUserSettings(userID int) (*UserSettings, error)
UpdateUserSettings(userID int, settings *UserSettings) error
```

```go
// Implémentation dans internal/models/user.go

func (r *userRepo) GetUserSettings(userID int) (*UserSettings, error) { ... }
func (r *userRepo) UpdateUserSettings(userID int, settings *UserSettings) error { ... }
```

### Contraintes techniques
- **Base de données** : PostgreSQL via `database/sql` et `pgx` driver
- **Migration** : suivre la convention existante (fichier `.up.sql` et `.down.sql`)
- **Pattern** : les autres méthodes du repo (`UpdateDisplayName`, etc.) servent de modèle
- **Validation** : `revision_correct_timer` et `revision_incorrect_timer` minimum 1 seconde ; `revision_question_count` minimum 1 question
- **Gestion du NULL** : si la colonne settings est NULL en base, `GetUserSettings` doit retourner les settings par défaut (pas une erreur)

### Tests à implémenter

#### Tests unitaires — `internal/models/user_test.go`
- **Scénario 1** : MergeUserSettings avec des settings nil → retourne les défauts
  - Input : `MergeUserSettings(nil)`
  - Output : `UserSettings{3, 5, 20}`
- **Scénario 2** : MergeUserSettings avec overrides partiels → fusion correcte
  - Input : `MergeUserSettings(&UserSettings{RevisionCorrectTimer: 2})`
  - Output : `UserSettings{2, 5, 20}` (seulement correct_timer changé)
- **Scénario 3** : MergeUserSettings avec tous les champs → retourne l'override exact
  - Input : `MergeUserSettings(&UserSettings{1, 2, 10})`
  - Output : `UserSettings{1, 2, 10}`

#### Tests d'intégration — `internal/handlers/auth_test.go` (dans tâche #002)
- `TestGetSettings_WithNoSettings_ReturnsDefaults` : utilisateur sans settings → 200 avec défauts
- `TestGetSettings_WithPartialSettings_ReturnsMerged` : utilisateur avec settings partiels → 200 avec fusion
- `TestGetSettings_WithFullSettings_ReturnsThem` : utilisateur avec settings complets → 200
- `TestUpdateSettings_StoresAndReturns` : PUT settings → GET vérifie la persistance
- `TestUpdateSettings_InvalidValues_Returns400` : timer à 0 → 400

### Documentation

#### Documentation à mettre à jour
- `docs/user-settings.md` : documenter la structure UserSettings et les endpoints (à créer dans tâche #003)

### Exemples d'utilisation

```sql
-- Migration UP
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings JSONB;

-- Lecture settings (Go)
SELECT settings FROM users WHERE id = $1;

-- Écriture settings (Go)
UPDATE users SET settings = $1::jsonb, updated_at = NOW() WHERE id = $2;
```

```go
// Utilisation dans le handler (task #002)
settings, err := userRepo.GetUserSettings(user.ID)
if err != nil { /* handle */ }
// settings.RevisionCorrectTimer == 3 par défaut
```
