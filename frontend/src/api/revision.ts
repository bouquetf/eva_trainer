import { getApiClient } from './client'
import type { RevisionZonesResponse, RevisionZone, RevisionLineOfFire, RevisionItemsResponse, RevisionFilter } from '../lib/revision/types'

export async function fetchRevisionZones(
  teamId: number,
  mapId: number,
): Promise<RevisionZone[]> {
  const api = getApiClient()
  const response = await api.get<RevisionZonesResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/revision-zones`,
  )
  return response.zones
}

export async function fetchRevisionItems(
  teamId: number,
  mapId: number,
  filter?: RevisionFilter,
): Promise<{ zones: RevisionZone[]; lines: RevisionLineOfFire[] }> {
  const api = getApiClient()
  const body: RevisionFilter = filter ?? { zones: true, lines: true }
  const response = await api.post<RevisionItemsResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/revision-items`,
    body,
  )
  return {
    zones: response.zones ?? [],
    lines: response.lines ?? [],
  }
}
