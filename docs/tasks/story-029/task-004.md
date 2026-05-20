# Tâche #004 - Story #029 : Tests backend et E2E pour l'indicateur d'étage

## Objectif
Implémenter les tests unitaires backend, les tests d'intégration backend, et les tests E2E Playwright pour la fonctionnalité d'indicateur d'étage.

## Contexte
- Story #029 : `docs/stories/story-029.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Tests backend

#### Tests unitaires (validation)

**Fichier** : `internal/models/zone_test.go` (ou créer si inexistant)

| # | Scénario | Données | Résultat attendu |
|---|----------|---------|-----------------|
| 1 | `validateFloor(nil)` | `nil` | `nil` (pas d'erreur) |
| 2 | `validateFloor(ptr("haut"))` | `"haut"` | `nil` |
| 3 | `validateFloor(ptr("bas"))` | `"bas"` | `nil` |
| 4 | `validateFloor(ptr(""))` | `""` | `nil` (chaîne vide = effacer) |
| 5 | `validateFloor(ptr("middle"))` | `"middle"` | `ErrInvalidFloorValue` |
| 6 | `validateFloor(ptr("HAUT"))` | `"HAUT"` | `ErrInvalidFloorValue` (case-sensitive) |

#### Tests d'intégration (API)

**Fichier** : `internal/handlers/zone_test.go` (à créer ou compléter)

| # | Scénario | Résultat attendu |
|---|----------|-----------------|
| 1 | Batch action `set_floor` avec `floor = "haut"` sur un polygone existant | Succès, vérifier en base que floor vaut "haut" |
| 2 | Batch action `set_floor` avec `floor = null` sur un polygone | Succès, vérifier que floor est NULL en base |
| 3 | Batch action `set_floor` avec `floor = "bas"` | Succès, vérifier "bas" |
| 4 | Batch action `set_floor` sur polygone inexistant | Erreur (404 ou 500 selon la convention) |
| 5 | Batch action `set_floor` avec `floor = "invalid"` | Erreur de validation |
| 6 | Créer template via publish → copier `floor` dans template_zones | Vérifier que le JSON de polygon_data contient "floor" |
| 7 | Dupliquer template → vérifier que les nouveaux polygones ont le floor préservé | Vérifier en base |

**Approche recommandée** : Suivre la structure des tests existants dans `internal/handlers/`. Utiliser `net/http/httptest` avec la base de test.

### Tests E2E

**Fichier** : `frontend/e2e/zones/floor-indicator.spec.ts`

Utiliser les helpers E2E existants (`frontend/e2e/helpers/db.ts`, `frontend/e2e/helpers/auth.ts`) pour :
- Créer une équipe, une carte, et une zone avec des polygones
- Se connecter en tant qu'éditeur
- Naviguer vers l'éditeur

#### Scénario E2E 1 : Clic droit → sélection "Haut" → indicateur visuel → rechargement

```typescript
test("clic droit sur forme → Haut → indicateur visible → persistant après rechargement", async ({ page }) => {
  // 1. Seed une équipe, une carte, et crée une zone avec un polygone
  // 2. Login en tant que membre de l'équipe
  // 3. Naviguer vers l'éditeur de la carte
  // 4. Cliquer sur le polygone pour le sélectionner
  // 5. Clic droit → le menu contextuel apparaît
  // 6. Cliquer sur "Étage > Haut"
  // 7. Vérifier que l'indicateur visuel apparaît (un élément SVG avec flèche)
  // 8. Recharger la page (F5)
  // 9. Vérifier que l'indicateur est toujours présent
});
```

**Sélecteurs** :
- Menu contextuel : `.context-menu` ou `div[style*="position: fixed"]` (voir implémentation actuelle)
- Option "Étage > Haut" : texte exact dans le menu
- Indicateur : élément SVG `<polygon>` dans le `<g>` du polygone (ou un data-attribute)

#### Scénario E2E 2 : Exclusivité Haut/Bas

```typescript
test("sélection Bas sur forme déjà en Haut → l'indicateur passe à Bas", async ({ page }) => {
  // 1. Même setup que le test 1
  // 2. Sélectionner le polygone, clic droit → "Étage > Haut"
  // 3. Vérifier indicateur Haut présent
  // 4. Clic droit à nouveau → "Étage > Bas"
  // 5. Vérifier que l'indicateur Haut a disparu et que l'indicateur Bas est présent
});
```

#### Scénario E2E 3 : Template duplication avec floor

```typescript
test("template public → l'indicateur d'étage est copié lors de la duplication", async ({ page }) => {
  // 1. Créer une zone avec floor = "haut" sur un polygone
  // 2. Publier le template (nécessite rôle captain ou cocaptain)
  // 3. Se déconnecter, se connecter en tant qu'autre équipe
  // 4. Dupliquer le template public
  // 5. Vérifier que la zone dupliquée a l'indicateur d'étage
});
```

#### Scénario E2E 4 : Effacer l'indicateur

```typescript
test("effacer supprime l'indicateur d'étage", async ({ page }) => {
  // 1. Créer une zone avec floor = "haut"
  // 2. Sélectionner le polygone, clic droit → "Étage > Effacer"
  // 3. Vérifier que l'indicateur a disparu
});
```

#### Scénario E2E 5 : Option Effacer absente quand aucun floor défini

```typescript
test("Effacer n'apparaît pas si aucun floor n'est défini", async ({ page }) => {
  // 1. Créer une zone sans floor
  // 2. Sélectionner le polygone, clic droit
  // 3. Vérifier que "Étage > Effacer" n'est PAS dans le menu
  // 4. Vérifier que "Étage > Haut" et "Étage > Bas" sont présents
});
```

### Helpers E2E à créer ou modifier

**Fichier** : `frontend/e2e/helpers/db.ts`

Ajouter un helper pour créer un polygone avec un floor spécifique :
```typescript
export async function setPolygonFloor(polygonId: number, floor: string | null): Promise<void>
```

Ou bien, modifier `seedZone` / `seedPolygon` pour accepter un paramètre `floor` optionnel.

### Contraintes techniques
- **Tests Go** : Utiliser `testing` standard, pas de framework externe. Suivre les conventions des tests existants.
- **Tests E2E** : Utiliser Playwright avec les patterns existants (`test`, `expect`, helpers d'authentification)
- **Données de seed** : Utiliser les helpers existants dans `frontend/e2e/helpers/db.ts`
- **Exclusivité** : Le test E2E #2 doit vérifier qu'un seul indicateur est visible à la fois
- **Template** : Le test E2E #3 nécessite de se connecter avec un captain pour publier, puis se déconnecter et se reconnecter avec une autre équipe

### Documentation
- Mettre à jour `docs/zone-editor.md` si elle existe, avec la section "Indicateur d'étage"
