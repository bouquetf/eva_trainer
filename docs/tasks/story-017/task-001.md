# Tâche #001 - Story #017 : Testing Strategy Document and Coverage Audit

## Objectif
Créer le document `docs/testing-strategy.md` détaillant la stratégie de tests complète du projet, et auditer la couverture existante pour identifier les lacunes.

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Toutes les stories (le document couvre l'ensemble du projet)
- Nécessaire pour : Tâches #002 à #006 (guide la priorisation des efforts de test)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Document de stratégie de tests
Créer `docs/testing-strategy.md` avec les sections suivantes :

1. **Introduction** : Objectif de la stratégie, périmètre du projet
2. **Piliers de test** :
   - Tests unitaires (Go backend, Vitest frontend)
   - Tests d'intégration (Go + testcontainers PostgreSQL)
   - Tests End-to-End (Playwright)
3. **Stack technique** :
   - Backend : `go test` (build tag `integration`), `testcontainers-go` pour DB
   - Frontend : `vitest` + `@testing-library/react`
   - E2E : Playwright avec Docker Compose
4. **Couverture attendue** :
   - **Unit backend** : services, validation, utilitaires purs, algorithmes
   - **Unit frontend** : stores Zustand, utilitaires, hooks TanStack Query
   - **Integration backend** : endpoints HTTP avec DB réelle, middlewares
   - **Integration frontend** : composants React connectés à des mocks API
   - **E2E** : happy paths et scénarios critiques par story
5. **Conventions et patterns** :
   - Nommage des fichiers : `*_test.go` / `*.test.ts` / `*.test.tsx`
   - Build tags : `//go:build integration` pour les tests d'intégration Go
   - Helpers : `setupTestDB()`, `setupRevisionTest()`, etc.
   - Mocks : httptest, fstest pour les fichiers
6. **Organisation des tests E2E par story** :
   - Story #002 : Discord auth (login, callback, refresh token, 404 Discord down)
   - Story #003 : Team CRUD (create, rename, delete)
   - Story #004 : Roles (assign, transfer, constraints)
   - Story #005 : Maps (upload, obsolete, purge, image serving)
   - Story #006 : Invitations (by pseudo, link, join requests)
   - Story #007 : Splits (admin CRUD, filtering)
   - Story #008 : API security (role verification, permissions)
   - Story #009 : Zone editing (polygon creation, naming, constraints)
   - Story #010 : Map locking (lock, timeout, conflict)
   - Story #011 : Publishing (zone publish, template, auto-update)
   - Story #012 : History (version snapshots, rollback, visibility)
   - Story #013 : Error handling (standard format, logging)
   - Story #014 : Template duplication (copy, independence)
   - Story #015 : Revision mode (zone display, validation, score)
   - Story #016 : Weighted probability (weight update, selection)
7. **Exécution en CI** :
   - Workflow GitHub Actions (unit → integration → E2E)
   - Parallélisation des jobs
   - Stratégie de cache (Go modules, node_modules, Playwright browsers)
   - Timeouts, retries, artifacts (screenshots E2E)
8. **Métriques et rapports** :
   - Sortie JSON des résultats (`go test -json`)
   - Rapport de couverture (`go test -coverprofile`, `vitest --coverage`)
   - Playwright HTML report
9. **Annexe : Audit de couverture**

#### Audit de couverture existante
Analyser les fichiers suivants pour établir un état des lieux :

**Backend Go :**
- `internal/handlers/` : Lister les fichiers `*_test.go` existants vs handlers sans tests
- `internal/models/` : Mêmes calculs
- `internal/middleware/` : Mêmes calculs
- `internal/services/` : Mêmes calculs
- `pkg/` : Mêmes calculs
- `api/` : Mêmes calculs

Pour chaque dossier, établir un tableau :
| Fichier source | Fichier test | Status | Scénarios couverts |
|----------------|-------------|--------|-------------------|

**Frontend :**
- `frontend/src/api/` : Fichiers `*.test.ts` vs sources
- `frontend/src/components/` : `*.test.tsx` vs sources
- `frontend/src/pages/` : `*.test.tsx` vs sources
- `frontend/src/store/` : `*.test.ts` vs sources
- `frontend/src/lib/` : `*.test.ts` vs sources

#### Règles additionnelles
- Vérifier que les tests d'intégration utilisent `testcontainers.PostgreSQL` (pattern existant)
- Vérifier que les tests unitaires n'ont pas de dépendance externe (DB, réseau)
- Vérifier que les tests d'intégration utilisent le build tag `//go:build integration`
- Vérifier que les tests existants passent avec `go test ./...` et `go test --tags=integration ./...`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/testing-strategy.md` | Créer | Document complet de stratégie de tests |
| Aucun fichier source modifié | - | Tâche documentaire uniquement |

### Contraintes techniques
- **Convention existante** : Les tests d'intégration Go utilisent `//go:build integration` en header
- **Convention existante** : Les tests unitaires Go n'ont PAS de build tag
- **Convention existante** : Les tests frontend utilisent `vitest` avec `@testing-library/react`
- **Infrastructure existante** : `testcontainers-go` pour les bases de données de test
- **Documentation** : Respecter le format Markdown des autres docs dans `docs/`

### Tests à implémenter

Aucun test à implémenter pour cette tâche (tâche documentaire + audit).

### Documentation

#### Documentation à créer
- `docs/testing-strategy.md` : Stratégie de tests complète

#### Documentation à mettre à jour
- `README.md` : Ajouter une section "Tests" pointant vers `docs/testing-strategy.md`

### Exemples d'API

Sans objet (tâche documentaire).
