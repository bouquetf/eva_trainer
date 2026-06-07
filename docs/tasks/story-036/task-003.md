# Tâche #003 - Story #036 : Frontend API + Interface Équipes Eva

## Objectif
Ajouter les fonctions API côté frontend pour les équipes eva et étendre la page `ArchivePage` avec une section de gestion des équipes eva (liste, création, modification, suppression, association à une équipe eva-trainer).

## Contexte
- Story #036 : `docs/stories/story-036.md`
- Dépend de : Tâche #002 (endpoints API backend)
- Nécessaire pour : Tâche #004 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : API client (archives.ts)

Ajouter les fonctions d'appel API pour les équipes eva dans `frontend/src/api/archives.ts` :

**Interfaces TypeScript :**
```typescript
export interface EvaTeam {
  id: number
  name: string
  tag: string
  logo_url: string | null
  eva_trainer_team_id: number | null
  created_at: string
  updated_at: string
}

export interface EvaTeamListResponse {
  teams: EvaTeam[]
}

export interface CreateEvaTeamPayload {
  name: string
  tag: string
  logo_url?: string | null
  eva_trainer_team_id?: number | null
}

export type UpdateEvaTeamPayload = CreateEvaTeamPayload
```

**Fonctions API :**
```typescript
const TEAMS_BASE = '/api/v1/archive/teams'

export async function listEvaTeams(): Promise<EvaTeamListResponse>
export async function getEvaTeam(id: number): Promise<{ team: EvaTeam }>
export async function createEvaTeam(data: CreateEvaTeamPayload): Promise<{ team: EvaTeam }>
export async function updateEvaTeam(id: number, data: UpdateEvaTeamPayload): Promise<{ team: EvaTeam }>
export async function deleteEvaTeam(id: number): Promise<{ message: string }>
```

**Pattern API :** Utiliser `getApiClient()` comme les autres fonctions du fichier (`.get<T>`, `.post<T>`, `.put<T>`, `.del<T>`).

#### Partie 2 : Interface utilisateur — Section Équipes Eva sur ArchivePage

Étendre `frontend/src/pages/ArchivePage.tsx` pour ajouter une section "Équipes Eva" avec un système d'onglets visuels (ligues locales / équipes eva).

**Approche proposée :**
- Ajouter 2 onglets en haut de la page : "Ligues locales" et "Équipes Eva"
- L'onglet actif est géré par `useState<'ligues' | 'teams'>('ligues')`
- Le contenu existant (ligues locales) reste dans l'onglet "Ligues locales"
- Le nouveau contenu (équipes eva) est dans l'onglet "Équipes Eva"

**Interface "Équipes Eva" (quand l'onglet teams est actif) :**

*Vue liste :*
- Tableau avec colonnes : Nom, Tag, Association eva-trainer, Actions
- Si `canWrite` : bouton "Nouvelle équipe eva" dans l'en-tête
- Chaque ligne a des boutons Modifier/Supprimer (si `canWrite`)
- Afficher "Aucune équipe eva" si la liste est vide
- La colonne "Association" affiche le nom/ID de l'équipe eva-trainer associée, ou "—" si aucune

*Formulaire de création/édition (modal) :*
- Champs :
  - **Nom** (input text, required)
  - **Tag** (input text, required, max 50 caractères)
  - **Association équipe eva-trainer** (select dropdown, optionnel)
    - Charger la liste des équipes eva-trainer depuis `GET /api/v1/teams` (endpoint public existant, accessible à tout utilisateur authentifié)
    - Afficher "Aucune" comme option par défaut (valeur vide)
    - Afficher le nom de chaque équipe dans le dropdown, stocker l'ID
- Validation : nom et tag requis
- Boutons : Annuler / Créer ou Modifier

*Confirmation de suppression (modal) :*
- Message : "Êtes-vous sûr de vouloir supprimer l'équipe eva [nom] ?"
- Boutons : Annuler / Supprimer
- Afficher une erreur si la suppression est bloquée (409 — équipe référencée)

**Gestion des états :**
- Loading : `<div className="empty-state"><p>Chargement...</p></div>`
- Erreur : `<div className="error-message">...</div>` (cliquable pour fermer)
- Succès : `<div className="success-message">...</div>` (cliquable pour fermer)
- React Query : Utiliser `useQuery` avec clé `['archive-eva-teams']` et `staleTime: 30_000`
- Mutations : `useMutation` avec `onSuccess: () => queryClient.invalidateQueries({ queryKey: ['archive-eva-teams'] })`

#### Partie 3 : Appel API pour lister les équipes eva-trainer

Utiliser l'endpoint `GET /api/v1/teams` (existant, accessible à tout utilisateur authentifié).

Ajouter une fonction dans `archives.ts` :
```typescript
export interface TrainerTeam {
  id: number
  name: string
}

export async function listTrainerTeams(): Promise<TrainerTeam[]>
```

Note : le endpoint `/api/v1/teams` retourne un tableau d'équipes. Adapter l'appel à la structure de réponse du handler existant (vérifier le type de retour, typiquement `{ teams: Team[] }` ou `[]Team`).

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter interfaces EvaTeam + fonctions CRUD + fonction listTrainerTeams |
| `frontend/src/pages/ArchivePage.tsx` | Modifier | Ajouter onglets + section Équipes Eva |
| `frontend/src/styles/admin.css` | Modifier (si besoin) | Style des onglets si pas déjà existant |

### Signatures

```typescript
// frontend/src/api/archives.ts

export interface EvaTeam {
  id: number
  name: string
  tag: string
  eva_trainer_team_id: number | null
  created_at: string
  updated_at: string
}

export interface EvaTeamListResponse {
  teams: EvaTeam[]
}

export interface CreateEvaTeamPayload {
  name: string
  tag: string
  eva_trainer_team_id?: number | null
}

export type UpdateEvaTeamPayload = CreateEvaTeamPayload

export async function listEvaTeams(): Promise<EvaTeamListResponse>
export async function getEvaTeam(id: number): Promise<{ team: EvaTeam }>
export async function createEvaTeam(data: CreateEvaTeamPayload): Promise<{ team: EvaTeam }>
export async function updateEvaTeam(id: number, data: UpdateEvaTeamPayload): Promise<{ team: EvaTeam }>
export async function deleteEvaTeam(id: number): Promise<{ message: string }>
```

### Contraintes techniques
- **Pattern Admin** : Utiliser les mêmes classes CSS que le reste de l'admin : `.admin-page`, `.admin-header`, `.admin-table`, `.btn-primary`, `.btn-secondary`, `.btn-danger`, `.btn-small`, `.modal-overlay`, `.modal-content`, `.modal-actions`, `.form-group`, `.empty-state`, `.error-message`, `.success-message`, `.admin-tabs` (classe existante pour les onglets)
- **React Query** : Utiliser `useQuery` et `useMutation` (comme le reste de la page)
- **Pas de nouveau fichier CSS** : Utiliser les styles existants
- **Pas de nouveau fichier page** : Étendre `ArchivePage.tsx` existant
- **Internationalisation** : Tout le texte en français (comme le reste de l'application)
- **Accessibilité** : Labels sur les champs de formulaire, aria-* si pertinent

### Tests à implémenter

Tests détaillés dans la Tâche #004.

**Vérification manuelle :**
```bash
cd frontend
npx tsc --noEmit
```

### Documentation

Mise à jour de `docs/archives-eva-teams.md` (créé dans Tâche #004) avec :
- Guide d'utilisation de l'interface (capture d'écran textuelle)
- Structure de la page ArchivePage
- Exemples d'utilisation de l'API depuis le frontend
