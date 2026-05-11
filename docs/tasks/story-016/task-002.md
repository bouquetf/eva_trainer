# Tâche #002 - Story #016 : Frontend - Afficher les zones sur la bonne version de carte

## Objectif
Mettre à jour le frontend pour que les zones provenant de différentes versions de carte (obsolete + active) s'affichent sur l'image de carte correcte, en utilisant le `map_id` de chaque zone plutôt que le `mapId` de configuration.

## Contexte
- Story #016 : `docs/stories/story-016.md` - Revision Mode with Weighted Probability
- Dépend de : Tâche #001 (backend inclut zones des cartes obsolètes avec `map_id` correct)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Modification du composant RevisionMap

Actuellement, `RevisionMap` reçoit `mapId` et `polygons` en props, et affiche l'image de la carte correspondant à `mapId`. Les polygones sont superposés sur cette image.

Avec l'ajout des cartes obsolètes, les zones peuvent provenir de différentes versions de carte (ex: zone_1 sur map v1, zone_5 sur map v0). Chaque zone a un `map_id` correct qui correspond à la version de carte dont elle provient.

**Changement :** Dans `Revision.tsx`, remplacer :
```tsx
<RevisionMap mapId={config.mapId} polygons={currentQuestion.polygons} />
```
par :
```tsx
<RevisionMap mapId={currentQuestion.map_id} polygons={currentQuestion.polygons} />
```

Ainsi, chaque zone est affichée sur l'image de la carte dont elle provient (la version correcte).

#### Aucun autre changement nécessaire

Les types `RevisionZone` incluent déjà `map_id`. Le composant `RevisionMap` affiche déjà l'image via `getMapImageUrl(mapId)`. Les polygones sont déjà en coordonnées absolues compatibles avec l'image correspondante. Le store `useRevisionStore` n'a pas besoin de modification.

**Cas nominaux :**
- Zone provenant de la carte active → affichée sur l'image de la carte active ✅
- Zone provenant d'une carte obsolète → affichée sur l'image de la carte obsolète ✅
- Zone provenant d'un template public → affichée sur l'image de la carte du template ✅

**Cas limites :**
- L'image de la carte obsolète existe toujours en base (elle a un webp_blob). Le blob est conservé quand la carte devient obsolète
- Si l'image ne se charge pas (404), `RevisionMap` affiche déjà un spinner (gestion d'erreur existante)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/Revision.tsx` | Modifier | Ligne 246 : utiliser `currentQuestion.map_id` au lieu de `config.mapId` |

### Signatures

```tsx
// AVANT (ligne 246)
<RevisionMap mapId={config.mapId} polygons={currentQuestion.polygons} />

// APRÈS
<RevisionMap mapId={currentQuestion.map_id} polygons={currentQuestion.polygons} />
```

### Contraintes techniques

- **Changement minimal** : Une seule ligne à modifier dans `Revision.tsx`
- **Type safe** : `currentQuestion.map_id` existe déjà dans l'interface `RevisionZone` (champ `map_id: number`)
- **Pas de régression** : Si toutes les zones viennent de la même carte (cas sans obsolètes), `currentQuestion.map_id` sera identique à `config.mapId` pour toutes les zones
- **Pas de modification du store** : Le store `useRevisionStore` n'a pas besoin de changement
- **Image loading** : Le hook `useImage` dans `RevisionMap` gère déjà le chargement différé. Changer de `mapId` entre deux questions provoquera un rechargement de l'image. C'est un comportement acceptable

### Tests à implémenter

#### Tests d'intégration (composant)

**Fichier** : `frontend/src/components/RevisionMap.test.tsx`

| Scénario | Description |
|----------|-------------|
| Affichage avec mapId obsolète | Vérifier que `getMapImageUrl` est appelé avec le bon `mapId` |
| Changement de mapId entre questions | Simuler un changement de `mapId` et vérifier que la nouvelle image est chargée |

### Documentation

Aucune documentation à créer pour cette tâche.
