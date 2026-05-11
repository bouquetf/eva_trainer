import { getApiClient } from './client'
import type { RevisionZonesResponse, RevisionZone } from '../lib/revision/types'

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
