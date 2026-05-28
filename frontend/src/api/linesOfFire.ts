import { getApiClient } from './client'
import type { LineOfFire, Cone } from '../lib/api/types'

export interface CreateLineOfFirePayload {
  name: string
  polygon_data: { cones: Cone[] }
}

export interface CreateLineOfFireResponse {
  line_of_fire: LineOfFire
}

export interface ListLinesOfFireResponse {
  lines_of_fire: LineOfFire[]
}

export type LineOfFireBatchAction =
  | { type: 'create_line_of_fire'; temp_id: string; name: string; polygon_data: { cones: Cone[] } }
  | { type: 'add_cone'; line_of_fire_id: number; cone: Cone }
  | { type: 'update_cone'; line_of_fire_id: number; cone_index: number; cone: Cone }
  | { type: 'remove_cone'; line_of_fire_id: number; cone_index: number }
  | { type: 'rename_line_of_fire'; line_of_fire_id: number; name: string }
  | { type: 'delete_line_of_fire'; line_of_fire_id: number }

const api = () => getApiClient()

export async function listLinesOfFire(teamId: number, mapId: number): Promise<ListLinesOfFireResponse> {
  return api().get<ListLinesOfFireResponse>(`/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire`)
}

export async function createLineOfFire(teamId: number, mapId: number, payload: CreateLineOfFirePayload): Promise<CreateLineOfFireResponse> {
  return api().post<CreateLineOfFireResponse>(`/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire`, payload)
}

export async function getLineOfFire(teamId: number, lineOfFireId: number): Promise<{ line_of_fire: LineOfFire }> {
  return api().get<{ line_of_fire: LineOfFire }>(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}`)
}

export async function deleteLineOfFire(teamId: number, lineOfFireId: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}`)
}

export async function setLineOfFireName(teamId: number, lineOfFireId: number, name: string): Promise<{ name: { id: number; line_of_fire_id: number; name: string } }> {
  return api().put<{ name: { id: number; line_of_fire_id: number; name: string } }>(
    `/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/name`,
    { name },
  )
}

export async function addLineOfFireCone(teamId: number, lineOfFireId: number, cone: Cone): Promise<void> {
  await api().post(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones`, cone)
}

export async function updateLineOfFireCone(teamId: number, lineOfFireId: number, coneIndex: number, cone: Cone): Promise<void> {
  await api().put(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}`, cone)
}

export async function removeLineOfFireCone(teamId: number, lineOfFireId: number, coneIndex: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/lines-of-fire/${lineOfFireId}/cones/${coneIndex}`)
}

export async function batchLineOfFireActions(teamId: number, mapId: number, actions: LineOfFireBatchAction[]): Promise<{ results: unknown[] }> {
  return api().post<{ results: unknown[] }>(`/api/v1/teams/${teamId}/maps/${mapId}/lines-of-fire/batch`, { actions })
}
