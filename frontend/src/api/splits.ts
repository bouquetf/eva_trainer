import { getApiClient } from './client'
import type { Map } from './maps'

export interface Split {
  id: number
  name: string
  map_names: string[]
}

export interface SplitResponse {
  split: Split
}

export interface SplitsResponse {
  splits: Split[]
}

export interface SplitMapsResponse {
  maps: Map[]
}

export interface DeleteSplitResponse {
  message: string
}

export async function createSplit(data: { name: string; map_names: string[] }): Promise<SplitResponse> {
  const api = getApiClient()
  return api.post<SplitResponse>('/api/v1/admin/splits', data)
}

export async function listSplits(): Promise<SplitsResponse> {
  const api = getApiClient()
  return api.get<SplitsResponse>('/api/v1/admin/splits')
}

export async function getSplit(id: number): Promise<SplitResponse> {
  const api = getApiClient()
  return api.get<SplitResponse>(`/api/v1/admin/splits/${id}`)
}

export async function updateSplit(id: number, data: { name: string; map_names: string[] }): Promise<SplitResponse> {
  const api = getApiClient()
  return api.put<SplitResponse>(`/api/v1/admin/splits/${id}`, data)
}

export async function deleteSplit(id: number): Promise<DeleteSplitResponse> {
  const api = getApiClient()
  return api.del<DeleteSplitResponse>(`/api/v1/admin/splits/${id}`)
}

export async function getSplitMaps(id: number): Promise<SplitMapsResponse> {
  const api = getApiClient()
  return api.get<SplitMapsResponse>(`/api/v1/admin/splits/${id}/maps`)
}
