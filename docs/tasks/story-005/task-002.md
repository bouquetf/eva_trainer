# Tâche #002 - Story #005 : Endpoints API de gestion des cartes (back-end)

## Objectif

Implémenter les endpoints API REST de gestion des cartes : upload, liste, détail, purge pour l'admin, et endpoint public de service d'images.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Tâche #001 (modèle Map, MapRepository)
- Nécessaire pour : Tâche #004 (frontend admin), Tâche #005 (tests)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Handler MapHandler :**
Créer `internal/handlers/map.go` avec les endpoints suivants :

| Méthode | Path | Auth | Rôle requis | Description |
|---------|------|------|-------------|-------------|
| POST | `/api/v1/admin/maps` | RequireAuth + RequireAdmin | admin | Uploader une nouvelle carte |
| GET | `/api/v1/admin/maps` | RequireAuth + RequireAdmin | admin | Lister toutes les cartes |
| GET | `/api/v1/admin/maps/:id` | RequireAuth + RequireAdmin | admin | Détail d'une carte |
| POST | `/api/v1/admin/maps/purge` | RequireAuth + RequireAdmin | admin | Purger les cartes obsolètes sans zone |
| GET | `/api/v1/maps/:id/image` | Aucune | public | Servir l'image webp d'une carte (endpoint public) |

**Partie 2 - Détail des endpoints :**

**POST `/api/v1/admin/maps`** - Upload d'une nouvelle carte
- Content-Type : `multipart/form-data`
- Champs du formulaire :
  - `name` (string, obligatoire) : Nom de la carte (ex: "Atlantis", "Polaris")
  - `file` (file, obligatoire) : Fichier image au format webp
- Validation :
  - `name` obligatoire → si vide, 400 `map_name_required`
  - `file` obligatoire → si aucun fichier, 400 `file_required`
  - Extension du fichier doit être `.webp` → si non, 400 `invalid_file_format`
  - Content-Type du fichier doit être `image/webp` → si non, 400 `invalid_file_format`
  - Taille max : 50 MB → si dépasse, 413 `file_too_large` (configurable via `c.Request.ParseMultipartForm(50 << 20)`)
- Appelle `mapRepo.Create(name, blobData)` qui gère l'obsolescence automatique
- Si une ancienne carte du même nom a été marquée obsolète, inclure dans la réponse : `"previous_version_obsoleted": { "id": oldID, "name": oldName }`
- Réponse 201 :
```json
{
  "map": { "id": 1, "name": "Atlantis", "added_at": "2025-01-01T00:00:00Z", "is_obsolete": false },
  "previous_version_obsoleted": null
}
```
- Erreur 400 : nom manquant, fichier manquant, format invalide
- Erreur 413 : fichier trop volumineux
- Erreur 500 : erreur interne (lecture fichier, insertion DB)

**GET `/api/v1/admin/maps`** - Liste des cartes
- Appelle `mapRepo.FindAll()` 
- Retourne la liste sans les blobs (éviter de transférer les images dans la liste)
- Réponse 200 :
```json
{
  "maps": [
    { "id": 1, "name": "Atlantis", "added_at": "...", "is_obsolete": false, "obsolete_of": null },
    { "id": 2, "name": "Polaris", "added_at": "...", "is_obsolete": true, "obsolete_of": 3 }
  ]
}
```

**GET `/api/v1/admin/maps/:id`** - Détail d'une carte
- Appelle `mapRepo.FindByID(id)`
- Inclut les métadonnées (sans le blob)
- Réponse 200 : `{ "map": { ... } }`
- Erreur 404 : `map_not_found`

**POST `/api/v1/admin/maps/purge`** - Purge des cartes obsolètes sans zone
- Appelle `mapRepo.FindObsoleteWithoutZones()` pour lister les cartes à purger
- Pour chaque carte retournée, appelle `mapRepo.Delete(id)`
- Utilise une transaction pour supprimer toutes les cartes éligibles
- Réponse 200 :
```json
{
  "message": "purged X obsolete maps without zones",
  "purged_ids": [3, 5, 7]
}
```
- Si aucune carte à purger : `{ "message": "no obsolete maps to purge", "purged_ids": [] }`

**GET `/api/v1/maps/:id/image`** - Service d'image de carte (endpoint public, sans authentification)
- Appelle `mapRepo.GetBlob(id)` pour récupérer uniquement le blob
- Retourne l'image avec le Content-Type `image/webp`
- Header `Cache-Control: public, max-age=31536000, immutable` pour le cache navigateur
- Réponse 200 : `image/webp` (corps binaire)
- Erreur 404 : `map_not_found` (si la carte n'existe pas ou si le blob est NULL)
- Note : cet endpoint est utilisable directement dans `<img src="/api/v1/maps/1/image">` par le frontend sans token

**Partie 3 - Routes :**
Ajouter dans `api/routes.go` :

```go
// Dans setupAdminRoutes, après les routes existantes
admin.POST("/maps", adminHandler.UploadMap)       // ou créer un mapHandler dédié
admin.GET("/maps", adminHandler.ListMaps)
admin.GET("/maps/:id", adminHandler.GetMapDetail)
admin.POST("/maps/purge", adminHandler.PurgeObsoleteMaps)
```

Et ajouter la route publique pour servir les images (sans middleware d'auth) directement sur le groupe `apiV1` :

```go
// Dans SetupRoutes, après setupAdminRoutes
mapHandler := handlers.NewMapHandler(models.NewMapRepository(database.DB()))
apiV1.GET("/maps/:id/image", mapHandler.ServeImage)
```

**Cas nominaux :**
- Admin upload une carte webp → HTTP 201, carte créée en base
- Admin upload une carte avec un nom existant → HTTP 201, ancienne carte marquée obsolète
- Admin liste les cartes → HTTP 200, liste complète
- Utilisateur authentifié accède à une image → HTTP 200, image webp servie
- Admin purge les cartes obsolètes → HTTP 200, cartes purgées

**Cas limites :**
- Upload sans fichier → 400 `file_required`
- Upload fichier non-webp → 400 `invalid_file_format`
- Upload fichier > 50MB → 413 `file_too_large`
- Image d'une carte inexistante → 404 `map_not_found`
- Purge sans carte éligible → 200 avec `purged_ids: []`
- Carte avec blob NULL → 404 `map_not_found`

**Gestion d'erreurs :**
Format JSON standard (identique aux handlers existants) :
```json
{ "error": "error_code", "message": "Description lisible" }
```

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | map_name_required | Nom de carte manquant |
| 400 | file_required | Fichier manquant |
| 400 | invalid_file_format | Format de fichier invalide (pas .webp) |
| 400 | invalid_map_id | ID de carte invalide |
| 404 | map_not_found | Carte introuvable |
| 413 | file_too_large | Fichier trop volumineux |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/map.go` | Créer | MapHandler avec les 5 endpoints |
| `api/routes.go` | Modifier | Ajouter routes admin maps et route publique d'image |

### Signatures

```go
// internal/handlers/map.go
package handlers

type MapHandler struct {
    mapRepo models.MapRepository
}

func NewMapHandler(mapRepo models.MapRepository) *MapHandler

func (h *MapHandler) UploadMap(c *gin.Context)
func (h *MapHandler) ListMaps(c *gin.Context)
func (h *MapHandler) GetMapDetail(c *gin.Context)
func (h *MapHandler) PurgeObsoleteMaps(c *gin.Context)
func (h *MapHandler) ServeImage(c *gin.Context)
```

```go
// Ajout dans api/routes.go
// Instanciation dans setupAdminRoutes
mapHandler := handlers.NewMapHandler(models.NewMapRepository(db))

// Routes admin
admin.POST("/maps", mapHandler.UploadMap)
admin.GET("/maps", mapHandler.ListMaps)
admin.GET("/maps/:id", mapHandler.GetMapDetail)
admin.POST("/maps/purge", mapHandler.PurgeObsoleteMaps)

// Route publique (dans SetupRoutes, pas de groupe admin)
apiV1.GET("/maps/:id/image", mapHandler.ServeImage)
```

### Contraintes techniques

- **Gin** : Utiliser `c.FormFile("file")` pour récupérer le fichier uploadé et `c.PostForm("name")` pour le nom.
- **Lecture fichier** : Ouvrir le fichier avec `file.Open()`, lire avec `io.ReadAll()`. Fermer avec `defer file.Close()`.
- **Multipart form** : Limiter la taille à 50MB avec `c.Request.ParseMultipartForm(50 << 20)`.
- **Content-Type** : Pour la réponse image, utiliser `c.Data(http.StatusOK, "image/webp", blob)`.
- **Validation format** : Vérifier l'extension `.webp` sur le nom du fichier ET le Content-Type `image/webp`.
- **Cache** : Ajouter `Cache-Control: public, max-age=31536000, immutable` sur l'endpoint d'image pour le cache navigateur.
- **Middleware** : Les routes admin utilisent `middleware.RequireAuth` + `middleware.RequireAdmin` (existants). La route d'image est publique (aucun middleware).
- **Purge** : Utiliser POST (pas DELETE) pour l'action de purge. Appelée `POST /api/v1/admin/maps/purge` sans body.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
