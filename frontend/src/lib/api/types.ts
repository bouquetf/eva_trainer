export type Point = {
  x: number;
  y: number;
};

export type ZoneShape = "polygon";

export type ZoneStyle = {
  fill: string;
  stroke: string;
  strokeWidth: number;
  strokeDasharray?: string;
};

export type ZoneDto = {
  id: string;
  mapId: string;
  label: string;
  type: ZoneShape;
  points: Point[];
  style?: ZoneStyle;
  createdAt: string;
  updatedAt: string;
};

export type MapDto = {
  id: string;
  name: string;
  imageUrl: string;
  width: number;
  height: number;
};

export type EditorTool = "select" | "draw-polygon" | "move-zone" | "edit-points";

export interface MapsApi {
  getMap(mapId: string): Promise<MapDto>;
  getZones(mapId: string): Promise<ZoneDto[]>;
  saveZones(mapId: string, zones: ZoneDto[]): Promise<ZoneDto[]>;
  deleteZone(mapId: string, zoneId: string): Promise<void>;
}
