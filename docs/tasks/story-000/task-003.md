# Tâche #003 - Story #000 : Docker Compose, Dockerfile et configuration environnement

## Objectif

Créer l'infrastructure de développement local avec Docker Compose (PostgreSQL + app Go), le Dockerfile multi-stage pour le backend, le Dockerfile.dev avec Air pour le hot-reload, et le fichier `.env.example` de configuration.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Tâche #001 (structure backend Go), Tâche #002 (migrations)
- Nécessaire pour : Tâche #005 (Tests), Tâche #006 (README)
- Ordre d'exécution : 3ème (après #001 et #002)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Mettre en place l'environnement de développement local avec Docker.

**Cas nominaux :**
- `docker-compose.yml` définit deux services : `app` (backend Go) et `db` (PostgreSQL)
- Le service `db` expose PostgreSQL sur le port 5432 avec un volume persistant
- Le service `app` expose l'API sur le port configuré (défaut 8080)
- `Dockerfile` construit l'application Go en multi-stage (build + run) pour la production
- `Dockerfile.dev` construit avec Air pour le hot-reload en développement
- `.env.example` liste toutes les variables d'environnement nécessaires avec des valeurs par défaut pour le dev

**Cas limites :**
- Le volume PostgreSQL est persistant (named volume) pour ne pas perdre les données entre les redémarrages
- Le service `app` attend que PostgreSQL soit prêt avant de démarrer (healthcheck condition)
- Le service `app` peut être redémarré indépendamment de `db` avec `docker compose restart app`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `Dockerfile` | Créer | Multi-stage build pour Go (alpine) |
| `Dockerfile.dev` | Créer | Build avec Air pour hot-reload |
| `.air.toml` | Créer | Configuration Air pour le hot-reload |
| `docker-compose.yml` | Créer | Services app + db |
| `.env.example` | Créer | Exemple de configuration |
| `.dockerignore` | Créer | Fichiers à exclure du contexte Docker |

### Contraintes techniques

- **Version Go** : 1.26 (images `golang:1.26-alpine`)

- **Dockerfile (production)** :
  ```dockerfile
  # Stage 1: Build
  FROM golang:1.26-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

  # Stage 2: Run
  FROM alpine:latest
  RUN adduser -D -g '' appuser
  COPY --from=builder /app/server /app/server
  EXPOSE ${PORT:-8080}
  USER appuser
  ENTRYPOINT ["/app/server"]
  ```

- **`.air.toml`** :
  ```toml
  root = "."
  tmp_dir = "tmp"

  [build]
    cmd = "go build -o ./tmp/server ./cmd/server"
    bin = "./tmp/server"
    include_ext = ["go"]
    exclude_dir = ["frontend", "tmp", "vendor", "docs", "maps", "ia"]
    exclude_regex = [".*_test.go"]
    exclude_unchanged = true
    log = "build-errors.log"
    stop_on_error = true
    send_interrupt = true
    kill_delay = "5s"

  [log]
    time = true

  [color]
    app = "yellow"
    build = "green"
    runner = "cyan"
  ```

- **Dockerfile.dev (hot-reload Air)** :
  ```dockerfile
  FROM golang:1.26-alpine
  WORKDIR /app

  # Install Air
  RUN go install github.com/air-verse/air@latest

  COPY go.mod go.sum ./
  RUN go mod download

  COPY . .

  EXPOSE ${PORT:-8080}
  CMD ["air", "-c", ".air.toml"]
  ```

- **docker-compose.yml** :
  ```yaml
  services:
    db:
      image: postgres:16-alpine
      environment:
        POSTGRES_USER: ${DB_USER:-eva_user}
        POSTGRES_PASSWORD: ${DB_PASSWORD:-eva_pass}
        POSTGRES_DB: ${DB_NAME:-eva_trainer}
      ports:
        - "${DB_PORT:-5432}:5432"
      volumes:
        - pgdata:/var/lib/postgresql/data
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-eva_user} -d ${DB_NAME:-eva_trainer}"]
        interval: 5s
        timeout: 5s
        retries: 5
      networks:
        - eva-network

    app:
      build:
        context: .
        dockerfile: Dockerfile.dev
      ports:
        - "${PORT:-8080}:8080"
      environment:
        - DATABASE_URL=postgres://${DB_USER:-eva_user}:${DB_PASSWORD:-eva_pass}@db:5432/${DB_NAME:-eva_trainer}?sslmode=disable
        - GIN_MODE=debug
        - PORT=${PORT:-8080}
      volumes:
        - .:/app
      depends_on:
        db:
          condition: service_healthy
      networks:
        - eva-network

  volumes:
    pgdata:

  networks:
    eva-network:
      driver: bridge
  ```

- **.env.example** :
  ```bash
  # Server
  PORT=8080
  GIN_MODE=debug

  # Database (individual)
  DB_HOST=localhost
  DB_PORT=5432
  DB_USER=eva_user
  DB_PASSWORD=eva_pass
  DB_NAME=eva_trainer
  DB_SSLMODE=disable
  DB_TIMEOUT=10

  # Database (URL - prioritaire sur les vars individuelles)
  DATABASE_URL=postgres://eva_user:eva_pass@localhost:5432/eva_trainer?sslmode=disable

  # CORS
  CORS_ORIGINS=*

  # Discord OAuth
  DISCORD_CLIENT_ID=
  DISCORD_CLIENT_SECRET=

  # JWT
  JWT_SECRET=change-me-in-production
  JWT_EXPIRY=24h
  REFRESH_TOKEN_EXPIRY=7d
  ```

- **.dockerignore** :
  ```
  .git
  .gitignore
  .env
  .env.*
  node_modules/
  frontend/node_modules/
  frontend/dist/
  docs/
  ia/
  maps/
  .idea/
  *.md
  ```

- **Ne PAS inclure** : Les configurations de déploiement production Clever Cloud (seront faites dans la story #001)
- **Ne PAS inclure** : Les vrais secrets (`.env` ne doit pas être commité)

### Documentation

- Ajouter des commentaires dans `docker-compose.yml` pour expliquer chaque service
- Ajouter un commentaire en haut de `.env.example` expliquant comment l'utiliser

### Exemples d'utilisation

```bash
# Démarrer l'environnement de développement
cp .env.example .env
docker compose up -d

# Vérifier que l'application répond
curl http://localhost:8080/health

# Exécuter les migrations (via golang-migrate CLI ou depuis le code)
# Option 1 : via un container dédié
# Option 2 : via une commande Go (à implémenter dans une tâche ultérieure)

# Voir les logs
docker compose logs -f app

# Arrêter l'environnement
docker compose down

# Supprimer les volumes (réinitialisation complète)
docker compose down -v
```
