import { getApiClient } from './client'
import type { LineOfFire, ConeData, Cone, BatchLineOfFireAction, BatchLineOfFireResponse } from '../lib/lines_of_fire/types'

interface ListLinesOfFireResponse {
  lines_of_fire: LineOfFire[]
}

interface LineOfFireResponse {
  line_of_fire: LineOfFire
}

interface ImportLineOfFirePayload {
  name: string
  polygon_data: ConeData
}

interface ImportLinesOfFireRequest {
  lines_of_fire: ImportLineOfFirePayload[]
}

interface ImportLinesOfFireResponse {
  lines_of_fire: LineOfFire[]
}

interface SetNameResponse {
  name: { id: number; line_of_fire_id: number; name: string }
}

const api = () => getApiClient()

export async function listLinesOfFire(teamId: number, mapId: number): Promise<LineOfFire[]> {
  const response = await api().get<ListLinesOfFireResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire`,
  )
  return response.lines_of_fire
}

export async function getLineOfFire(teamId: number, lineOfFireId: number): Promise<LineOfFire> {
  const response = await api().get<LineOfFireResponse>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}`,
  )
  return response.line_of_fire
}

export async function createLineOfFire(
  teamId: number,
  mapId: number,
  name: string,
  polygonData: ConeData,
): Promise<LineOfFire> {
  const response = await api().post<LineOfFireResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire`,
    { name, polygon_data: polygonData },
  )
  return response.line_of_fire
}

export async function deleteLineOfFire(teamId: number, lineOfFireId: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}`)
}

export async function setLineOfFireName(
  teamId: number,
  lineOfFireId: number,
  name: string,
): Promise<void> {
  await api().put<SetNameResponse>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/name`,
    { name },
  )
}

export async function addLineOfFireCone(
  teamId: number,
  lineOfFireId: number,
  cone: Cone,
): Promise<LineOfFire> {
  const response = await api().post<LineOfFireResponse>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones`,
    cone,
  )
  return response.line_of_fire
}

export async function updateLineOfFireCone(
  teamId: number,
  lineOfFireId: number,
  coneIndex: number,
  cone: Cone,
): Promise<LineOfFire> {
  const response = await api().put<LineOfFireResponse>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}`,
    cone,
  )
  return response.line_of_fire
}

export async function removeLineOfFireCone(
  teamId: number,
  lineOfFireId: number,
  coneIndex: number,
): Promise<LineOfFire> {
  const response = await api().del<LineOfFireResponse>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}`,
  )
  return response.line_of_fire
}

export async function batchLinesOfFireActions(
  teamId: number,
  mapId: number,
  actions: BatchLineOfFireAction[],
): Promise<BatchLineOfFireResponse> {
  return api().post<BatchLineOfFireResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire/batch`,
    { actions },
  )
}

export async function importLinesOfFire(
  teamId: number,
  mapId: number,
  lines: ImportLineOfFirePayload[],
): Promise<LineOfFire[]> {
  const response = await api().post<ImportLinesOfFireResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire/import`,
    { lines_of_fire: lines } as ImportLinesOfFireRequest,
  )
  return response.lines_of_fire
}

export async function exportLinesOfFire(
  teamId: number,
  mapId: number,
): Promise<LineOfFire[]> {
  const response = await api().get<ImportLinesOfFireResponse>(
    `/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire/export`,
  )
  return response.lines_of_fire
}
