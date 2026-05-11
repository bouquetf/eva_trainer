# Tâche #002 - Story #015 : Frontend - API client, types et validation de réponse

## Objectif

Créer l'API client frontend pour récupérer les zones de révision, définir les types TypeScript, et implémenter la fonction de validation de réponse (insensible à la casse et aux accents).

## Contexte

- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation
- Dépend de : Tâche #001 (endpoint backend)
- Nécessaire pour : Tâche #003 (store), Tâche #004 (page)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Types Revision

Créer le fichier `frontend/src/lib/revision/types.ts` avec les types suivants :

```typescript
export interface Point {
  x: number
  y: number
}

export interface ZonePolygon {
  id: number
  zone_id: number
  points: Point[]
  order_index: number
  created_at: string
}

export interface RevisionZone {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygons: ZonePolygon[]
  name: string
}

export interface RevisionZonesResponse {
  zones: RevisionZone[]
}

export interface RevisionConfig {
  totalQuestions: number      // nombre de questions configuré (défaut: 20)
  mapId: number               // ID de la carte sélectionnée
  mapName: string             // nom de la carte
  teamId: number              // ID de l'équipe
}
```

**Note :** `RevisionZone` inclut `polygons` au format `ZonePolygon[]` pour permettre l'affichage des polygones sur la carte via React-Konva (Tâche #004). Les types `Point` et `ZonePolygon` sont définis ici plutôt que réimportés depuis `lib/api/types` pour éviter la dépendance aux types dépréciés de l'éditeur.

#### Partie 2 - API client

Créer le fichier `frontend/src/api/revision.ts` :

```typescript
import { getApiClient } from './client'
import type { RevisionZonesResponse, RevisionZone } from '../lib/revision/types'

export interface RevisionZonesResponse {
  zones: RevisionZone[]
}

export async function fetchRevisionZones(
  teamId: number,
  mapId: number,
): Promise<RevisionZone[]> {
  const api = getApiClient()
  const response = await api.get<RevisionZonesResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/revision-zones`,
  )
  return response.zones
}
```

#### Partie 3 - Fonction de validation de réponse

Créer le fichier `frontend/src/lib/revision/validation.ts` avec une fonction pure :

```typescript
/**
 * Valide une réponse utilisateur contre le nom attendu d'une zone.
 * La validation est insensible à la casse et aux accents.
 * Aucune autre tolérance n'est appliquée.
 */
export function validateAnswer(userAnswer: string, correctAnswer: string): boolean {
  const normalizedUser = normalizeForComparison(userAnswer)
  const normalizedCorrect = normalizeForComparison(correctAnswer)
  return normalizedUser === normalizedCorrect
}

/**
 * Normalise une chaîne pour la comparaison :
 * - Supprime les espaces superflus (trim)
 * - Met en minuscules
 * - Supprime les accents et diacritiques
 * - Supprime les caractères non-alphanumériques (sauf espaces)
 */
function normalizeForComparison(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')  // supprime les diacritiques
}
```

**Cas nominaux :**
- `validateAnswer("Top", "top")` → `true` (casse différente)
- `validateAnswer("ÉLÉPHANT", "elephant")` → `true` (accents)
- `validateAnswer("  top  ", "top")` → `true` (espaces superflus)
- `validateAnswer("Pigeon", "Pigeon")` → `true` (identique)

**Cas limites :**
- `validateAnswer("top secret", "topsecret")` → `false` (espace compte)
- `validateAnswer("top", "top ")` → `true` (trim applied to user answer)
- `validateAnswer("", "top")` → `false` (vide)
- `validateAnswer("TOP", "top")` → `true`
- `validateAnswer("Côte d'Azur", "cote d'azur")` → `true` (accent + apostrophe)
- `validateAnswer("Top!", "top")` → `false` (ponctuation non tolérée)

**Gestion d'erreurs :**
- Chaînes vides → retourner `false` si l'une est vide (après trim)
- Chaînes null/undefined → traiter comme chaîne vide

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/revision/types.ts` | Créer | Types TypeScript pour la révision |
| `frontend/src/api/revision.ts` | Créer | Fonction `fetchRevisionZones` |
| `frontend/src/lib/revision/validation.ts` | Créer | Fonction `validateAnswer` et `normalizeForComparison` |

### Signatures

```typescript
// frontend/src/lib/revision/types.ts
export interface Point {
  x: number
  y: number
}

export interface ZonePolygon {
  id: number
  zone_id: number
  points: Point[]
  order_index: number
  created_at: string
}

export interface RevisionZone {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygons: ZonePolygon[]
  name: string
}

export interface RevisionZonesResponse {
  zones: RevisionZone[]
}

export interface RevisionConfig {
  totalQuestions: number
  mapId: number
  mapName: string
  teamId: number
}

// frontend/src/api/revision.ts
export async function fetchRevisionZones(teamId: number, mapId: number): Promise<RevisionZone[]>

// frontend/src/lib/revision/validation.ts
export function validateAnswer(userAnswer: string, correctAnswer: string): boolean
```

### Contraintes techniques

- **Fonction pure** : `validateAnswer` ne doit avoir aucun effet de bord, aucune dépendance externe
- **Normalization Unicode** : Utiliser `String.prototype.normalize('NFD')` pour décomposer les caractères accentués, puis supprimer les diacritiques avec la regex `[\u0300-\u036f]`
- **API client** : Suivre le pattern des autres fichiers dans `frontend/src/api/` (importer `getApiClient`, exporter des fonctions async)
- **Pas de dépendance React** : Le module `validation.ts` doit être pur JavaScript/TypeScript, réutilisable partout

### Tests à implémenter

#### Tests unitaires (validation)

- **Fichier** : `frontend/src/lib/revision/validation.test.ts`

| Scénario | Entrée | Attendu |
|----------|--------|---------|
| Même mot, casse différente | `validateAnswer("Top", "top")` | `true` |
| Accents | `validateAnswer("ÉLÉPHANT", "elephant")` | `true` |
| Espaces superflus | `validateAnswer("  top  ", "top")` | `true` |
| Identique | `validateAnswer("Pigeon", "Pigeon")` | `true` |
| Espace dans le mot | `validateAnswer("top secret", "topsecret")` | `false` |
| Chaîne vide | `validateAnswer("", "top")` | `false` |
| Les deux vides | `validateAnswer("", "")` | `true` |
| Ponctuation | `validateAnswer("Top!", "top")` | `false` |
| Accents composés | `validateAnswer("Côte d'Azur", "cote d'azur")` | `true` |
| Tirets | `validateAnswer("arrière-cour", "arriere-cour")` | `true` |
| Chiffres | `validateAnswer("Zone 42", "zone 42")` | `true` |
| Chiffres vs sans | `validateAnswer("Zone42", "zone 42")` | `false` |

#### Tests API (integration)

- **Fichier** : `frontend/src/api/revision.test.ts`
- Tester que `fetchRevisionZones` retourne bien un tableau de `RevisionZone`
- Tester que l'appel échoue proprement avec un token invalide

### Documentation

Aucune documentation à créer pour cette tâche.
