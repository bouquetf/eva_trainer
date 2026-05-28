import { getApiClient } from './client'
import type {
  LineOfFireVersion,
  Zone,
  VersionPage,
  VersionFilters,
} from '../lib/api/types'

const api = () => getApiClient()

export async function listMapVersions(
  teamId: number,
  mapId: number,
  filters?: VersionFilters,
): Promise<VersionPage> {
  const params = new URLSearchParams()
  if (filters?.page != null) params.set('page', String(filters.page))
  if (filters?.limit != null) params.set('limit', String(filters.limit))
  if (filters?.zone_id != null) params.set('zone_id', String(filters.zone_id))
  if (filters?.editor_id != null) params.set('editor_id', String(filters.editor_id))

  const query = params.toString()
  const path = `/api/v1/teams/${teamId}/maps/${mapId}/versions${query ? `?${query}` : ''}`
  return api().get<VersionPage>(path)
}

export async function listLineOfFireVersions(
  teamId: number,
  lineOfFireId: number,
): Promise<{ versions: LineOfFireVersion[]; total: number }> {
  return api().get<{ versions: LineOfFireVersion[]; total: number }>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/versions`,
  )
}

export async function listZoneVersions(
  teamId: number,
  zoneId: number,
  page?: number,
  limit?: number,
): Promise<VersionPage> {
  const params = new URLSearchParams()
  if (page != null) params.set('page', String(page))
  if (limit != null) params.set('limit', String(limit))

  const query = params.toString()
  const path = `/api/v1/teams/${teamId}/zones/${zoneId}/versions${query ? `?${query}` : ''}`
  return api().get<VersionPage>(path)
}

export async function restoreVersion(
  teamId: number,
  zoneId: number,
  versionId: number,
): Promise<{ zone: Zone }> {
  return api().post<{ zone: Zone }>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/versions/${versionId}/restore`,
  )
}
