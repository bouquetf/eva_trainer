export type Point = {
  x: number;
  y: number;
};

export type ZonePolygon = {
  id: number;
  zone_id: number;
  points: Point[];
  order_index: number;
  created_at: string;
};

export type Zone = {
  id: number;
  team_id: number;
  map_id: number;
  polygons: ZonePolygon[];
  name: string;
  created_at: string;
  updated_at: string;
};

/** @deprecated imageUrl/width/height will be removed in Task #005 */
export type MapDto = {
  id: number;
  name: string;
  added_at: string;
  is_obsolete: boolean;
  /** @deprecated Use getMapImageUrl() instead */
  imageUrl: string;
  /** @deprecated Use naturalWidth/naturalHeight from loaded img element */
  width: number;
  /** @deprecated Use naturalWidth/naturalHeight from loaded img element */
  height: number;
};

export type EditorTool = "select" | "draw-polygon" | "move-zone" | "edit-points";

/** @deprecated Use Zone instead. Kept for transition during Task #005. */
export type ZoneDto = {
  id: string;
  mapId: string;
  label: string;
  type: "polygon";
  points: Point[];
  style?: ZoneStyle;
  createdAt: string;
  updatedAt: string;
};

/** @deprecated Use ZonePolygon instead. Kept for transition during Task #005. */
export type ZoneShape = "polygon";

export type ZoneStyle = {
  fill: string;
  stroke: string;
  strokeWidth: number;
  strokeDasharray?: string;
};

/** @deprecated Use real API functions (zones.ts, locks.ts, maps.ts) instead. */
export interface MapsApi {
  getMap(mapId: string): Promise<MapDto>;
  getZones(mapId: string): Promise<ZoneDto[]>;
  saveZones(mapId: string, zones: ZoneDto[]): Promise<ZoneDto[]>;
  deleteZone(mapId: string, zoneId: string): Promise<void>;
}
