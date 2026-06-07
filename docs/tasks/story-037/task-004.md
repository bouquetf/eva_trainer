# Tâche #004 - Story #037 : Frontend — Section Divisions + Classement

## Objectif
Créer la section Divisions dans la page Archives, avec CRUD des divs, assignation/désassignation d'équipes eva, et affichage du classement.

## Contexte
- Story #037 : `docs/stories/story-037.md`
- Dépend de : Tâche #003 (API backend)
- Nécessaire pour : Tâche #005 (E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : API Client

Ajouter les fonctions API dans `frontend/src/api/archives.ts` pour interagir avec les endpoints divs.

#### Partie 2 : Composant DivisionsSection

Créer un composant `DivisionsSection` similaire à `EvaTeamsSection` (même pattern React Query + formulaires modaux).

**Fonctionnalités :**
1. **Sélection de ligue** : Un sélecteur pour choisir la ligue (période) dont on veut gérer les divisions
2. **Liste des divs** : Tableau des divs de la ligue sélectionnée avec boutons Modifier / Supprimer
3. **Création de div** : Formulaire modal avec champ "Nom" + ligue_id (déduit du sélecteur)
4. **Modification de div** : Formulaire modal pré-rempli pour changer le nom
5. **Assignation d'équipe** : Dans le détail d'une div, afficher les équipes déjà assignées + un sélecteur pour en ajouter
6. **Désassignation** : Bouton pour retirer une équipe de la div
7. **Classement** : Tableau affichant le classement avec colonnes :
   - Équipe (nom + tag)
   - BO : J, G, N, P, Pts
   - Matchs : J, G, N, P, Pts
8. **Permissions** : Les boutons d'écriture sont cachés si `canWrite` est false

#### Partie 3 : Intégration dans ArchivePage

Ajouter un troisième onglet "Divisions" dans `ArchivePage.tsx`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter types et fonctions API Div |
| `frontend/src/pages/DivisionsSection.tsx` | Créer | Composant de gestion des divisions |
| `frontend/src/pages/DivisionsSection.test.tsx` | Créer | Tests unitaires |
| `frontend/src/pages/ArchivePage.tsx` | Modifier | Ajouter onglet Divisions |
| `frontend/src/pages/ArchivePage.test.tsx` | Modifier | Tester le nouvel onglet |

### Signatures (API Client)

```typescript
// Types à ajouter dans archives.ts

export interface Div {
  id: number
  ligue_id: number
  nom: string
}

export interface DivListResponse {
  divs: Div[]
}

export interface DivDetailResponse {
  div: Div
  teams: EvaTeam[]
}

export interface CreateDivPayload {
  ligue_id: number
  nom: string
}

export interface StandingEntry {
  team_id: number
  team_name: string
  team_tag: string
  bos_played: number
  bo_won: number
  bo_drawn: number
  bo_lost: number
  bo_points: number
  matches_played: number
  match_won: number
  match_drawn: number
  match_lost: number
  match_points: number
}

export interface StandingsResponse {
  standings: StandingEntry[]
}
```

```typescript
// Fonctions API

const DIVISIONS_BASE = '/api/v1/archive/divisions'

export async function listDivs(ligueId: number): Promise<DivListResponse> {
  const api = getApiClient()
  return api.get<DivListResponse>(`${DIVISIONS_BASE}?ligue_id=${ligueId}`)
}

export async function getDiv(id: number): Promise<DivDetailResponse> {
  const api = getApiClient()
  return api.get<DivDetailResponse>(`${DIVISIONS_BASE}/${id}`)
}

export async function createDiv(data: CreateDivPayload): Promise<{ div: Div }> {
  const api = getApiClient()
  return api.post<{ div: Div }>(DIVISIONS_BASE, data)
}

export async function updateDiv(id: number, data: { nom: string }): Promise<{ div: Div }> {
  const api = getApiClient()
  return api.put<{ div: Div }>(`${DIVISIONS_BASE}/${id}`, data)
}

export async function deleteDiv(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${DIVISIONS_BASE}/${id}`)
}

export async function assignTeamToDiv(divId: number, equipeEvaId: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.post<{ message: string }>(`${DIVISIONS_BASE}/${divId}/assign`, { equipe_eva_id: equipeEvaId })
}

export async function removeTeamFromDiv(divId: number, teamId: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${DIVISIONS_BASE}/${divId}/teams/${teamId}`)
}

export async function getDivStandings(divId: number): Promise<StandingsResponse> {
  const api = getApiClient()
  return api.get<StandingsResponse>(`${DIVISIONS_BASE}/${divId}/standings`)
}
```

### Composant DivisionsSection

Pattern à suivre : exactement le même que `EvaTeamsSection.tsx` (même style, même structure React Query).

**Structure :**

```
┌─ Sélecteur de ligue (dropdown) ──────────────────────────────┐
│  [Ligue #1 — 2026-01-01 → 2026-06-30 (Split 1 2026)]  ▼     │
└──────────────────────────────────────────────────────────────┘

┌─ Liste des divs (quand une ligue est sélectionnée) ──────────┐
│  [Nouvelle div]                                                │
│                                                               │
│  ┌──────────────┬──────────────────┬────────────────────────┐ │
│  │ Nom          │ Équipes          │ Actions                │ │
│  ├──────────────┼──────────────────┼────────────────────────┤ │
│  │ Groupe A     │ Team1, Team2     │ Modifier  Supprimer    │ │
│  │ Groupe B     │ Team3            │ Modifier  Supprimer    │ │
│  └──────────────┴──────────────────┴────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘

┌─ Classement (quand on clique sur une div) ───────────────────┐
│  Classement Groupe A                                          │
│  ┌──────────┬───────────┬──────────────────┬────────────────┐ │
│  │ Équipe   │ BO        │ Matchs           │                │ │
│  │          │ J G N P Pts│ J  G  N  P  Pts │                │ │
│  ├──────────┼───────────┼──────────────────┼────────────────┤ │
│  │ Team1    │ 2 2 0 0 2 │ 6  5  1  0  5   │                │ │
│  │ Team2    │ 2 0 0 2 0 │ 6  0  1  5  0   │                │ │
│  └──────────┴───────────┴──────────────────┴────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

**Détail d'implémentation :**

1. **État local** :
   - `selectedLigueId: number | null` — ligue sélectionnée
   - `divs: Div[]` — liste des divs (via useQuery)
   - `selectedDivId: number | null` — div sélectionnée pour le détail/classement
   - Form states pour création/édition de div
   - `showAssignForm: boolean` + `assignDivId: number | null` — formulaire d'assignation
   - `selectedTeamId: number | ''` — équipe sélectionnée pour l'assignation

2. **Requêtes** :
   - `useQuery(['archive-ligues', page])` → lister les ligues pour le sélecteur
   - `useQuery(['archive-divs', selectedLigueId], ...)` → lister les divs
   - `useQuery(['archive-div-detail', selectedDivId], ...)` → détail d'une div
   - `useQuery(['archive-standings', selectedDivId], ...)` → classement
   - `useQuery(['archive-eva-teams'])` → lister les équipes eva pour l'assignation

3. **Mutations** (créer, modifier, supprimer div + assigner/retirer équipe) :
   - Même pattern `useMutation` que `EvaTeamsSection`
   - `onSuccess` → invalider les queries concernées

4. **Classement** : S'affiche directement dans la vue détail de la div, en dessous de la liste des équipes.

### Tests unitaires

Ajouter `frontend/src/pages/DivisionsSection.test.tsx` :

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import DivisionsSection from './DivisionsSection'

vi.mock('../api/archives', () => ({
  listLigues: vi.fn(),
  listDivs: vi.fn(),
  getDiv: vi.fn(),
  createDiv: vi.fn(),
  updateDiv: vi.fn(),
  deleteDiv: vi.fn(),
  assignTeamToDiv: vi.fn(),
  removeTeamFromDiv: vi.fn(),
  getDivStandings: vi.fn(),
  listEvaTeams: vi.fn(),
}))

// ... tests suivant le même pattern que ArchivePage.test.tsx
```

**Scénarios de test :**
1. Affiche le sélecteur de ligue
2. Ne montre pas la liste des divs si aucune ligue sélectionnée
3. Liste les divs quand une ligue est sélectionnée
4. Ouvre le formulaire de création de div
5. Crée une div
6. Modifie une div
7. Supprime une div avec confirmation
8. Affiche les équipes d'une div
9. Ouvre le formulaire d'assignation
10. Cache les boutons d'écriture si `canWrite` est false
11. Affiche le classement quand on sélectionne une div

### Contraintes techniques
- **Framework** : React 18, TypeScript, TanStack React Query
- **Style** : Utiliser les classes CSS existantes (`.admin-table`, `.admin-header`, `.modal-overlay`, `.modal-content`, `.form-group`, `.btn-primary`, etc.) définies dans `frontend/src/styles/admin.css`
- **Pattern** : Suivre exactement le pattern de `EvaTeamsSection.tsx` (même imports, même structure de composant, mêmes hooks React Query)
- **i18n** : Pas de i18n — utiliser le français directement
- **Requêtes** : Utiliser `getApiClient()` depuis `../api/client`
- **Permissions** : Utiliser la prop `canWrite` pour conditionner les actions d'écriture

### Documentation

Créer `docs/archives-divisions.md` :

```markdown
# Archives — Divisions et classement

## Divs

Une division (div) est un groupe d'équipes eva au sein d'une ligue (période de compétition).

### Endpoints API

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/api/v1/archive/divs?ligue_id=N` | Lister les divs d'une ligue |
| GET | `/api/v1/archive/divs/:id` | Détail d'une div (avec équipes) |
| POST | `/api/v1/archive/divs` | Créer une div |
| PUT | `/api/v1/archive/divs/:id` | Modifier une div |
| DELETE | `/api/v1/archive/divs/:id` | Supprimer une div |
| POST | `/api/v1/archive/divs/:id/assign` | Assigner une équipe à une div |
| DELETE | `/api/v1/archive/divs/:id/teams/:teamId` | Retirer une équipe d'une div |
| GET | `/api/v1/archive/divs/:id/standings` | Classement d'une div |

### Classement

Le classement est calculé automatiquement à partir des résultats de BO saisis :

- **Classement BO** : 1 point par BO gagné, 0 par BO perdu
- **Classement match** : 1 point par match gagné, 0 par match perdu

Les deux classements sont affichés séparément dans le tableau.
Les équipes sont triées par points BO décroissants, puis points match décroissants.

### Modèle de données

- `divs` : id, ligue_id, nom
- `equipe_eva_div` : equipe_eva_id, div_id
```
