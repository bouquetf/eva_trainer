# Prompt pour agent de code — MVP éditeur de zones polygonales sur image WebP

Tu es un agent de code senior frontend. Génère un MVP fonctionnel d’un éditeur web de zones polygonales sur image WebP.

## Objectif

Créer une application frontend permettant de charger/afficher une image WebP de type carte, puis de dessiner une ou plusieurs zones polygonales par-dessus l’image. Les zones doivent être stockées sous forme de coordonnées réutilisables, afin de pouvoir les réafficher plus tard exactement au même endroit.

Le backend réel n’existe pas encore. Il faut donc créer une librairie interne qui mock l’API backend, avec une interface propre et facilement remplaçable plus tard par de vrais appels HTTP.

## Stack technique attendue

Utiliser :

- React
- TypeScript
- Vite
- SVG natif pour l’overlay interactif
- CSS simple ou Tailwind si déjà configuré facilement
- localStorage pour simuler la persistance côté mock backend

Ne pas utiliser de backend réel.
Ne pas utiliser de librairie lourde de dessin type Fabric.js ou Konva pour ce MVP.
Ne pas modifier l’image source : les zones doivent être un calque au-dessus de l’image.

## Fonctionnalités du MVP

### 1. Affichage d’une image

L’application doit permettre d’afficher une image WebP.

Pour le MVP, prévoir une image par défaut dans :

```txt
public/maps/atlantis-map.webp
```

L’image doit être affichée dans un conteneur responsive.

Au-dessus de l’image, afficher un SVG transparent servant de calque de dessin.

Le SVG doit utiliser un `viewBox` basé sur les dimensions naturelles de l’image, par exemple :

```html
<svg viewBox="0 0 imageWidth imageHeight">
```

L’objectif est que les coordonnées restent valables même si l’image est affichée à une taille différente.

### 2. Outils disponibles

L’interface doit proposer un panneau d’outils avec au minimum les modes suivants :

```ts
type EditorTool =
  | "select"
  | "draw-polygon"
  | "move-zone"
  | "edit-points";
```

Le panneau d’outils doit permettre de :

- sélectionner une zone existante
- créer une zone polygonale avec l’outil de cerclage de zone
- déplacer une zone sélectionnée
- déplacer les points d’une zone existante
- copier une zone sélectionnée
- coller une zone copiée pour en créer une nouvelle
- appliquer une symétrie verticale à une zone sélectionnée
- appliquer une symétrie horizontale à une zone sélectionnée

Prévoir aussi les boutons globaux :

- “Sauvegarder”
- “Recharger”
- “Réinitialiser”
- “Tout sélectionner”
- “Désélectionner”
- “Supprimer la sélection”

### 3. Outil de cerclage de zone polygonale

Le MVP doit gérer des polygones, pas seulement des rectangles.

L’outil principal de création de zone est `draw-polygon`.

Interaction attendue :

1. L’utilisateur sélectionne l’outil de cerclage de zone.
2. Premier clic gauche sur l’image : crée le premier point du polygone.
3. Les clics gauche suivants : ajoutent les points suivants.
4. Le clic du milieu de la souris termine la sélection de la zone.
5. `Ctrl + clic gauche` efface le dernier point du polygone en cours de création.

Comportement attendu :

- un polygone en cours de dessin doit être affiché visuellement
- les points déjà créés doivent être visibles
- les segments entre les points doivent être visibles
- lorsque le curseur bouge, afficher éventuellement un segment temporaire entre le dernier point et le curseur
- le clic du milieu doit finaliser le polygone uniquement s’il contient au moins 3 points
- si le polygone contient moins de 3 points, le clic du milieu ne doit pas créer de zone
- après finalisation, créer une `ZoneDto` de type `"polygon"`
- après finalisation, vider l’état du polygone temporaire
- empêcher le comportement par défaut du clic du milieu dans le navigateur si nécessaire

Important :

- utiliser les événements souris adaptés pour gérer le clic du milieu, par exemple `onMouseDown`, `onPointerDown` ou `onAuxClick`
- vérifier `event.button === 1` pour le clic du milieu
- vérifier `event.button === 0` pour le clic gauche
- vérifier `event.ctrlKey` pour `Ctrl + clic gauche`
- sur macOS, prévoir aussi `event.metaKey` comme équivalent de Ctrl si pertinent

Exemple de stockage d’un polygone :

```ts
points: [
  { x: 420, y: 650 },
  { x: 610, y: 610 },
  { x: 780, y: 700 },
  { x: 720, y: 820 },
  { x: 500, y: 790 }
]
```

### 4. Rendu visuel des zones

Chaque zone doit être affichée au-dessus de l’image avec :

- remplissage vert semi-transparent
- contour vert
- contour éventuellement en pointillés ou avec des points
- label visible au centre approximatif de la zone ou proche du premier point

Le rendu doit évoquer une zone sélectionnée comme sur une carte tactique.

Exemple de style :

```ts
fill: "rgba(0, 255, 150, 0.25)"
stroke: "#00e676"
strokeWidth: 4
```

Pour le rendu SVG :

- utiliser `<polygon>` pour les zones finalisées
- utiliser `<polyline>` ou `<polygon>` pour la zone temporaire en cours de dessin
- utiliser des `<circle>` pour afficher les points temporaires ou les points éditables

### 5. Liste des zones

Afficher un panneau latéral ou inférieur avec la liste des zones créées.

Pour chaque zone, afficher :

- son ID
- son label
- le nombre de points
- un indicateur visuel si la zone est sélectionnée
- un bouton supprimer

### 6. Sélection simple d’une zone

Quand l’utilisateur clique sur une zone avec l’outil `select` :

- la zone devient sélectionnée
- toutes les autres zones sont désélectionnées
- son style change légèrement
- ses coordonnées sont affichées dans un panneau de détail

Le panneau de détail doit afficher le JSON de la zone sélectionnée.

### 7. Sélection multiple de zones

Ajouter le support de sélection de plusieurs zones.

Interactions attendues :

#### Sélection/désélection avec touche modificatrice

L’utilisateur doit pouvoir sélectionner plusieurs zones avec :

- `Ctrl + clic` sur Windows/Linux
- `Cmd + clic` sur macOS
- idéalement `Shift + clic` aussi pour une sélection additive simple

Comportement attendu :

- clic simple sur une zone : sélection unique de cette zone
- `Ctrl/Cmd + clic` sur une zone non sélectionnée : ajoute la zone à la sélection
- `Ctrl/Cmd + clic` sur une zone déjà sélectionnée : retire la zone de la sélection
- `Shift + clic` sur une zone : ajoute la zone à la sélection
- clic sur le fond de l’image : désélectionne tout, sauf si l’utilisateur est en train de dessiner une nouvelle zone

#### État de sélection

Ne pas stocker `selected` dans les `ZoneDto`.

La sélection est un état purement UI, par exemple :

```ts
const [selectedZoneIds, setSelectedZoneIds] = useState<string[]>([]);
```

ou :

```ts
const [selectedZoneIds, setSelectedZoneIds] = useState<Set<string>>(new Set());
```

Le modèle persisté des zones ne doit pas contenir l’état de sélection.

#### Rendu visuel de la sélection multiple

Chaque zone sélectionnée doit avoir un style distinct :

- contour plus épais
- couleur légèrement plus lumineuse
- éventuellement un halo ou une ombre SVG
- badge ou label indiquant qu’elle est sélectionnée

Les zones non sélectionnées gardent le style normal.

Si plusieurs zones sont sélectionnées, elles doivent toutes être visibles comme sélectionnées en même temps.

#### Panneau de détails en sélection multiple

Le panneau `ZoneDetails` doit gérer trois cas :

1. aucune zone sélectionnée  
   Afficher : “Aucune zone sélectionnée.”

2. une seule zone sélectionnée  
   Afficher le JSON complet de la zone.

3. plusieurs zones sélectionnées  
   Afficher :
   - le nombre de zones sélectionnées
   - la liste des IDs sélectionnés
   - le JSON des zones sélectionnées ou un résumé lisible
   - un bouton “Supprimer la sélection”

Exemple :

```txt
3 zones sélectionnées

- zone-001
- zone-002
- zone-003
```

#### Liste des zones et sélection multiple

Dans `ZoneList`, l’utilisateur doit pouvoir :

- cliquer sur une zone pour faire une sélection unique
- faire `Ctrl/Cmd + clic` pour ajouter/retirer une zone de la sélection
- voir clairement quelles zones sont sélectionnées
- supprimer une zone individuellement
- supprimer toutes les zones sélectionnées

### 8. Déplacement d’une zone sélectionnée

Ajouter un mode `move-zone`.

Comportement attendu :

- l’utilisateur sélectionne une zone existante
- il active l’outil `move-zone`
- il clique-glisse sur la zone sélectionnée
- toute la zone se déplace
- tous les points de la zone sont translatés du même delta `{ dx, dy }`
- le déplacement doit rester exprimé dans le repère naturel de l’image
- à la fin du déplacement, `updatedAt` est mis à jour

Si plusieurs zones sont sélectionnées :

- le déplacement peut déplacer toutes les zones sélectionnées ensemble
- si cette fonctionnalité est trop coûteuse pour le MVP, documenter clairement dans le code que seul le déplacement d’une zone à la fois est supporté
- de préférence, supporter le déplacement groupé des zones sélectionnées

Important :

- ne pas modifier l’image
- ne pas stocker une position séparée de la zone
- déplacer une zone revient à modifier les coordonnées de tous ses points

Exemple :

```ts
function translateZone(zone: ZoneDto, dx: number, dy: number): ZoneDto {
  return {
    ...zone,
    points: zone.points.map(point => ({
      x: point.x + dx,
      y: point.y + dy
    })),
    updatedAt: new Date().toISOString()
  };
}
```

### 9. Déplacement des points d’une zone existante

Ajouter un mode `edit-points`.

Comportement attendu :

- l’utilisateur sélectionne une zone existante
- il active l’outil `edit-points`
- les sommets de la zone sélectionnée apparaissent sous forme de poignées visuelles
- l’utilisateur peut cliquer-glisser un point
- seul ce point se déplace
- la forme du polygone est mise à jour en temps réel
- à la fin du déplacement, `updatedAt` est mis à jour

Si plusieurs zones sont sélectionnées :

- afficher les points de toutes les zones sélectionnées
- chaque point doit pouvoir être déplacé individuellement
- chaque poignée doit permettre d’identifier la zone et l’index du point

Exemple de callback possible :

```ts
onPointDrag(zoneId: string, pointIndex: number, newPoint: Point): void
```

### 10. Copier/coller une zone existante

Ajouter la possibilité de copier/coller une zone existante.

Comportement attendu :

- l’utilisateur sélectionne une zone
- il clique sur “Copier”
- l’application stocke une copie de la zone dans un état UI interne, pas dans le backend
- l’utilisateur clique sur “Coller”
- une nouvelle zone est créée à partir de la zone copiée
- la nouvelle zone doit avoir :
  - un nouvel ID
  - un nouveau label, par exemple `Zone 4` ou `Copie de Zone 2`
  - des dates `createdAt` et `updatedAt` mises à jour
  - les mêmes points, mais légèrement décalés pour être visible
- la nouvelle zone est ajoutée à la liste des zones
- la nouvelle zone devient sélectionnée

Décalage recommandé au collage :

```ts
const pasteOffset = { dx: 20, dy: 20 };
```

Si plusieurs zones sont sélectionnées :

- pour le MVP, copier/coller la première zone sélectionnée suffit
- si le code reste simple, permettre de copier/coller plusieurs zones sélectionnées

La donnée copiée doit rester côté UI. Elle ne doit pas être persistée tant que l’utilisateur ne clique pas sur “Sauvegarder”.

### 11. Symétrie verticale ou horizontale d’une zone sélectionnée

Lorsque l’on sélectionne une zone, il faut pouvoir appliquer une symétrie verticale ou horizontale.

Ajouter deux actions dans le panneau d’outils ou dans le panneau de détails :

- “Symétrie horizontale”
- “Symétrie verticale”

Définition attendue :

- symétrie horizontale : miroir haut/bas par rapport à l’axe horizontal passant par le centre de la bounding box de la zone
- symétrie verticale : miroir gauche/droite par rapport à l’axe vertical passant par le centre de la bounding box de la zone

Pour calculer la bounding box :

```ts
const minX = Math.min(...points.map(p => p.x));
const maxX = Math.max(...points.map(p => p.x));
const minY = Math.min(...points.map(p => p.y));
const maxY = Math.max(...points.map(p => p.y));

const centerX = (minX + maxX) / 2;
const centerY = (minY + maxY) / 2;
```

Symétrie verticale :

```ts
const mirroredX = centerX - (point.x - centerX);
// ou
const mirroredX = 2 * centerX - point.x;
```

Symétrie horizontale :

```ts
const mirroredY = centerY - (point.y - centerY);
// ou
const mirroredY = 2 * centerY - point.y;
```

Important :

- la symétrie doit modifier les coordonnées des points
- elle ne doit pas modifier l’image
- elle doit mettre à jour `updatedAt`
- si plusieurs zones sont sélectionnées, appliquer la symétrie à chaque zone individuellement, chacune autour de sa propre bounding box
- conserver l’ordre des points

Créer des utilitaires dédiés, par exemple :

```ts
mirrorZoneHorizontally(zone: ZoneDto): ZoneDto
mirrorZoneVertically(zone: ZoneDto): ZoneDto
```

### 12. Suppression d’une ou plusieurs zones

L’utilisateur doit pouvoir supprimer :

- une zone depuis la liste
- une zone sélectionnée depuis le panneau de détail
- plusieurs zones sélectionnées avec “Supprimer la sélection”

La suppression doit mettre à jour :

- l’état local des zones
- l’état local de sélection
- la persistance lors de la sauvegarde

Pour le MVP, la suppression peut rester locale jusqu’au clic sur “Sauvegarder”, mais il faut que l’UX soit cohérente.

### 13. Persistance mockée

Créer une librairie interne qui simule l’API backend.

Elle doit être placée dans un dossier du type :

```txt
src/lib/api/
```

Avec une interface publique claire, par exemple :

```ts
export interface MapsApi {
  getMap(mapId: string): Promise<MapDto>;
  getZones(mapId: string): Promise<ZoneDto[]>;
  saveZones(mapId: string, zones: ZoneDto[]): Promise<ZoneDto[]>;
  deleteZone(mapId: string, zoneId: string): Promise<void>;
}
```

Implémenter ensuite un mock :

```txt
src/lib/api/mockMapsApi.ts
```

Le mock doit :

- utiliser `localStorage`
- retourner des Promises pour simuler des appels asynchrones
- ajouter un petit délai artificiel de 100 à 300 ms
- exposer exactement la même interface que le futur backend

L’application ne doit jamais utiliser directement `localStorage` dans les composants React. Toute la persistance doit passer par la librairie `api`.

### 14. Modèle de données

Créer des types TypeScript propres.

Exemple :

```ts
export type Point = {
  x: number;
  y: number;
};

export type ZoneShape = "polygon";

export type ZoneDto = {
  id: string;
  mapId: string;
  label: string;
  type: ZoneShape;
  points: Point[];
  style?: ZoneStyle;
  createdAt: string;
  updatedAt: string;
};

export type ZoneStyle = {
  fill: string;
  stroke: string;
  strokeWidth: number;
  strokeDasharray?: string;
};

export type MapDto = {
  id: string;
  name: string;
  imageUrl: string;
  width: number;
  height: number;
};
```

Important : ne pas ajouter `selected`, `isSelected`, `highlighted` ou autre état UI dans `ZoneDto`.

### 15. Conversion coordonnées souris → coordonnées image

Créer une fonction utilitaire dédiée, par exemple :

```ts
screenPointToImagePoint(event, svgElement, imageWidth, imageHeight)
```

Elle doit convertir les coordonnées écran en coordonnées dans le repère original de l’image.

Principe :

```ts
const rect = svg.getBoundingClientRect();

const x = (event.clientX - rect.left) * (imageWidth / rect.width);
const y = (event.clientY - rect.top) * (imageHeight / rect.height);
```

Arrondir les coordonnées à l’entier ou à une décimale.

Créer également un utilitaire pour les deltas de déplacement :

```ts
screenDeltaToImageDelta(dxScreen, dyScreen, svgElement, imageWidth, imageHeight)
```

ou calculer les deltas à partir de deux points convertis dans le repère image.

### 16. Utilitaires géométriques attendus

Créer des utilitaires dédiés dans :

```txt
src/lib/geometry/
```

Prévoir au minimum :

```ts
screenPointToImagePoint(...)
createZoneFromPolygonPoints(...)
translateZone(zone, dx, dy)
moveZonePoint(zone, pointIndex, newPoint)
getZoneBoundingBox(zone)
mirrorZoneHorizontally(zone)
mirrorZoneVertically(zone)
offsetZone(zone, dx, dy)
getPolygonCentroid(points)
```

Ces fonctions doivent être testables indépendamment des composants React.

### 17. Organisation de code souhaitée

Proposer une structure de projet propre, par exemple :

```txt
src/
  App.tsx
  main.tsx

  components/
    MapEditor/
      MapEditor.tsx
      MapCanvas.tsx
      ZoneOverlay.tsx
      ZoneList.tsx
      ZoneDetails.tsx
      Toolbar.tsx
      PointHandles.tsx

  lib/
    api/
      mapsApi.ts
      mockMapsApi.ts
      types.ts

    geometry/
      coordinates.ts
      zoneFactory.ts
      transforms.ts

  styles/
    global.css
```

### 18. Composants attendus

#### `MapEditor`

Composant principal.

Responsabilités :

- charger la map
- charger les zones
- gérer l’état des zones
- gérer la sauvegarde
- gérer l’outil actif
- gérer la sélection simple et multiple
- maintenir `selectedZoneIds`
- maintenir le polygone temporaire en cours de dessin
- maintenir l’état de copie interne
- gérer “tout sélectionner”
- gérer “désélectionner”
- gérer “supprimer la sélection”
- gérer “copier”
- gérer “coller”
- gérer “symétrie horizontale”
- gérer “symétrie verticale”

#### `MapCanvas`

Responsabilités :

- afficher l’image
- afficher le SVG overlay
- gérer le dessin du polygone
- gérer le clic du milieu pour terminer une zone
- gérer `Ctrl + clic gauche` pour supprimer le dernier point temporaire
- gérer le clic sur zone
- gérer le clic sur fond
- gérer le drag d’une zone en mode `move-zone`
- gérer le drag de points en mode `edit-points`
- appeler les callbacks nécessaires :
  - `onPolygonPointAdded`
  - `onPolygonLastPointRemoved`
  - `onPolygonCompleted`
  - `onZoneClick`
  - `onBackgroundClick`
  - `onZoneDrag`
  - `onPointDrag`

Le callback de clic sur zone doit recevoir l’événement ou au minimum des informations permettant de savoir si `Ctrl`, `Cmd` ou `Shift` était pressé :

```ts
onZoneClick(zoneId: string, options: {
  additive: boolean;
  toggle: boolean;
}): void
```

ou bien :

```ts
onZoneClick(zoneId: string, event: React.MouseEvent): void
```

#### `ZoneOverlay`

Responsabilités :

- afficher toutes les zones
- afficher la zone temporaire pendant le dessin
- appliquer le style sélectionné/non sélectionné
- accepter `selectedZoneIds`
- permettre que plusieurs zones soient sélectionnées visuellement en même temps
- afficher les labels des zones

#### `PointHandles`

Responsabilités :

- afficher les poignées des points en mode `edit-points`
- afficher les poignées uniquement pour les zones sélectionnées
- gérer le drag de chaque point
- remonter `zoneId`, `pointIndex` et les nouvelles coordonnées

#### `ZoneList`

Responsabilités :

- afficher la liste des zones
- sélectionner une zone
- ajouter/retirer une zone de la sélection via `Ctrl/Cmd + clic`
- afficher les zones sélectionnées
- supprimer une zone
- supprimer la sélection

#### `ZoneDetails`

Responsabilités :

- afficher le JSON de la zone sélectionnée si une seule zone est sélectionnée
- afficher un résumé et/ou le JSON des zones sélectionnées si plusieurs zones sont sélectionnées
- afficher un état vide si aucune zone n’est sélectionnée
- proposer une action “Supprimer la sélection” si au moins une zone est sélectionnée
- proposer les actions de transformation sur la sélection :
  - symétrie horizontale
  - symétrie verticale
  - copier
  - coller si une zone a été copiée

#### `Toolbar`

Responsabilités :

- afficher et changer l’outil actif
- sauvegarder
- recharger
- reset
- tout sélectionner
- désélectionner
- supprimer la sélection
- copier
- coller
- symétrie horizontale
- symétrie verticale
- afficher le nombre de zones sélectionnées
- afficher le nombre de points du polygone temporaire si l’outil `draw-polygon` est actif

## UX minimale attendue

L’application doit être utilisable sans explication.

Prévoir une interface simple :

```txt
+----------------------------------------------------------------------------------+
| Tool: Select | Draw polygon | Move zone | Edit points                             |
| Save | Reload | Reset | Select all | Clear selection | Copy | Paste | Flip H | Flip V |
+-------------------------------------------+--------------------------------------+
|                                           | Zones                                |
|        Image + SVG overlay                 | - Zone 1                             |
|                                           | - Zone 2                             |
|                                           |                                      |
|                                           | Détails / sélection                  |
+-------------------------------------------+--------------------------------------+
```

Sur mobile ou petit écran, le panneau peut passer sous l’image.

### Indications utilisateur

Afficher une courte aide contextuelle selon l’outil actif.

Exemples :

- `draw-polygon` : “Clic gauche : ajouter un point. Ctrl+clic : supprimer le dernier point. Clic du milieu : terminer la zone.”
- `select` : “Clique sur une zone pour la sélectionner. Ctrl/Cmd+clic pour sélectionner plusieurs zones.”
- `move-zone` : “Glisse une zone sélectionnée pour la déplacer.”
- `edit-points` : “Glisse les points visibles pour modifier la forme.”

## Critères d’acceptation

Le MVP est considéré comme terminé si :

1. L’image WebP s’affiche correctement.
2. L’overlay SVG est parfaitement aligné avec l’image.
3. Je peux sélectionner l’outil de cerclage de zone.
4. Le premier clic gauche crée le premier point d’un polygone temporaire.
5. Les clics gauches suivants ajoutent des points au polygone temporaire.
6. `Ctrl + clic gauche` supprime le dernier point du polygone temporaire.
7. Le clic du milieu termine la zone si elle contient au moins 3 points.
8. La zone créée est stockée comme un polygone.
9. La zone est affichée avec un style vert semi-transparent.
10. Je peux créer plusieurs zones polygonales.
11. Je peux sélectionner une zone avec un clic simple.
12. Je peux sélectionner plusieurs zones avec `Ctrl/Cmd + clic`.
13. Je peux ajouter une zone à la sélection avec `Shift + clic`.
14. Je peux retirer une zone de la sélection avec `Ctrl/Cmd + clic` si elle est déjà sélectionnée.
15. Je peux voir plusieurs zones sélectionnées visuellement en même temps.
16. Je peux voir le JSON complet quand une seule zone est sélectionnée.
17. Je peux voir un résumé ou le JSON des zones sélectionnées quand plusieurs zones sont sélectionnées.
18. Je peux désélectionner toutes les zones.
19. Je peux sélectionner toutes les zones.
20. Je peux supprimer une zone.
21. Je peux supprimer toutes les zones sélectionnées.
22. Je peux déplacer une zone sélectionnée.
23. Je peux déplacer plusieurs zones sélectionnées ensemble, ou le code documente explicitement que seul le déplacement d’une zone est supporté dans le MVP.
24. Je peux déplacer les points d’une zone sélectionnée.
25. Je peux copier une zone sélectionnée.
26. Je peux coller une zone copiée pour créer une nouvelle zone.
27. La zone collée possède un nouvel ID et est légèrement décalée.
28. Je peux appliquer une symétrie horizontale à une zone sélectionnée.
29. Je peux appliquer une symétrie verticale à une zone sélectionnée.
30. La symétrie modifie les coordonnées des points et non l’image.
31. Je peux sauvegarder les zones.
32. Je peux recharger la page et retrouver les zones sauvegardées.
33. Les composants React ne connaissent pas `localStorage`.
34. Toute la persistance passe par la librairie interne `mapsApi`.
35. Le code est typé proprement en TypeScript.
36. L’API mockée pourra être remplacée plus tard par une vraie API HTTP sans réécrire les composants.
37. L’état de sélection n’est jamais persisté dans `ZoneDto`.
38. Les coordonnées sont toujours stockées dans le repère naturel de l’image, pas dans le repère écran.

## Contraintes importantes

- Ne pas coder de backend.
- Ne pas dépendre du format interne d’une librairie de dessin.
- Les coordonnées doivent être stockées dans le repère naturel de l’image, pas dans le repère écran.
- Le système doit rester simple, lisible et facilement extensible.
- Prévoir des noms de fonctions et composants explicites.
- Ajouter quelques commentaires uniquement là où la logique de coordonnées peut être ambiguë.
- Garder la sélection comme un état UI séparé des données métier persistées.
- Garder la zone copiée comme un état UI séparé des données métier persistées.
- Éviter les optimisations prématurées : l’objectif est un MVP clair et fonctionnel.
- Gérer proprement les événements souris pour éviter les conflits entre dessin, sélection, déplacement de zone et déplacement de points.
- En mode `draw-polygon`, les clics sur le fond servent à créer des points, pas à désélectionner.
- En mode `select`, les clics sur le fond peuvent désélectionner.

## Livrable attendu

Génère le code complet du MVP.

Inclure :

- les fichiers React/TypeScript nécessaires
- les types
- le mock d’API
- les composants
- les utilitaires de coordonnées
- les utilitaires de transformation géométrique
- un style minimal propre
- les instructions pour lancer le projet avec `npm install` puis `npm run dev`

Commence par créer une version simple mais réellement fonctionnelle.
