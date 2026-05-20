# Tâche #002 - Story #029 : Frontend — Types, API client, et menu contextuel "Étage"

## Objectif
Ajouter le type `floor` sur `ZonePolygon`, la nouvelle action batch `set_floor`, et le sous-menu contextuel "Étage" avec les options Haut/Bas/Effacer.

## Contexte
- Story #029 : `docs/stories/story-029.md`
- Dépend de : Tâche #001 (Backend — modèle, migration, batch action)
- Nécessaire pour : Tâche #003 (Visuels), Tâche #004 (Tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Types :**
- `ZonePolygon` reçoit un champ optionnel `floor?: 'haut' | 'bas'` (string union ou undefined)

**API Client :**
- Le type `BatchAction` reçoit un nouveau variant `set_floor` :
  ```typescript
  | { type: 'set_floor'; polygon_id: number; floor: 'haut' | 'bas' | null }
  ```

**Menu contextuel avec sous-menus imbriqués :**
- Le composant `ContextMenu` doit être modifié pour supporter les sous-menus imbriqués
- Quand l'utilisateur fait un clic droit sur un polygone (et qu'un polygone est sélectionné), le menu "Étage" apparaît avec un sous-menu
- Le sous-menu "Étage" contient jusqu'à 3 options :
  - "Haut" → définit floor = "haut" sur le(s) polygone(s) sélectionné(s)
  - "Bas" → définit floor = "bas" sur le(s) polygone(s) sélectionné(s)
  - "Effacer" → définit floor = null (visible uniquement si AU MOINS UN des polygones sélectionnés a un floor non-null)
- Les options Haut et Bas sont mutuellement exclusives : si le polygone est déjà "Haut", l'option "Haut" reste disponible (elle est idempotente)
- L'action s'applique à TOUS les polygones sélectionnés (pas seulement celui cliqué)

**Sauvegarde :**
- Après sélection d'une option, envoyer une action batch `set_floor` pour chaque polygone sélectionné (ou une seule action par polygone dans le tableau)
- Ne pas sauvegarder immédiatement dans le state local optimiste — le rafraîchissement viendra du re-fetch (ou on update le state local pour réactivité)

**State management :**
- Quand la réponse du batch arrive, mettre à jour le state local `zones` et `editingZone` pour refléter le changement de floor
- Le `updateEditingPolygon` peut être utilisé pour mettre à jour le polygone dans la zone courante

**Cas nominaux :**
- Clic droit sur un polygone sélectionné → menu avec "Étage → Haut/Bas"
- Sélectionner "Haut" → floor = "haut" sur TOUS les polygones sélectionnés
- Sélectionner "Bas" → floor = "bas" sur tous les polygones sélectionnés
- Sélectionner "Effacer" → floor = null sur tous les polygones sélectionnés

**Cas limites :**
- Aucun polygone sélectionné → ne PAS montrer le sous-menu Étage (les options existantes "Copier", etc. sont déjà conditionnées par `selectedPolygonIndexes.size > 0`)
- Polygone n'ayant pas de floor → Effacer n'apparaît pas
- Un polygone a floor="haut", un autre n'a pas de floor → Effacer apparaît (car au moins un a un floor)
- Le sous-menu Étage est toujours visible dans le menu contextuel lorsqu'au moins un polygone est sélectionné (même si aucun floor défini)

**Gestion d'erreurs :**
- Si la sauvegarde échoue (erreur API), afficher une alerte "Erreur lors de la mise à jour de l'étage"
- Le state local n'est pas mis à jour en cas d'erreur (optimistic update avec rollback si nécessaire)
- L'état du menu reste inchangé si l'API échoue

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/api/types.ts` | Modifier | Ajouter `floor?: 'haut' \| 'bas'` à `ZonePolygon` |
| `frontend/src/api/zones.ts` | Modifier | Ajouter `set_floor` à `BatchAction` |
| `frontend/src/components/MapEditor/ContextMenu.tsx` | Modifier | Ajouter le support des sous-menus imbriqués (`ContextMenuItem.children`) |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter `handleSetFloor`, modifier `handleZoneRightClick` pour sous-menu Étage |

### Signatures

```typescript
// frontend/src/components/MapEditor/ContextMenu.tsx — MODIFIER ContextMenuItem
interface ContextMenuItem {
  label: string;
  action?: () => void;
  disabled?: boolean;
  children?: ContextMenuItem[]; // AJOUTÉ : sous-menu imbriqué
}

// frontend/src/lib/api/types.ts — MODIFIER ZonePolygon
export type ZonePolygon = {
  id: number;
  zone_id: number;
  points: Point[];
  order_index: number;
  floor?: 'haut' | 'bas';
  created_at: string;
};

// frontend/src/api/zones.ts — MODIFIER BatchAction
export type BatchAction =
  | { type: 'create_zone'; temp_id: string; polygons: { points: Point[] }[]; name?: string }
  | { type: 'add_polygon'; zone_id: number; points: Point[] }
  | { type: 'remove_polygon'; zone_id: number; polygon_id: number }
  | { type: 'update_points'; zone_id: number; polygon_id: number; points: Point[] }
  | { type: 'add_point'; polygon_id: number; point: Point; index_after: number }
  | { type: 'remove_point'; polygon_id: number; point_index: number; zone_id: number }
  | { type: 'rename'; zone_id: number; name: string }
  | { type: 'delete_zone'; zone_id: number }
  | { type: 'set_floor'; polygon_id: number; floor: 'haut' | 'bas' | null };

// frontend/src/components/MapEditor/MapEditor.tsx — AJOUTER dans le composant
const handleSetFloor = useCallback(
  async (floor: 'haut' | 'bas' | null) => {
    // Voir corps ci-dessous
  },
  [selectedPolygonIndexes, editingZone, zones, setZones, setEditingZone, teamId, mapId, ensureLock],
);
```

**Corps de `handleSetFloor` :**
```typescript
const handleSetFloor = useCallback(
  async (floor: 'haut' | 'bas' | null) => {
    if (!editingZone || selectedPolygonIndexes.size === 0) return;

    const polygonIds = Array.from(selectedPolygonIndexes)
      .map((i) => editingZone.polygons[i])
      .filter(Boolean)
      .map((p) => p.id)
      .filter((id) => id > 0); // only persisted polygons

    if (polygonIds.length === 0) return;

    const ok = await ensureLock();
    if (!ok) return;

    const actions: BatchAction[] = polygonIds.map((polygonId) => ({
      type: 'set_floor',
      polygon_id: polygonId,
      floor,
    }));

    try {
      await batchActions(teamId, mapId, actions);
      // Mise à jour locale optimiste : on met à jour les polygones concernés
      setEditingZone((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          polygons: prev.polygons.map((p) =>
            polygonIds.includes(p.id) ? { ...p, floor: floor ?? undefined } : p,
          ),
        };
      });
      setZones((prev) =>
        prev.map((z) =>
          z.id === editingZone.id
            ? {
                ...z,
                polygons: z.polygons.map((p) =>
                  polygonIds.includes(p.id) ? { ...p, floor: floor ?? undefined } : p,
                ),
              }
            : z,
        ),
      );
    } catch {
      showError("Erreur lors de la mise à jour de l'étage.");
    }
  },
  [editingZone, selectedPolygonIndexes, teamId, mapId, ensureLock],
);
```

**Modification de `handleZoneRightClick` :**
```typescript
const handleZoneRightClick = useCallback(
  (_zoneId: number, clientX: number, clientY: number) => {
    const items: ContextMenuItem[] = [];

    if (selectedPolygonIndexes.size > 0) {
      items.push({
        label: "Copier",
        action: () => handleCopy(),
      });
    }

    if (clipboardPolygons.length > 0) {
      items.push({
        label: "Coller",
        action: () => handlePaste(),
      });
    }

    // --- Sous-menu Étage ---
    if (selectedPolygonIndexes.size > 0 && editingZone) {
      const hasFloor = Array.from(selectedPolygonIndexes).some(
        (i) => editingZone.polygons[i]?.floor != null,
      );

      const floorChildren: ContextMenuItem[] = [
        { label: "Haut", action: () => handleSetFloor("haut") },
        { label: "Bas", action: () => handleSetFloor("bas") },
      ];
      if (hasFloor) {
        floorChildren.push({
          label: "Effacer",
          action: () => handleSetFloor(null),
        });
      }

      items.push({
        label: "Étage",
        children: floorChildren,
      });
    }
    // --- Fin sous-menu Étage ---

    items.push(
      { label: "Flip H", action: () => handleFlipH() },
      { label: "Flip V", action: () => handleFlipV() },
    );

    setContextMenu({ x: clientX, y: clientY, items });
  },
  [selectedPolygonIndexes, clipboardPolygons, editingZone, handleCopy, handlePaste, handleSetFloor, handleFlipH, handleFlipV],
);
```

### Contraintes techniques
- **Framework** : React + TypeScript
- **Immutabilité** : Toujours créer de nouveaux objets plutôt que de muter
- **`showError`** : Utiliser la fonction `showError` déjà définie dans `MapEditor.tsx` ligne 59
- **Sous-menus** : Le composant `ContextMenu` doit être modifié pour supporter les items ayant un tableau `children`. Un item avec `children` n'a pas d'`action` directe mais ouvre un sous-menu au survol (hover).
- **Design du sous-menu** : 
  - L'item parent "Étage" s'affiche avec une petite flèche ">" à droite pour indiquer la présence d'un sous-menu
  - Le sous-menu apparaît à droite de l'item parent, avec le même style que le menu principal (fond #1e1e1e, bordure #333)
  - Le sous-menu reste ouvert tant que la souris est sur l'item parent ou le sous-menu lui-même
  - Un délai de ~300ms avant fermeture pour éviter les fermetures intempestives pendant le survol
- **Modifications de ContextMenu.tsx :**
  - `ContextMenuItem`: ajouter `children?: ContextMenuItem[]` optionnel
  - Rendu : si `item.children` existe, rendre l'item différemment (sans `onClick` final direct) avec un indicateur ">" 
  - Gérer le state `hoveredItem` pour savoir quel sous-menu afficher
  - Ajouter un effet de fermeture du sous-menu quand la souris quitte la zone parent+enfant
  - Le sous-menu est positionné en absolute par rapport à l'item parent (top: 0, left: 100%)
- **`ensureLock`** : Appeler `ensureLock()` avant toute action batch — suivre le pattern des autres handlers
- **Pas de test** : Les tests sont dans la Tâche #004

### Tests
Aucun test dans cette tâche (les tests sont dans la Tâche #004).

### Documentation
- Aucune documentation à créer dans cette tâche
