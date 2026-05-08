import { getApiClient } from './client'

export interface TeamMemberWithUser {
  user_id: number
  display_name: string
  role: 'captain' | 'cocaptain' | 'editor' | 'member'
  joined_at: string
}

export interface MembersListResponse {
  members: TeamMemberWithUser[]
}

export interface UpdateRolePayload {
  role: 'member' | 'editor' | 'cocaptain'
}

export interface TransferCaptaincyPayload {
  user_id: number
}

export async function listMembers(teamId: number): Promise<MembersListResponse> {
  const api = getApiClient()
  return api.get<MembersListResponse>(`/api/v1/teams/${teamId}/members`)
}

export async function updateMemberRole(
  teamId: number,
  userId: number,
  data: UpdateRolePayload,
): Promise<{ message: string; user_id: number; role: string }> {
  const api = getApiClient()
  return api.put<{ message: string; user_id: number; role: string }>(
    `/api/v1/teams/${teamId}/members/${userId}/role`,
    data,
  )
}

export async function transferCaptaincy(
  teamId: number,
  data: TransferCaptaincyPayload,
): Promise<{ message: string; new_captain_id: number }> {
  const api = getApiClient()
  return api.post<{ message: string; new_captain_id: number }>(
    `/api/v1/teams/${teamId}/transfer-captaincy`,
    data,
  )
}
