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

export async function requestJoinTeam(
  teamId: number,
): Promise<JoinRequest> {
  const api = getApiClient()
  return api.post<JoinRequest>(`/api/v1/teams/${teamId}/join-requests`, {})
}

export async function listPendingJoinRequests(
  teamId: number,
): Promise<JoinRequestsListResponse> {
  const api = getApiClient()
  return api.get<JoinRequestsListResponse>(`/api/v1/teams/${teamId}/join-requests`)
}

export async function getMyJoinRequest(
  teamId: number,
): Promise<JoinRequest> {
  const api = getApiClient()
  return api.get<JoinRequest>(`/api/v1/teams/${teamId}/join-requests/mine`)
}

export async function cancelMyJoinRequest(
  teamId: number,
): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`/api/v1/teams/${teamId}/join-requests/mine`)
}

export async function acceptJoinRequest(
  teamId: number,
  requestId: number,
): Promise<{ message: string; user_id: number }> {
  const api = getApiClient()
  return api.put<{ message: string; user_id: number }>(
    `/api/v1/teams/${teamId}/join-requests/${requestId}/accept`,
  )
}

export async function declineJoinRequest(
  teamId: number,
  requestId: number,
): Promise<{ message: string; user_id: number }> {
  const api = getApiClient()
  return api.put<{ message: string; user_id: number }>(
    `/api/v1/teams/${teamId}/join-requests/${requestId}/decline`,
  )
}
