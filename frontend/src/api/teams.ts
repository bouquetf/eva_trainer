import { getApiClient } from './client'

export interface Team {
  id: number
  name: string
  captain_id: number
  created_at: string
  updated_at: string
}

export interface TeamMember {
  user_id: number
  role: 'captain' | 'cocaptain' | 'editor' | 'member'
  display_name: string
  joined_at: string
}

export interface TeamDetail {
  team: Team
  members: TeamMember[]
}

export interface TeamListItem {
  team: Team
  member_count: number
  my_role: string
}

export interface CreateTeamPayload {
  name: string
}

export interface UpdateTeamPayload {
  name: string
}

export interface DeleteTeamPayload {
  confirm_with_name: string
}

export async function createTeam(data: CreateTeamPayload): Promise<Team> {
  const api = getApiClient()
  return api.post<Team>('/api/v1/teams', data)
}

export async function listMyTeams(): Promise<TeamListItem[]> {
  const api = getApiClient()
  return api.get<TeamListItem[]>('/api/v1/teams')
}

export async function getTeam(id: number): Promise<TeamDetail> {
  const api = getApiClient()
  return api.get<TeamDetail>(`/api/v1/teams/${id}`)
}

export async function updateTeam(id: number, data: UpdateTeamPayload): Promise<Team> {
  const api = getApiClient()
  return api.put<Team>(`/api/v1/teams/${id}`, data)
}

export async function deleteTeam(id: number, data: DeleteTeamPayload): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`/api/v1/teams/${id}`, data)
}

export async function leaveTeam(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.post<{ message: string }>(`/api/v1/teams/${id}/leave`)
}
