import { getApiClient } from './client'

export interface User {
  id: number
  discord_id: string
  discord_username: string
  display_name: string
  is_admin: boolean
  created_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export function getLoginUrl(): string {
  const baseURL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  const redirect = window.location.pathname !== '/login' ? window.location.pathname : '/'
  return `${baseURL}/api/v1/auth/discord/login?redirect=${encodeURIComponent(redirect)}`
}

export async function fetchMe(): Promise<User> {
  const api = getApiClient()
  return api.get<User>('/api/v1/auth/me')
}

export async function refreshToken(token: string): Promise<RefreshResponse> {
  const api = getApiClient()
  return api.post<RefreshResponse>('/api/v1/auth/refresh', { refresh_token: token })
}

export async function updateDisplayName(displayName: string): Promise<{ display_name: string }> {
  const api = getApiClient()
  return api.put<{ display_name: string }>('/api/v1/auth/display-name', { display_name: displayName })
}

export interface UserSettings {
  revision_correct_timer: number
  revision_incorrect_timer: number
  revision_question_count: number
}

export async function fetchSettings(): Promise<UserSettings> {
  const api = getApiClient()
  return api.get<UserSettings>('/api/v1/auth/settings')
}

export async function updateSettings(settings: Partial<UserSettings>): Promise<UserSettings> {
  const api = getApiClient()
  return api.put<UserSettings>('/api/v1/auth/settings', settings)
}
