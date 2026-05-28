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

export interface RevisionLineOfFire {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygon_data: { cones: { origin: { x: number; y: number }; left_end: { x: number; y: number }; right_end: { x: number; y: number } }[] }
  name: string
}

export interface RevisionItemsResponse {
  zones?: RevisionZone[]
  lines?: RevisionLineOfFire[]
}

export interface ConePoint {
  x: number
  y: number
}

export interface RevisionCone {
  origin: ConePoint
  left_end: ConePoint
  right_end: ConePoint
}

export interface RevisionConeData {
  cones: RevisionCone[]
}

export type RevisionItem = RevisionZone & { type: 'zone' } | RevisionLineOfFire & { type: 'line_of_fire' }

export interface RevisionFilter {
  zones: boolean
  lines: boolean
}

export interface RevisionConfig {
  totalQuestions: number
  mapId: number
  mapName: string
  teamId: number
  correctTimerSeconds: number
  incorrectTimerSeconds: number
  filterMode: 'zones' | 'lines' | 'both'
}
