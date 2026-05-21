# Tâche #004 - Story #031 : Tests E2E pour la réorganisation de l'interface d'équipe

## Objectif
Écrire les tests E2E Playwright couvrant les 4 critères d'acceptation liés à la réorganisation de l'interface d'équipe.

## Contexte
- Story #031 : `docs/stories/story-031.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un fichier de test E2E `interface.spec.ts` dans le répertoire `frontend/e2e/teams/` pour valider les changements d'interface de la Story #031.

**Cas nominaux (4 tests) :**

1. **Test : Le bloc "Gestion des rôles" n'est plus affiché**
   - Se connecter en tant que capitaine
   - Naviguer vers `/teams/{testTeam.id}`
   - Vérifier que le texte "Gestion des rôles" n'apparaît **nulle part** sur la page
   - Vérifier que le texte "En tant que Capitaine, vous pouvez" n'apparaît pas non plus

2. **Test : "Demandes d'adhésion" apparaît dans le bloc "Invitations"**
   - Se connecter en tant que capitaine
   - Naviguer vers `/teams/{testTeam.id}`
   - Vérifier que "Demandes d'adhésion" n'apparaît PAS comme un bloc `.section` séparé
   - Vérifier que "Demandes d'adhésion" apparaît dans la page (c'est une sous-section d'Invitations)
   - Optionnel : utiliser la structure DOM pour vérifier qu'il est à l'intérieur du bloc Invitations

3. **Test : "Membres" apparaît après "Cartes" dans l'ordre du DOM**
   - Se connecter en tant que capitaine
   - Naviguer vers `/teams/{testTeam.id}`
   - Récupérer l'ordre des sections visibles
   - Vérifier que "Cartes" précède "Membres" dans l'ordre du DOM
   - Utiliser `page.locator('.section-title, .collapsible-section .header')` ou similaire pour obtenir l'ordre

4. **Test : L'accordéon fonctionne (ouverture/fermeture)**
   - Se connecter en tant que capitaine
   - Naviguer vers `/teams/{testTeam.id}`
   - Vérifier que la section "Membres" est fermée par défaut (contenu non visible)
   - Cliquer sur l'en-tête de la section "Membres"
   - Vérifier que le contenu des membres devient visible
   - Cliquer à nouveau sur l'en-tête
   - Vérifier que le contenu redevient caché

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/interface.spec.ts` | Créer | Nouveau fichier de test E2E |

### Structure du test

```typescript
import { test, expect } from '../fixtures'
import { loginAsUser } from '../helpers/auth'

test.describe('Interface d\'équipe réorganisée', () => {

  test('031.1: Le bloc "Gestion des rôles" est supprimé', async ({ page, adminUser, testTeam }) => {
    // Arrange
    await page.goto('/login')
    await loginAsUser(page, adminUser)
    
    // Act
    await page.goto(`/teams/${testTeam.id}`)
    
    // Assert
    await expect(page.getByText('Gestion des rôles')).not.toBeVisible()
    await expect(page.getByText(/En tant que Capitaine/)).not.toBeVisible()
  })

  test('031.2: "Demandes d\'adhésion" est intégré dans "Invitations"', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login')
    await loginAsUser(page, adminUser)
    await page.goto(`/teams/${testTeam.id}`)

    // Trouver le bloc Invitations (via le titre)
    const invitationsBlock = page.locator('.collapsible-section').filter({ hasText: /^Invitations/ })
    await expect(invitationsBlock).toBeVisible()

    // Vérifier que "Demandes d'adhésion" est un enfant du bloc Invitations
    await expect(invitationsBlock.locator('h3:has-text("Demandes d\'adhésion")')).toBeVisible()

    // Vérifier qu'il n'y a PAS de bloc section autonome avec "Demandes d'adhésion"
    await expect(page.locator('.section').filter({ hasText: /^Demandes d'adhésion/ })).toHaveCount(0)
  })

  test('031.3: Les sections sont dans l\'ordre : Informations, Cartes, Membres', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login')
    await loginAsUser(page, adminUser)
    await page.goto(`/teams/${testTeam.id}`)

    // Récupérer tous les titres de section dans l'ordre du DOM
    // Les sections Informations et Cartes utilisent h2.section-title
    // Les sections Membres et suivantes utilisent .title dans CollapsibleSection
    const sectionTitles = page.locator('h2.section-title, .title')
    const titles = await sectionTitles.allTextContents()
    
    const cartesIndex = titles.findIndex(t => t.trim().startsWith('Cartes'))
    const membresIndex = titles.findIndex(t => t.trim().startsWith('Membres'))
    
    expect(cartesIndex).toBeGreaterThanOrEqual(0)
    expect(membresIndex).toBeGreaterThanOrEqual(0)
    expect(cartesIndex).toBeLessThan(membresIndex)
  })

  test('031.4: L\'accordéon ouvre et ferme les sections', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login')
    await loginAsUser(page, adminUser)
    await page.goto(`/teams/${testTeam.id}`)

    // Trouver l'en-tête de la section Membres
    const membresSection = page.locator('.collapsible-section').filter({ hasText: 'Membres' }).first()
    const membresHeader = membresSection.locator('.header')

    // Vérifier que Membres est fermé par défaut (aria-expanded="false")
    await expect(membresHeader).toHaveAttribute('aria-expanded', 'false')

    // Cliquer sur l'en-tête pour ouvrir
    await membresHeader.click()

    // Vérifier que le contenu est visible (aria-expanded="true")
    await expect(membresHeader).toHaveAttribute('aria-expanded', 'true')

    // Cliquer à nouveau pour fermer
    await membresHeader.click()

    // Vérifier que le contenu est masqué à nouveau
    await expect(membresHeader).toHaveAttribute('aria-expanded', 'false')
  })
}
```

**Notes sur l'implémentation :**
- Adapter les assertions en fonction de la structure DOM réelle du composant `CollapsibleSection`
- Utiliser les bons sélecteurs : `.collapsible-section` pour le conteneur, `.header` pour la zone cliquable, `.content` pour le panneau de contenu
- Pour vérifier l'état ouvert/fermé, on peut vérifier la présence de classes CSS, l'attribut `aria-expanded`, ou la visibilité du contenu

### Contraintes techniques
- **Framework** : Playwright (déjà configuré dans le projet)
- **Fixtures** : Utiliser `adminUser` et `testTeam` depuis `../fixtures` (déjà disponibles)
- **Helpers** : Utiliser `loginAsUser` depuis `../helpers/auth` (déjà disponible)
- **Conventions** : Suivre les conventions des tests E2E existants dans `frontend/e2e/teams/` (mêmes helpers, même structure de test)
- **Data cleanup** : Les fixtures gèrent déjà le cleanup automatique (`adminUser` et `testTeam` sont nettoyés après chaque test)
- **Indépendance** : Chaque test doit être indépendant et ne pas dépendre d'un autre test

### Documentation

Aucune documentation à créer pour cette tâche.
