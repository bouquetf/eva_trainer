import { getApiClient } from './client'

export interface Template {
  id: number
  original_team_id: number
  map_id: number
  is_public: boolean
  created_at: string
  updated_at: string
}

export interface PublishStatus {
  published: boolean
  template?: Template
}

export async function publishMap(teamId: number, mapId: number): Promise<Template> {
  const api = getApiClient()
  return api.post(`/api/v1/teams/${teamId}/maps/${mapId}/publish`)
}

export async function getPublishStatus(teamId: number, mapId: number): Promise<PublishStatus> {
  const api = getApiClient()
  return api.get(`/api/v1/teams/${teamId}/maps/${mapId}/publish-status`)
}
