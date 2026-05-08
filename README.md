# Eva Trainer

Application d'entraînement à la reconnaissance de zones sur cartes Eva.

## Technologies

- **Backend** : Go 1.26 + Gin
- **Frontend** : React 19 + Vite + TypeScript
- **Base de données** : PostgreSQL 16
- **Conteneurisation** : Docker Compose
- **Librairies frontend** : React-Konva, Zustand, TanStack Query, React Router DOM

## Prérequis

- Go 1.26+
- Node.js 20+
- Docker Desktop
- npm

## Démarrage rapide

```bash
cp .env.example .env
docker compose up -d
cd frontend && npm install && npm run dev
```

Accès :
- Frontend (hot-reload) : http://localhost:5173
- API : http://localhost:8080/api/v1/
- Health check : http://localhost:8080/health

## Architecture

```
/
├── main.go               # Point d'entrée unique du serveur (//go:embed frontend/dist)
├── cmd/
│   └── migrate/           # Commande de migration
├── api/                   # Définition des routes API
├── internal/
│   ├── handlers/          # Handlers HTTP
│   ├── models/            # Modèles de données
│   └── middleware/        # Middlewares
├── pkg/
│   ├── config/            # Configuration centralisée (godotenv, validation)
│   └── database/          # Connexion PostgreSQL + migrations
├── db/migrations/         # Migrations SQL (golang-migrate)
├── frontend/              # Application React
│   ├── src/
│   │   ├── components/    # Composants réutilisables
│   │   ├── pages/         # Pages de l'application
│   │   ├── routes/        # Configuration React Router
│   │   ├── store/         # Stores Zustand
│   │   ├── api/           # Hooks TanStack Query
│   │   └── utils/         # Utilitaires
│   └── ...
├── scripts/               # Scripts de test
├── docs/                  # Documentation
├── maps/                  # Fichiers webp des cartes
├── Dockerfile             # Multi-stage (frontend + Go + alpine)
├── Dockerfile.dev
└── docker-compose.yml
```

## Variables d'environnement

Les variables sont définies dans `.env.example`. Voir la [documentation complète](docs/configuration.md) pour le détail.

| Variable | Rôle |
|---|---|
| `PORT` | Port du serveur (défaut: 8080) |
| `DATABASE_URL` | URL de connexion PostgreSQL |
| `DB_HOST`, `DB_PORT`, etc. | Configuration DB individuelle |
| `DISCORD_CLIENT_ID` | Client ID Discord OAuth |
| `DISCORD_CLIENT_SECRET` | Client Secret Discord OAuth |
| `JWT_SECRET` | Clé secrète JWT (obligatoire) |
| `CORS_ORIGINS` | Origines CORS autorisées |

## Commandes utiles

```bash
go test ./...                           # Tests unitaires
go test --tags=integration ./...        # Tests intégration
bash scripts/test.sh                    # Tous les tests
bash scripts/test-e2e.sh                # Tests E2E (Docker)
cd frontend && npm run build            # Build frontend
docker compose logs -f                  # Logs Docker
```

## Documentation

- [Configuration](docs/configuration.md) — toutes les variables d'environnement
- [Déploiement](docs/deployment.md) — guide de déploiement local et production
- [Spécification fonctionnelle](docs/specification/specification.md)
- [Schéma de base de données](docs/specification/database_schema.md)
- [Architecture technique](docs/architecture.md)
- [Planning d'implémentation](docs/stories/workflow.md)
