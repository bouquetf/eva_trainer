# Tâche #006 - Story #000 : Documentation et README

## Objectif

Créer le fichier README.md principal du projet avec les instructions de démarrage, l'architecture technique, et le diagramme d'architecture.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Tâche #001 (Backend Go+Gin), Tâche #004 (Frontend structure), Tâche #005 (Tests, pour connaître les commandes)
- Nécessaire pour : Tâche #007 (Validation)
- Ordre d'exécution : 6ème (après #005)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer la documentation principale du projet.

**Cas nominaux :**
- README.md complet avec description du projet et instructions de démarrage
- `docs/architecture.md` avec diagramme technique Mermaid et description des couches
- Structure des répertoires documentée
- Technologies utilisées listées avec versions

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `README.md` | Créer | Documentation principale du projet |
| `docs/architecture.md` | Créer | Diagramme et description de l'architecture |

### Structure du README.md

Le README doit contenir les sections suivantes :

1. **Titre et description** : Eva Trainer - Application d'entraînement à la reconnaissance de zones sur cartes Eva
2. **Technologies** :
   - Backend : Go 1.26 + Gin
   - Frontend : React 19 + Vite + TypeScript
   - Base de données : PostgreSQL 16
   - Conteneurisation : Docker Compose
   - Librairies frontend : React-Konva, Zustand, TanStack Query, React Router DOM
3. **Prérequis** : Go 1.26+, Node.js 20+, Docker Desktop, npm
4. **Démarrage rapide** :
   ```bash
   cp .env.example .env
   docker compose up -d
   go run cmd/migrate/main.go
   cd frontend && npm install && npm run dev
   ```
5. **Architecture** :
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
6. **Variables d'environnement** : Lister les variables principales et leur rôle (référer à `.env.example`)
7. **Commandes utiles** : `go test ./...`, `go test --tags=integration ./...`, `bash scripts/test.sh`, `bash scripts/test-e2e.sh`, `npm run build`, `docker compose logs -f`
8. **Liens utiles** :
   - Spécification fonctionnelle : `docs/specification/specification.md`
   - Schéma de base de données : `docs/specification/database_schema.md`
   - Planning d'implémentation : `docs/stories/workflow.md`
   - Architecture technique : `docs/architecture.md`

### Structure du fichier docs/architecture.md

1. **Diagramme d'architecture (développement)** :
   ```mermaid
   graph TD
     subgraph Développement
       Dev[Browser] --> Vite[Vite Dev Server :5173]
       Dev --> GoDev[Go + Gin API :8080]
       GoDev --> PG[(PostgreSQL :5432)]
       GoDev --> Discord[Discord OAuth API]
     end

     subgraph Frontend Stack
       React[React App]
       Router[React Router]
       Zustand[Zustand Store]
       RQ[TanStack Query]
       Konva[React-Konva Canvas]
     end

     subgraph Backend Stack
       Gin[Gin Router]
       H[HTTP Handlers]
       MW[Middleware]
       M[Models]
       DB[pkg/database]
       Mig[golang-migrate]
     end

     Dev --> Frontend Stack
     Gin --> H
     H --> MW
     H --> M
     M --> DB
     DB --> Mig
     DB --> PG
   ```

2. **Diagramme d'architecture (production)** :
   ```mermaid
   graph TD
     Prod[Browser] --> GoProd[Go + Gin API (sert aussi le frontend)]
     GoProd --> PG[(PostgreSQL)]
     GoProd --> Discord[Discord OAuth API]

     subgraph Frontend servi par Go
       Static[Fichiers statiques compilés]
     end

     subgraph Backend
       Gin
       H
       MW
       M
       DB
       Mig
     end

     Prod --> Frontend servi par Go
     GoProd --> Backend
     Backend --> PG
   ```

3. **Description des couches** :
   - **Frontend** : React 19 + TypeScript, Vite (dev) / servi par Go (prod)
   - **Backend** : Go 1.26 + Gin (API REST)
   - **Base de données** : PostgreSQL 16
   - **Migrations** : golang-migrate
4. **Flux de données** : Client → Gin Router → Middleware → Handler → Models → Database → PostgreSQL
5. **Patterns** : MVC/MVVM, Repository pattern pour l'accès DB, Middleware chain pour la sécurité

### Contraintes techniques

- **Format** : Markdown propre, bien structuré
- **Images** : Ne pas inclure d'images réelles, utiliser Mermaid
- **Badges** : Ne pas inclure de badges
- **Liens** : Liens relatifs vers les fichiers du projet
- **Version Go** : Mentionner Go 1.26 (sans le patch)
- **Commande Docker** : Utiliser `docker compose` (nouvelle syntaxe) dans la documentation

### Documentation

- `README.md` à la racine
- `docs/architecture.md`
- Le README référence `docs/specification/specification.md`, `docs/specification/database_schema.md`, `docs/stories/workflow.md`, `docs/architecture.md`