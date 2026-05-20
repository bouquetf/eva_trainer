export interface Point {
  x: number
  y: number
}

export interface ZonePolygon {
  id: number
  zone_id: number
  points: Point[]
  order_index: number
  floor?: 'haut' | 'bas'
  created_at: string
}

export interface RevisionZone {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygons: ZonePolygon[]
  name: string
}

export interface RevisionZonesResponse {
  zones: RevisionZone[]
}

export interface RevisionConfig {
  totalQuestions: number
  mapId: number
  mapName: string
  teamId: number
  correctTimerSeconds: number
  incorrectTimerSeconds: number
}
