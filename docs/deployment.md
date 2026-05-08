# Déploiement

## Prérequis

- **Docker** 24+ avec Docker Compose V2
- **Go** 1.26+
- **Node.js** 20+
- **npm** 10+

## Architecture des points d'entrée

Depuis la story #001, le point d'entrée Go a été centralisé :

- **`main.go`** (racine) : nouveau point d'entrée unique du serveur. Remplace l'ancien `cmd/server/main.go`.
- **`cmd/migrate/`** : commande CLI pour exécuter les migrations de base de données uniquement.

Le `main.go` racine utilise `//go:embed frontend/dist/*` pour embarquer les fichiers compilés du frontend dans le binaire.

## Déploiement local (développement)

### Option 1 : Application complète avec Docker Compose

```bash
cp .env.example .env
# Éditer .env avec vos clés Discord et JWT

docker compose up -d
```

Cela démarre PostgreSQL et l'application Go. Les migrations sont exécutées automatiquement via le script `scripts/start-dev.sh`.

Le frontend est servi par Vite en mode développement (port 5173, avec hot-reload). Accès :
- Frontend : http://localhost:5173
- API : http://localhost:8080/api/v1/
- Health : http://localhost:8080/health

### Option 2 : Backend Go seul + PostgreSQL

```bash
# Terminal 1 : Démarrer PostgreSQL uniquement
docker compose up db -d

# Terminal 2 : Démarrer le backend Go (avec hot-reload via Air)
go run .
```

Le backend écoute sur le port configuré (défaut : 8080).

### Option 3 : Frontend indépendant

```bash
# Terminal 1 : Backend
docker compose up app -d

# Terminal 2 : Frontend Vite (hot-reload)
cd frontend && npm install && npm run dev
```

## Déploiement local (production-like)

Construire et exécuter une version identique à la production :

```bash
# 1. Build du frontend
cd frontend && npm install && npm run build

# 2. Build du binaire Go (depuis la racine)
cd .. && go build -o server .

# 3. Lancer en mode production
GIN_MODE=release ./server
```

### Comportement en mode production

- **`/assets/*`** : sert les fichiers statiques du frontend compilé (JS, CSS, etc.)
- **`/`** : sert `index.html`
- **Toute route non-API** (ne commençant pas par `/api/` ou `/health`) : sert `index.html` (SPA fallback, nécessaire pour le routing côté client)
- **Routes API** (`/api/v1/*`) : inchangées
- Le binaire embarque le frontend via `//go:embed`, pas de besoin de copier `frontend/dist/` manuellement

## Déploiement production (Clever Cloud)

Le projet utilise un [Dockerfile multi-stage](../Dockerfile) adapté au déploiement sur Clever Cloud.

### Étapes

1. **Pousser le code** vers le repository lié à Clever Cloud
2. **Configurer les variables d'environnement** dans le dashboard Clever Cloud :
   - `PORT` : défini automatiquement par Clever Cloud (généralement `8080`)
   - `GIN_MODE` : `release`
   - `CORS_ORIGINS` : l'origine de votre frontend (ex. `https://mon-domaine.com`)
   - Toutes les variables obligatoires (`JWT_SECRET`, `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`)
   - Variables de base de données (via le addon PostgreSQL Clever Cloud)
3. **HTTPS** : géré automatiquement par Clever Cloud (aucune configuration nécessaire)
4. **Base de données** : utiliser l'addon PostgreSQL de Clever Cloud. La variable `DATABASE_URL` est automatiquement fournie.

### Dockerfile

Le [Dockerfile](../Dockerfile) comprend 3 stages :

1. **`frontend`** (node:20-alpine) : installe les dépendances et build le frontend
2. **`builder`** (golang:1.26-alpine) : compile le binaire Go en CGO_ENABLED=0
3. **`runtime`** (alpine) : copie le binaire et l'exécute avec un utilisateur non-root

### Variables d'environnement Clever Cloud

| Variable | Source |
|----------|--------|
| `PORT` | Définie par Clever Cloud (ne pas surcharger) |
| `DATABASE_URL` | Fournie par l'addon PostgreSQL (écrase les vars `DB_*`) |
| `GIN_MODE` | `release` (à configurer dans le dashboard) |
| `CORS_ORIGINS` | L'origine de votre frontend (à configurer dans le dashboard) |
| `JWT_SECRET` | À configurer dans le dashboard |
| `DISCORD_CLIENT_ID` | À configurer dans le dashboard |
| `DISCORD_CLIENT_SECRET` | À configurer dans le dashboard |

## Scripts utiles

```bash
# Tests unitaires
go test ./...

# Tests intégration (nécessite une base PostgreSQL)
go test --tags=integration ./...

# Tous les tests (script complet)
bash scripts/test.sh

# Tests E2E (Docker Compose)
bash scripts/test-e2e.sh

# Build frontend uniquement
cd frontend && npm run build

# Logs Docker
docker compose logs -f
```
