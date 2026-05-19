# Tâche #003 - Story #027 : Tests automatisés pour la fonctionnalité copier/coller

## Objectif
Implémenter les tests unitaires pour la fonction `duplicatePolygonsWithOffset`, les tests d'intégration pour les handlers copier/coller dans l'éditeur, et les tests E2E pour le parcours complet.

## Contexte
- Story #027 : `docs/stories/story-027.md`
- Dépend de : Tâche #001, Tâche #002
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Tests unitaires (transforms) :**
- Tester `duplicatePolygonsWithOffset` avec :
  - Un seul polygone sélectionné → vérifier le deep clone et le décalage de **30px**
  - Plusieurs polygones sélectionnés → vérifier que tous sont dupliqués
  - Aucun polygone sélectionné (Set vide) → retourner un tableau vide
  - Offset personnalisé → vérifier que les points sont décalés de la valeur fournie
  - Vérifier que l'original n'est pas muté (immutabilité)
  - Vérifier que les nouveaux polygones ont `id: 0`

**Tests d'intégration (MapEditor) :**
- Tester le handler `handleCopy` :
  - Copie les polygones sélectionnés dans le clipboard avec leur `zone_id` source
  - Ne copie rien si `selectedPolygonIndexes` est vide
  - Remplace le clipboard si on recopie (et met à jour `clipboardSourceZoneId`)
- Tester le handler `handlePaste` :
  - Ajoute les polygones du clipboard à la zone d'origine avec décalage de 30px
  - Sélectionne automatiquement les nouveaux polygones
  - Ne fait rien si le clipboard est vide
  - Ne fait rien si aucune zone n'est en cours d'édition
  - Passe à la zone source si l'utilisateur a changé de zone (`clipboardSourceZoneId !== editingZone.id`)

**Tests E2E (Playwright) :**
- Créer un fichier `frontend/e2e/editor/copy-paste.spec.ts`
- Scénario E2E :
  1. Se connecter, naviguer vers l'éditeur d'une équipe/carte
  2. Créer une nouvelle zone via l'UI (bouton de création dans le panneau latéral)
  3. Ajouter un polygone en cliquant sur le canvas (mode dessin)
  4. Sélectionner le polygone
  5. Appuyer sur Ctrl+C
  6. Appuyer sur Ctrl+V
  7. Vérifier qu'un nouveau polygone apparaît dans la même zone
  8. Vérifier que le nouveau polygone est décalé (coordonnées différentes)
- Pas de helper de seed DB : la zone est créée entièrement via l'UI Playwright

**Cas nominaux :**
- Un polygone sélectionné → duplication avec offset de 30px
- Plusieurs polygones sélectionnés → tous dupliqués

**Cas limites :**
- Aucune sélection → copier ne fait rien
- Clipboard vide → coller ne fait rien
- Polygone à 3 points → dupliqué correctement

**Gestion d'erreurs :**
- Les tests doivent couvrir le cas où `zone` est `null`/`undefined` (pour le handler paste)
- Les tests doivent couvrir le cas où la zone source n'existe plus (supprimée entre copie et collage)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/geometry/transforms.test.ts` | Modifier | Ajouter les tests pour `duplicatePolygonsWithOffset` |
| `frontend/src/components/MapEditor/MapEditor.test.tsx` | Modifier | Ajouter les tests d'intégration pour `handleCopy`/`handlePaste` |
| `frontend/e2e/editor/copy-paste.spec.ts` | Créer | Test E2E complet copier/coller |

### Signatures

```typescript
// Dans transforms.test.ts

describe('duplicatePolygonsWithOffset', () => {
  it('should duplicate a single selected polygon with default offset (30px)', () => { ... });
  it('should duplicate multiple selected polygons', () => { ... });
  it('should return empty array when no polygons selected', () => { ... });
  it('should use custom offset when provided', () => { ... });
  it('should not mutate the original zone', () => { ... });
  it('should set id to 0 on duplicated polygons', () => { ... });
  it('should keep zone_id from original zone', () => { ... });
});
```

### Contraintes techniques

#### Tests unitaires
- **Framework** : Vitest (déjà configuré)
- **Pattern** : Suivre exactement le pattern des tests existants dans `transforms.test.ts` : utiliser les fonctions `mockPoint`, `mockPolygon`, `mockZone` helpers déjà définies ou en définir de nouvelles si nécessaire
- **Import** : `PASTE_OFFSET` et `duplicatePolygonsWithOffset` depuis `../transforms`
- **Offset** : Tous les tests par défaut doivent vérifier l'offset à 30px. Utiliser `PASTE_OFFSET` plutôt que la valeur littérale `30`.

#### Tests d'intégration
- **Framework** : Vitest + React Testing Library
- **Pattern** : Suivre le pattern de `MapEditor.test.tsx`
- **Note** : L'éditeur nécessite beaucoup de mocks (API, auth store, etc.). Les tests des handlers doivent être unitaires (tester directement `handleCopy`/`handlePaste` en simulant l'état React) plutôt que des tests de rendu complets.

#### Tests E2E
- **Framework** : Playwright
- **Setup** : Utiliser `loginAsUser()` pour l'authentification, `seedTeam()` pour créer une équipe et une carte (mais pas de zone)
- **Création de zone** : La zone est créée **via l'UI Playwright** (pas de helper DB). Le test doit :
  1. Naviguer vers l'éditeur
  2. Cliquer sur le bouton de création de zone dans le panneau latéral
  3. Cliquer sur le canvas pour ajouter des points (mode dessin)
  4. Compléter le polygone (clic droit → "Terminer la forme")
- **Sélecteurs** : Utiliser des sélecteurs de données stables (ex: data-testid si disponible, sinon classes CSS, textes, rôles ARIA)

### Tests à implémenter

#### Tests unitaires

**Fichier** : `frontend/src/lib/geometry/transforms.test.ts`

**Scénario 1 – Dupliquer un seul polygone :**
- Données : zone avec 1 polygone de 3 points, selectedIndexes = new Set([0])
- Résultat attendu : 1 polygone, points décalés de PASTE_OFFSET (30), id = 0, zone_id = zone.id

**Scénario 2 – Dupliquer plusieurs polygones :**
- Données : zone avec 3 polygones, selectedIndexes = new Set([0, 2])
- Résultat attendu : 2 polygones, tous décalés de PASTE_OFFSET

**Scénario 3 – Aucun polygone sélectionné :**
- Données : zone avec 2 polygones, selectedIndexes = new Set()
- Résultat attendu : []

**Scénario 4 – Offset personnalisé :**
- Données : zone avec 1 polygone, selectedIndexes = new Set([0]), offsetX = 50, offsetY = 30
- Résultat attendu : points décalés de (50, 30)

**Scénario 5 – Immutabilité :**
- Données : zone avec 1 polygone
- Résultat attendu : modifier le résultat ne modifie pas l'original

**Scénario 6 – id = 0 :**
- Données : zone avec 1 polygone ayant id = 42
- Résultat attendu : polygone dupliqué a id = 0

**Scénario 7 – zone_id conservé :**
- Données : zone avec zone_id = 15, 1 polygone, selectedIndexes = new Set([0])
- Résultat attendu : polygone dupliqué a zone_id = 15

#### Tests d'intégration

**Fichier** : `frontend/src/components/MapEditor/MapEditor.test.tsx`

Ajouter des tests pour les handlers :
1. `handleCopy` copie les polygones sélectionnés dans `clipboardPolygons`
2. `handleCopy` stocke `clipboardSourceZoneId = editingZone.id`
3. `handlePaste` ajoute à la zone d'origine (pas la zone en cours d'édition si différente)

#### Tests E2E

**Fichier** : `frontend/e2e/editor/copy-paste.spec.ts`

```typescript
import { test, expect } from '@playwright/test';
import { loginAsUser } from '../helpers/auth';
import { seedTeam } from '../helpers/db';

test.describe('Copy/Paste in Editor', () => {
  test('should copy and paste a polygon in the same zone', async ({ page }) => {
    // Arrange : login + seed team + navigate to editor
    // La zone est créée via l'UI :
    //   1. Cliquer "Nouvelle zone" dans ZoneList
    //   2. Saisir un nom, valider
    //   3. Cliquer sur le canvas 3+ fois pour dessiner un polygone
    //   4. Clic droit → "Terminer la forme"
    // Act : Ctrl+C → Ctrl+V
    // Assert : un nouveau polygone apparaît dans la même zone
  });

  test('should copy and paste from context menu', async ({ page }) => {
    // Arrange : zone with polygon (via UI)
    // Act : right-click → "Copier", right-click → "Coller"
    // Assert : polygon duplicated
  });
});
```

### Documentation
- Aucune documentation à créer

### Dépendances
- Cette tâche dépend de la Tâche #001 et de la Tâche #002
