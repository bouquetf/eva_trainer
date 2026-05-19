# Tâche #003 - Story #025 : Tests automatisés

## Objectif
Implémenter les tests automatisés pour valider le comportement adaptatif de RevisionMap : test unitaire de la logique de calcul d'échelle, test d'intégration du rendu avec différentes tailles de conteneur, et test E2E de vérification visuelle.

## Contexte
- Story #025 : `docs/stories/story-025.md`
- Dépend de : Tâches #001 et #002 (les tests ne peuvent passer que si l'implémentation est terminée)
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter les tests décrits ci-dessous.

### Description fonctionnelle

Trois niveaux de tests :

1. **Test unitaire** : Tester la logique mathématique de calcul d'échelle isolément
2. **Test d'intégration** : Tester que le composant `RevisionMap` se rend correctement avec différentes tailles de conteneur
3. **Test E2E** : Vérifier que la carte tient entièrement dans le viewport sans scroll

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/RevisionMap.test.tsx` | Modifier | Remplacer le test placeholder par de vrais tests |
| `frontend/e2e/revision/adaptive.spec.ts` | Créer | Nouveau test E2E pour le comportement adaptatif |

### Tests à implémenter

#### Tests unitaires (frontend/src/components/RevisionMap.test.tsx)

Puisque la logique de scale est encapsulée dans le composant et utilise `useEffect` + `ResizeObserver` (API native), nous allons utiliser une approche de **rendu contrôlé** en mockant les dimensions du conteneur.

**Approche recommandée :** Extraire la formule de calcul d'échelle dans une fonction pure, OU tester via le rendu avec des dimensions mockées.

**Option A (recommandée) :** Extraire une fonction utilitaire :

```typescript
// Dans un nouveau fichier ou dans RevisionMap.tsx
export function calcAdaptiveScale(
  containerWidth: number,
  containerHeight: number,
  imageWidth: number,
  imageHeight: number
): number {
  if (containerWidth <= 0 || containerHeight <= 0 || imageWidth <= 0 || imageHeight <= 0) {
    return 1
  }
  return Math.min(containerWidth / imageWidth, containerHeight / imageHeight)
}
```

**Scénarios :**

| # | Scénario | containerW | containerH | imageW | imageH | Attendu |
|---|----------|-----------|-----------|-------|-------|--------|
| 1 | Container plus large que haut (16:9), image carrée | 1600 | 900 | 800 | 800 | min(2.0, 1.125) = 1.125 |
| 2 | Container plus haut que large (portrait), image paysage | 800 | 1200 | 1600 | 900 | min(0.5, 1.333) = 0.5 |
| 3 | Container identique à l'image | 800 | 600 | 800 | 600 | min(1.0, 1.0) = 1.0 |
| 4 | Container très petit | 100 | 100 | 800 | 600 | min(0.125, 0.167) = 0.125 |
| 5 | Container beaucoup plus grand que l'image | 3000 | 2000 | 800 | 600 | min(3.75, 3.333) = 3.333 |
| 6 | Container avec hauteur nulle | 800 | 0 | 800 | 600 | 1 (valeur par défaut, pas de division par zéro) |
| 7 | Image avec hauteur nulle | 800 | 600 | 800 | 0 | 1 (valeur par défaut, pas de division par zéro) |

Si tu choisis l'Option A, ces tests vont dans un `describe('calcAdaptiveScale')`.

**Option B :** Tester via le rendu du composant en mockant `containerRef.clientWidth` et `containerRef.clientHeight` :
- Mock `ResizeObserver` qui appelle `updateScale` avec des dimensions contrôlées
- Vérifier que le `Stage` (react-konva) a les bonnes dimensions

Cette approche est plus complexe et plus proche de l'intégration. À utiliser si l'extraction de la fonction pure n'est pas souhaitée.

**Critères :**
- Minimum 2 scénarios de test
- Coverage de la formule mathématique
- Test des cas limites (division par zéro)

#### Test d'intégration

Dans le même fichier `RevisionMap.test.tsx` :

- Mock de `useImage` pour retourner une fausse image avec des dimensions connues
- Mock de `ResizeObserver` pour contrôler les dimensions du conteneur
- Rendu du composant avec des `polygons` simulés
- **Scénario 1 :** Vérifier que le composant rend bien un `Stage` react-konva
- **Scénario 2 :** Vérifier que la carte s'affiche sans erreur avec des données valides

**Note :** react-konva nécessite un environnement Canvas (node-canvas ou JSDOM). Si l'environnement de test actuel ne supporte pas Konva, limite-toi au test unitaire pur (Option A).

#### Test E2E (frontend/e2e/revision/adaptive.spec.ts)

Suivre le pattern des tests E2E existants dans `frontend/e2e/revision/` :

```typescript
import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { startRevisionSession, makeRevisionZone } from '../helpers/revision';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Carte adaptative à l\'écran', () => {
  test('3.1: La carte tient entièrement dans le viewport sans scroll', async ({ page, adminUser }) => {
    // 1. Créer une team, une map et des zones
    // 2. Lancer une session de révision
    // 3. Vérifier que la page ne scroll pas
    // 4. Vérifier que la carte est entièrement visible
  });
});
```

**Détails du scénario E2E :**

1. **Setup :** Créer team + map + zones de révision
2. **Action :** Naviguer vers l'écran de révision active (via `startRevisionSession`)
3. **Assertions :**
   - Vérifier `await expect(page.locator('html')).not.toHaveClass(/scroll/i)` ou équivalent
   - Vérifier que l'input est visible (la carte ne le cache pas)
   - Vérifier que le conteneur de la carte (`revision-canvas-wrapper`) a une hauteur > 0 et est visible
   - Vérifier que le stage react-konva est rendu (présence d'un élément canvas dans le wrapper)
4. **Teardown :** Nettoyer les données via `queryDB`

**Alternative :** Si la détection de scroll est difficile, tu peux vérifier que `.revision-active-screen .canvas-area` n'a pas de scrollbar visible :
```typescript
const canvasArea = page.locator('.revision-active-screen .canvas-area')
const hasScrollbar = await canvasArea.evaluate(el => el.scrollHeight > el.clientHeight)
expect(hasScrollbar).toBe(false)
```

### Contraintes techniques

- **Framework de test** : Vitest pour les tests unitaires/integration, Playwright pour les E2E (suivre les patterns existants)
- **Mock de ResizeObserver** : Dans les tests Vitest, `ResizeObserver` n'est pas disponible — utiliser un mock :
  ```typescript
  global.ResizeObserver = class FakeResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  ```
- **Mock de useImage** : La bibliothèque `use-image` doit être mockée pour les tests d'intégration
- **E2E** : Utiliser le helper `startRevisionSession` qui mocke déjà l'API des zones (cf. `frontend/e2e/helpers/revision.ts`)

### Documentation

Aucune documentation à créer ou modifier. Les tests constituent la documentation du comportement attendu.
