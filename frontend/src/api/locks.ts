import { getApiClient } from './client'

export interface LockStatus {
  locked: boolean
  lock: {
    id: number
    team_id: number
    map_id: number
    editor_id: number
    locked_at: string
    expires_at: string
  } | null
}

const api = () => getApiClient()

export async function acquireLock(
  teamId: number,
  mapId: number,
): Promise<{ lock: LockStatus['lock'] }> {
  return api().post<{ lock: LockStatus['lock'] }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lock`,
  )
}

export async function releaseLock(teamId: number, mapId: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/maps/${mapId}/lock`)
}

export async function refreshLock(
  teamId: number,
  mapId: number,
): Promise<{ lock: LockStatus['lock'] }> {
  return api().put<{ lock: LockStatus['lock'] }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lock`,
  )
}

export async function getLockStatus(teamId: number, mapId: number): Promise<LockStatus> {
  return api().get<LockStatus>(`/api/v1/teams/${teamId}/maps/${mapId}/lock`)
}
