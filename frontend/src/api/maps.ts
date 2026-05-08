import { getApiClient } from './client'

export interface Map {
  id: number
  name: string
  added_at: string
  is_obsolete: boolean
  obsolete_of: number | null
}

export interface UploadMapResponse {
  map: Map
  previous_version_obsoleted: { id: number; name: string } | null
}

export interface AdminMapsResponse {
  maps: Map[]
}

export interface MapDetailResponse {
  map: Map
}

export interface PurgeResponse {
  message: string
  purged_ids: number[]
}

export async function uploadMap(name: string, file: File): Promise<UploadMapResponse> {
  const api = getApiClient()
  const formData = new FormData()
  formData.append('name', name)
  formData.append('file', file)
  return api.upload<UploadMapResponse>('/api/v1/admin/maps', formData)
}

export async function listAdminMaps(): Promise<AdminMapsResponse> {
  const api = getApiClient()
  return api.get<AdminMapsResponse>('/api/v1/admin/maps')
}

export async function getMapDetail(mapId: number): Promise<MapDetailResponse> {
  const api = getApiClient()
  return api.get<MapDetailResponse>(`/api/v1/admin/maps/${mapId}`)
}

export async function purgeObsoleteMaps(): Promise<PurgeResponse> {
  const api = getApiClient()
  return api.post<PurgeResponse>('/api/v1/admin/maps/purge')
}

export function getMapImageUrl(mapId: number): string {
  const baseURL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  return `${baseURL}/api/v1/maps/${mapId}/image`
}
