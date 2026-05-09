# Tâche #006 - Story #006 : Frontend API client

## Objectif
Créer les fonctions API TypeScript pour les invitations et les demandes d'adhésion, en suivant le pattern existant dans `frontend/src/api/`.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #005 (Routes)
- Nécessaire pour : Tâche #007 (UI invitations), Tâche #008 (UI demandes)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer/modifier les fichiers API clients suivants :

1. **`frontend/src/api/invitations.ts`** : Toutes les fonctions API liées aux invitations
2. **`frontend/src/api/joinRequests.ts`** : Toutes les fonctions API liées aux demandes d'adhésion
3. **Modifier `frontend/src/api/teams.ts`** : Splitter `listMyTeams` vers `listMyTeams` (appelle `/api/v1/myteams`) et ajouter `listAllTeams` (appelle `/api/v1/teams` avec pagination `?page=1&per_page=20`)

### Cas nominaux
- Inviter par display_name → retourne l'invitation
- Générer un lien d'invitation → retourne le lien complet
- Accepter une invitation → message de succès
- Demander à rejoindre → retourne la demande
- Voir le statut de ma demande → retourne la demande
- Rechercher des utilisateurs → retourne une liste

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/invitations.ts` | Créer | API client invitations |
| `frontend/src/api/joinRequests.ts` | Créer | API client demandes d'adhésion |
| `frontend/src/api/teams.ts` | Modifier | Ajouter `listAllTeams`, renommer `listMyTeams` → `GET /api/v1/myteams` |

### Contenu de `frontend/src/api/invitations.ts`

```typescript
import { getApiClient } from './client'

export interface TeamInvitation {
  id: number
  team_id: number
  team_name?: string
  invited_user_id?: number
  invited_user_display_name?: string
  invited_by_user_id?: number
  invited_by_display_name?: string
  token?: string
  invite_url?: string
  status: 'pending' | 'accepted' | 'declined' | 'expired' | 'cancelled'
  expires_at: string
  created_at: string
}

export interface InviteByDisplayNamePayload {
  display_name: string
}

export interface InviteLinkInfo {
  team_id: number
  team_name: string
  expires_at: string
  is_expired: boolean
}

export interface UserSearchResult {
  id: number
  display_name: string
  discord_username: string
}

export interface UserSearchResponse {
  users: UserSearchResult[]
}

// Inviter par display_name/discord_username
export async function inviteByDisplayName(
  teamId: number,
  data: InviteByDisplayNamePayload,
): Promise<TeamInvitation> {
  const api = getApiClient()
  return api.post<TeamInvitation>(`/api/v1/teams/${teamId}/invitations`, data)
}

// Générer un lien d'invitation
export async function generateInviteLink(
  teamId: number,
): Promise<TeamInvitation> {
  const api = getApiClient()
  return api.post<TeamInvitation>(`/api/v1/teams/${teamId}/invitations/link`, {})
}

// Lister les invitations en attente d'une équipe (capitaine/cocapitaine)
export async function listPendingInvitations(
  teamId: number,
): Promise<{ invitations: TeamInvitation[] }> {
  const api = getApiClient()
  return api.get<{ invitations: TeamInvitation[] }>(`/api/v1/teams/${teamId}/invitations/pending`)
}

// Lister mes invitations en attente
export async function listMyInvitations(): Promise<TeamInvitation[]> {
  const api = getApiClient()
  return api.get<TeamInvitation[]>('/api/v1/invitations')
}

// Accepter une invitation (par ID)
export async function acceptInvitation(
  invitationId: number,
): Promise<{ message: string; team_id: number; team_name: string }> {
  const api = getApiClient()
  return api.post<{ message: string; team_id: number; team_name: string }>(
    `/api/v1/invitations/${invitationId}/accept`,
  )
}

// Refuser une invitation
export async function declineInvitation(
  invitationId: number,
): Promise<{ message: string }> {
  const api = getApiClient()
  return api.post<{ message: string }>(`/api/v1/invitations/${invitationId}/decline`)
}

// Voir les infos d'un lien d'invitation (sans auth)
export async function getInviteLinkInfo(
  token: string,
): Promise<InviteLinkInfo> {
  const api = getApiClient()
  return api.get<InviteLinkInfo>(`/api/v1/invite-links/${token}`)
}

// Accepter une invitation par lien
export async function acceptInviteLink(
  token: string,
): Promise<{ message: string; team_id: number; team_name: string }> {
  const api = getApiClient()
  return api.post<{ message: string; team_id: number; team_name: string }>(
    `/api/v1/invite-links/${token}/accept`,
  )
}

// Rechercher des utilisateurs (autocomplétion)
export async function searchUsers(query: string): Promise<UserSearchResponse> {
  const api = getApiClient()
  return api.get<UserSearchResponse>(`/api/v1/users/search?q=${encodeURIComponent(query)}`)
}
```

### Contenu de `frontend/src/api/joinRequests.ts`

```typescript
import { getApiClient } from './client'

export interface JoinRequest {
  id: number
  team_id: number
  user_id: number
  display_name?: string
  status: 'pending' | 'accepted' | 'declined' | 'cancelled'
  created_at: string
}

export interface JoinRequestsListResponse {
  requests: JoinRequest[]
}

// Demander à rejoindre une équipe
export async function requestJoinTeam(
  teamId: number,
): Promise<JoinRequest> {
  const api = getApiClient()
  return api.post<JoinRequest>(`/api/v1/teams/${teamId}/join-requests`, {})
}

// Lister les demandes en attente (capitaine/cocapitaine)
export async function listPendingJoinRequests(
  teamId: number,
): Promise<JoinRequestsListResponse> {
  const api = getApiClient()
  return api.get<JoinRequestsListResponse>(`/api/v1/teams/${teamId}/join-requests`)
}

// Voir ma demande
export async function getMyJoinRequest(
  teamId: number,
): Promise<JoinRequest> {
  const api = getApiClient()
  return api.get<JoinRequest>(`/api/v1/teams/${teamId}/join-requests/mine`)
}

// Annuler ma demande
export async function cancelMyJoinRequest(
  teamId: number,
): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`/api/v1/teams/${teamId}/join-requests/mine`)
}

// Accepter une demande (capitaine/cocapitaine)
export async function acceptJoinRequest(
  teamId: number,
  requestId: number,
): Promise<{ message: string; user_id: number }> {
  const api = getApiClient()
  return api.put<{ message: string; user_id: number }>(
    `/api/v1/teams/${teamId}/join-requests/${requestId}/accept`,
  )
}

// Refuser une demande (capitaine/cocapitaine)
export async function declineJoinRequest(
  teamId: number,
  requestId: number,
): Promise<{ message: string; user_id: number }> {
  const api = getApiClient()
  return api.put<{ message: string; user_id: number }>(
    `/api/v1/teams/${teamId}/join-requests/${requestId}/decline`,
  )
}
```

### Modifications de `frontend/src/api/teams.ts`

```typescript
// Remplacer la fonction existante listMyTeams :
export async function listMyTeams(): Promise<TeamListItem[]> {
  const api = getApiClient()
  return api.get<TeamListItem[]>('/api/v1/myteams')  // Changement d'URL
}

// Ajouter la nouvelle fonction listAllTeams (découverte) :
export interface TeamListQuery {
  page?: number
  per_page?: number
  q?: string
}

export interface TeamListResponse {
  teams: TeamListItem[]
  total: number
  page: number
  per_page: number
}

export async function listAllTeams(params?: TeamListQuery): Promise<TeamListResponse> {
  const api = getApiClient()
  const query = params ? '?' + new URLSearchParams(
    Object.entries(params).filter(([_, v]) => v !== undefined).map(([k, v]) => [k, String(v)])
  ).toString() : ''
  return api.get<TeamListResponse>(`/api/v1/teams${query}`)
}
```

### Contraintes techniques
- **Pattern** : Même pattern que `teams.ts` et `roles.ts` (import `getApiClient`, fonctions async avec typage explicite)
- **ApiClient** : Utiliser `api.get<T>()`, `api.post<T>()`, `api.put<T>()`, `api.del<T>()` comme dans les fichiers existants
- **Gestion d'erreurs** : Les erreurs API sont gérées par TanStack Query dans les composants (pas de try/catch dans les fonctions API)
- **Nommage** : Types exportés avec `export interface`, fonctions exportées avec `export async function`

### Tests à implémenter

- **Fichier** : `frontend/src/api/invitations.test.ts`
- Test basique de la structure des fonctions (vérifier que les URLs sont correctes)
- Mock du `getApiClient` et vérifier les appels

- **Fichier** : `frontend/src/api/joinRequests.test.ts`
- Même approche

### Documentation
- Aucune documentation spécifique pour cette tâche
