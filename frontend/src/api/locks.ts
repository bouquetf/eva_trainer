import { getApiClient } from './client'
import { useAuthStore } from '../store/auth'

export interface LockDto {
  id: number
  team_id: number
  map_id: number
  editor_id: number
  editor_name?: string
  locked_at: string
  expires_at: string
  remaining_seconds: number
}

export interface LockStatus {
  locked: boolean
  lock: LockDto | null
}

const api = () => getApiClient()

export async function acquireLock(
  teamId: number,
  mapId: number,
): Promise<{ lock: LockDto }> {
  return api().post<{ lock: LockDto }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lock`,
  )
}

export async function releaseLock(teamId: number, mapId: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/maps/${mapId}/lock`)
}

export async function refreshLock(
  teamId: number,
  mapId: number,
): Promise<{ lock: LockDto }> {
  return api().put<{ lock: LockDto }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lock`,
  )
}

export async function getLockStatus(teamId: number, mapId: number): Promise<LockStatus> {
  return api().get<LockStatus>(`/api/v1/teams/${teamId}/maps/${mapId}/lock`)
}

export function releaseLockSync(teamId: number, mapId: number): void {
  try {
    const token = useAuthStore.getState().getAccessToken()
    fetch(`/api/v1/teams/${teamId}/maps/${mapId}/lock/release`, {
      method: 'POST',
      keepalive: true,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
  } catch {
    // best-effort, silencieux
  }
}
