# Tâche #004 - Story #009 : API Client Zones et types frontend

## Objectif

Créer le client API frontend pour les zones et les locks, mettre à jour les types TypeScript pour supporter le modèle multi-polygones, et remplacer les appels mock par des appels API réels dans l'éditeur.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Tâche #002 (API endpoints backend opérationnels)
- Nécessaire pour : Tâche #005 (éditeur multi-polygones fonctionnel)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Mise à jour des types

Mettre à jour `frontend/src/lib/api/types.ts` pour refléter le nouveau modèle multi-polygones :

- Renommer/supprimer l'ancien `ZoneDto` (qui était un polygone unique)
- Créer un nouveau modèle `Zone` qui contient un tableau de polygones
- Mettre à jour `MapDto` pour correspondre au backend (ID numérique, pas de width/height calculés)

**Nouveaux types :**

```typescript
// Point reste inchangé
export type Point = {
  x: number;
  y: number;
};

// Un polygone individuel dans une zone
export type ZonePolygon = {
  id: number;
  zone_id: number;
  points: Point[];
  order_index: number;
  created_at: string;
};

// Une zone complète avec ses polygones et son nom
export type Zone = {
  id: number;
  team_id: number;
  map_id: number;
  polygons: ZonePolygon[];
  name: string;
  created_at: string;
  updated_at: string;
};

// MapDto mis à jour (provenant du backend)
// Note : width/height sont déterminés côté client depuis l'image chargée
// (pas stockés en base). Voir getMapImageUrl().
export type MapDto = {
  id: number;
  name: string;
  added_at: string;
  is_obsolete: boolean;
};
```

#### Partie 2 - Client API Zones

Créer `frontend/src/api/zones.ts` avec toutes les fonctions API pour les zones et les locks.

#### Partie 3 - Client API Locks

Créer `frontend/src/api/locks.ts` avec les fonctions API pour le verrouillage.

#### Partie 4 - Mise à jour de l'éditeur

Dans `frontend/src/components/MapEditor/MapEditor.tsx` :
- Remplacer l'import de `mockMapsApi` par les nouvelles fonctions API
- Adapter les appels `getMap`, `getZones`, `saveZones`, `deleteZone`
- Initialiser le chargement avec l'ID de carte réel (depuis l'URL ou props)
- Ajouter la gestion d'erreurs avec affichage utilisateur (toast/notification)

**Cas nominaux :**
- Chargement de l'éditeur → appel GET zones → affichage des zones
- Création d'une zone → appel POST zone → ajout à l'état local
- Suppression d'une zone → appel DELETE zone → retrait de l'état local

**Cas limites :**
- Erreur API 401 → redirection vers login (déjà géré par le client)
- Erreur API 403 → message "permission insuffisante"
- Erreur API 423 → message "zone verrouillée par un autre éditeur"
- Erreur API 409 → message "conflit" (nom déjà pris, lock déjà acquis)
- Erreur réseau → message "impossible de contacter le serveur"
- Chargement en cours → afficher un spinner

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/api/types.ts` | Modifier | Mettre à jour les types Zone, ZonePolygon, MapDto |
| `frontend/src/api/zones.ts` | Créer | Fonctions API pour les zones |
| `frontend/src/api/locks.ts` | Créer | Fonctions API pour les locks |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Remplacer mockMapsApi par les vrais appels API |
| `frontend/src/lib/api/mockMapsApi.ts` | Supprimer | Plus nécessaire (ou garder pour tests) |

### Signatures

```typescript
// frontend/src/api/zones.ts

export interface CreateZonePayload {
  polygons?: { points: Point[] }[];
  name?: string;
}

export interface CreateZoneResponse {
  zone: Zone;
}

export interface ListZonesResponse {
  zones: Zone[];
}

export interface AddPolygonPayload {
  points: Point[];
}

export interface AddPolygonResponse {
  polygon: ZonePolygon;
}

export interface RemovePolygonResponse {
  polygon_deleted: boolean;
  zone_deleted: boolean;
}

export interface UpdatePointsPayload {
  points: Point[];
}

export interface AddPointPayload {
  point: Point;
  index_after: number;
}

export interface SetNamePayload {
  name: string;
}

// Fonctions
export async function listZones(teamId: number, mapId: number): Promise<ListZonesResponse>
export async function getZone(teamId: number, zoneId: number): Promise<{ zone: Zone }>
export async function createZone(teamId: number, mapId: number, payload: CreateZonePayload): Promise<CreateZoneResponse>
export async function deleteZone(teamId: number, zoneId: number): Promise<void>
export async function addPolygon(teamId: number, zoneId: number, payload: AddPolygonPayload): Promise<AddPolygonResponse>
export async function removePolygon(teamId: number, zoneId: number, polygonId: number): Promise<RemovePolygonResponse>
export async function updatePolygonPoints(teamId: number, zoneId: number, polygonId: number, payload: UpdatePointsPayload): Promise<AddPolygonResponse>
export async function addPolygonPoint(teamId: number, zoneId: number, polygonId: number, payload: AddPointPayload): Promise<AddPolygonResponse>
export async function removePolygonPoint(teamId: number, zoneId: number, polygonId: number, pointIndex: number): Promise<AddPolygonResponse>
export async function setZoneName(teamId: number, zoneId: number, payload: SetNamePayload): Promise<{ name: { id: number; zone_id: number; name: string } }>

// Types pour le batch d'actions (stratégie batch queue)
export type BatchAction =
  | { type: "create_zone"; temp_id: string; polygons: { points: Point[] }[]; name?: string }
  | { type: "add_polygon"; zone_id: number; points: Point[] }
  | { type: "remove_polygon"; polygon_id: number }
  | { type: "update_points"; polygon_id: number; points: Point[] }
  | { type: "add_point"; polygon_id: number; point: Point; index_after: number }
  | { type: "remove_point"; polygon_id: number; point_index: number }
  | { type: "rename"; zone_id: number; name: string }
  | { type: "delete_zone"; zone_id: number };

export interface BatchResult {
  results: unknown[]; // tableau hétérogène, une entrée par action dans l'ordre
}

export async function batchActions(teamId: number, mapId: number, actions: BatchAction[]): Promise<BatchResult>
```

```typescript
// frontend/src/api/locks.ts

export interface LockStatus {
  locked: boolean;
  lock: {
    id: number;
    team_id: number;
    map_id: number;
    editor_id: number;
    locked_at: string;
    expires_at: string;
  } | null;
}

export async function acquireLock(teamId: number, mapId: number): Promise<{ lock: LockStatus['lock'] }>
export async function releaseLock(teamId: number, mapId: number): Promise<void>
export async function refreshLock(teamId: number, mapId: number): Promise<{ lock: LockStatus['lock'] }>
export async function getLockStatus(teamId: number, mapId: number): Promise<LockStatus>

// URL de l'image de la carte (existe déjà dans maps.ts)
// export function getMapImageUrl(mapId: number): string
```

### Contraintes techniques

- **API Client** : Utiliser `getApiClient()` depuis `frontend/src/api/client.ts` qui gère déjà l'authentification JWT et le refresh token.
- **Gestion d'erreurs** : Les fonctions API doivent laisser les exceptions `ApiError` remonter. L'éditeur les attrapera et les affichera.
- **Suppression mock** : Ne pas supprimer `mockMapsApi.ts` immédiatement — le marquer comme déprécié (`@deprecated`) pour permettre la transition. Il pourra être supprimé dans une tâche ultérieure.
- **Types** : Les types doivent correspondre exactement à la réponse JSON du backend (Task #002). Vérifier la casse (snake_case pour les clés JSON).
- **Compatibilité éditeur** : L'éditeur actuel utilise `ZoneDto { id, mapId, label, type, points[], style, createdAt, updatedAt }`. Le nouveau modèle `Zone` avec `polygons[]` et `name` est différent. L'adaptation de l'éditeur sera faite dans la Tâche #005. Cette tâche se concentre sur la création du client API et la mise à jour des types.

### Tests

#### Tests unitaires
- **Fichier** : `frontend/src/api/zones.test.ts`
- Scénario 1 : `listZones` avec réponse succès → retourne la liste des zones
  - Mock : `apiClient.get` retourne `{ zones: [...] }`
  - Résultat attendu : tableau de `Zone`
- Scénario 2 : `createZone` avec payload → appelle POST avec le bon body
  - Mock : `apiClient.post` est appelé avec les bons arguments
  - Résultat attendu : `{ zone: Zone }`

### Documentation

Aucune documentation à créer pour cette tâche.
