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

// --- Divisions ---

export interface Div {
  id: number
  ligue_id: number
  nom: string
  teams?: EvaTeam[]
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

export interface BO {
  id: number
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  vainqueur_id: number | null
  format: string
  forfait: number
  created_at: string
  updated_at: string
}

export interface GenerateRoundRobinResponse {
  bos: BO[]
}

export async function generateRoundRobin(divId: number): Promise<GenerateRoundRobinResponse> {
  const api = getApiClient()
  return api.post<GenerateRoundRobinResponse>(`${DIVISIONS_BASE}/${divId}/generate-round-robin`)
}

// --- BOs (Best-of) ---

export interface Ban {
  id: number
  bo_id: number
  equipe_id: number
  map_id: number
  created_at: string
}

export interface Pick {
  id: number
  bo_id: number
  equipe_id: number
  map_id: number
  created_at: string
}

export interface Match {
  id: number
  bo_id: number
  map_id: number
  score_equipe1: number
  score_equipe2: number
  forfait: number
  vainqueur_id: number | null
  commentaires: string | null
  created_at: string
  updated_at: string
}

export interface BODetail {
  BO: BO
  matchs: Match[]
  bans: Ban[]
  picks: Pick[]
}

export interface BOListItem extends BO {
  equipe1_name: string
  equipe2_name: string
}

export interface BOListResponse {
  bos: BOListItem[]
  total: number
}

export interface BOListDetailResponse {
  bos: BODetail[]
  total: number
}

export interface BODetailResponse {
  bo: BODetail
}

export interface CreateBOPayload {
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  format: string
  forfait?: number
  bans?: { equipe_id: number; map_id: number; ordre: number }[]
  picks?: { equipe_id?: number | null; map_id: number; ordre: number }[]
  matchs?: { map_id: number; score_equipe1: number; score_equipe2: number; forfait?: number; commentaires?: string | null }[]
}

export interface UpdateBOPayload {
  format: string
  forfait?: number
  bans?: { equipe_id: number; map_id: number; ordre: number }[]
  picks?: { equipe_id?: number | null; map_id: number; ordre: number }[]
  matchs?: { map_id: number; score_equipe1: number; score_equipe2: number; forfait?: number; commentaires?: string | null }[]
}

export interface LigueMapsResponse {
  maps: { id: number; name: string }[]
}

const BOS_BASE = '/api/v1/archive/bos'

export async function listBOs(ligueId: number, details = false): Promise<BOListResponse | BOListDetailResponse> {
  const api = getApiClient()
  return api.get<BOListResponse | BOListDetailResponse>(`${BOS_BASE}?ligue_id=${ligueId}${details ? '&details=true' : ''}`)
}

export async function getBO(id: number): Promise<BODetailResponse> {
  const api = getApiClient()
  return api.get<BODetailResponse>(`${BOS_BASE}/${id}`)
}

export async function createBO(data: CreateBOPayload): Promise<BODetailResponse> {
  const api = getApiClient()
  return api.post<BODetailResponse>(BOS_BASE, data)
}

export async function updateBO(id: number, data: UpdateBOPayload): Promise<BODetailResponse> {
  const api = getApiClient()
  return api.put<BODetailResponse>(`${BOS_BASE}/${id}`, data)
}

export async function deleteBO(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`${BOS_BASE}/${id}`)
}

export async function getLigueMaps(ligueId: number): Promise<LigueMapsResponse> {
  const api = getApiClient()
  return api.get<LigueMapsResponse>(`/api/v1/archive/ligues/${ligueId}/maps`)
}

// --- Card Stats ---

export interface OpponentCardStats {
  team_id: number
  team_name: string
  avg_score: number
  median_score: number
}

export interface CardStats {
  map_id: number
  map_name: string
  wins: number
  losses: number
  winrate: number
  avg_score: number
  median_score: number
  pick_rate: number
  ban_rate: number
  opponents: OpponentCardStats[]
}

export interface CardStatsResponse {
  stats: CardStats[]
}

const CARD_STATS_BASE = '/api/v1/archive/card-stats'

export async function getCardStats(teamId: number): Promise<CardStatsResponse> {
  const api = getApiClient()
  return api.get<CardStatsResponse>(`${CARD_STATS_BASE}?team_id=${teamId}`)
}

// --- Head-to-Head Stats ---

export interface HeadToHeadTeamStats {
  wins: number
  losses: number
  winrate: number
  avg_score: number
  median_score: number
}

export interface HeadToHeadMapStats {
  map_id: number
  map_name: string
  team1_stats: HeadToHeadTeamStats
  team2_stats: HeadToHeadTeamStats
}

export interface HeadToHeadStatsResponse {
  team1_id: number
  team2_id: number
  stats: HeadToHeadMapStats[]
}

export async function getHeadToHeadStats(team1Id: number, team2Id: number): Promise<HeadToHeadStatsResponse> {
  const api = getApiClient()
  return api.get<HeadToHeadStatsResponse>(`/api/v1/archive/head-to-head-stats?team1_id=${team1Id}&team2_id=${team2Id}`)
}
