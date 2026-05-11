import { getApiClient } from './client'

export interface Template {
  id: number
  original_team_id: number
  original_team_name: string
  map_id: number
  is_public: boolean
  map_name: string
  created_at: string
  updated_at: string
}

export interface TemplateZone {
  id: number
  template_id: number
  polygon_data: unknown
  name: string
  order_index: number
}

export interface TemplateWithZones {
  template: Template
  zones: TemplateZone[]
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

export async function listPublicTemplates(): Promise<{ templates: TemplateWithZones[] }> {
  const api = getApiClient()
  return api.get('/api/v1/templates/public')
}

export async function duplicateTemplate(teamId: number, templateId: number): Promise<{ zones: unknown[] }> {
  const api = getApiClient()
  return api.post(`/api/v1/teams/${teamId}/templates/${templateId}/duplicate`)
}
