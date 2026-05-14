import { getApiClient, ApiError } from './client'
import { useAuthStore } from '../store/auth'
import type { User } from './auth'

export interface PermissionWithRoles {
  id: number
  feature: string
  description: string
  roles: string[]
}

export interface AdminPermissionsResponse {
  permissions: PermissionWithRoles[]
  roles: string[]
}

export interface AdminUsersResponse {
  users: User[]
}

export interface TeamMembership {
  team_id: number
  team_name: string
  role: string
}

export interface UserDetailResponse {
  user: User
  memberships: TeamMembership[]
}

export interface Team {
  id: number
  name: string
  captain_id: number
}

export interface AdminTeamsResponse {
  teams: Team[]
}

export interface RolePermissionsResponse {
  role: string
  permissions: { id: number; feature: string; description: string }[]
}

export async function listAdminUsers(): Promise<AdminUsersResponse> {
  const api = getApiClient()
  return api.get<AdminUsersResponse>('/api/v1/admin/users')
}

export async function getUserDetail(userId: number): Promise<UserDetailResponse> {
  const api = getApiClient()
  return api.get<UserDetailResponse>(`/api/v1/admin/users/${userId}`)
}

export async function listAdminPermissions(): Promise<AdminPermissionsResponse> {
  const api = getApiClient()
  return api.get<AdminPermissionsResponse>('/api/v1/admin/permissions')
}

export async function getRolePermissions(role: string): Promise<RolePermissionsResponse> {
  const api = getApiClient()
  return api.get<RolePermissionsResponse>(`/api/v1/admin/permissions/${role}`)
}

export async function listAdminTeams(): Promise<AdminTeamsResponse> {
  const api = getApiClient()
  return api.get<AdminTeamsResponse>('/api/v1/admin/teams')
}

export async function promoteUser(userId: number): Promise<void> {
  const api = getApiClient()
  return api.post(`/api/v1/admin/users/${userId}/promote`)
}

export async function updateUserRole(userId: number, teamId: number, role: string): Promise<void> {
  const api = getApiClient()
  return api.put(`/api/v1/admin/users/${userId}/teams/${teamId}/role`, { role })
}

export async function addUserToTeam(userId: number, teamId: number, role: string): Promise<void> {
  const api = getApiClient()
  return api.post(`/api/v1/admin/users/${userId}/teams`, { team_id: teamId, role })
}

export async function removeUserFromTeam(userId: number, teamId: number): Promise<void> {
  const api = getApiClient()
  return api.del(`/api/v1/admin/users/${userId}/teams/${teamId}`)
}

// --- Import / Export ---

export interface ExportPolygon {
  points: { x: number; y: number }[]
}

export interface ExportZone {
  name: string
  polygons: ExportPolygon[]
}

export interface ExportMapData {
  map_name: string
  team: string
  export_date: string
  zones: ExportZone[]
}

export interface ImportPolygon {
  points: { x: number; y: number }[]
}

export interface ImportZone {
  name: string
  polygons: ImportPolygon[]
}

export interface ImportMapPayload {
  map_name: string
  zones: ImportZone[]
}

export interface ImportMapError {
  map_name: string
  error: string
}

export interface ImportResponse {
  imported_count: number
  maps_processed: number
  errors?: ImportMapError[]
}

const ADMIN_BASE = '/api/v1/admin'

function getBaseURL(): string {
  return import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
}

function getAuthHeaders(): Record<string, string> {
  const token = useAuthStore.getState().accessToken
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function exportTeamMap(teamId: number, mapId: number): Promise<Blob> {
  const res = await fetch(`${getBaseURL()}${ADMIN_BASE}/teams/${teamId}/export/${mapId}`, {
    headers: getAuthHeaders(),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  return res.blob()
}

export async function exportAllTeamMaps(teamId: number): Promise<Blob> {
  const res = await fetch(`${getBaseURL()}${ADMIN_BASE}/teams/${teamId}/export`, {
    headers: getAuthHeaders(),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  return res.blob()
}

export async function importTeamZones(teamId: number, data: ImportMapPayload[]): Promise<ImportResponse> {
  const api = getApiClient()
  return api.post<ImportResponse>(`${ADMIN_BASE}/teams/${teamId}/import`, data)
}
