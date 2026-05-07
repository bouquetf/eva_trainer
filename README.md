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
go run cmd/migrate/main.go
cd frontend && npm install && npm run dev
```

## Architecture

```
/
├── cmd/                # Points d'entrée Go
│   ├── server/         # Serveur API
│   └── migrate/        # Commande de migration
├── api/                # Définition des routes API
├── internal/           # Code interne Go
│   ├── handlers/       # Handlers HTTP
│   ├── models/         # Modèles de données
│   └── middleware/     # Middlewares
├── pkg/database/       # Connexion PostgreSQL + migrations
├── db/migrations/      # Migrations SQL (golang-migrate)
├── frontend/           # Application React
│   ├── src/
│   │   ├── components/  # Composants réutilisables
│   │   ├── pages/       # Pages de l'application
│   │   ├── routes/      # Configuration React Router
│   │   ├── store/       # Stores Zustand
│   │   ├── api/         # Hooks TanStack Query
│   │   └── utils/       # Utilitaires
│   └── ...
├── scripts/            # Scripts de test
├── docs/               # Documentation
│   ├── specification/   # Spécification fonctionnelle
│   ├── stories/         # User stories
│   └── tasks/           # Tâches de découpage
├── maps/               # Fichiers webp des cartes
├── Dockerfile
├── Dockerfile.dev
└── docker-compose.yml
```

## Variables d'environnement

Les variables sont définies dans `.env.example`. Principales :

| Variable | Rôle |
|---|---|
| `PORT` | Port du serveur (défaut: 8080) |
| `DATABASE_URL` | URL de connexion PostgreSQL |
| `DB_HOST`, `DB_PORT`, etc. | Configuration DB individuelle |
| `DISCORD_CLIENT_ID` | Client ID Discord OAuth |
| `DISCORD_CLIENT_SECRET` | Client Secret Discord OAuth |
| `JWT_SECRET` | Clé secrète JWT |
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

## Liens utiles

- [Spécification fonctionnelle](docs/specification/specification.md)
- [Schéma de base de données](docs/specification/database_schema.md)
- [Planning d'implémentation](docs/stories/workflow.md)
- [Architecture technique](docs/architecture.md)
