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

export async function inviteByDisplayName(
  teamId: number,
  data: InviteByDisplayNamePayload,
): Promise<TeamInvitation> {
  const api = getApiClient()
  return api.post<TeamInvitation>(`/api/v1/teams/${teamId}/invitations`, data)
}

export async function generateInviteLink(
  teamId: number,
): Promise<TeamInvitation> {
  const api = getApiClient()
  return api.post<TeamInvitation>(`/api/v1/teams/${teamId}/invitations/link`, {})
}

export async function listPendingInvitations(
  teamId: number,
): Promise<{ invitations: TeamInvitation[] }> {
  const api = getApiClient()
  return api.get<{ invitations: TeamInvitation[] }>(`/api/v1/teams/${teamId}/invitations/pending`)
}

export async function listMyInvitations(): Promise<TeamInvitation[]> {
  const api = getApiClient()
  return api.get<TeamInvitation[]>('/api/v1/invitations')
}

export async function acceptInvitation(
  invitationId: number,
): Promise<{ message: string; team_id: number; team_name: string }> {
  const api = getApiClient()
  return api.post<{ message: string; team_id: number; team_name: string }>(
    `/api/v1/invitations/${invitationId}/accept`,
  )
}

export async function declineInvitation(
  invitationId: number,
): Promise<{ message: string }> {
  const api = getApiClient()
  return api.post<{ message: string }>(`/api/v1/invitations/${invitationId}/decline`)
}

export async function getInviteLinkInfo(
  token: string,
): Promise<InviteLinkInfo> {
  const api = getApiClient()
  return api.get<InviteLinkInfo>(`/api/v1/invite-links/${token}`)
}

export async function acceptInviteLink(
  token: string,
): Promise<{ message: string; team_id: number; team_name: string }> {
  const api = getApiClient()
  return api.post<{ message: string; team_id: number; team_name: string }>(
    `/api/v1/invite-links/${token}/accept`,
  )
}

export async function cancelInvitation(
  teamId: number,
  invitationId: number,
): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`/api/v1/teams/${teamId}/invitations/${invitationId}`)
}

export async function searchUsers(query: string): Promise<UserSearchResponse> {
  const api = getApiClient()
  return api.get<UserSearchResponse>(`/api/v1/users/search?q=${encodeURIComponent(query)}`)
}
