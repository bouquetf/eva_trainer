# Tâche #005 - Story #038 : Tests E2E

## Objectif
Créer les tests End-to-End (Playwright) pour le parcours complet de saisie, consultation et modification d'un BO.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Tâche #003 (Frontend section UI BO)
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Test E2E : Saisie d'un BO complet → consultation → modification**

1. **Préparation** (helpers) :
   - Créer une ligue avec des données de test (ligue, split, équipes eva, maps)
   - Ajouter les helpers nécessaires dans `frontend/e2e/helpers/db.ts` si besoin

2. **Test : Création d'un BO** :
   - Se connecter en tant qu'archiviste
   - Naviguer vers `/archives`
   - Cliquer sur l'onglet "BOs"
   - Sélectionner une ligue (dropdown période)
   - Cliquer sur "Nouveau BO"
   - Remplir le formulaire :
     - Sélectionner equipe1 et equipe2
     - Choisir format "BO3"
     - Ajouter 1 ban (équipe + carte)
     - Ajouter 2 picks (carte + équipe facultative)
     - Ajouter 2 matchs (carte + scores)
   - Soumettre
   - Vérifier que le BO apparaît dans la liste
   - Vérifier que le vainqueur est calculé correctement

3. **Test : Consultation des détails** :
   - Cliquer sur le BO dans la liste
   - Vérifier l'affichage des bans, picks et matchs
   - Vérifier le score et le vainqueur

4. **Test : Modification d'un BO** :
   - Cliquer sur "Modifier" sur un BO existant
   - Modifier les scores d'un match
   - Ajouter un nouveau match
   - Supprimer un ban
   - Soumettre
   - Vérifier que les modifications sont visibles dans les détails
   - Vérifier que le vainqueur du BO a été recalculé si nécessaire

5. **Test : Suppression d'un BO** :
   - Cliquer sur "Supprimer"
   - Confirmer la suppression
   - Vérifier que le BO n'est plus dans la liste

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter helpers si nécessaire (createBO, createMap, etc.) |
| `frontend/e2e/archives/bo.spec.ts` | Créer | Test E2E pour le parcours BO |

### Contraintes techniques
- **Framework** : Playwright, config existante dans le projet (`frontend/playwright.config.ts`)
- **Pattern** : Suivre exactement les tests E2E existants :
  - Utiliser `test.describe` et `test` de `@playwright/test`
  - Utiliser `loginAsUser` depuis `frontend/e2e/helpers/auth.ts`
  - Utiliser les helpers DB pour préparer les données
  - Sélecteurs : data-testid, getByRole, getByText (comme dans les tests existants)
- **Helper DB** : Ajouter dans `frontend/e2e/helpers/db.ts` les fonctions :
  - `createMap(name: string)` (si pas déjà existant)
  - Utiliser `createEvaTeam()` et `createLigue()` s'ils existent déjà
  - Vérifier et réutiliser les helpers existants avant d'en créer

### Exemple de squelette

```typescript
import { test, expect } from '@playwright/test'
import { loginAsUser } from '../helpers/auth'
import { createLigue } from '../helpers/db' // adapter selon helpers existants

test.describe('BO Management', () => {
  test('create a BO with bans, picks and matches', async ({ page }) => {
    // Arrange
    await loginAsUser(page, 'archivist@test.com') // ou utiliser le bon helper
    await page.goto('/archives')

    // Sélectionner l'onglet BOs
    await page.getByRole('button', { name: 'BOs' }).click()

    // Sélectionner une ligue
    await page.getByLabel('Période').selectOption(String(ligueId))

    // Créer un nouveau BO
    await page.getByRole('button', { name: 'Nouveau BO' }).click()

    // Remplir le formulaire...
    await page.getByLabel('Équipe 1').selectOption(String(team1Id))
    await page.getByLabel('Équipe 2').selectOption(String(team2Id))
    await page.getByLabel('Format').selectOption('BO3')

    // Ajouter un ban
    await page.getByRole('button', { name: 'Ajouter un ban' }).click()
    // ... remplir les champs du ban

    // Ajouter des picks et matchs...

    // Soumettre
    await page.getByRole('button', { name: 'Enregistrer' }).click()

    // Vérifier
    await expect(page.getByText(/Team1.*vs.*Team2/)).toBeVisible()
  })
})
```
