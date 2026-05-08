# Map Management

## 1. Vue d'ensemble

Le système de gestion des cartes permet aux administrateurs de gérer les images de fond de jeu au format WebP. Les cartes servent de support pour les zones et la révision. Les administrateurs peuvent uploader de nouvelles cartes, gérer leur cycle de vie, et purger les versions obsolètes.

## 2. Cycle de vie d'une carte

- **Active** : Une carte active peut être utilisée pour créer et modifier des zones.
- **Obsolète** : Une carte devient obsolète quand une nouvelle version du même nom est uploadée. Les zones existantes restent visibles, mais plus de création ou modification sur l'ancienne version.
- **Purge** : Les cartes obsolètes sans zone associée peuvent être supprimées définitivement.

```
[Upload] → Active → [Nouvel upload même nom] → Obsolète → [Purge] → Supprimée
```

## 3. Upload d'une carte

- **Format accepté** : WebP uniquement (extension `.webp` et MIME type `image/webp`)
- **Taille maximale** : 50 MB
- **Comportement** : Si une carte avec le même nom existe déjà, l'ancienne version est automatiquement marquée obsolète (`is_obsolete = true`, `obsolete_of` pointe vers la nouvelle version)
- **Accès** : Administrateur uniquement (authentification + rôle admin requis)

## 4. API Endpoints

### 4.1 Upload d'une carte

**POST** `/api/v1/admin/maps`

- **Auth** : Bearer token (admin requis)
- **Content-Type** : `multipart/form-data`

**Paramètres :**

| Champ | Type | Description |
|-------|------|-------------|
| `name` | string | Nom de la carte (obligatoire) |
| `file` | file | Fichier image au format WebP (obligatoire) |

**Exemple curl :**

```bash
curl -X POST http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer <token>" \
  -F "name=Atlantis" \
  -F "file=@atlantis-map.webp"
```

**Réponse 201 (succès) :**

```json
{
  "map": {
    "id": 1,
    "name": "Atlantis",
    "added_at": "2025-01-01T14:30:00Z",
    "is_obsolete": false,
    "obsolete_of": null
  },
  "previous_version_obsoleted": null
}
```

Si une ancienne version a été marquée obsolète :

```json
{
  "map": {
    "id": 2,
    "name": "Atlantis",
    "added_at": "2025-01-02T10:00:00Z",
    "is_obsolete": false,
    "obsolete_of": null
  },
  "previous_version_obsoleted": {
    "id": 1,
    "name": "Atlantis"
  }
}
```

**Codes d'erreur :**

| Code | Erreur | Description |
|------|--------|-------------|
| 400 | `file_too_large` | Fichier dépasse 50 MB |
| 400 | `map_name_required` | Nom de carte manquant |
| 400 | `file_required` | Aucun fichier fourni |
| 400 | `invalid_file_format` | Format autre que WebP |
| 401 | `missing_authorization` | Token manquant |
| 403 | `forbidden` | Utilisateur non administrateur |
| 500 | `internal_error` | Erreur serveur |

### 4.2 Liste des cartes

**GET** `/api/v1/admin/maps`

- **Auth** : Bearer token (admin requis)

**Exemple curl :**

```bash
curl http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer <token>"
```

**Réponse 200 :**

```json
{
  "maps": [
    {
      "id": 1,
      "name": "Atlantis",
      "added_at": "2025-01-01T14:30:00Z",
      "is_obsolete": true,
      "obsolete_of": 2
    },
    {
      "id": 2,
      "name": "Atlantis",
      "added_at": "2025-01-02T10:00:00Z",
      "is_obsolete": false,
      "obsolete_of": null
    }
  ]
}
```

### 4.3 Détail d'une carte

**GET** `/api/v1/admin/maps/:id`

- **Auth** : Bearer token (admin requis)

**Exemple curl :**

```bash
curl http://localhost:8080/api/v1/admin/maps/1 \
  -H "Authorization: Bearer <token>"
```

**Réponse 200 :**

```json
{
  "map": {
    "id": 1,
    "name": "Atlantis",
    "added_at": "2025-01-01T14:30:00Z",
    "is_obsolete": false,
    "obsolete_of": null
  }
}
```

**Codes d'erreur :**

| Code | Erreur | Description |
|------|--------|-------------|
| 400 | `invalid_map_id` | ID non numérique |
| 404 | `map_not_found` | Carte inexistante |

### 4.4 Purge des cartes obsolètes

**POST** `/api/v1/admin/maps/purge`

- **Auth** : Bearer token (admin requis)
- **Comportement** : Supprime définitivement toutes les cartes marquées obsolètes qui n'ont aucune zone associée

**Exemple curl :**

```bash
curl -X POST http://localhost:8080/api/v1/admin/maps/purge \
  -H "Authorization: Bearer <token>"
```

**Réponse 200 :**

```json
{
  "message": "purged 2 obsolete maps without zones",
  "purged_ids": [1, 3]
}
```

### 4.5 Image d'une carte (publique)

**GET** `/api/v1/maps/:id/image`

- **Auth** : Aucune (public)
- **Cache** : `Cache-Control: public, max-age=31536000, immutable`

**Exemple HTML :**

```html
<img src="/api/v1/maps/1/image" alt="Atlantis" />
```

**Réponse 200** : Contenu binaire de l'image au format WebP avec header `Content-Type: image/webp`.

**Codes d'erreur :**

| Code | Erreur | Description |
|------|--------|-------------|
| 404 | `map_not_found` | Carte inexistante ou sans données image |

## 5. Seed automatique

Au démarrage de l'application, si la table `maps` est vide, les cartes du dossier `maps/` sont automatiquement insérées en base.

**Comportement :**
- Vérification du `COUNT(*)` de la table `maps` au démarrage
- Si la table est vide, lecture des fichiers `.webp` du dossier `maps/`
- Création d'une entrée en base pour chaque fichier valide
- Le seed est **idempotent** : si des cartes existent déjà, aucune action n'est effectuée
- En cas d'échec de lecture d'un fichier, un warning est loggé et le seed continue avec les fichiers suivants

## 6. Structure des fichiers de maps

Les fichiers webp initiaux sont stockés dans le dossier `maps/` à la racine du projet.

**Convention de nommage :** `{nom}-map.webp`

Exemples :
- `atlantis-map.webp` → nom extrait : "Atlantis"
- `lunar-outpost-map.webp` → nom extrait : "Lunar Outpost"

**Règle d'extraction du nom :**
1. Suppression du suffixe `-map.webp` (si présent)
2. Suppression du suffixe `.webp` restant
3. Remplacement des tirets `-` par des espaces
4. Mise en titre (première lettre de chaque mot en majuscule)

## 7. Schéma de base de données

Table `maps` :

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` | Identifiant unique |
| `name` | `VARCHAR(255)` | `NOT NULL` | Nom de la carte |
| `webp_blob` | `BYTEA` | `NOT NULL` | Données binaires de l'image |
| `added_at` | `TIMESTAMP` | `NOT NULL DEFAULT NOW()` | Date d'ajout |
| `obsolete_of` | `INTEGER` | `REFERENCES maps(id) ON DELETE SET NULL` | ID de la version plus récente |
| `is_obsolete` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` | Indique si la carte est obsolète |

## 8. Fichiers source

| Fichier | Description |
|---------|-------------|
| `internal/models/map.go` | Modèle Map et implémentation du repository |
| `internal/handlers/map.go` | Handlers HTTP (upload, liste, détail, purge, image) |
| `api/routes.go` | Déclaration des routes API |
| `db/seed/maps.go` | Seed automatique des cartes au démarrage |
| `db/migrations/000006_drop_current_version_id.up.sql` | Migration de nettoyage du schéma |
| `frontend/src/api/maps.ts` | Client API frontend |
| `frontend/src/pages/AdminDashboard.tsx` | Interface d'administration (onglet Cartes) |
| `frontend/src/styles/admin.css` | Styles de l'interface |
| `maps/*.webp` | Fichiers de cartes préchargés |

## 9. Tests

| Fichier | Type | Description |
|---------|------|-------------|
| `internal/models/map_test.go` | Unitaire | Tests du repository Map |
| `internal/handlers/map_test.go` | Intégration | Tests des endpoints API |
| `db/seed/maps_test.go` | Unitaire + Intégration | Tests du seed et de l'extraction de nom |
| `scripts/test-e2e.sh` | E2E | Tests de bout en bout via curl |
