# Tâche #003 - Story #001 : Documentation configuration et déploiement

## Objectif

Créer la documentation de configuration (`docs/configuration.md`) et de déploiement (`docs/deployment.md`) couvrant l'ensemble des variables d'environnement, les modes de fonctionnement (dev/prod), et les instructions de déploiement local et production (Clever Cloud).

## Contexte

- Story #001 : `docs/stories/story-001.md` - Configuration and Deployment
- Dépend de : Tâche #001 (système de configuration), Tâche #002 (CORS + production static serving)
- Nécessaire pour : Aucune (dernière tâche de la story)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Fichier 1 - `docs/configuration.md` :**
Document de référence pour toutes les variables d'environnement du projet.

Sections :
1. **Principe de chargement** : godotenv (.env) + variables système, priorité, validation au démarrage
2. **Variables obligatoires** : JWT_SECRET, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET
3. **Variables optionnelles** : PORT, GIN_MODE, DB_*, CORS_ORIGINS, JWT_EXPIRY, REFRESH_TOKEN_EXPIRY
4. **Fichier `.env`** : Copier depuis `.env.example`, structure, exemple complet
5. **Validation** : Ce qui se passe si une variable obligatoire manque (serveur refuse de démarrer avec la liste des champs manquants)
6. **Exemples** : Configuration dev typique (CORS=*, GIN_MODE=debug), configuration prod typique (GIN_MODE=release, CORS restreint)

**Fichier 2 - `docs/deployment.md` :**
Guide de déploiement pour les différents environnements.

Sections :
1. **Prérequis** : Docker, Go 1.26+, Node.js 20+
2. **Déploiement local (dev)** :
   - Démarrer uniquement PostgreSQL : `docker compose up db -d`
   - Démarrer l'application complète : `docker compose up -d`
   - Mode dev (frontend via Vite, backend Go indépendant) : expliquer les 2 terminaux :
     ```bash
     # Terminal 1 : Backend
     go run .                      # ou docker compose up app -d
     
     # Terminal 2 : Frontend
     cd frontend && npm run dev
     ```
3. **Déploiement local (production-like)** :
   ```bash
   cd frontend && npm install && npm run build
   go build -o server .             # main.go à la racine
   GIN_MODE=release ./server
   ```
4. **Déploiement production (Clever Cloud)** :
   - Utiliser l'image Docker (Dockerfile multi-stage existant)
   - Variables d'environnement à configurer dans le dashboard Clever Cloud
   - HTTPS géré par Clever Cloud (aucune action requise)
   - CORS : uniquement origine locale (pas de configuration spécifique nécessaire)
   - Pas besoin de build séparé, le Dockerfile gère tout

**Cas nominaux :**
- Documentation complète et auto-suffisante pour un nouveau développeur
- Toutes les variables documentées avec leur type, défaut, obligatoire/optionnel
- Instructions de déploiement testables (reproductibles)

**Cas limites :**
- Mentionner que `DATABASE_URL` écrase les vars DB individuelles
- Mentionner le comportement si aucun `.env` n'est présent (vars système utilisées)
- Mentionner le comportement du SPA fallback en production
- Mentionner le nouveau point d'entrée `main.go` à la racine

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/configuration.md` | Créer | Documentation des variables d'environnement |
| `docs/deployment.md` | Créer | Guide de déploiement |
| `README.md` | Modifier | Ajouter les liens vers les deux nouveaux fichiers docs dans la section "Liens utiles" |

### Contraintes techniques

- **Format** : Markdown propre et bien structuré
- **Exemples** : Inclure des blocs de code bash pour les commandes
- **Tableau des variables** : Dans `docs/configuration.md`, inclure un tableau complet avec colonnes : Variable, Type, Défaut, Obligatoire, Description
- **Liens** : Liens relatifs vers les fichiers du projet (`.env.example`, `docker-compose.yml`, etc.)
- **Ne PAS** inclure de diagrammes Mermaid dans ces documents
- **Ne PAS** dupliquer le README ; les nouveaux fichiers sont des références détaillées
- **main.go racine** : Documenter le nouveau point d'entrée (plus de `cmd/server/`)
- **Dockerfile** : Mentionner le multi-stage build (frontend + Go + alpine)

### Documentation

- `docs/configuration.md` : Référence exhaustive des variables
- `docs/deployment.md` : Guide de déploiement
- `README.md` : Ajouter section "Documentation" avec les liens

### Exemple de tableau pour configuration.md

```markdown
## Variables obligatoires

| Variable | Type | Description |
|----------|------|-------------|
| `JWT_SECRET` | string | Clé secrète pour signer les tokens JWT (min 32 caractères) |
| `DISCORD_CLIENT_ID` | string | Client ID de l'application Discord |
| `DISCORD_CLIENT_SECRET` | string | Client Secret de l'application Discord |

## Variables optionnelles

| Variable | Type | Défaut | Description |
|----------|------|--------|-------------|
| `PORT` | string | `8080` | Port d'écoute du serveur |
| `GIN_MODE` | string | `debug` | Mode Gin (`debug` ou `release`) |
| `CORS_ORIGINS` | string | `*` | Origines CORS autorisées (séparées par des virgules, `*` pour toutes) |
| ... | ... | ... | ... |
```
