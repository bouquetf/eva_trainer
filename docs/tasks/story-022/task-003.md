# Tâche #003 - Story #022 : Mettre à jour les tests E2E pour la nouvelle redirection

## Objectif
Mettre à jour les tests E2E existants qui dépendent de l'ancien comportement de la page d'accueil, et ajouter un test dédié à la redirection `/` → `/teams`.

## Contexte
- Story #022 : `docs/stories/story-022.md`
- Dépend de : Tâche #001, Tâche #002 (les modifications doivent être appliquées d'abord)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Tests à modifier :**

1. **`frontend/e2e/auth/login.spec.ts` - test 1.2** : Après le callback OAuth, la redirection par défaut est maintenant `/teams` au lieu de `/`. Il faut :
   - Changer `await page.waitForURL('/')` en `await page.waitForURL('/teams')`
   - Changer l'assertion `await expect(page.getByText(regularUser.discordId)).toBeVisible()` — ce texte n'est plus affiché sur `/teams`. Remplacer par une assertion pertinente pour la page TeamsList, par exemple un titre "Mes équipes" ou vérifier que l'URL est bien `/teams`.

2. **`frontend/e2e/auth/logout.spec.ts` - tests 2.1 et 2.2** : Ces tests utilisent `page.goto('/')` après `loginAsUser()` puis cherchent le bouton de déconnexion. Comme `/` redirige maintenant vers `/teams`, adapter les tests :
   - Test 2.1 : Naviguer vers `/teams` au lieu de `/` après login, trouver le bouton de déconnexion (dans la barre latérale de AppLayout), cliquer dessus, vérifier la redirection vers `/login`
   - Test 2.2 : Même pattern

3. **Nouveau test `frontend/e2e/auth/home-redirect.spec.ts`** (créer) :
   - Test 1 : Utilisateur authentifié navigateur vers `/` → redirigé vers `/teams` (vérifier l'URL)
   - Test 2 : Utilisateur non authentifié navigue vers `/` → redirigé vers `/login` (AuthGuard)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/auth/login.spec.ts` | Modifier | Test 1.2 : adapter attentes URL et contenu |
| `frontend/e2e/auth/logout.spec.ts` | Modifier | Tests 2.1, 2.2 : naviguer vers `/teams` au lieu de `/` |
| `frontend/e2e/auth/home-redirect.spec.ts` | Créer | Nouveau test dédié à la redirection `/` → `/teams` |

### Détails des modifications

#### 1. `frontend/e2e/auth/login.spec.ts` — Test 1.2

```typescript
test('1.2: Callback après authentification Discord', async ({ page, regularUser }) => {
  const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
  const token = await generateJWT(regularUser.id, secret);

  await page.goto(`/callback#access_token=${token}&refresh_token=e2e-refresh-token`);

  // Changé : on attend /teams au lieu de /
  await page.waitForURL('/teams');

  // Changé : on vérifie la présence de l'URL /teams plutôt qu'un texte spécifique
  expect(page.url()).toContain('/teams');
});
```

#### 2. `frontend/e2e/auth/logout.spec.ts` — Tests 2.1 et 2.2

**Test 2.1** :
```typescript
test('2.1: Déconnexion volontaire', async ({ page, regularUser }) => {
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  // Changé : naviguer vers /teams au lieu de /
  await page.goto('/teams');

  // Le bouton de déconnexion est maintenant dans la barre latérale (AppLayout),
  // juste en dessous de "Paramètres", avec la classe .sidebar-logout
  const logoutBtn = page.getByRole('button', { name: /déconnexion/i });
  await expect(logoutBtn).toBeVisible();

  await logoutBtn.click();
  // Après logout(), window.location.href = '/' → page non authentifiée
  await page.waitForURL('/');

  // Sur la page d'accueil non authentifiée, le bouton Discord est visible
  const loginBtn = page.getByRole('button', { name: /connexion.*discord/i });
  await expect(loginBtn).toBeVisible();
});
```

**Test 2.2** :
```typescript
test('2.2: Accès après déconnexion', async ({ page, regularUser }) => {
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  // Changé : naviguer vers /teams au lieu de /
  await page.goto('/teams');

  await page.getByRole('button', { name: /déconnexion/i }).click();
  await page.waitForURL('/');

  await page.goto('/teams');
  await page.waitForURL(/\/login/);
});
```

#### 3. `frontend/e2e/auth/home-redirect.spec.ts` — Nouveau fichier

```typescript
import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Redirection page d\'accueil', () => {
  test('1.1: Utilisateur authentifié accède à / → redirigé vers /teams', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/');
    await page.waitForURL('/teams');
    expect(page.url()).toContain('/teams');
  });

  test('1.2: Utilisateur non authentifié accède à / → redirigé vers /login', async ({ page }) => {
    await page.goto('/');
    // AuthGuard redirige vers /login?redirect=/... ou /login
    await page.waitForURL(/\/login/);
    expect(page.url()).toContain('/login');
  });

  test('1.3: Admin authentifié accède à / → redirigé vers /teams', async ({ page, adminUser }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/');
    await page.waitForURL('/teams');
    expect(page.url()).toContain('/teams');
  });
});
```

### Contraintes techniques
- **Framework** : Playwright (config existante dans le projet)
- **Fixtures** : Utiliser `regularUser` et `adminUser` des fixtures existantes (`frontend/e2e/fixtures.ts`)
- **Helpers** : Utiliser `loginAsUser` de `frontend/e2e/helpers/auth.ts`
- **Exécution** : Lancer les tests avec `cd frontend && npx vitest run` (ou la commande E2E appropriée — voir AGENTS.md pour la commande exacte)

### Tests à implémenter

Les tests E2E sont décrits ci-dessus. Aucun test unitaire supplémentaire nécessaire.

### Documentation

Aucune documentation à créer.
