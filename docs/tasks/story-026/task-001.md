# Tâche #001 - Story #026 : Titre de l'application "Eva-Trainer"

## Objectif
Modifier le fichier `index.html` du frontend pour remplacer le titre "frontend" par "Eva-Trainer", et ajouter un test E2E vérifiant que le titre est correct sur toutes les pages.

## Contexte
- Story #026 : `docs/stories/story-026.md`
- Dépend de : Rien
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Le titre de l'onglet du navigateur doit afficher "Eva-Trainer" au lieu de "frontend".

**Changement :**
- La balise `<title>` dans `index.html` passe de `frontend` à `Eva-Trainer`
- Le changement est immédiat : après rechargement de n'importe quelle page, `document.title` doit valoir `Eva-Trainer`
- Le titre s'applique à toutes les pages de l'application (pas de SPA dynamic title, juste le `<title>` statique)

**Cas nominaux :**
- Navigation sur `/login` → `document.title === "Eva-Trainer"`
- Navigation sur `/teams` → `document.title === "Eva-Trainer"`
- Navigation sur une page d'édition → `document.title === "Eva-Trainer"`
- Navigation sur une page de révision → `document.title === "Eva-Trainer"`

**Cas limites :**
- Chargement initial de l'application (avant hydration React) → le titre est déjà "Eva-Trainer" (défini dans le HTML statique)
- Après une erreur 404 ou une redirection → le titre reste "Eva-Trainer"

**Gestion d'erreurs :**
- Aucune gestion d'erreur nécessaire (modification statique)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/index.html` | Modifier | Remplacer `<title>frontend</title>` par `<title>Eva-Trainer</title>` |
| `frontend/e2e/auth/home-redirect.spec.ts` | Modifier | Ajouter un test de vérification du titre après redirection |
| OU `frontend/e2e/general/app-title.spec.ts` | Créer | Nouveau fichier de test E2E (selon préférence : dans un répertoire général) |

### Contraintes techniques
- **Fichier concerné** : `frontend/index.html` ligne 7
- **Changement minimal** : une seule ligne à modifier (`<title>frontend</title>` → `<title>Eva-Trainer</title>`)
- **Pas de package supplémentaire** : le titre est statique, pas besoin de react-helmet ou équivalent
- **Test E2E** : Utiliser Playwright avec les fixtures existantes

### Tests à implémenter

#### Test E2E

Le test vérifie que `document.title` contient "Eva-Trainer" sur n'importe quelle page après navigation.

Approche recommandée : ajouter un test dans un nouveau fichier dédié OU dans un fichier existant (ex: `auth/home-redirect.spec.ts`).

**Option A (recommandée) : Créer `frontend/e2e/general/app-title.spec.ts`**

```typescript
import { test, expect } from '../fixtures';

test.describe('Feature: App Title', () => {
  test('1.1: Le titre de l\'onglet affiche "Eva-Trainer" sur la page de login', async ({ page }) => {
    await page.goto('/login');
    await expect(page).toHaveTitle('Eva-Trainer');
  });

  test('1.2: Le titre de l\'onglet affiche "Eva-Trainer" après redirection vers /teams', async ({ page, regularUser }) => {
    await page.goto('/login');
    await page.getByLabel('Discord ID').fill(regularUser.discordId);
    await page.getByLabel('Token').fill(regularUser.token);
    await page.getByRole('button', { name: /se connecter|login/i }).click();
    await page.waitForURL('/teams');
    await expect(page).toHaveTitle('Eva-Trainer');
  });
});
```

**Option B : Ajouter un test existant** dans `frontend/e2e/auth/home-redirect.spec.ts` :
- Après le test 1.3, ajouter `test('1.4: Le titre de l\'onglet affiche "Eva-Trainer"')`

À toi de choisir l'approche. L'important est que la couverture fonctionnelle soit assurée.

### Documentation
Aucune documentation à créer ou modifier (simple modification de configuration).

### Vérification
1. Build le frontend : `npm run build` (dans `frontend/`)
2. Lancer les tests E2E : `npm run test:e2e` (à partir du root du projet)
3. Vérifier que le test de titre passe
4. Vérifier manuellement que l'onglet du navigateur affiche "Eva-Trainer"
