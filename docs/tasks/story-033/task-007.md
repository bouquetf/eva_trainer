# Tâche #007 - Story #033 : Intégration éditeur de carte (context menu, rendu canvas, édition)

## Objectif
Intégrer les lignes de tir dans le MapEditor : menu contextuel "Créer une ligne de tir", rendu sur le canvas Konva, sélection, modification, suppression, nommage, et ajout de cônes par double-clic.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #006 (frontend API)
- Nécessaire pour : Tâche #009 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**1. Rendu des lignes de tir sur le canvas**

Créer un composant `LineOfFireOverlay.tsx` dans `frontend/src/components/MapEditor/` qui :
- Reçoit la liste des `LineOfFire` en props
- Pour chaque ligne, pour chaque cône :
  - Dessine une ligne de l'origine vers `left_end`
  - Dessine une ligne de l'origine vers `right_end`
  - Dessine un arc (ou une ligne droite) entre `left_end` et `right_end` pour former le cône
  - Si `left_end == right_end` : une simple ligne droite de l'origine vers ce point unique
- Style : couleur distincte des zones (ex: rouge/orangé semi-transparent)
- Si la ligne est sélectionnée, les poignées (origin, left_end, right_end) sont des cercles draggables

**2. Menu contextuel "Créer une ligne de tir"**

Dans `MapEditor.tsx`, modifier `handleMapRightClick` :
- Ajouter un item "Créer une ligne de tir" dans le menu contextuel (quand on ne dessine pas de polygone)
- Action : ouvrir un dialog de saisie de nom, puis créer 3 points (origin au clic, left_end et right_end à poser ensuite)
- Alternative plus simple : créer une ligne vide (1 cône par défaut avec les 3 points au même endroit) et passer en mode édition

**3. Mode placement d'une ligne de tir**

Après clic droit → "Créer une ligne de tir" :
- Premier clic : définit l'origine
- Deuxième clic : définit `left_end`
- Troisième clic : définit `right_end`
- Si left_end et right_end sont au même point → segment droit (pas de cône visible)
- Dialog pour saisir le nom
- Sauvegarde via l'API

**4. Ajout d'un cône par double-clic sur l'origine**

Dans `LineOfFireOverlay.tsx` :
- Si l'utilisateur double-clique sur l'origine d'un cône existant : ajouter un nouveau cône à la même ligne (même origine)
- Le nouveau cône a la même origine, left_end et right_end par défaut au même point que l'origine
- L'utilisateur peut ensuite déplacer les extrémités

**5. Sélection et manipulation**

- Clic sur une ligne de tir → la sélectionne (comme les zones)
- Quand une ligne de tir est sélectionnée :
  - Poignées draggables sur origin, left_end, right_end de chaque cône
  - Appui sur Suppr → demande confirmation → supprime
  - Menu contextuel sur la ligne : Renommer, Supprimer, Ajouter un cône

**6. Liste des lignes de tir**

Dans `ZoneList.tsx` (ou un nouveau composant) :
- Afficher les lignes de tir dans une section séparée sous les zones
- Chaque ligne avec son nom, icône spécifique (🔫 ou autre)
- Clic pour sélectionner, clic droit pour menu contextuel
- Bouton ✕ pour supprimer

**7. Sauvegarde**

- Intégrer les lignes de tir dans le système de batch existant
- Les modifications sur les lignes de tir suivent le même pattern de `isDirty` et sauvegarde que les zones
- Ajouter des actions batch spécifiques pour les lignes de tir

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/LineOfFireOverlay.tsx` | Créer | Rendu canvas des lignes de tir |
| `frontend/src/components/MapEditor/LineOfFireNameDialog.tsx` | Créer | Dialog de saisie de nom |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Intégration lignes de tir |
| `frontend/src/components/MapEditor/MapCanvas.tsx` | Modifier | Passer les lignes de tir au composant d'overlay |
| `frontend/src/components/MapEditor/ZoneList.tsx` | Modifier | Afficher les lignes de tir |
| `frontend/src/components/MapEditor/LineOfFireOverlay.module.css` | Créer | Styles |

### Contraintes techniques
- **Rendu Konva** : Utiliser `react-konva` avec `Line`, `Circle`, `Group` pour dessiner les cônes.
- **Couleur** : Utiliser une couleur distincte (ex: `#ff6600` avec opacité 0.3 pour le remplissage du cône, `#ff4400` pour les contours).
- **Interaction** : Suivre le pattern de sélection/édition existant (`selectedPolygonIndexes`, `editingZone`, etc.).
- **Drag** : Utiliser `onDragMove` pour les poignées, avec mise à jour de l'état local et marquage `isDirty`.
- **Double-clic** : Détecter `onDblClick` sur les cercles d'origine pour ajouter un cône.
- **Lock** : Vérifier le lock avant de permettre la création/édition (via `ensureLock()`).
- **Batch** : Les actions de lignes de tir doivent pouvoir être intégrées dans le même flux batch que les zones, ou bien utiliser un batch séparé (via `batchLineOfFireActions`).

### Comportement attendu

| Action | Résultat |
|--------|----------|
| Clic droit sur la carte → "Créer une ligne de tir" | Passe en mode placement (3 clics pour les 3 points) |
| 3 points placés + nom saisi | Ligne de tir créée et sauvegardée |
| Double-clic sur l'origine d'un cône | Nouveau cône ajouté à la même ligne |
| Drag de l'origine | Tous les cônes de la ligne sont déplacés (même origine) |
| Drag de left_end/right_end | Le point d'extrémité se déplace |
| Clic droit sur une ligne de tir | Menu : Renommer, Supprimer, Ajouter cône |

### Tests

Voir Tâche #009 pour les tests E2E. Cette tâche se concentre sur l'implémentation UI.
