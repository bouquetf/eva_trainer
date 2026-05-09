# Splits (Saisons)

## 1. Vue d'ensemble

Un split représente une **saison** du jeu Eva. Chaque split contient un sous-ensemble de cartes jouables, identifiées par leur nom.

Les splits permettent de filtrer les cartes pendant la révision pour se concentrer sur un ensemble spécifique de cartes.

## 2. Structure d'un split

Un split est défini par :

- **Nom** : Identifiant du split (ex: "Season 1", "Prerelease")
- **Noms de cartes** : Liste des noms de cartes Eva incluses dans ce split (ex: `["Atlantis", "Lunar Outpost", "Polaris"]`)

## 3. Gestion des splits

La création, la modification et la suppression des splits sont réservées aux **administrateurs**.

### 3.1 API Admin

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/v1/admin/splits` | Créer un split |
| GET | `/api/v1/admin/splits` | Lister tous les splits |
| GET | `/api/v1/admin/splits/:id` | Détail d'un split |
| PUT | `/api/v1/admin/splits/:id` | Modifier un split |
| DELETE | `/api/v1/admin/splits/:id` | Supprimer un split |
| GET | `/api/v1/admin/splits/:id/maps` | Cartes d'un split |

### 3.2 API Publique (authentifiée)

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/v1/splits` | Lister les splits |
| GET | `/api/v1/splits/:id/maps` | Cartes d'un split |

Les endpoints publics nécessitent un token d'authentification et la permission `split:view` (attribuée à tous les rôles : member, editor, cocaptain, captain).

## 4. Exemples d'utilisation

### 4.1 Créer un split

```bash
curl -X POST http://localhost:8080/api/v1/admin/splits \
  -H "Authorization: Bearer <token-admin>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Season 1",
    "map_names": ["Atlantis", "Lunar Outpost", "Polaris"]
  }'
```

**Réponse 201 :**

```json
{
  "split": {
    "id": 1,
    "name": "Season 1",
    "map_names": ["Atlantis", "Lunar Outpost", "Polaris"]
  }
}
```

**Codes d'erreur :**

| Code | Erreur | Description |
|------|--------|-------------|
| 400 | `invalid_request_body` | JSON invalide, nom manquant ou map_names manquant |
| 400 | `split_name_required` | Nom vide |
| 400 | `split_map_names_required` | Liste map_names vide |
| 401 | `missing_authorization` | Token manquant |
| 403 | `forbidden` | Utilisateur non administrateur |
| 500 | `internal_error` | Erreur serveur |

### 4.2 Lister les splits

```bash
curl http://localhost:8080/api/v1/admin/splits \
  -H "Authorization: Bearer <token-admin>"
```

**Réponse 200 :**

```json
{
  "splits": [
    {
      "id": 1,
      "name": "Season 1",
      "map_names": ["Atlantis", "Lunar Outpost", "Polaris"]
    }
  ]
}
```

### 4.3 Détail d'un split

```bash
curl http://localhost:8080/api/v1/admin/splits/1 \
  -H "Authorization: Bearer <token-admin>"
```

**Réponse 200 :**

```json
{
  "split": {
    "id": 1,
    "name": "Season 1",
    "map_names": ["Atlantis", "Lunar Outpost", "Polaris"]
  }
}
```

**Codes d'erreur :**

| Code | Erreur | Description |
|------|--------|-------------|
| 404 | `split_not_found` | Split inexistant |

### 4.4 Modifier un split

```bash
curl -X PUT http://localhost:8080/api/v1/admin/splits/1 \
  -H "Authorization: Bearer <token-admin>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Season 1 Updated",
    "map_names": ["Atlantis", "Polaris"]
  }'
```

**Réponse 200 :**

```json
{
  "split": {
    "id": 1,
    "name": "Season 1 Updated",
    "map_names": ["Atlantis", "Polaris"]
  }
}
```

### 4.5 Supprimer un split

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/splits/1 \
  -H "Authorization: Bearer <token-admin>"
```

**Réponse 200 :**

```json
{
  "message": "split deleted"
}
```

### 4.6 Résoudre les cartes d'un split

```bash
curl http://localhost:8080/api/v1/admin/splits/1/maps \
  -H "Authorization: Bearer <token-admin>"
```

**Réponse 200 :**

```json
{
  "maps": [
    {"id": 1, "name": "Atlantis", "added_at": "2024-01-01T14:30:00Z", "is_obsolete": false, "obsolete_of": null},
    {"id": 2, "name": "Atlantis", "added_at": "2024-06-01T10:00:00Z", "is_obsolete": true, "obsolete_of": 3},
    {"id": 3, "name": "Lunar Outpost", "added_at": "2024-01-01T14:30:00Z", "is_obsolete": false, "obsolete_of": null},
    {"id": 4, "name": "Polaris", "added_at": "2024-01-01T14:30:00Z", "is_obsolete": false, "obsolete_of": null}
  ]
}
```

### 4.7 Lister les splits (public)

```bash
curl http://localhost:8080/api/v1/splits \
  -H "Authorization: Bearer <token-utilisateur>"
```

**Réponse 200 :**

```json
{
  "splits": [
    {
      "id": 1,
      "name": "Season 1",
      "map_names": ["Atlantis", "Lunar Outpost", "Polaris"]
    }
  ]
}
```

### 4.8 Cartes d'un split (public)

```bash
curl http://localhost:8080/api/v1/splits/1/maps \
  -H "Authorization: Bearer <token-utilisateur>"
```

**Réponse 200 :** (même format que l'endpoint admin)

### 4.9 Filtrer les cartes par nom

```bash
curl "http://localhost:8080/api/v1/admin/maps?name=Atlantis" \
  -H "Authorization: Bearer <token-admin>"
```

**Réponse 200 :**

```json
{
  "maps": [
    {"id": 1, "name": "Atlantis", "added_at": "2024-01-01T14:30:00Z", "is_obsolete": false, "obsolete_of": null},
    {"id": 2, "name": "Atlantis", "added_at": "2024-06-01T10:00:00Z", "is_obsolete": true, "obsolete_of": 2}
  ]
}
```

## 5. Résolution des cartes par split

Contrairement à une jointure par ID, un split référence les cartes par leur **nom**. Cela signifie que :

- Toutes les versions d'une carte (actuelles **et** obsolètes) ayant ce nom sont retournées
- Si "Atlantis" est dans un split et qu'il existe 3 versions uploadées, les 3 apparaîtront dans la résolution
- Ce comportement est intentionnel : pendant la révision, l'utilisateur peut tomber sur n'importe quelle version d'une carte appartenant au split

## 6. Filtrage par split pendant la révision

Lors de la révision, l'utilisateur peut cocher l'option "cartes du split" pour filtrer l'affichage des cartes. Seules les cartes dont le nom est présent dans le split sélectionné sont proposées.

**Note importante** : Comme le split référence les **noms** de cartes (pas les IDs), réviser les cartes d'un split peut inclure des cartes obsolètes. Par exemple, si "Atlantis" est dans un split et qu'il existe deux versions (v1 obsolète, v2 actuelle), les deux peuvent apparaître pendant la révision.

## 7. Schéma de base de données

Table `splits` :

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` | Identifiant unique |
| `name` | `VARCHAR(255)` | `NOT NULL` | Nom du split |
| `map_names` | `TEXT[]` | `NOT NULL` | Liste des noms de cartes |

## 8. Fichiers source

| Fichier | Description |
|---------|-------------|
| `internal/models/split.go` | Modèle Split et implémentation du repository |
| `internal/handlers/split.go` | Handlers HTTP (CRUD, résolution maps) |
| `internal/handlers/map.go` | Handler `ListMaps` avec filtre `?name=` |
| `internal/middleware/permission.go` | Middleware `RequireGlobalPermission` |
| `api/routes.go` | Déclaration des routes API (admin + public) |
| `db/migrations/000008_add_split_permission.up.sql` | Migration ajout permission `split:view` |
| `db/migrations/000010_drop_split_dates.up.sql` | Migration suppression colonnes dates |
| `db/seed/permissions.go` | Seed de `split:view` pour les rôles |

## 9. Tests

| Fichier | Type | Description |
|---------|------|-------------|
| `internal/models/split_test.go` | Intégration | Tests du repository Split |
| `internal/handlers/split_test.go` | Intégration | Tests des endpoints API Split |
