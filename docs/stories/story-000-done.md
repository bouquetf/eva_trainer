# Story #000 : Project Initialization and Scaffolding

## Description
En tant que développeur, je veux initialiser la structure du projet avec les technologies choisies afin de pouvoir commencer le développement des fonctionnalités.

## Critères d'acceptation
- [ ] Initialisation du projet Go avec le framework Gin (structure MVC/MVVM)
- [ ] Configuration de la base de données PostgreSQL avec driver approprié
- [ ] Initialisation du frontend avec Vite + ReactJS
- [ ] Installation des librairies frontend : React-Konva, Zustand, TanStack Query, React Router DOM
- [ ] Configuration des routes frontend dans `/src/routes/` ou `/src/App.jsx` (React Router)
- [ ] Configuration du routeur backend Gin dans `/api/routes.go` (définition des endpoints de base)
- [ ] Création du fichier `docker-compose.yml` pour le développement local (PostgreSQL + app)
- [ ] Mise en place du schéma de base de données initial (selon `docs/specification/database_schema.md`)
- [ ] Configuration des variables d'environnement (exemple `.env.example`)
- [ ] Structure des dossiers backend : `/cmd`, `/internal`, `/pkg`, `/api`, `/models`, `/middleware`
- [ ] Structure des dossiers frontend : `/src`, `/components`, `/pages`, `/store`, `/api`, `/utils`
- [ ] README.md initial avec instructions de démarrage

## Tests automatisés
- Test unitaire : Vérification que l'application démarre (endpoint de santé `/health`)
- Test d'intégration : Connexion à la base de données PostgreSQL au démarrage
- Test E2E : Démarrage avec `docker-compose up` → accès à http://localhost:PORT/health

## Documentation
- README.md (à créer)
- docs/architecture.md (à créer - diagramme technique)
- Voir specification.md section 5 (Technique)

## Valeur utilisateur
Fournit une base de code propre et structurée pour accueillir toutes les fonctionnalités de l'application Eva Trainer.

## Dépendances
- Aucune

## Notes techniques
- **Backend** : Go + Gin, structure :
  ```
  /
  ├── cmd/server/main.go (initialisation du routeur Gin)
  ├── internal/handlers/
  ├── internal/models/
  ├── internal/middleware/
  ├── pkg/database/
  ├── api/routes.go (définition des routes API : /api/v1/...)
  ├── Dockerfile
  └── docker-compose.yml
  ```
- **Frontend** : React + Vite, structure :
  ```
  /frontend
  ├── src/components/
  ├── src/pages/
  ├── src/routes/ (React Router configuration)
  ├── src/store/ (Zustand)
  ├── src/api/ (TanStack Query)
  ├── src/utils/
  ├── src/App.jsx (avec <BrowserRouter>)
  └── vite.config.js
  ```
- **Base de données** : Scripts d'initialisation dans `/db/migrations/`
