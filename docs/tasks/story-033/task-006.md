# Tâche #006 - Story #033 : Frontend API hooks + types

## Objectif
Créer le module API frontend `frontend/src/api/linesOfFire.ts` avec tous les hooks nécessaires, et les types TypeScript associés.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #003 (handler HTTP), Tâche #005 (endpoint POST révision)
- Nécessaire pour : Tâche #007 (éditeur), Tâche #008 (révision)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Types (`frontend/src/lib/api/types.ts`)**

Ajouter les types suivants :

```typescript
export interface Cone {
  origin: Point
  left_end: Point
  right_end: Point
}

export interface ConeData {
  cones: Cone[]
}

export interface LineOfFire {
  id: number
  team_id: number
  map_id: number
  name: string
  polygon_data: ConeData
  order_index: number
  created_at: string
  updated_at: string
}
```

Mettre à jour `RevisionZone` dans `frontend/src/lib/revision/types.ts` pour ajouter le champ `type` et renommer en `RevisionItem` :
```typescript
// frontend/src/lib/revision/types.ts
export interface RevisionItem {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygons: ZonePolygon[]
  name: string
  type: 'zone' | 'line_of_fire'
}

export interface RevisionFilter {
  zones: boolean
  lines: boolean
}
```

**API hooks (`frontend/src/api/linesOfFire.ts`)**

```typescript
import { getApiClient } from './client'
import type { Point, LineOfFire, Cone } from '../lib/api/types'

export interface CreateLineOfFirePayload {
  name: string
  polygon_data: { cones: Cone[] }
}

export interface CreateLineOfFireResponse {
  line_of_fire: LineOfFire
}

export interface ListLinesOfFireResponse {
  lines_of_fire: LineOfFire[]
}

export type LineOfFireBatchAction =
  | { type: 'create_line_of_fire'; temp_id: string; name: string; polygon_data: { cones: Cone[] } }
  | { type: 'add_cone'; line_of_fire_id: number; cone: Cone }
  | { type: 'update_cone'; line_of_fire_id: number; cone_index: number; cone: Cone }
  | { type: 'remove_cone'; line_of_fire_id: number; cone_index: number }
  | { type: 'rename_line_of_fire'; line_of_fire_id: number; name: string }
  | { type: 'delete_line_of_fire'; line_of_fire_id: number }

// Fonctions
export async function listLinesOfFire(teamId: number, mapId: number): Promise<ListLinesOfFireResponse>
export async function createLineOfFire(teamId: number, mapId: number, payload: CreateLineOfFirePayload): Promise<CreateLineOfFireResponse>
export async function getLineOfFire(teamId: number, lineOfFireId: number): Promise<{ line_of_fire: LineOfFire }>
export async function deleteLineOfFire(teamId: number, lineOfFireId: number): Promise<void>
export async function setLineOfFireName(teamId: number, lineOfFireId: number, name: string): Promise<{ name: { id: number; line_of_fire_id: number; name: string } }>
export async function addLineOfFireCone(teamId: number, lineOfFireId: number, cone: Cone): Promise<void>
export async function updateLineOfFireCone(teamId: number, lineOfFireId: number, coneIndex: number, cone: Cone): Promise<void>
export async function removeLineOfFireCone(teamId: number, lineOfFireId: number, coneIndex: number): Promise<void>
export async function batchLineOfFireActions(teamId: number, mapId: number, actions: LineOfFireBatchAction[]): Promise<{ results: unknown[] }>
```

**Hook révision (`frontend/src/api/revision.ts`)**

Ajouter la fonction suivante :

```typescript
import { getApiClient } from './client'
import type { RevisionItem, RevisionFilter } from '../lib/revision/types'

export interface FetchRevisionItemsPayload {
  zones: boolean
  lines: boolean
}

export interface FetchRevisionItemsResponse {
  items: RevisionItem[]
}

export async function fetchRevisionItems(
  teamId: number,
  mapId: number,
  filter: RevisionFilter
): Promise<FetchRevisionItemsResponse> {
  const response = await getApiClient().post(`/api/v1/teams/${teamId}/maps/${mapId}/revision-items`, filter)
  return response.data
}
```

### Endpoints API correspondants

| Fonction | Méthode | URL |
|----------|---------|-----|
| `listLinesOfFire` | GET | `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire` |
| `createLineOfFire` | POST | `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire` |
| `getLineOfFire` | GET | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}` |
| `deleteLineOfFire` | DELETE | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}` |
| `setLineOfFireName` | PUT | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/name` |
| `addLineOfFireCone` | POST | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones` |
| `updateLineOfFireCone` | PUT | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}` |
| `removeLineOfFireCone` | DELETE | `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}` |
| `batchLineOfFireActions` | POST | `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire/batch` |
| `fetchRevisionItems` | POST | `/api/v1/teams/${teamId}/maps/${mapId}/revision-items` |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/linesOfFire.ts` | Créer | API hooks |
| `frontend/src/lib/api/types.ts` | Modifier | Ajouter types LineOfFire, Cone, ConeData |
| `frontend/src/lib/revision/types.ts` | Modifier | Ajouter `RevisionItem` + `RevisionFilter`, renommer `RevisionZone`→`RevisionItem` |
| `frontend/src/api/revision.ts` | Modifier | Ajouter `fetchRevisionItems` POST hook |

### Contraintes techniques
- **Pattern** : Suivre le pattern existant de `frontend/src/api/zones.ts` (utilisation de `getApiClient()`, typage fort).
- **Client API** : Utiliser les méthodes `.get()`, `.post()`, `.put()`, `.del()` du client existant.
- **Pas de TanStack Query** : Comme `zones.ts`, utiliser des fonctions async simples (pas de hooks React Query pour l'éditeur, car l'éditeur gère son propre état).
