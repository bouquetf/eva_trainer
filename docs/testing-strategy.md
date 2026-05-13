# Stratégie de tests automatisés

## 1. Introduction

Ce document définit la stratégie de tests pour le projet Eva Trainer. L'objectif est d'assurer la qualité et la stabilité de l'application via une couverture fonctionnelle à 100%, couvrant les tests unitaires, d'intégration et End-to-End.

### Périmètre
- **Backend** : Go 1.26 + Gin (API REST)
- **Frontend** : React 19 + TypeScript + Vite
- **Base de données** : PostgreSQL 16
- **E2E** : Playwright

## 2. Piliers de test

### Tests unitaires (Go + Vitest)
- **Backend** : logique métier pure, services, validation, algorithmes, utilitaires.
- **Frontend** : stores Zustand, utilitaires de validation, hooks TanStack Query.
- Pas de dépendance externe (DB, réseau, filesystem).

### Tests d'intégration (Go + testcontainers)
- Endpoints HTTP Gin avec base PostgreSQL réelle (testcontainers).
- Middlewares (auth, permissions, recovery).
- Repositories (CRUD, contraintes, transactions).
- Build tag `//go:build integration` pour isolation.

### Tests End-to-End (Playwright)
- Parcours utilisateur complets avec UI navigateur.
- Happy paths et scénarios critiques par story.
- Container WireMock pour mock Discord OAuth.

## 3. Stack technique

### Backend
- **Runner** : `go test`
- **Build tags** : `integration` pour tests avec DB
- **DB de test** : `testcontainers-go` (PostgreSQL 16 Alpine)
- **Mocks** : `httptest`, `fstest.MapFS`
- **Reports** : `-coverprofile`, `-json`

### Frontend
- **Runner** : `vitest`
- **Rendu** : `@testing-library/react` + `jsdom`
- **Mocks** : `vi.mock()`, `vi.fn()`
- **Setup** : `frontend/src/test-setup.ts` (import `@testing-library/jest-dom/vitest`)

### E2E
- **Framework** : Playwright (chromium)
- **Mock OAuth** : WireMock (container Docker)
- **Base URL** : `http://localhost:8080`
- **Setup** : Docker Compose (app + db + discord-mock)

## 4. Couverture attendue

### 4.1 Backend Go

#### Handlers (`internal/handlers/`) — Tests d'intégration
| Handler | Test | Status | Priorité |
|---------|------|--------|----------|
| `admin.go` | `admin_test.go` | ✅ Existe | — |
| `auth.go` | `auth_test.go` | ✅ Existe | — |
| `edit_lock.go` | `edit_lock_test.go` | ✅ Existe | — |
| `health.go` | `health_test.go` | ✅ Existe | — |
| `history.go` | ❌ | ❌ Manquant | Haute |
| `map.go` | `map_test.go` | ✅ Existe | — |
| `map_catalog.go` | ❌ | ❌ Manquant | Haute |
| `revision.go` | `revision_test.go` | ✅ Existe | — |
| `split.go` | `split_test.go` | ✅ Existe | — |
| `team.go` | `team_test.go` | ✅ Existe | — |
| `team_invitation.go` | ❌ | ❌ Manquant | Haute |
| `team_join_request.go` | ❌ | ❌ Manquant | Haute |
| `team_roles.go` | `team_roles_test.go` | ✅ Existe | — |
| `template.go` | `template_test.go` | ✅ Existe | — |
| `zone.go` | `zone_test.go` | ✅ Existe | — |

**Résultat** : 11/15 handlers couverts (73%)

#### Modèles/Repositories (`internal/models/`) — Tests d'intégration
| Fichier | Test | Status | Priorité |
|---------|------|--------|----------|
| `edit_history.go` | ❌ | ❌ Manquant | Haute |
| `edit_lock.go` | `edit_lock_test.go` | ✅ Existe | — |
| `map.go` | `map_test.go` | ✅ Existe | — |
| `map_catalog.go` | ❌ | ❌ Manquant | Haute |
| `permission.go` | `permission_test.go` | ✅ Existe | — |
| `refresh_token.go` | `refresh_token_test.go` | ✅ Existe | — |
| `role_permission.go` | ❌ | ❌ Manquant | Moyenne |
| `split.go` | `split_test.go` | ✅ Existe | — |
| `team.go` | `team_test.go` | ✅ Existe | — |
| `team_invitation.go` | `team_invitation_test.go` | ✅ Existe | — |
| `team_join_request.go` | `team_join_request_test.go` | ✅ Existe | — |
| `team_member.go` | ❌ | ❌ Manquant | Moyenne |
| `template.go` | ❌ | ❌ Manquant | Haute |
| `user.go` | `user_test.go` | ✅ Existe | — |
| `zone.go` | `zone_test.go` | ✅ Existe | — |
| `zone_version.go` | ❌ | ❌ Manquant | Haute |

**Résultat** : 10/16 repositories couverts (63%)

#### Middlewares (`internal/middleware/`) — Tests unitaires
| Fichier | Test | Status | Priorité |
|---------|------|--------|----------|
| `admin.go` | `admin_test.go` | ✅ Existe | — |
| `auth.go` | `auth_test.go` | ✅ Existe | — |
| `errors.go` | `errors_test.go` | ✅ Existe | — |
| `permission.go` | `permission_test.go` | ✅ Existe | — |
| `recovery.go` | `recovery_test.go` | ✅ Existe | — |
| `team.go` | ❌ | ❌ Manquant | Moyenne |

**Résultat** : 5/6 middlewares couverts (83%)

#### Services (`internal/services/`) — Tests unitaires
| Fichier | Test | Status | Priorité |
|---------|------|--------|----------|
| `discord.go` | ❌ | ❌ Manquant | Moyenne |
| `jwt.go` | `jwt_test.go` | ✅ Existe | — |
| `lock_cleanup.go` | ❌ | ❌ Manquant | Faible |
| `logger.go` | `logger_test.go` | ✅ Existe | — |
| `template_sync.go` | `template_sync_test.go` | ✅ Existe | — |

**Résultat** : 3/5 services couverts (60%)

#### Packages (`pkg/`) — Tests unitaires
| Fichier | Test | Status |
|---------|------|--------|
| `apierror/errors.go` | `errors_test.go` | ✅ Existe |
| `config/config.go` | `config_test.go` | ✅ Existe |
| `database/config.go` | `config_test.go` | ✅ Existe |
| `database/database.go` | `database_test.go` | ✅ Existe |
| `database/migrations.go` | `migrations_test.go` | ✅ Existe |

**Résultat** : 5/5 packages couverts (100%)

#### Routes (`api/`) — Tests unitaires
| Fichier | Test | Status |
|---------|------|--------|
| `routes.go` | `routes_test.go`, `routes_debug_test.go` | ✅ Existe |

**Résultat** : 1/1 couvert (100%)

### 4.2 Frontend

#### API hooks (`frontend/src/api/`) — Tests unitaires
| Source | Test | Status | Priorité |
|--------|------|--------|----------|
| `admin.ts` | ❌ | ❌ Manquant | Faible |
| `auth.ts` | ❌ | ❌ Manquant | Faible |
| `client.ts` | ❌ | ❌ Manquant | Faible |
| `invitations.ts` | `invitations.test.ts` | ✅ Existe | — |
| `joinRequests.ts` | `joinRequests.test.ts` | ✅ Existe | — |
| `locks.ts` | `locks.test.ts` | ✅ Existe | — |
| `map-catalog.ts` | ❌ | ❌ Manquant | Faible |
| `maps.ts` | ❌ | ❌ Manquant | Faible |
| `revision.ts` | `revision.test.ts` | ✅ Existe | — |
| `roles.ts` | `roles.test.ts` | ✅ Existe | — |
| `splits.ts` | ❌ | ❌ Manquant | Faible |
| `teams.ts` | `teams.test.ts` | ✅ Existe | — |
| `templates.ts` | ❌ | ❌ Manquant | Faible |
| `versions.ts` | `versions.test.ts` | ✅ Existe | — |
| `zones.ts` | `zones.test.ts` | ✅ Existe | — |

**Résultat** : 8/15 hooks couverts (53%)

#### Composants (`frontend/src/components/`) — Tests unitaires
| Source | Test | Status | Priorité |
|--------|------|--------|----------|
| `AdminGuard.tsx` | ❌ | ❌ Manquant | Moyenne |
| `AppLayout.tsx` | ❌ | ❌ Manquant | Moyenne |
| `AuthGuard.tsx` | ❌ | ❌ Manquant | Moyenne |
| `ConfirmDialog.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/ContextMenu.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/LockIndicator.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/MapCanvas.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/MapEditor.tsx` | `MapEditor.test.tsx` | ✅ Existe | — |
| `MapEditor/PointHandles.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/PublishButton.tsx` | `PublishButton.test.tsx` | ✅ Existe | — |
| `MapEditor/Toolbar.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/ZoneDetails.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/ZoneList.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/ZoneNameDialog.tsx` | ❌ | ❌ Manquant | Faible |
| `MapEditor/ZoneOverlay.tsx` | ❌ | ❌ Manquant | Faible |
| `MemberList.tsx` | `MemberList.test.tsx` | ✅ Existe | — |
| `RestoreConfirmModal.tsx` | ❌ | ❌ Manquant | Faible |
| `RevisionMap.tsx` | `RevisionMap.test.tsx` | ✅ Existe | — |
| `RevisionScore.tsx` | ❌ | ❌ Manquant | Faible |
| `TeamsLayout.tsx` | ❌ | ❌ Manquant | Faible |
| `TransferCaptaincyDialog.tsx` | ❌ | ❌ Manquant | Faible |
| `VersionBrowser.tsx` | ❌ | ❌ Manquant | Faible |

**Résultat** : 5/22 composants couverts (23%)

#### Pages (`frontend/src/pages/`) — Tests unitaires
| Source | Test | Status | Priorité |
|--------|------|--------|----------|
| `AdminDashboard.tsx` | ❌ | ❌ Manquant | Faible |
| `Callback.tsx` | ❌ | ❌ Manquant | Faible |
| `CreateTeam.tsx` | ❌ | ❌ Manquant | Faible |
| `Editor.tsx` | ❌ | ❌ Manquant | Haute |
| `History.tsx` | ❌ | ❌ Manquant | Faible |
| `Home.tsx` | ❌ | ❌ Manquant | Faible |
| `InviteAccept.tsx` | ❌ | ❌ Manquant | Faible |
| `Login.tsx` | ❌ | ❌ Manquant | Faible |
| `NotFound.tsx` | ❌ | ❌ Manquant | Faible |
| `ProfileSettings.tsx` | `ProfileSettings.test.tsx` | ✅ Existe | — |
| `Revision.tsx` | `Revision.test.tsx` | ✅ Existe | — |
| `TeamSettings.tsx` | `TeamSettings.test.tsx` | ✅ Existe | — |
| `TeamsList.tsx` | ❌ | ❌ Manquant | Faible |

**Résultat** : 3/13 pages couvertes (23%)

#### Stores (`frontend/src/store/`) — Tests unitaires
| Source | Test | Status |
|--------|------|--------|
| `auth.ts` | `auth.test.ts` | ✅ Existe |
| `revision.ts` | `revision.test.ts` | ✅ Existe |

**Résultat** : 2/2 stores couverts (100%)

#### Utilitaires (`frontend/src/lib/`) — Tests unitaires
| Source | Test | Status | Priorité |
|--------|------|--------|----------|
| `geometry/coordinates.ts` | ❌ | ❌ Manquant | Faible |
| `geometry/transforms.ts` | ❌ | ❌ Manquant | Faible |
| `geometry/zoneFactory.ts` | ❌ | ❌ Manquant | Faible |
| `revision/validation.ts` | `validation.test.ts` | ✅ Existe | — |

**Résultat** : 1/4 fichiers de logique couverts (25%)

### 4.3 E2E (Playwright)

Les tests E2E seront implémentés dans `frontend/e2e/` via les tâches #002 à #005.

**Couverture prévue :**
- Authentification Discord (login, logout, refresh token, Discord down)
- Gestion des équipes (création, rôles, invitations, transfert)
- Upload et gestion des cartes (admin)
- Édition de zones (création polygones, nommage, verrouillage)
- Templates (publication, duplication, auto-update)
- Historique (versions, rollback, restrictions)
- Mode révision (validation casse/accents, score, poids, zones obsolètes)

## 5. Conventions et patterns

### 5.1 Backend Go

**Nommage :**
- Tests unitaires : `*_test.go` (même package, pas de build tag)
- Tests d'intégration : `*_test.go` avec `//go:build integration` en header

**Build tags :**
```go
//go:build integration

package handlers
```

**Helpers de test :**
- `setupTestDB()` dans `internal/models/setup_test.go` — PostgreSQL via testcontainers
- Migration files listées explicitement
- Nettoyage via `t.Cleanup()` ou fonction retournée

**Pattern handler test :**
```go
func TestXxx(t *testing.T) {
    gin.SetMode(gin.TestMode)
    db, cleanup := setupTestXxx(t)
    defer cleanup()
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    // ...
}
```

**Pattern repository test :**
```go
//go:build integration

func TestRepo_Create(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := NewXxxRepository(db)
    // ...
}
```

### 5.2 Frontend

**Nommage :**
- Tests unitaires : `*.test.ts` ou `*.test.tsx`
- Même répertoire que le fichier testé

**Pattern store test (Zustand) :**
```typescript
import { useXxxStore } from './xxx'

beforeEach(() => {
  useXxxStore.setState({ /* état initial */ })
})

it('should ...', () => {
  useXxxStore.getState().someAction()
  expect(useXxxStore.getState().someValue).toBe(...)
})
```

**Pattern API hook test (TanStack Query) :**
```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
```

**Pattern component test :**
```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
```

### 5.3 E2E Playwright

**Organisation :**
```
frontend/e2e/
├── auth/
├── teams/
├── maps/
├── zones/
├── templates/
├── history/
├── revision/
├── helpers/
│   ├── auth.ts
│   ├── db.ts
│   └── api.ts
└── fixtures.ts
```

**Pattern :**
```typescript
import { test, expect } from '../fixtures';

test('scenario description', async ({ page, adminUser }) => {
  await page.goto('/');
  // ...
});
```

## 6. CI Pipeline

### Workflow GitHub Actions (`.github/workflows/ci.yml`)

1. **lint** : golangci-lint + ESLint (parallélisé)
2. **unit-backend** : `go test ./...` + coverage
3. **unit-frontend** : `npx vitest run --coverage`
4. **integration-backend** : `go test --tags=integration ./...` (PostgreSQL service)
5. **e2e** : Playwright (chromium) sur Docker Compose complet

### Ordre d'exécution
```
lint ───────────────────────┐
unit-backend ───────────────┤
unit-frontend ──────────────┤
                            ↓
                    integration-backend
                            ↓
                         e2e
```

### Caches
- Go modules (`go.sum`)
- node_modules (`package-lock.json`)
- Playwright browsers

### Artifacts
- Coverage Go
- Coverage Vitest
- Rapport HTML Playwright
- Screenshots d'échec Playwright
- Logs Docker Compose

## 7. Commandes

```bash
# Backend
go test ./...                                   # Tests unitaires
go test --tags=integration ./...                # Tests intégration
go test -coverprofile=coverage.out ./...        # Coverage
go tool cover -html=coverage.out                # Visualisation

# Frontend
cd frontend && npx vitest run                   # Tests unitaires
cd frontend && npx vitest run --coverage        # Coverage
cd frontend && npx vitest --ui                  # UI mode

# E2E
cd frontend && npx playwright test              # E2E headless
cd frontend && npx playwright test --ui         # E2E UI mode
cd frontend && npx playwright test --debug      # E2E debug

# Tout
bash scripts/test.sh                            # Tout
```

## 8. Annexes

### 8.1 Synthèse de l'audit

#### Backend Go — 59% global

| Couche | Couverts | Total | % |
|--------|----------|-------|---|
| Handlers | 11 | 15 | 73% |
| Models | 10 | 16 | 63% |
| Middleware | 5 | 6 | 83% |
| Services | 3 | 5 | 60% |
| Packages | 5 | 5 | 100% |
| Routes | 1 | 1 | 100% |

#### Frontend — 31% global

| Couche | Couverts | Total | % |
|--------|----------|-------|---|
| API | 8 | 15 | 53% |
| Components | 5 | 22 | 23% |
| Pages | 3 | 13 | 23% |
| Stores | 2 | 2 | 100% |
| Lib | 1 | 4 | 25% |

### 8.2 Priorités de gap filling

#### Haute priorité (fonctionnalités sans test)
1. `internal/handlers/history.go` — Historique, rollback, visibilité éditeur
2. `internal/handlers/team_invitation.go` — Invitations par pseudo, lien, expiration
3. `internal/handlers/team_join_request.go` — Demandes d'adhésion, acceptation/refus
4. `internal/handlers/map_catalog.go` — Catalogue de cartes (listes publiques)
5. `internal/models/template.go` — CRUD template, publication, auto-update
6. `internal/models/zone_version.go` — Snapshots zone_versions, rollback
7. `internal/models/edit_history.go` — Audit trail des modifications de zone
8. `frontend/src/pages/Editor.tsx` — Page éditeur de zone

#### Priorité moyenne
1. `internal/middleware/team.go` — Middleware d'appartenance à une équipe
2. `internal/models/role_permission.go` — Mapping rôle ↔ permissions
3. `internal/models/team_member.go` — Gestion des membres d'équipe
4. `internal/services/discord.go` — Service d'appel API Discord
5. `frontend/src/api/` — Hooks API manquants
6. `frontend/src/components/AdminGuard.tsx`, `AuthGuard.tsx`, `AppLayout.tsx`

#### Priorité faible (UI secondaire)
- Tous les composants MapEditor sub (sauf MapEditor et PublishButton déjà testés)
- Toutes les pages sauf Editor (couverture minimale acceptable)
- `internal/services/lock_cleanup.go` — Service de nettoyage (utile mais simple)
- `frontend/src/lib/geometry/` — Utilitaires géométriques

### 8.3 Migrations nécessaires dans setup_test.go

```go
var migrationFiles = []string{
    "000001_initial_schema.up.sql",
    "000002_add_refresh_tokens.up.sql",
    "000003_add_display_name.up.sql",
    "000004_add_discord_username.up.sql",
    "000007_add_team_invitations_and_join_requests.up.sql",
    "000011_add_zone_polygon_constraints.up.sql",
    // Migrations manquantes à ajouter :
    // "000005_seed_permissions.up.sql",
    // "000006_drop_current_version_id.up.sql",
    // "000008_add_split_permission.up.sql",
    // "000009_add_map_catalog.up.sql",
    // "000010_drop_split_dates.up.sql",
    // "000012_add_zone_permissions.up.sql",
    // "000013_add_zone_versions.up.sql",
    // "000014_fix_edit_history_permissions.up.sql",
    // "000015_add_template_is_public.up.sql",
}
```
