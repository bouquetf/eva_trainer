import { getApiClient } from './client'
import type { Point, Zone, ZonePolygon } from '../lib/api/types'

export interface CreateZonePayload {
  polygons?: { points: Point[] }[]
  name?: string
}

export interface CreateZoneResponse {
  zone: Zone
}

export interface ListZonesResponse {
  zones: Zone[]
}

export interface AddPolygonPayload {
  points: Point[]
}

export interface AddPolygonResponse {
  polygon: ZonePolygon
}

export interface RemovePolygonResponse {
  polygon_deleted: boolean
  zone_deleted: boolean
}

export interface UpdatePointsPayload {
  points: Point[]
}

export interface AddPointPayload {
  point: Point
  index_after: number
}

export interface SetNamePayload {
  name: string
}

export type BatchAction =
  | { type: 'create_zone'; temp_id: string; polygons: { points: Point[] }[]; name?: string }
  | { type: 'add_polygon'; zone_id: number; points: Point[] }
  | { type: 'remove_polygon'; polygon_id: number }
  | { type: 'update_points'; polygon_id: number; points: Point[] }
  | { type: 'add_point'; polygon_id: number; point: Point; index_after: number }
  | { type: 'remove_point'; polygon_id: number; point_index: number; zone_id: number }
  | { type: 'rename'; zone_id: number; name: string }
  | { type: 'delete_zone'; zone_id: number }

export interface BatchResult {
  results: unknown[]
}

const api = () => getApiClient()

export async function listZones(teamId: number, mapId: number): Promise<ListZonesResponse> {
  return api().get<ListZonesResponse>(`/api/v1/teams/${teamId}/maps/${mapId}/zones`)
}

export async function getZone(teamId: number, zoneId: number): Promise<{ zone: Zone }> {
  return api().get<{ zone: Zone }>(`/api/v1/teams/${teamId}/zones/${zoneId}`)
}

export async function createZone(
  teamId: number,
  mapId: number,
  payload: CreateZonePayload,
): Promise<CreateZoneResponse> {
  return api().post<CreateZoneResponse>(`/api/v1/teams/${teamId}/maps/${mapId}/zones`, payload)
}

export async function deleteZone(teamId: number, zoneId: number): Promise<void> {
  await api().del(`/api/v1/teams/${teamId}/zones/${zoneId}`)
}

export async function addPolygon(
  teamId: number,
  zoneId: number,
  payload: AddPolygonPayload,
): Promise<AddPolygonResponse> {
  return api().post<AddPolygonResponse>(`/api/v1/teams/${teamId}/zones/${zoneId}/polygons`, payload)
}

export async function removePolygon(
  teamId: number,
  zoneId: number,
  polygonId: number,
): Promise<RemovePolygonResponse> {
  return api().del<RemovePolygonResponse>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/polygons/${polygonId}`,
  )
}

export async function updatePolygonPoints(
  teamId: number,
  zoneId: number,
  polygonId: number,
  payload: UpdatePointsPayload,
): Promise<AddPolygonResponse> {
  return api().put<AddPolygonResponse>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/polygons/${polygonId}/points`,
    payload,
  )
}

export async function addPolygonPoint(
  teamId: number,
  zoneId: number,
  polygonId: number,
  payload: AddPointPayload,
): Promise<AddPolygonResponse> {
  return api().post<AddPolygonResponse>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/polygons/${polygonId}/points`,
    payload,
  )
}

export async function removePolygonPoint(
  teamId: number,
  zoneId: number,
  polygonId: number,
  pointIndex: number,
): Promise<AddPolygonResponse> {
  return api().del<AddPolygonResponse>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/polygons/${polygonId}/points/${pointIndex}`,
  )
}

export async function setZoneName(
  teamId: number,
  zoneId: number,
  payload: SetNamePayload,
): Promise<{ name: { id: number; zone_id: number; name: string } }> {
  return api().put<{ name: { id: number; zone_id: number; name: string } }>(
    `/api/v1/teams/${teamId}/zones/${zoneId}/name`,
    payload,
  )
}

export async function batchActions(
  teamId: number,
  mapId: number,
  actions: BatchAction[],
): Promise<BatchResult> {
  return api().post<BatchResult>(`/api/v1/teams/${teamId}/maps/${mapId}/zones/batch`, { actions })
}
