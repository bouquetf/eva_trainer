# Tâche #005 - Story #009 : Éditeur multi-polygones avec nommage et sauvegarde automatique (frontend)

## Objectif

Mettre à jour le MapEditor frontend pour supporter les zones multi-polygones, l'attribution d'un nom lors de la création, la sauvegarde automatique des modifications, et l'intégration du verrouillage d'édition.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Tâche #004 (nouveaux types et API client), Tâche #002 (API backend)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle — Architecture SVG conservée

**Important :** L'éditeur existant utilise **SVG** (pas React-Konva). `MapCanvas.tsx` superpose un élément `<svg>` avec `viewBox` sur une `<img>`. Les coordonnées sont converties écran → image via `getScreenCTM()` dans `coordinates.ts`. Cette architecture est **conservée**. Aucune migration vers Konva.

#### Partie 1 - Mise à jour du modèle de données interne

L'éditeur utilise actuellement un état `ZoneDto[]` où chaque entrée est un polygone unique avec un label.
Remplacer par un état qui reflète le nouveau modèle `Zone` (une zone = plusieurs polygones + un nom).

**URL de l'éditeur :** `/teams/:id/editor/:mapId` (ex: `/teams/1/editor/3`)
- `:id` = team ID
- `:mapId` = map ID (numérique)

**Chargement initial de l'éditeur :**
1. Récupérer `teamId` et `mapId` depuis les paramètres de route
2. Acquérir le lock via `acquireLock(teamId, mapId)`
3. Charger les zones via `listZones(teamId, mapId)`
4. Charger l'image de la carte via `getMapImageUrl(mapId)` dans un `<img>` caché
5. À la fin du chargement de l'image, lire `naturalWidth` et `naturalHeight` pour définir le `viewBox` SVG
6. Stocker les dimensions dans l'état local (`imageWidth`, `imageHeight`)

**État interne :**
```typescript
// Dans MapEditor.tsx
const [zones, setZones] = useState<Zone[]>([]);
const [mapInfo, setMapInfo] = useState<{ id: number; name: string } | null>(null);
const [imageSize, setImageSize] = useState<{ width: number; height: number } | null>(null);
```

**Détermination des dimensions du viewBox (côté client) :**
```typescript
const loadImageDimensions = (url: string): Promise<{ width: number; height: number }> => {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve({ width: img.naturalWidth, height: img.naturalHeight });
    img.onerror = reject;
    img.src = url;
  });
};
```

**Migration des données :**
- Les zones existantes dans localStorage (MockMapsApi) sont dans l'ancien format `ZoneDto[]`. À l'initialisation, les données viennent maintenant de l'API réelle (Task #004), donc plus de problème de format legacy.

#### Partie 2 - Rendu multi-polygones (SVG)

Mettre à jour `ZoneOverlay.tsx` pour rendre chaque zone comme un groupe de polygones SVG :

- Chaque `Zone` contient `polygons[]`. Rendre chaque polygone via un `<polygon>` SVG.
- Chaque polygone doit être cliquable et sélectionnable.
- Le nom de la zone est affiché via `<text>` au centroïde moyen de tous ses polygones.
- Adapter `PointHandles.tsx` pour qu'il fonctionne avec `ZonePolygon[]` : chaque poignée de point doit savoir à quel polygone elle appartient (`zoneId + polygonId`).

```typescript
// Structure de rendu dans ZoneOverlay (SVG, pas Konva)
<g key={zone.id}>
  {zone.polygons.map((polygon, idx) => (
    <polygon key={polygon.id || idx} points={polygon.points.map(p => `${p.x},${p.y}`).join(' ')} />
  ))}
  <text x={centroid.x} y={centroid.y} textAnchor="middle" dominantBaseline="middle">
    {zone.name}
  </text>
</g>
```

#### Partie 3 - Nommage de la zone lors de la création

Quand l'utilisateur termine un polygone (double-clic ou clic droit "Terminer la forme") :

1. Si aucun polygone n'existe encore (création d'une nouvelle zone) :
   - Afficher une boîte de dialogue modale ou un champ dans le panneau latéral pour saisir le nom
   - Le champ doit avoir une validation : max 250 caractères, non vide
   - Valider l'unicité via API (`setZoneName` attrape l'erreur 409)
   - Si l'utilisateur ferme sans nommer, le polygone temporaire est conservé mais pas sauvegardé en base (conformément à la spec : "une nouvelle zone n'est sauvegardée que si elle contient aussi un nom")

2. Si la zone existe déjà (ajout d'un polygone à une zone existante) :
   - Sauvegarder immédiatement le nouveau polygone via `addPolygon(teamId, zoneId, payload)`

#### Partie 4 - Sauvegarde automatique (stratégie batch queue)

Supprimer le bouton "Save" manuel. Remplacer par une file d'attente de batch.

**Principe général :**
1. Chaque action utilisateur (add polygon, remove polygon, add point, rename, etc.) est empaquetée comme une **action** et ajoutée à une pile côté client.
2. Dès qu'au moins une action est dans la pile, un envoi batch est déclenché (après un court délai d'attente de 300ms pour grouper les actions rapides).
3. Le batch est envoyé à l'API via `POST /api/v1/teams/:id/maps/:mapId/zones/batch`.
4. Si le batch réussit :
   - Mettre à jour l'état local avec les IDs réels retournés par le serveur (mapping `temp_id` → `zone_id`)
   - Si de nouvelles actions ont été ajoutées pendant l'envoi, les envoyer immédiatement
5. Si le batch échoue :
   - Afficher une notification d'erreur : "Échec de la sauvegarde, nouvelle tentative dans 5s"
   - Re-tenter automatiquement après 5 secondes (sans perdre les actions en attente)
   - Si la re-tentative réussit : notification de succès "Sauvegarde rétablie"
   - Si la re-tentative échoue après 5 essais : notification persistante d'erreur critique

**Structure d'une action :**
```typescript
type BatchAction =
  | { type: "create_zone"; temp_id: string; polygons: { points: Point[] }[]; name?: string }
  | { type: "add_polygon"; zone_id: number; points: Point[] }
  | { type: "remove_polygon"; polygon_id: number }
  | { type: "update_points"; polygon_id: number; points: Point[] }
  | { type: "add_point"; polygon_id: number; point: Point; index_after: number }
  | { type: "remove_point"; polygon_id: number; point_index: number }
  | { type: "rename"; zone_id: number; name: string }
  | { type: "delete_zone"; zone_id: number }
```

**File d'attente :**
```typescript
const pendingActions = useRef<BatchAction[]>([]);
const isSending = useRef(false);
const retryCount = useRef(0);
```

**Déclenchement :**
- Après chaque ajout d'action, lancer un `setTimeout` de 300ms.
- Si aucune nouvelle action n'arrive pendant ce délai, envoyer le batch.
- Si des actions arrivent pendant l'envoi, les accumuler et les envoyer après le succès du batch en cours.

**Mapping temp_id → zone_id :**
- Le backend retourne un tableau `results` où chaque résultat pour une action `create_zone` contient `{ temp_id: string, zone_id: number, ... }`.
- Le frontend met à jour l'état local : remplacer l'ID temporaire par l'ID réel dans toutes les zones et polygones.
- Les actions suivantes qui référencent cette zone doivent utiliser le nouvel ID.

**État de sauvegarde** : Afficher un indicateur visuel global :
- "Sauvegardé" (vert) : pas d'actions en attente, dernier batch OK
- "Sauvegarde..." (orange) : batch en cours d'envoi
- "Échec de sauvegarde" (rouge) : dernière tentative échouée (avec indication "nouvelle tentative dans X s")

#### Partie 5 - Intégration du verrouillage

Au chargement de l'éditeur :

1. Essayer d'acquérir le lock via `acquireLock(teamId, mapId)`
2. Si succès → l'éditeur est en mode édition
3. Si 409 → afficher "Cette carte est en cours d'édition par un autre éditeur" → mode lecture seule
4. Pendant l'édition, rafraîchir le lock toutes les 10 minutes (via `refreshLock`)
5. Au départ de l'éditeur (unmount), libérer le lock via `releaseLock(teamId, mapId)`
6. Si la libération échoue (réseau), le lock expirera automatiquement après 15 minutes

**UI du lock :**
- Afficher un badge "Verrouillé par [éditeur]" si locké par un autre
- Afficher "Édition en cours" si locké par soi-même
- Désactiver les contrôles d'édition si pas de lock

**Cas nominaux :**
- Chargement de l'éditeur → lock acquis → zones chargées → édition possible
- Création d'une zone avec nom → zone créée en base → polygone sauvegardé
- Ajout d'un polygone à une zone existante → polygone sauvegardé immédiatement
- Suppression du dernier polygone → zone supprimée
- Déplacement d'un polygone → points mis à jour après 500ms de debounce

**Cas limites :**
- Création d'une zone sans nom → polygone temporaire conservé (état local uniquement), pas de sauvegarde
- Fermeture du navigateur → lock libéré après timeout (15 min)
- Perte de connexion → erreur affichée, tentative de re-sauvegarde
- Lock perdu pendant l'édition → message d'erreur, passage en lecture seule
- Carte obsolète → pas de création/modification possible (vérifié par le backend, mais aussi par le frontend)

#### Partie 6 - Mise à jour des routes et de la page Editor

1. **Route** : Mettre à jour `frontend/src/routes/index.tsx` :
   - Remplacer `<Route path="/editor" element={<Editor />} />`
   - Par `<Route path="/teams/:id/editor/:mapId" element={<Editor />} />`

2. **Page Editor** : Mettre à jour `frontend/src/pages/Editor.tsx` :
   - Utiliser `useParams()` pour récupérer `id` (teamId) et `mapId`
   - Passer ces paramètres au `MapEditor`

3. **MapEditor** : Accepter `teamId: number` et `mapId: number` en props :
   ```typescript
   interface MapEditorProps {
     teamId: number;
     mapId: number;
   }
   ```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Refonte pour multi-polygones, auto-save, lock |
| `frontend/src/components/MapEditor/ZoneOverlay.tsx` | Modifier | Rendu multi-polygones par zone |
| `frontend/src/components/MapEditor/ZoneList.tsx` | Modifier | Afficher les zones avec nom et nombre de polygones |
| `frontend/src/components/MapEditor/ZoneDetails.tsx` | Modifier | Afficher tous les polygones d'une zone sélectionnée |
| `frontend/src/components/MapEditor/Toolbar.tsx` | Modifier | Supprimer le bouton Save, ajouter indicateur lock |
| `frontend/src/components/MapEditor/MapCanvas.tsx` | Modifier | Ajuster pour le nouveau modèle multi-polygones |
| `frontend/src/components/MapEditor/ZoneNameDialog.tsx` | Créer | Dialogue de saisie du nom de zone |
| `frontend/src/components/MapEditor/LockIndicator.tsx` | Créer | Indicateur de verrouillage |
| `frontend/src/lib/geometry/transforms.ts` | Modifier | Adapter les transformations pour le nouveau modèle |
| `frontend/src/lib/geometry/zoneFactory.ts` | Modifier | Adapter la factory pour le nouveau modèle |
| `frontend/src/routes/index.tsx` | Modifier | Route `/teams/:id/editor/:mapId` |
| `frontend/src/pages/Editor.tsx` | Modifier | Récupérer teamId et mapId depuis useParams() |

### Comportement détaillé de l'auto-save

```
État local ←→ API
  ↓ (debounce 500ms pour drag)
PendingQueue: Map<action_id, Promise>
  ↓
Success → update local state with DB ids
Error → retry (3x), then show error
```

**Gestion des IDs temporaires (pattern existant) :**
- Les nouvelles zones non sauvegardées ont un ID temporaire basé sur le pattern existant : `` `zone-${Math.random().toString(36).substr(2, 9)}` `` (voir `zoneFactory.ts` ligne 11)
- Après la création API, remplacer l'ID temporaire par l'ID numérique réel de la DB (dans le state local)
- Les polygones non sauvegardés ont `id = 0` ou `id = undefined`
- Après la sauvegarde, assigner l'ID réel (`number`) reçu de l'API

### Contraintes techniques

- **Batch queue** : Utiliser une pile (`BatchAction[]`) + un flag `isSending`. Pas de debounce sur les actions individuelles, mais un délai de groupement de 300ms avant d'envoyer le batch.
- **Tentatives** : Retry automatique 5 secondes après un échec. Max 5 tentatives avant erreur critique. Chaque tentative envoie TOUTES les actions en attente (anciennes + nouvelles).
- **État de sauvegarde** : Créer un type `SaveStatus = 'idle' | 'saving' | 'error' | 'critical_error'` et l'utiliser dans l'UI.
- **Drag de polygone** : Les actions de déplacement de points sont ajoutées à la pile comme les autres actions (pas de debounce spécifique). Le délai de groupement de 300ms s'applique.
- **Verrouillage** : Pendant l'envoi d'un batch, l'utilisateur peut continuer à éditer. Les nouvelles actions s'ajoutent à la pile et seront envoyées dans le prochain batch.
- **Dimensions viewBox** : Déterminées côté client via `new Image()` + `naturalWidth/naturalHeight`. Pas de stockage en base. Le MapDto du backend n'a pas besoin de ces champs.
- **URL éditeur** : `/teams/:id/editor/:mapId` avec `:id` = team ID et `:mapId` = map ID numériques. Utiliser `useParams()` de react-router-dom.
- **Lock** : Appeler `acquireLock` au montage et `releaseLock` au démontage du composant. Stocker le `lock.editor_id` dans l'état. Vérifier `lockRepo.IsLockedByEditor` avant chaque modification (via l'API backend, qui renverra 423).
- **Nom obligatoire** : Le bouton ou action de finalisation de la zone ne doit pas être actif tant que le nom n'est pas saisi.
- **Validation nom** : 250 caractères max, non vide. Vérification visuelle immédiate (rouge/vert) avant envoi API.
- **Gestion d'erreur lock** : Si l'acquisition du lock échoue (409), désactiver tous les outils d'édition et afficher un message clair.
- **Refresh lock** : Utiliser `setInterval` avec `refreshLock` toutes les 10 minutes. Ne pas oublier de clear l'interval au unmount.

### Tests

#### Tests d'intégration
- **Fichier** : `frontend/src/components/MapEditor/MapEditor.test.tsx`
- Scénario 1 : L'éditeur charge et acquiert le lock → zones affichées
- Scénario 2 : Ajout d'un polygone à une zone existante → appel API addPolygon
- Scénario 3 : Création d'une zone avec nom → appel API createZone
- Scénario 4 : Suppression du dernier polygone → appel API deleteZone
- Scénario 5 : Lock déjà pris → message d'erreur, contrôles désactivés

### Documentation

- `docs/zone-editor.md` : Créer la documentation utilisateur pour l'éditeur de zones (comment créer, nommer, modifier des zones polygonales).

### Exemples d'API

```http
# Exemple : Acquérir un lock
POST /api/v1/teams/1/maps/2/lock

# Exemple : Créer une zone avec un polygone et un nom
POST /api/v1/teams/1/maps/2/zones
{
  "polygons": [{ "points": [{ "x": 100, "y": 200 }, { "x": 300, "y": 400 }, { "x": 500, "y": 100 }] }],
  "name": "top"
}

# Exemple : Ajouter un polygone à une zone existante
POST /api/v1/teams/1/zones/5/polygons
{
  "points": [{ "x": 100, "y": 200 }, { "x": 300, "y": 400 }, { "x": 500, "y": 100 }]
}

# Exemple : Supprimer un polygone
DELETE /api/v1/teams/1/zones/5/polygons/12

# Exemple : Ajouter un point à un polygone
POST /api/v1/teams/1/zones/5/polygons/12/points
{
  "point": { "x": 250, "y": 300 },
  "index_after": 1
}

# Exemple : Définir le nom d'une zone
PUT /api/v1/teams/1/zones/5/name
{
  "name": "secret"
}
```
