export interface Point {
  x: number
  y: number
}

export interface Cone {
  origin: Point
  left_end: Point
  right_end: Point
  arc_bulge?: number
}

export interface ConeData {
  cones: Cone[]
}

export interface LineOfFire {
  id: number
  team_id: number
  map_id: number
  name: string
  polygon_data: ConeData
  order_index: number
  created_at: string
  updated_at: string
}

export interface RevisionLineOfFire {
  id: number
  source: 'team' | 'template'
  map_id: number
  polygon_data: ConeData
  name: string
}

export interface LineOfFireListItem {
  id: number
  name: string
  polygon_data: ConeData
  order_index: number
}

export type BatchLineOfFireAction =
  | { type: 'create'; temp_id: string; name: string; polygon_data: ConeData }
  | { type: 'delete'; line_of_fire_id: number }
  | { type: 'rename'; line_of_fire_id: number; name: string }
  | { type: 'add_cone'; line_of_fire_id: number; cone: Cone }
  | { type: 'update_cone'; line_of_fire_id: number; cone_index: number; cone: Cone }
  | { type: 'remove_cone'; line_of_fire_id: number; cone_index: number }
  | { type: 'update_polygon_data'; line_of_fire_id: number; polygon_data: ConeData }

export interface BatchLineOfFireResult {
  temp_id?: string
  line_of_fire_id?: number
  deleted?: boolean
}

export interface BatchLineOfFireResponse {
  results: BatchLineOfFireResult[]
}
