# Tâche #002 - Story #017 : Playwright E2E Framework and CI Pipeline

## Objectif
Installer et configurer Playwright pour les tests E2E, créer les helpers et fixtures partagés, et configurer le pipeline CI (GitHub Actions) pour exécuter automatiquement tous les tests.

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Tâche #001 (stratégie définie)
- Nécessaire pour : Tâches #003, #004, #005 (E2E tests par feature)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Playwright - Installation et configuration

Installer Playwright dans le projet (`frontend/`) et créer la configuration :

**Fichiers :**
- `frontend/playwright.config.ts` : Configuration principale
- `frontend/e2e/tsconfig.json` : TypeScript config pour les tests E2E
- `frontend/e2e/global-setup.ts` : Setup global (démarrage Docker Compose si nécessaire)
- `frontend/e2e/global-teardown.ts` : Teardown global (arrêt Docker Compose)
- `frontend/e2e/helpers/auth.ts` : Helpers d'authentification (génération JWT, création utilisateur)
- `frontend/e2e/helpers/db.ts` : Helpers base de données (insertion directe pour seed)
- `frontend/e2e/helpers/api.ts` : Helpers API (requêtes typées avec gestion d'erreur)

**Playwright config :**
```typescript
// frontend/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: [
    ['html', { outputFolder: 'e2e-report' }],
    ['list'],
    ['json', { outputFile: 'e2e-results.json' }],
  ],
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  globalSetup: require.resolve('./e2e/global-setup.ts'),
  globalTeardown: require.resolve('./e2e/global-teardown.ts'),
  timeout: 60000,
  expect: {
    timeout: 10000,
  },
});
```

**Helpers d'authentification :**
```typescript
// frontend/e2e/helpers/auth.ts
export async function createTestUser(discordId: string, isAdmin?: boolean): Promise<TestUser>
export async function generateJWT(userId: number, secret: string): Promise<string>
export async function loginAsUser(page: Page, user: TestUser): Promise<void>
export async function loginAsAdmin(page: Page): Promise<void>
```

**Helpers API :**
```typescript
// frontend/e2e/helpers/api.ts
export async function apiGet<T>(path: string, token?: string): Promise<ApiResponse<T>>
export async function apiPost<T>(path: string, body?: unknown, token?: string): Promise<ApiResponse<T>>
export async function apiPut<T>(path: string, body?: unknown, token?: string): Promise<ApiResponse<T>>
export async function apiDelete<T>(path: string, token?: string): Promise<ApiResponse<T>>
```

**Helpers DB :**
```typescript
// frontend/e2e/helpers/db.ts
export async function queryDB(query: string, params?: unknown[]): Promise<unknown[]>
export async function seedUser(discordId: string, opts?: SeedUserOptions): Promise<TestUser>
export async function seedTeam(name: string, captainId: number): Promise<TestTeam>
export async function seedMap(name: string): Promise<TestMap>
export async function cleanDatabase(): Promise<void>
```

#### 2. Test fixtures partagés

Créer `frontend/e2e/fixtures.ts` :
```typescript
import { test as base } from '@playwright/test';

interface TestFixtures {
  adminUser: TestUser;
  regularUser: TestUser;
  testTeam: TestTeam;
  testMap: TestMap;
}

export const test = base.extend<TestFixtures>({
  adminUser: async ({ }, use) => {
    const user = await seedUser('e2e-admin', { isAdmin: true });
    await use(user);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [user.discordId]);
  },
  regularUser: async ({ }, use) => {
    const user = await seedUser('e2e-user-' + Date.now(), { isAdmin: false });
    await use(user);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [user.discordId]);
  },
  testTeam: async ({ adminUser }, use) => {
    const team = await seedTeam('E2E Team ' + Date.now(), adminUser.id);
    await use(team);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  },
  testMap: async ({ }, use) => {
    const map = await seedMap('E2E-Map-' + Date.now());
    await use(map);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
  },
});

export { expect } from '@playwright/test';
```

#### 3. Pipeline CI (GitHub Actions)

Créer `.github/workflows/ci.yml` :

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.26'
  NODE_VERSION: '20'
  JWT_SECRET: 'test-secret-key-for-ci-at-least-32-characters'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
        working-directory: frontend
      - run: npm run lint
        working-directory: frontend

  unit-backend:
    name: Unit Tests (Backend)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      - run: go test ./... -v -count=1 -short
      - run: go test ./... -coverprofile=coverage.out -covermode=atomic
      - uses: actions/upload-artifact@v4
        with:
          name: go-coverage
          path: coverage.out

  unit-frontend:
    name: Unit Tests (Frontend)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
        working-directory: frontend
      - run: npx vitest run --coverage
        working-directory: frontend

  integration-backend:
    name: Integration Tests (Backend)
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 5s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      - run: go test --tags=integration ./... -v -count=1 -timeout 300s
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable

  e2e:
    name: E2E Tests (Playwright)
    runs-on: ubuntu-latest
    timeout-minutes: 30
    services:
      discord-mock:
        image: wiremock/wiremock:latest
        ports:
          - 8099:8080
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
        working-directory: frontend
      - name: Install Playwright
        run: npx playwright install chromium
        working-directory: frontend
      - name: Build frontend
        run: npm run build
        working-directory: frontend
      - name: Configure WireMock mappings
        run: |
          mkdir -p dev/wiremock/mappings
          # Mappings injectés via checkpoint CI ou via script
      - name: Start Docker Compose
        run: |
          docker compose up -d
          # Config DISCORD_API_BASE_URL pointe vers le service discord-mock
      - name: Wait for app
        run: |
          for i in {1..30}; do
            curl -s http://localhost:8080/health && break
            sleep 2
          done
      - name: Run E2E tests
        run: npx playwright test
        working-directory: frontend
        env:
          DISCORD_API_BASE_URL: http://localhost:8099
      - name: Upload E2E report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-report
          path: frontend/e2e-report/
      - name: Upload E2E results JSON
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-results
          path: frontend/e2e-results.json
      - name: Docker Compose logs
        if: failure()
        run: docker compose logs
      - name: Cleanup
        if: always()
        run: docker compose down
```

**Workflow principes :**
- Jobs `lint`, `unit-backend`, `unit-frontend` s'exécutent en parallèle
- `integration-backend` commence après les jobs unitaires
- `e2e` commence après tous les autres jobs
- Upload d'artifacts : coverage, e2e-report, screenshots en cas d'échec
- Cache Go modules et node_modules

#### 4. Mock Discord OAuth avec WireMock

Ajouter un service WireMock dans `docker-compose.yml` pour simuler l'API Discord OAuth :

```yaml
services:
  # ... services existants (db, app)
  
  discord-mock:
    image: wiremock/wiremock:latest
    ports:
      - "8099:8080"
    volumes:
      - ./dev/wiremock:/home/wiremock
    command: ["--global-response-templating", "--verbose"]
```

Créer les mappings WireMock dans `dev/wiremock/mappings/` :

**`dev/wiremock/mappings/discord-oauth-token.json`** :
```json
{
  "request": {
    "method": "POST",
    "urlPath": "/api/oauth2/token"
  },
  "response": {
    "status": 200,
    "jsonBody": {
      "access_token": "mock-access-token",
      "token_type": "Bearer",
      "expires_in": 604800,
      "refresh_token": "mock-refresh-token",
      "scope": "identify"
    },
    "headers": {
      "Content-Type": "application/json"
    }
  }
}
```

**`dev/wiremock/mappings/discord-user-info.json`** :
```json
{
  "request": {
    "method": "GET",
    "urlPath": "/api/users/@me"
  },
  "response": {
    "status": 200,
    "jsonBody": {
      "id": "{{request.query.discord_id}}",
      "username": "E2E Test User",
      "discriminator": "0",
      "avatar": null
    },
    "headers": {
      "Content-Type": "application/json"
    }
  }
}
```

**`dev/wiremock/mappings/discord-down.json`** (pour les tests d'erreur) :
```json
{
  "request": {
    "method": "GET",
    "urlPath": "/api/users/@me",
    "headers": {
      "X-Simulate-Error": { "equalTo": "503" }
    }
  },
  "response": {
    "status": 503,
    "body": "Service Unavailable"
  }
}
```

Configurer l'application pour pointer vers WireMock en mode E2E :
- Variable d'env `DISCORD_API_BASE_URL=http://discord-mock:8080` dans Docker Compose
- Le serveur Go utilise cette URL pour les appels Discord OAuth

#### 5. Migration complète - Remplacement des scripts shell

Les scripts shell existants sont supprimés et remplacés par Playwright :

- `scripts/test-e2e.sh` : **Supprimer** (entièrement remplacé par `npm run test:e2e`)
- `scripts/test.sh` : Mettre à jour pour inclure `npm run test:e2e` après les tests Go

Nouveau `scripts/test.sh` :
```bash
#!/bin/bash
set -e

echo "=== Unit Tests ==="
go test ./...

echo "=== Integration Tests ==="
go test --tags=integration ./...

echo "=== Frontend Unit Tests ==="
cd frontend && npx vitest run

echo "=== E2E Tests ==="
npx playwright test

echo "=== All tests passed ==="
```

#### 6. Frontend package.json - scripts E2E

Ajouter dans `frontend/package.json` :
```json
{
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:debug": "playwright test --debug",
    "test:e2e:report": "playwright show-report e2e-report"
  }
}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/playwright.config.ts` | Créer | Configuration Playwright |
| `frontend/e2e/global-setup.ts` | Créer | Setup global E2E (Docker Compose) |
| `frontend/e2e/global-teardown.ts` | Créer | Teardown global E2E |
| `frontend/e2e/helpers/auth.ts` | Créer | Helpers d'authentification pour tests |
| `frontend/e2e/helpers/db.ts` | Créer | Helpers base de données pour seed |
| `frontend/e2e/helpers/api.ts` | Créer | Helpers API typés |
| `frontend/e2e/fixtures.ts` | Créer | Fixtures Playwright partagées |
| `frontend/e2e/tsconfig.json` | Créer | TypeScript config pour e2e |
| `.github/workflows/ci.yml` | Créer | Pipeline CI complet |
| `frontend/package.json` | Modifier | Ajouter scripts test:e2e |
| `scripts/test-e2e.sh` | Supprimer | Remplacé par Playwright |
| `scripts/test.sh` | Modifier | Ajouter frontend + E2E tests |
| `docker-compose.yml` | Modifier | Ajouter service discord-mock (WireMock) |
| `dev/wiremock/mappings/discord-oauth-token.json` | Créer | Mapping WireMock token OAuth |
| `dev/wiremock/mappings/discord-user-info.json` | Créer | Mapping WireMock user info |
| `dev/wiremock/mappings/discord-down.json` | Créer | Mapping pour simulation erreur Discord |

### Contraintes techniques
- **Playwright** : Version latest, navigateur chromium uniquement (rapidité)
- **Pas de dépendance externe** : Les helpers E2E doivent seed/clean les données directement en DB (pas d'appel API pour le setup)
- **Idempotence** : Chaque test doit créer ses propres données et les nettoyer
- **Isolation** : Les fixtures utilisent des IDs uniques (timestamp) pour éviter les collisions
- **Sécurité** : JWT_SECRET en variable d'environnement CI, jamais en dur dans le code
- **Reports** : Toujours uploader les artifacts même en cas d'échec (screenshots pour debug)
- **Timeout global** : 60s par test, 30min pour le job E2E
- **Migration complète** : `scripts/test-e2e.sh` est supprimé, remplacé par `playwright test`
- **WireMock** : L'API Discord est mockée via un container WireMock dans Docker Compose

### Tests à implémenter

Aucun test E2E fonctionnel dans cette tâche (uniquement l'infrastructure). Voir Tâches #003 à #005.

### Documentation

- `README.md` : Ajouter la section "Tests" pointant vers `docs/testing-strategy.md`
- `docs/testing-strategy.md` : Mise à jour si nécessaire (section CI + E2E setup)
