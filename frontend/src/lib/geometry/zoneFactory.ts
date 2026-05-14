import type { Point, Zone } from "../api/types";

export function createTempId(): string {
  return `zone-${Math.random().toString(36).substr(2, 9)}`;
}

export function createZone(
  teamId: number,
  mapId: number,
  name: string,
  points: Point[],
): Zone {
  const now = new Date().toISOString();
  return {
    id: 0,
    team_id: teamId,
    map_id: mapId,
    polygons: [
      {
        id: 0,
        zone_id: 0,
        points: [...points],
        order_index: 0,
        created_at: now,
      },
    ],
    name,
    created_at: now,
    updated_at: now,
  };
}

export function cloneZone(zone: Zone): Zone {
  const now = new Date().toISOString();
  return {
    ...zone,
    id: 0,
    polygons: zone.polygons.map((p) => ({
      ...p,
      id: 0,
      points: p.points.map((pt) => ({ ...pt })),
    })),
    name: `Copie de ${zone.name}`,
    created_at: now,
    updated_at: now,
  };
}

export function addPolygonToZone(zone: Zone, points: Point[]): Zone {
  const now = new Date().toISOString();
  return {
    ...zone,
    polygons: [
      ...zone.polygons,
      {
        id: 0,
        zone_id: zone.id,
        points: [...points],
        order_index: zone.polygons.length,
        created_at: now,
      },
    ],
    updated_at: now,
  };
}

export function removePolygonFromZone(zone: Zone, polygonIndex: number): Zone | null {
  if (zone.polygons.length <= 1) return null;
  const newPolygons = zone.polygons.filter((_, i) => i !== polygonIndex).map((p, i) => ({
    ...p,
    order_index: i,
  }));
  return {
    ...zone,
    polygons: newPolygons,
    updated_at: new Date().toISOString(),
  };
}

export function updatePolygonPoints(
  zone: Zone,
  polygonIndex: number,
  newPoints: Point[],
): Zone {
  const newPolygons = zone.polygons.map((p, i) =>
    i === polygonIndex ? { ...p, points: [...newPoints] } : p,
  );
  return {
    ...zone,
    polygons: newPolygons,
    updated_at: new Date().toISOString(),
  };
}

export function getZoneById(zones: Zone[], id: number): Zone | undefined {
  return zones.find((z) => z.id === id);
}

export function getZoneByTempId(zones: Zone[], tempId: string): Zone | undefined {
  return zones.find((z) => String(z.id) === tempId);
}

export function replaceTempId(zones: Zone[], tempId: string, realId: number): Zone[] {
  return zones.map((z) => {
    if (String(z.id) === tempId) {
      return {
        ...z,
        id: realId,
        polygons: z.polygons.map((p, _i) => ({
          ...p,
          zone_id: realId,
        })),
      };
    }
    return z;
  });
}
