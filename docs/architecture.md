# Architecture technique

## Diagramme d'architecture (développement)

```mermaid
graph TD
  subgraph Développement
    Dev[Browser] --> Vite[Vite Dev Server :5173]
    Dev --> GoDev[Go + Gin API :8080]
    GoDev --> PG[(PostgreSQL :5432)]
    GoDev --> Discord[Discord OAuth API]
  end

  subgraph Frontend_Stack
    React[React App]
    Router[React Router]
    Zustand[Zustand Store]
    RQ[TanStack Query]
    Konva[React-Konva Canvas]
  end

  subgraph Backend_Stack
    Gin[Gin Router]
    H[HTTP Handlers]
    MW[Middleware]
    M[Models]
    DB[pkg/database]
    Mig[golang-migrate]
  end

  Dev --> Frontend_Stack
  Gin --> H
  H --> MW
  H --> M
  M --> DB
  DB --> Mig
  DB --> PG
```

## Diagramme d'architecture (production)

```mermaid
graph TD
  Prod[Browser] --> GoProd[Go + Gin API - sert aussi le frontend]
  GoProd --> PG[(PostgreSQL)]
  GoProd --> Discord[Discord OAuth API]

  subgraph Frontend_servi_par_Go
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

  Prod --> Frontend_servi_par_Go
  GoProd --> Backend
  Backend --> PG
```

## Description des couches

- **Frontend** : React 19 + TypeScript, Vite en développement / servi par Go en production
- **Backend** : Go 1.26 + Gin (API REST)
- **Base de données** : PostgreSQL 16
- **Migrations** : golang-migrate

## Flux de données

Client → Gin Router → Middleware → Handler → Models → Database → PostgreSQL

## Patterns

- **MVC / MVVM** : Architecture inspirée modèle-vue-contrôleur
- **Repository pattern** : Accès à la base de données via `pkg/database`
- **Middleware chain** : Sécurité et logging via la chaîne de middlewares Gin
