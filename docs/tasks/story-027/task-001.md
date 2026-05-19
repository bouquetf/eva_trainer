# Tâche #001 - Story #027 : Fonction de duplication de polygones + état clipboard + handlers copier/coller

## Objectif
Implémenter la fonction utilitaire `duplicatePolygonsWithOffset` dans `transforms.ts` et l'état clipboard (copied polygons) ainsi que les handlers de copier/coller dans `MapEditor.tsx`.

## Contexte
- Story #027 : `docs/stories/story-027.md`
- Dépend de : Aucune (première tâche)
- Nécessaire pour : Tâche #002 (UI), Tâche #003 (Tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Copier (Ctrl+C / Cmd+C) :**
- Lorsque l'utilisateur appuie sur Ctrl+C (ou Cmd+C), les polygones actuellement sélectionnés (`selectedPolygonIndexes`) sont copiés dans un état clipboard interne.
- Rien ne change visuellement (pas de feedback visuel de copie).
- Si aucun polygone n'est sélectionné, la copie ne fait rien.
- Les polygones copiés conservent leur `zone_id` d'origine (pour savoir dans quelle zone ils appartiennent).
- Re-copier écrase le clipboard précédent.

**Coller (Ctrl+V / Cmd+V) :**
- Lorsque l'utilisateur appuie sur Ctrl+V (ou Cmd+V), les polygones précédemment copiés sont dupliqués avec un décalage de **30px en X et 30px en Y**.
- La forme collée appartient **toujours à la zone d'origine** (celle depuis laquelle la copie a été faite), pas forcément la zone en cours d'édition.
  - Si `editingZone.id === clipboard.sourceZoneId` → ajouter les polygones à `editingZone` (cas normal)
  - Si `editingZone.id !== clipboard.sourceZoneId` → basculer l'édition sur la zone source, ajouter les polygones à cette zone
  - Si la zone source n'existe plus (supprimée) → ne rien faire
- Les nouveaux polygones ont leur `id` à 0 (non persistés, nouveaux polygones).
- Les nouveaux polygones sont automatiquement sélectionnés après collage.
- Si le clipboard est vide ou si la zone source n'existe plus, le collage ne fait rien.

**Cas nominaux :**
- Copier un seul polygone sélectionné → coller → un nouveau polygone décalé de 30px apparaît dans la même zone
- Copier plusieurs polygones (Ctrl+Click pour sélection multiple) → coller → tous les polygones sont dupliqués avec décalage
- Copier depuis zone A, sélectionner zone B, coller → bascule automatique sur zone A, polygones ajoutés à A

**Cas limites :**
- Aucun polygone sélectionné → Ctrl+C ne fait rien
- Clipboard vide → Ctrl+V ne fait rien
- Aucune zone en cours d'édition → Ctrl+V ne fait rien
- Copier, supprimer la zone source, coller → ne fait rien (zone source introuvable)
- Copier puis copier à nouveau → le second copier remplace le clipboard (pas d'accumulation)

**Gestion d'erreurs :**
- Pas d'erreur utilisateur : les actions copier/coller sont silencieuses si les préconditions ne sont pas remplies
- La sauvegarde est gérée séparément (via le bouton sauvegarder ou Ctrl+S)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/geometry/transforms.ts` | Modifier | Ajouter la constante `PASTE_OFFSET` exportée + fonction `duplicatePolygonsWithOffset` |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter l'état clipboard, `sourceZoneId`, les handlers `handleCopy`/`handlePaste` (wrappés `useCallback`) |

### Signatures

```typescript
// Dans transforms.ts (AJOUTER à la fin du fichier)

export const PASTE_OFFSET = 30;

/**
 * Crée une copie profonde des polygones sélectionnés avec un décalage.
 * @param zone - La zone source contenant les polygones à dupliquer
 * @param selectedIndexes - Les indices des polygones à dupliquer
 * @param offsetX - Décalage horizontal en pixels (défaut: PASTE_OFFSET)
 * @param offsetY - Décalage vertical en pixels (défaut: PASTE_OFFSET)
 * @returns Un tableau de nouveaux ZonePolygon avec id=0 et points décalés
 */
export function duplicatePolygonsWithOffset(
  zone: Zone,
  selectedIndexes: Set<number>,
  offsetX: number = PASTE_OFFSET,
  offsetY: number = PASTE_OFFSET,
): ZonePolygon[]
```

```typescript
// Dans MapEditor.tsx (ajouter au niveau des useState, vers ligne 74)

import { PASTE_OFFSET, duplicatePolygonsWithOffset } from "../../lib/geometry/transforms";

// Nouveaux états
const [clipboardPolygons, setClipboardPolygons] = useState<ZonePolygon[]>([]);
const [clipboardSourceZoneId, setClipboardSourceZoneId] = useState<number | null>(null);
```

```typescript
// Handlers (wrappés dans useCallback, à placer avant handlePolygonClick)

const handleCopy = useCallback(() => {
  if (!editingZone || selectedPolygonIndexes.size === 0) return;
  const copied = Array.from(selectedPolygonIndexes)
    .sort((a, b) => a - b)
    .map((i) => editingZone.polygons[i])
    .filter(Boolean)
    .map((poly) => ({
      ...poly,
      id: 0,
      points: poly.points.map((p) => ({ ...p })),
      created_at: new Date().toISOString(),
    }));
  setClipboardPolygons(copied);
  setClipboardSourceZoneId(editingZone.id);
}, [editingZone, selectedPolygonIndexes]);

const handlePaste = useCallback(() => {
  if (!editingZone || clipboardPolygons.length === 0 || clipboardSourceZoneId === null) return;

  // Déterminer la zone cible : toujours la zone d'origine de la copie
  let targetZone = editingZone;
  if (editingZone.id !== clipboardSourceZoneId) {
    // Chercher la zone source dans la liste complète des zones
    const sourceZone = zones.find((z) => z.id === clipboardSourceZoneId);
    if (!sourceZone) return; // zone source supprimée
    targetZone = sourceZone;
    // Basculer l'édition sur la zone source
    startEditZone(sourceZone);
  }

  updateEditingPolygon((z) => {
    const maxOrder = z.polygons.reduce((m, p) => Math.max(m, p.order_index), -1);
    const newPolys = clipboardPolygons.map((poly, i) => {
      const offsetIndex = i; // tous les polygones collés ont le même décalage cumulé
      const offset = (offsetIndex + 1) * PASTE_OFFSET;
      return {
        ...poly,
        zone_id: z.id,
        order_index: maxOrder + 1 + i,
        points: poly.points.map((p) => ({
          x: Math.round(p.x + offset),
          y: Math.round(p.y + offset),
        })),
      };
    });
    return {
      ...z,
      polygons: [...z.polygons, ...newPolys],
      updated_at: new Date().toISOString(),
    };
  });

  // Sélectionner les nouveaux polygones
  const startIndex = targetZone.polygons.length;
  setSelectedPolygonIndexes(
    new Set(clipboardPolygons.map((_, i) => startIndex + i)),
  );
}, [editingZone, clipboardPolygons, clipboardSourceZoneId, zones, updateEditingPolygon, startEditZone]);
```

### Contraintes techniques
- **Framework** : React + TypeScript
- **Pattern** : Suivre exactement le pattern des fonctions existantes dans `transforms.ts` (immutabilité, `Math.round` sur les coordonnées, mise à jour de `updated_at`)
- **Style** : Utiliser les mêmes conventions que les fonctions existantes (pas de commentaires superflus, types importés depuis `../../lib/api/types`)
- **Immutabilité** : Ne jamais muter l'état existant, toujours créer de nouveaux objets/tableaux (deep clone)
- **`useCallback`** : `handleCopy` et `handlePaste` doivent être wrappés dans `useCallback` avec les dépendances appropriées, pour pouvoir être passés comme dépendances au `handleZoneRightClick` (Tâche #002)
- **Offset** : `PASTE_OFFSET = 30` exporté depuis `transforms.ts`
- **Import** : MapEditor.tsx importe `PASTE_OFFSET` depuis `transforms.ts`
- **Pas de code mort** : `duplicatePolygonsWithOffset` est la seule fonction nouvelle dans `transforms.ts`
- **Décalage progressif** : si on colle 2 polygones à la fois, le second est décalé de 60px (2×30) pour éviter la superposition entre les collés eux-mêmes

### Tests implémentés dans cette tâche

Aucun test dans cette tâche (les tests sont dans la Tâche #003).

### Documentation
- Aucune documentation à créer dans cette tâche

### Dépendances
- Cette tâche doit être exécutée avant la Tâche #002
