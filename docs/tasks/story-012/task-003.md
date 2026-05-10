# Tâche #003 - Story #012 : Client API Versions et types TypeScript (Frontend)

## Objectif

Créer les types TypeScript et les fonctions API frontend pour les versions de zone et la restauration.

## Contexte

- Story #012 : `docs/stories/story-012.md` - Versions de zone, historique et rollback
- Dépend de : Tâche #002 (API endpoints backend définis)
- Nécessaire pour : Tâche #004 (frontend UI)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Types TypeScript

Ajouter dans `frontend/src/lib/api/types.ts` les types suivants :

```typescript
// Snapshot d'une version de zone
export type VersionSnapshot = {
  polygons: {
    points: Point[];
    order_index: number;
  }[];
  name: string;
}

// Entrée de version (retournée par l'API)
export type ZoneVersion = {
  id: number;
  zone_id: number;
  zone_name: string;
  editor_id: number | null;
  editor_name: string;
  version_data: VersionSnapshot;
  created_at: string;  // ISO8601
}

// Page de versions (liste paginée)
export type VersionPage = {
  versions: ZoneVersion[];
  total: number;
  page: number;
  limit: number;
}

// Filtres pour la liste de versions
export type VersionFilters = {
  zone_id?: number;
  editor_id?: number;
  page?: number;
  limit?: number;
}
```

#### Partie 2 - Client API Versions

Créer `frontend/src/api/versions.ts` avec les fonctions API suivantes :

1. **`listMapVersions(teamId, mapId, filters?)`** : GET `/api/v1/teams/:id/maps/:mapId/versions`
   - Query params : zone_id, editor_id, page, limit
   - Retourne `VersionPage`

2. **`listZoneVersions(teamId, zoneId, page?, limit?)`** : GET `/api/v1/teams/:id/zones/:zoneId/versions`
   - Query params : page, limit
   - Retourne `VersionPage`

3. **`restoreVersion(teamId, zoneId, versionId)`** : POST `/api/v1/teams/:id/zones/:zoneId/versions/:versionId/restore`
   - Retourne `{ zone: Zone }` (la zone restaurée)

**Cas nominaux :**
- Liste des versions de la carte → toutes les versions paginées
- Liste des versions d'une zone → versions de cette zone uniquement
- Restauration réussie → zone mise à jour retournée

**Cas limites :**
- Erreur API 403 → `ApiError` remonté (pas l'auteur, zone modifiée par un autre)
- Erreur API 404 → version introuvable
- Erreur API 423 → lock requis
- Filtres optionnels non fournis → pas de query param correspondant

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/api/types.ts` | Modifier | Ajouter VersionSnapshot, ZoneVersion, VersionPage, VersionFilters |
| `frontend/src/api/versions.ts` | Créer | Fonctions API listMapVersions, listZoneVersions, restoreVersion |

### Signatures

```typescript
// frontend/src/api/versions.ts
import type { Zone, VersionPage, VersionFilters } from '../lib/api/types'

export async function listMapVersions(
  teamId: number,
  mapId: number,
  filters?: VersionFilters,
): Promise<VersionPage>

export async function listZoneVersions(
  teamId: number,
  zoneId: number,
  page?: number,
  limit?: number,
): Promise<VersionPage>

export async function restoreVersion(
  teamId: number,
  zoneId: number,
  versionId: number,
): Promise<{ zone: Zone }>
```

### Contraintes techniques

- **API Client** : Utiliser `getApiClient()` depuis `frontend/src/api/client.ts` qui gère déjà l'authentification JWT et le refresh token.
- **Convention** : Suivre le même pattern que `zones.ts` et `locks.ts` : fonction `api()` locale, chaque fonction exportée utilise `api().get/post`.
- **Types** : Les types doivent correspondre exactement à la réponse JSON du backend (Task #002).
- **Query params optionnels** : Ne pas inclure les paramètres `undefined` ou `null` dans l'URL.

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `frontend/src/api/versions.test.ts`
- Scénario 1 : `listMapVersions` sans filtres → appelle GET avec page=1 par défaut
- Scénario 2 : `listMapVersions` avec filtres → query params dans l'URL
- Scénario 3 : `listZoneVersions` → appelle le bon endpoint avec zoneId
- Scénario 4 : `restoreVersion` → appelle POST avec le bon versionId
- Scénario 5 : Erreur API 403 → ApiError bien rejetée

### Documentation

Aucune documentation à créer pour cette tâche.
