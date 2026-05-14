# Tâche #001 - Story #020 : Mode lecture seule dans MapEditor et ZoneList

## Objectif

Ajouter un mode "lecture seule" (readOnly) aux composants `MapEditor` et `ZoneList` afin qu'ils puissent afficher les zones sans aucune fonctionnalité d'édition.

## Contexte

- Story #020 : `docs/stories/story-020.md` - Vue liste des zones en lecture
- Dépend de : Story #009 (composants MapEditor, ZoneList existants)
- Nécessaire pour : Tâche #002 (page dédiée), Tâche #003 (tests)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter une prop `readOnly?: boolean` aux composants suivants :

#### Partie 1 - MapEditor.tsx

**Fichier** : `frontend/src/components/MapEditor/MapEditor.tsx`

1. **Interface** : Ajouter `readOnly?: boolean` à `MapEditorProps`
2. **Initialisation** (useEffect init, ~ligne 162) :
   - Si `readOnly === true` : NE PAS appeler `ensureLock()`, NE PAS acquérir de verrou
   - Charger les zones normalement via `listZones(teamId, mapId)`
   - Charger l'image de la carte normalement
   - Ne pas initialiser le rafraîchissement du verrou ni le compte à rebours
3. **Toolbar** (~ligne 881) :
   - Si `readOnly === true` : NE PAS rendre `<Toolbar>`
   - Rendre à la place un en-tête minimaliste : nom de la carte + indication "Lecture seule" + lien "Retour" vers `/teams/${teamId}`
4. **MapCanvas** (~ligne 896) :
   - Si `readOnly === true` : passer `editingZone ? [editingZone] : []` (comme dans l'éditeur — une zone à la fois, celle sélectionnée dans la liste)
   - Si `readOnly === true` : le MapCanvas ne reçoit plus d'interactions édition — les callbacks `onPolygonPointAdded`, `onPolygonLastPointRemoved`, `onZoneDrag`, `onPointDrag`, `onPointDelete`, `onPointInsert`, `onMapDoubleClick` sont des no-ops
   - `onPolygonClick` et `onBackgroundClick` et `onMapRightClick` sont conservés (sélection, clic droit) — mais le contextMenu ne sera pas rendu
   - `onZoneRightClick` devient no-op (pas de menu contextuel d'édition)
5. **ZoneList** (~ligne 976) :
   - Passer `readOnly` à ZoneList
   - Passer `onSelectZone` (toujours utile pour sélectionner une zone et la voir sur la carte)
   - Passer des no-ops pour `onDeleteZone`, `onCreateZone`, `onRenameZone` (ou mieux : rendre ces props optionnelles dans ZoneList)
6. **UI édition** (~lignes 949-992) :
   - Si `readOnly === true` : NE PAS rendre les boutons "Sauvegarder"/"Annuler", ni VersionBrowser
   - **Garder le sélecteur de cartes** (`<select>` maps) mais simplifié : navigation directe vers `/teams/${teamId}/zones/${targetMapId}` (pas de vérification isDirty, pas de releaseLockSync, pas de dialog)
7. **ContextMenu** (~ligne 1035) : NE PAS rendre
8. **ConfirmDialog / RestoreConfirmModal** (~lignes 1044-1062) : NE PAS rendre
9. **Bannière version** (~ligne 996) : NE PAS rendre (pas de version browsing en lecture seule)

**Comportement attendu en mode readOnly :**
- La carte affiche la zone sélectionnée avec ses polygones et son nom (comme dans l'éditeur)
- L'utilisateur clique sur une zone dans la liste → elle s'affiche sur la carte (highlight + polygones)
- Un sélecteur de cartes simplifié permet de naviguer entre les cartes en restant en mode readOnly
- Aucun outil d'édition n'est accessible (pas de toolbar, pas de context menu, pas d'édition de points)
- Aucun verrou n'est acquis
- Aucune modification possible depuis cet écran

#### Partie 2 - ZoneList.tsx

**Fichier** : `frontend/src/components/MapEditor/ZoneList.tsx`

1. **Interface** : Ajouter `readOnly?: boolean` à `ZoneListProps`
2. **Props optionnelles** : Rendre `onDeleteZone`, `onCreateZone`, `onRenameZone` optionnelles (signature `?:`)
3. **Mode readOnly** :
   - NE PAS rendre le bloc "Nouvelle zone" (lignes 85-120) : pas d'input, pas de bouton +
   - NE PAS rendre les boutons "X" de suppression (lignes 185-195)
   - Désactiver le double-clic pour renommer (handleDoubleClick ne fait rien si readOnly)
   - Conserver l'affichage des noms, du compte polygones/points
   - Conserver le surlignage de la zone sélectionnée
   - Afficher les zones sans aucune action destructive possible

**Cas nominaux :**
- Lecture seule : la liste affiche toutes les zones avec leurs noms et statistiques
- Lecture seule : clic sur une zone → `onSelectZone(id)` est appelé (sélection visuelle)
- Lecture seule : les zones s'affichent sur la carte via MapCanvas

**Cas limites :**
- `readOnly` + `zones.length === 0` : afficher "Aucune zone."
- Si `onDeleteZone`/`onCreateZone`/`onRenameZone` ne sont pas fournies (undefined) et `readOnly` est `false` : comportement normal (elles sont requises hors readOnly, mais rendues optionnelles pour la sécurité)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter prop readOnly, conditionner logique lock/UI |
| `frontend/src/components/MapEditor/ZoneList.tsx` | Modifier | Ajouter prop readOnly, masquer UI édition |

### Signatures

```typescript
// MapEditor.tsx
interface MapEditorProps {
  teamId: number;
  mapId: number;
  readOnly?: boolean;
}
```

```typescript
// ZoneList.tsx
interface ZoneListProps {
  zones: Zone[];
  editingZoneId: number | null;
  onSelectZone: (id: number) => void;
  onDeleteZone?: (id: number) => void;
  onCreateZone?: (name: string) => void;
  onRenameZone?: (id: number, name: string) => void;
  readOnly?: boolean;
}
```

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **Style** : Ne pas modifier le style CSS existant. Utiliser les classes existantes (`.zone-list`, `.editor-container`, etc.)
- **Ne pas modifier** : MapCanvas.tsx, ZoneOverlay.tsx, Toolbar.tsx, ContextMenu.tsx
- **Propagation** : Ne PAS modifier Editor.tsx — il continue d'utiliser MapEditor sans readOnly

### Tests à implémenter

Ces tests sont définis dans la Tâche #003. Ne pas les implémenter ici. Seule l'implémentation du mode readOnly est demandée.

### Documentation

Ne pas créer de documentation ici. Voir Tâche #004.
