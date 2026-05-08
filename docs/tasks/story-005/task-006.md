# Tâche #006 - Story #005 : Documentation Map Management

## Objectif

Créer la documentation complète de la fonctionnalité de gestion des cartes, couvrant l'upload, l'obsolescence, la purge, le seed automatique et l'API de service d'images.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Tâche #001, #002, #003 (toutes les implémentations)
- Ordre d'exécution : 6ème (après implémentation complète)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Créer `docs/map-management.md` :**

Documenter la fonctionnalité de gestion des cartes avec les sections suivantes :

#### 1. Vue d'ensemble
Description du système de gestion des cartes : les cartes sont des images de fond de jeu au format WebP, utilisées pour les zones et la révision. Les administrateurs peuvent uploader de nouvelles cartes et gérer leur cycle de vie.

#### 2. Cycle de vie d'une carte
- **Active** : Une carte active peut être utilisée pour créer/modifier des zones
- **Obsolète** : Une carte devient obsolète quand une nouvelle version du même nom est uploadée. Les zones existantes restent visibles, mais plus de création/modification
- **Purge** : Les cartes obsolètes sans zone associée peuvent être supprimées définitivement

#### 3. Upload d'une carte
- Format accepté : WebP uniquement
- Taille maximale : 50 MB
- Comportement : si le nom existe déjà, l'ancienne version est marquée obsolète
- Accès : admin uniquement
- Endpoint : `POST /api/v1/admin/maps`
- Corps : multipart/form-data avec `name` (string) et `file` (file)

#### 4. API Endpoints

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| POST | `/api/v1/admin/maps` | Admin | Upload d'une nouvelle carte |
| GET | `/api/v1/admin/maps` | Admin | Liste des cartes |
| GET | `/api/v1/admin/maps/:id` | Admin | Détail d'une carte |
| POST | `/api/v1/admin/maps/purge` | Admin | Purge des cartes obsolètes |
| GET | `/api/v1/maps/:id/image` | Aucune (public) | Image d'une carte (utilisable en `<img>`) |

Pour chaque endpoint :
- Description détaillée
- Exemple de requête (curl)
- Exemple de réponse
- Codes d'erreur possibles

#### 5. Exemples d'utilisation

**Upload avec curl :**
```bash
curl -X POST http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer <token>" \
  -F "name=Atlantis" \
  -F "file=@atlantis-map.webp"
```

**Récupération d'une image dans une balise img :**
```html
<img src="/api/v1/maps/1/image" alt="Atlantis" />
```

**Purge des cartes obsolètes :**
```bash
curl -X POST http://localhost:8080/api/v1/admin/maps/purge \
  -H "Authorization: Bearer <token>"
```

**Liste des cartes :**
```bash
curl http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer <token>"
```

#### 6. Seed automatique
- Au démarrage de l'application, si la table `maps` est vide, les cartes du dossier `/maps` sont automatiquement insérées en base
- Les fichiers doivent être au format `nom-carte-map.webp`
- Le seed est idempotent (ne se réexécute pas si des cartes existent déjà)

#### 7. Structure des fichiers de maps
- Les fichiers webp initiaux sont stockés dans le dossier `maps/` à la racine du projet
- Convention de nommage : `{nom}-map.webp` (ex: `atlantis-map.webp`, `lunar-outpost-map.webp`)
- Le seed extrait le nom de la carte à partir du nom du fichier

#### 8. Schéma de base de données (table `maps`)

| Colonne | Type | Description |
|---------|------|-------------|
| id | SERIAL PRIMARY KEY | Identifiant unique |
| name | VARCHAR(255) | Nom de la carte |
| webp_blob | BYTEA | Données binaires de l'image |
| added_at | TIMESTAMP | Date d'ajout |
| obsolete_of | INTEGER FK → maps.id | ID de la version plus récente (si obsolète) |
| is_obsolete | BOOLEAN | Indique si la carte est obsolète |

**Partie 2 - Mettre à jour le README ou la table des matières :**

Si une table des matières ou un index de documentation existe, ajouter un lien vers `docs/map-management.md`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/map-management.md` | Créer | Documentation complète de la gestion des cartes |

### Contraintes techniques

- **Format** : Markdown standard, utilisable sur GitHub.
- **Exemples précis** : Tous les exemples de code doivent être testés et fonctionnels.
- **Liens** : Utiliser des chemins relatifs pour lier vers la spécification et les autres docs.
- **Public cible** : Développeurs et administrateurs. Le style doit être technique mais accessible.

### Tests à implémenter

Aucun test pour cette tâche.

### Documentation

Cette tâche consiste à créer la documentation elle-même.
