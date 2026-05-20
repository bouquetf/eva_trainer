# Tâche #003 - Story #029 : Frontend — Indicateurs visuels d'étage (Haut/Bas)

## Objectif
Afficher un indicateur visuel (flèche + trait) au-dessus ou en dessous de chaque polygone en fonction de son attribut `floor`, dans l'éditeur ET en mode révision.

## Contexte
- Story #029 : `docs/stories/story-029.md`
- Dépend de : Tâche #002 (Types + API + menu contextuel)
- Nécessaire pour : Tâche #004 (Tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Indicateur "Haut" :**
- Une flèche vers le haut (↑) avec un trait horizontal au-dessus de la forme
- Placé au-dessus du polygone (décalé vers le haut par rapport au centroid + hauteur du polygone)
- Le pictogramme est centré horizontalement sur le polygone

**Indicateur "Bas" :**
- Une flèche vers le bas (↓) avec un trait horizontal en dessous de la forme
- Placé en dessous du polygone (décalé vers le bas)

**Comportement général :**
- L'indicateur est visible pour tous les polygones ayant `floor` défini
- L'indicateur est visible dans l'éditeur ET en mode révision (`readOnly`)
- L'indicateur ne doit pas interférer avec les clics sur les polygones (pointer-events: none)
- Taille fixe : ~30 unités image (dans le repère SVG natif)

**Désign visuel :**
- Flèche : dessinée comme un triangle (polygon) de couleur blanc/clair
- Trait horizontal : une ligne au-dessus/en dessous de la flèche
- Couleur : blanc (#ffffff) avec une légère opacité (0.8), ou la couleur de contour du polygone

**Positionnement proportionnel :**
- Calculer le centroid du polygone (moyenne des X et Y de tous les points)
- Calculer la boîte englobante (bounding box) : `minY`, `maxY`, `height = maxY - minY`
- L'offset de l'indicateur est proportionnel à la hauteur du polygone : `offset = Math.max(20, height * 0.15)`
- Pour "Haut" : placer l'indicateur au-dessus du polygone à `y = minY - offset`, centré sur le centroid X
- Pour "Bas" : placer l'indicateur en dessous du polygone à `y = maxY + offset`, centré sur le centroid X
- L'icône a une taille fixe de 30 unités image (ne change pas avec la taille du polygone)

**Cas nominaux :**
- Polygone avec `floor = "haut"` → flèche ↑ + trait au-dessus
- Polygone avec `floor = "bas"` → flèche ↓ + trait en dessous
- Polygone sans floor → pas d'indicateur

**Cas limites :**
- Polygone très petit ou très grand → l'indicateur a une taille fixe dans le repère image
- Polygone proche du bord de l'image → l'indicateur peut dépasser (pas de clipping spécifique requis)
- Plusieurs polygones dans une zone avec des floors différents → chaque polygone a son propre indicateur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/ZoneOverlay.tsx` | Modifier | Ajouter le rendu des indicateurs d'étage pour chaque polygone |
| `frontend/src/components/MapEditor/MapCanvas.tsx` | Modifier | Propager `readOnly` à ZoneOverlay (optionnel, déjà géré) |

### Détail d'implémentation

Dans `ZoneOverlay.tsx`, après le `<polygon>` principal de chaque polygone (ligne ~87), ajouter le rendu conditionnel du floor indicator :

```typescript
// À l'intérieur du <g> de chaque polygone, APRÈS le polygone principal
{/* Indicateur d'étage proportionnel à la taille du polygone */}
{polygon.floor && (() => {
  const pts = polygon.points;
  const cx = pts.reduce((s, p) => s + p.x, 0) / pts.length;
  const minY = Math.min(...pts.map(p => p.y));
  const maxY = Math.max(...pts.map(p => p.y));
  const height = maxY - minY;
  const iconSize = 30;
  const offset = Math.max(20, height * 0.15);

  if (polygon.floor === 'haut') {
    const y = minY - offset;
    const arrowTip = y - iconSize * 0.6;
    return (
      <g key={`floor-${polygon.id}-${pIdx}`} style={{ pointerEvents: 'none' }}>
        <line x1={cx - iconSize} y1={y} x2={cx + iconSize} y2={y} stroke="white" strokeWidth={2} />
        <polygon
          points={`${cx},${arrowTip} ${cx - iconSize * 0.4},${y} ${cx + iconSize * 0.4},${y}`}
          fill="white"
        />
      </g>
    );
  }

  if (polygon.floor === 'bas') {
    const y = maxY + offset;
    const arrowTip = y + iconSize * 0.6;
    return (
      <g key={`floor-${polygon.id}-${pIdx}`} style={{ pointerEvents: 'none' }}>
        <line x1={cx - iconSize} y1={y} x2={cx + iconSize} y2={y} stroke="white" strokeWidth={2} />
        <polygon
          points={`${cx},${arrowTip} ${cx - iconSize * 0.4},${y} ${cx + iconSize * 0.4},${y}`}
          fill="white"
        />
      </g>
    );
  }

  return null;
})()}
```

**Note sur `MapCanvas.tsx` : `ZoneOverlay` est rendu dans tous les cas (lecture seule ou non). Les indicateurs seront donc visibles en mode révision automatiquement, puisqu'ils font partie du rendu de `ZoneOverlay`. Aucune modification de `MapCanvas.tsx` n'est nécessaire.**

### Contraintes techniques
- **SVG** : Utiliser des éléments SVG natifs (`<line>`, `<polygon>`) dans le `<g>` du polygone
- **`pointer-events: none`** : L'indicateur ne doit pas capturer les clics — important pour ne pas interférer avec la sélection
- **Ordre de rendu** : L'indicateur est rendu APRÈS le polygone (dans l'ordre SVG), donc au-dessus visuellement
- **Pas de dépendance externe** : Utiliser uniquement des primitives SVG, pas de librairie de rendu additionnelle
- **Couleur** : Blanc (#ffffff) pour la visibilité sur fond sombre du polygone
- **Opacité** : Pas d'opacité réduite — l'indicateur doit être bien visible
- **Taille** : `iconSize = 30` unités image (fixe)
- **Offset proportionnel** : `offset = Math.max(20, height * 0.15)` où `height = maxY - minY` du polygone. L'offset a un minimum de 20 unités pour les petits polygones et grandit proportionnellement pour les grands polygones.
- **Clarté du code** : Nommer les variables intermédiaires (`height`, `offset`, `arrowTip`) pour la lisibilité

### Tests
Aucun test dans cette tâche (les tests sont dans la Tâche #004).

### Documentation
- Aucune documentation à créer dans cette tâche
