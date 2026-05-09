import { getApiClient } from './client'

export interface MapCatalogEntry {
  id: number
  name: string
  created_at: string
}

export interface ListCatalogResponse {
  entries: MapCatalogEntry[]
}

export interface CatalogEntryResponse {
  entry: MapCatalogEntry
}

export async function listCatalog(): Promise<ListCatalogResponse> {
  const api = getApiClient()
  return api.get<ListCatalogResponse>('/api/v1/admin/map-catalog')
}

export async function createCatalogEntry(name: string): Promise<CatalogEntryResponse> {
  const api = getApiClient()
  return api.post<CatalogEntryResponse>('/api/v1/admin/map-catalog', { name })
}

export async function updateCatalogEntry(id: number, name: string): Promise<CatalogEntryResponse> {
  const api = getApiClient()
  return api.put<CatalogEntryResponse>(`/api/v1/admin/map-catalog/${id}`, { name })
}

export async function deleteCatalogEntry(id: number): Promise<{ message: string }> {
  const api = getApiClient()
  return api.del<{ message: string }>(`/api/v1/admin/map-catalog/${id}`)
}
