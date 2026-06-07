import { getApiClient } from './client'

export interface LigueLocale {
  id: number
  nom: string
  emplacement: string
  created_at: string
  updated_at: string
}

export interface LigueLocaleListResponse {
  ligues_locales: LigueLocale[]
}

export interface LigueDetail {
  id: number
  date_debut: string
  date_fin: string
  split_name: string
}

export interface LigueLocaleResponse {
  ligue_locale: LigueLocale
  ligues: LigueDetail[]
}

export interface CreateLigueLocalePayload {
  nom: string
  emplacement: string
}

export type UpdateLigueLocalePayload = CreateLigueLocalePayload

const BASE = '/api/v1/archive/leagues-locales'

export async function listLiguesLocales(emplacement?: string): Promise<LigueLocaleListResponse> {
  const api = getApiClient()
  const path = emplacement ? `${BASE}?emplacement=${encodeURIComponent(emplacement)}` : BASE
  return api.get<LigueLocaleListResponse>(path)
}

export async function getLigueLocale(id: number): Promise<LigueLocaleResponse> {
  const api = getApiClient()
  return api.get<LigueLocaleResponse>(`${BASE}/${id}`)
}

export async function createLigueLocale(data: CreateLigueLocalePayload): Promise<LigueLocaleResponse> {
  const api = getApiClient()
  return api.post<LigueLocaleResponse>(BASE, data)
}

export async function updateLigueLocale(id: number, data: UpdateLigueLocalePayload): Promise<LigueLocaleResponse> {
  const api = getApiClient()
  return api.put<LigueLocaleResponse>(`${BASE}/${id}`, data)
}

export async function deleteLigueLocale(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${BASE}/${id}`)
}

export interface LigueListItem {
  id: number
  date_debut: string
  date_fin: string
  split_name: string
  nb_locales: number
}

export interface LigueListResponse {
  items: LigueListItem[]
  total: number
  page: number
  per_page: number
}

export interface CreateLiguePayload {
  date_debut: string
  date_fin: string
}

const LIGUES_BASE = '/api/v1/archive/ligues'

export async function listLigues(page = 1, perPage = 5): Promise<LigueListResponse> {
  const api = getApiClient()
  return api.get<LigueListResponse>(`${LIGUES_BASE}?page=${page}&per_page=${perPage}`)
}

export async function createLigue(data: CreateLiguePayload): Promise<{ ligue: LigueListItem }> {
  const api = getApiClient()
  return api.post<{ ligue: LigueListItem }>(LIGUES_BASE, data)
}

export async function refreshLigue(id: number): Promise<{ message: string; associated: number }> {
  const api = getApiClient()
  return api.post<{ message: string; associated: number }>(`${LIGUES_BASE}/${id}/refresh`)
}

export async function importLigue(localeId: number, ligueId: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.post<{ message: string }>(`${BASE}/${localeId}/import-ligue`, { ligue_id: ligueId })
}

export interface UpdateLiguePayload {
  date_debut: string
  date_fin: string
}

export async function updateLigue(id: number, data: UpdateLiguePayload): Promise<{ ligue: LigueListItem }> {
  const api = getApiClient()
  return api.put<{ ligue: LigueListItem }>(`${LIGUES_BASE}/${id}`, data)
}

export async function deleteLigue(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${LIGUES_BASE}/${id}`)
}

// --- Eva Teams ---

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

export interface EvaTeamResponse {
  team: EvaTeam
}

export interface CreateEvaTeamPayload {
  name: string
  tag: string
  eva_trainer_team_id?: number | null
}

export type UpdateEvaTeamPayload = CreateEvaTeamPayload

// --- Trainer Teams (from GET /api/v1/teams for association dropdown) ---

export interface TrainerTeam {
  id: number
  name: string
  captain_id: number
}

export interface TrainerTeamItem {
  team: TrainerTeam
  member_count: number
}

export interface TrainerTeamsResponse {
  teams: TrainerTeamItem[]
  total: number
  page: number
  per_page: number
}

const EVA_TEAMS_BASE = '/api/v1/archive/teams'

export async function listEvaTeams(): Promise<EvaTeamListResponse> {
  const api = getApiClient()
  return api.get<EvaTeamListResponse>(EVA_TEAMS_BASE)
}

export async function createEvaTeam(data: CreateEvaTeamPayload): Promise<EvaTeamResponse> {
  const api = getApiClient()
  return api.post<EvaTeamResponse>(EVA_TEAMS_BASE, data)
}

export async function updateEvaTeam(id: number, data: UpdateEvaTeamPayload): Promise<EvaTeamResponse> {
  const api = getApiClient()
  return api.put<EvaTeamResponse>(`${EVA_TEAMS_BASE}/${id}`, data)
}

export async function deleteEvaTeam(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${EVA_TEAMS_BASE}/${id}`)
}

export async function listTrainerTeams(): Promise<TrainerTeamsResponse> {
  const api = getApiClient()
  return api.get<TrainerTeamsResponse>('/api/v1/teams?per_page=100')
}
