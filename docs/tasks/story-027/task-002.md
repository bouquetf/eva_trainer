# Tâche #002 - Story #027 : Raccourcis clavier + menu contextuel + popup d'aide

## Objectif
Intégrer les raccourcis clavier Ctrl+C/Ctrl+V pour copier/coller dans l'éditeur, ajouter les entrées "Copier" et "Coller" dans le menu contextuel, et documenter les nouveaux raccourcis dans la popup d'aide.

## Contexte
- Story #027 : `docs/stories/story-027.md`
- Dépend de : Tâche #001 (fonction `duplicatePolygonsWithOffset`, état `clipboardPolygons`, handlers `handleCopy`/`handlePaste`)
- Nécessaire pour : Tâche #003 (Tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Raccourcis clavier (dans `MapEditor.tsx`) :**
- Ajouter les handlers pour Ctrl+C / Cmd+C et Ctrl+V / Cmd+V dans le `useEffect` existant qui gère les raccourcis clavier (lignes 317-345).
- Ctrl+C : appeler `handleCopy()`
- Ctrl+V : appeler `handlePaste()`
- Les raccourcis doivent être désactivés quand un champ INPUT a le focus (`document.activeElement?.tagName === "INPUT"`) — déjà géré par le guard existant.
- Les raccourcis doivent être ignorés en mode `readOnly`.

**Menu contextuel (dans `handleZoneRightClick` de `MapEditor.tsx`) :**
- Ajouter les entrées "Copier" et "Coller" dans le menu contextuel qui s'affiche sur un clic droit sur une forme (`handleZoneRightClick`).
- "Copier" : visible uniquement quand un ou plusieurs polygones sont sélectionnés (`selectedPolygonIndexes.size > 0`).
- "Coller" : visible uniquement quand le clipboard n'est pas vide (`clipboardPolygons.length > 0`).
- Les entrées "Copier" et "Coller" apparaissent **avant** les entrées existantes "Flip H" / "Flip V", sans séparateur.

**Popup d'aide (dans `EditorHelpPopup.tsx`) :**
- Ajouter 2 nouvelles entrées dans le tableau `SHORTCUTS` :

| Commande | Raccourci | Contexte |
|----------|-----------|----------|
| Copier la/les forme(s) sélectionnée(s) | Ctrl+C / Cmd+C | Au moins 1 polygone sélectionné |
| Coller la/les forme(s) copiée(s) | Ctrl+V / Cmd+V | Clipboard non vide |

**Cas nominaux :**
- Sélectionner un polygone → Ctrl+C → Ctrl+V → polygone dupliqué avec décalage de 30px
- Clic droit sur une zone → menu avec "Copier", "Coller", "Flip H", "Flip V"
- Ouvrir la popup d'aide → les raccourcis copier/coller sont documentés

**Cas limites :**
- Ctrl+C alors qu'aucun polygone n'est sélectionné → rien ne se passe
- Ctrl+V alors que le clipboard est vide → rien ne se passe
- Clic droit sans sélection → "Copier" absent, "Coller" absent si clipboard vide
- Ctrl+C alors qu'un champ INPUT a le focus → pas de copie (le navigateur gère le copier natif)

**Gestion d'erreurs :**
- Aucune gestion d'erreur spécifique (les handlers de copier/coller sont déjà implémentés dans la Tâche #001 et gèrent les cas vides silencieusement)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter Ctrl+C/Ctrl+V dans le `useEffect` des raccourcis + entrées dans `handleZoneRightClick` |
| `frontend/src/components/MapEditor/EditorHelpPopup.tsx` | Modifier | Ajouter les 2 nouvelles entrées dans `SHORTCUTS` |

### Modifications dans MapEditor.tsx

**Dans le `useEffect` des raccourcis clavier (lignes 317-345) :**
```typescript
// À ajouter après le bloc Ctrl+Z, avant le bloc Delete
if ((e.ctrlKey || e.metaKey) && e.key === "c") {
  if (!editingZone || selectedPolygonIndexes.size === 0) return;
  e.preventDefault();
  handleCopy();
  return;
}

if ((e.ctrlKey || e.metaKey) && e.key === "v") {
  if (!editingZone || clipboardPolygons.length === 0) return;
  e.preventDefault();
  handlePaste();
  return;
}
```

**Dans `handleZoneRightClick` (lignes 853-862) :**
```typescript
const handleZoneRightClick = useCallback(
  (_zoneId: number, clientX: number, clientY: number) => {
    const items = [];

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

    items.push(
      { label: "Flip H", action: () => handleFlipH() },
      { label: "Flip V", action: () => handleFlipV() },
    );

    setContextMenu({ x: clientX, y: clientY, items });
  },
  [selectedPolygonIndexes, clipboardPolygons, handleCopy, handlePaste, handleFlipH, handleFlipV],
);
```

Note : `handleCopy` et `handlePaste` sont déjà wrappées dans `useCallback` par la Tâche #001, elles sont utilisables directement comme dépendances.

### Modifications dans EditorHelpPopup.tsx

**Dans le tableau `SHORTCUTS` :**
Ajouter ces 2 entrées après la ligne de "Supprimer polygone(s) sélectionné(s)" :

```typescript
{ command: "Copier la/les forme(s) sélectionnée(s)", shortcut: "Ctrl+C / Cmd+C", context: "Au moins 1 polygone sélectionné" },
{ command: "Coller la/les forme(s) copiée(s)", shortcut: "Ctrl+V / Cmd+V", context: "Clipboard non vide" },
```

### Contraintes techniques
- **React** : Les handlers `handleCopy`/`handlePaste` sont déjà `useCallback` (Tâche #001)
- **Raccourcis** : Suivre exactement le pattern existant (`e.ctrlKey || e.metaKey`, `e.preventDefault()`)
- **Menu contextuel** : Pas de séparateur entre copier/coller et Flip H/Flip V
- **Dépendances du `useEffect`** : Le `useEffect` (lignes 317-345) n'a pas de tableau de dépendances explicite (il se base sur le comportement par défaut). Les conditions utilisent `editingZone`, `selectedPolygonIndexes`, `clipboardPolygons` qui sont des états React.

### Tests implémentés dans cette tâche

Aucun test dans cette tâche (les tests sont dans la Tâche #003).

### Documentation
- Aucune documentation à créer (la popup d'aide tient lieu de documentation des raccourcis)

### Dépendances
- Cette tâche dépend de la Tâche #001 (les handlers `handleCopy`/`handlePaste` et l'état `clipboardPolygons` doivent exister)
