# Tâche #007 - Story #000 : Validation finale (Definition of Done)

## Objectif

Valider que l'ossature complète du projet est fonctionnelle et que tous les critères d'acceptation de la Story #000 sont remplis.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Tâche #005 (Tests), Tâche #006 (Documentation)
- Nécessaire pour : Aucune (dernière tâche)
- Ordre d'exécution : 7ème et dernière

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Exécuter la validation finale de l'ossature du projet en vérifiant tous les points de la Definition of Done.

**Cas nominaux :**
- Tous les services Docker démarrent correctement
- Le endpoint de santé répond
- Le frontend est accessible
- La base de données est connectable
- Les migrations s'exécutent
- La structure des dossiers est complète
- Les librairies sont installées et les routeurs configurés
- La documentation est créée

**Cas limites :**
- Si un test échoue, documenter l'erreur et proposer un correctif
- Si la structure est incomplète, créer les dossiers manquants

### Critères de validation

#### 8.1 Infrastructure Docker
- [ ] `docker compose up -d` démarre les services `app` et `db` sans erreur
- [ ] Le service `db` est healthy (pg_isready)
- [ ] Le service `app` est running et ne redémarre pas en boucle

#### 8.2 Backend Go
- [ ] `curl http://localhost:8080/health` retourne `{"status": "healthy"}` (HTTP 200)
- [ ] Le groupe `/api/v1/` est configuré (vide pour l'instant, ne retourne pas 404 sur `/api/v1`)
- [ ] `go build ./cmd/server` compile sans erreur
- [ ] `go vet ./...` ne remonte aucun warning

#### 8.3 Base de données
- [ ] `go run cmd/migrate/main.go` exécute les migrations sans erreur
- [ ] La table `users` existe dans PostgreSQL (vérifiable via `\dt` ou `psql`)
- [ ] La table `schema_migrations` est créée par golang-migrate
- [ ] Les index sont créés

#### 8.4 Frontend
- [ ] `cd frontend && npm run build` compile sans erreur
- [ ] `cd frontend && npx tsc --noEmit` ne retourne pas d'erreur TypeScript
- [ ] Les librairies sont installées (vérifiable dans `package.json`)
- [ ] Les répertoires `pages/`, `routes/`, `store/`, `api/`, `utils/` existent dans `frontend/src/`
- [ ] Les fichiers `Home.tsx` et `NotFound.tsx` existent dans `pages/`

#### 8.5 Tests
- [ ] `go test ./...` passe (tests unitaires)
- [ ] `go test --tags=integration ./...` passe (tests d'intégration, nécessite PostgreSQL)
- [ ] `bash scripts/test.sh` s'exécute sans erreur

#### 8.6 Documentation
- [ ] `README.md` existe à la racine avec les sections requises
- [ ] `docs/architecture.md` existe avec le diagramme Mermaid

#### 8.7 Structure du projet
- [ ] Structure backend complète : `cmd/`, `api/`, `internal/`, `pkg/`, `db/migrations/`
- [ ] Structure frontend complète : `components/`, `pages/`, `routes/`, `store/`, `api/`, `utils/`

### Actions en cas d'échec

- **Docker ne démarre pas** : Vérifier les logs avec `docker compose logs`, s'assurer que les ports ne sont pas déjà utilisés
- **Health endpoint ne répond pas** : Vérifier que `main.go` est correct, que le port est le bon
- **Migrations échouent** : Vérifier la connexion DB, s'assurer que PostgreSQL est healthy
- **Frontend ne compile pas** : Vérifier les imports, `package.json`, la configuration TypeScript
- **Tests échouent** : Analyser l'erreur, s'assurer que les packages sont téléchargés

### Livrables

Une fois la validation passée, créer un fichier `VALIDATED.md` dans `docs/tasks/story-000/` listant les points vérifiés et le résultat, ou marquer tous les points comme OK dans ce document.

### Contraintes

- Ne pas modifier le code des tâches précédentes pendant cette validation (sauf si un bug critique est découvert)
- Si un bug est découvert, le signaler dans le rapport plutôt que de le corriger immédiatement (sauf s'il bloque la validation)