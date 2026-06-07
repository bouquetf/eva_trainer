# Gestion des Équipes Eva (Archives)

## Présentation

La fonctionnalité "Équipes Eva" permet de gérer les équipes participantes aux ligues archivées.
Chaque équipe eva possède un nom, un tag, et peut optionnellement être associée à une équipe
eva-trainer existante.

## Modèle de données

Table `eva_teams` :

| Colonne               | Type        | Contraintes          |
|------------------------|-------------|----------------------|
| `id`                   | SERIAL      | PRIMARY KEY          |
| `name`                 | TEXT        | NOT NULL             |
| `tag`                  | VARCHAR(50) | NOT NULL             |
| `eva_trainer_team_id`  | INTEGER     | UNIQUE, NULLABLE, FK → teams(id) |
| `created_at`           | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| `updated_at`           | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

- Le tag est limité à 50 caractères.
- `eva_trainer_team_id` est UNIQUE : une équipe eva-trainer ne peut être associée qu'à une seule équipe eva.
- La suppression d'une équipe eva est bloquée si des références FK pointent vers elle (`HasReferences`).

## API REST

Base : `/api/v1/archive/teams`

### `GET /api/v1/archive/teams`

Liste toutes les équipes eva.

**Réponse** :
```json
{
  "teams": [
    {
      "id": 1,
      "name": "Équipe A",
      "tag": "EQA",
      "eva_trainer_team_id": null,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

### `POST /api/v1/archive/teams`

Crée une nouvelle équipe eva.

**Corps** :
```json
{
  "name": "Équipe A",
  "tag": "EQA",
  "eva_trainer_team_id": null
}
```

`eva_trainer_team_id` est optionnel (omettre ou `null` pour aucune association).

**Réponse** : `201 Created` avec l'objet `team`.

**Erreurs** :
- `400` : nom ou tag vide, tag trop long (> 50 caractères).
- `404` : `eva_trainer_team_id` référence une équipe eva-trainer inexistante.
- `409` : `eva_trainer_team_id` déjà associé à une autre équipe eva (contrainte UNIQUE).

### `GET /api/v1/archive/teams/:id`

Détail d'une équipe eva.

**Réponse** : `200 OK` avec l'objet `team`.
**Erreur** : `404` si l'ID n'existe pas.

### `PUT /api/v1/archive/teams/:id`

Modifie une équipe eva.

**Corps** : identique à POST.
**Réponse** : `200 OK` avec l'objet `team` mis à jour.
**Erreurs** : `400`, `404`, `409` (cf. POST).

### `DELETE /api/v1/archive/teams/:id`

Supprime une équipe eva.

**Réponse** : `200 OK` avec `{"message": "Équipe eva supprimée"}`.
**Erreurs** :
- `404` : ID inexistant.
- `409` : l'équipe est référencée par d'autres tables (FK), message : `"Impossible de supprimer cette équipe eva : elle est référencée par des ligues ou matchs"`.

## Association avec une équipe eva-trainer

L'association utilise l'endpoint `GET /api/v1/teams` existant pour afficher la liste des équipes
eva-trainer disponibles. Le champ `eva_trainer_team_id` est optionnel.

## Interface utilisateur

Les équipes eva sont accessibles via l'onglet "Équipes Eva" sur la page Archives
(`/archives`).

Fonctionnalités disponibles (selon les droits d'accès) :
- **Liste** : tableau avec colonnes Nom, Tag, Équipe eva-trainer associée.
- **Création** : bouton "Nouvelle équipe eva" → formulaire (nom, tag, association).
- **Modification** : bouton "Modifier" → formulaire pré-rempli.
- **Suppression** : bouton "Supprimer" → confirmation modale. Bloquée si références existent.

## Tests

- **Modèle** (`internal/models/eva_team_test.go`) : 12 tests (create, validations, find, update, delete).
- **Handler** (`internal/handlers/archive_test.go`) : 11 tests (CRUD complet, validation, 404).
- Exécution : `go test -tags=integration ./internal/models/ ./internal/handlers/`
