import type { Point, ZoneDto, ZoneStyle } from "../api/types";

export function createZoneFromPolygonPoints(
  mapId: string,
  label: string,
  points: Point[],
  style?: ZoneStyle
): ZoneDto {
  const now = new Date().toISOString();
  return {
    id: `zone-${Math.random().toString(36).substr(2, 9)}`,
    mapId,
    label,
    type: "polygon",
    points: [...points],
    style: style || {
      fill: "rgba(0, 255, 150, 0.25)",
      stroke: "#00e676",
      strokeWidth: 4,
    },
    createdAt: now,
    updatedAt: now,
  };
}

export function cloneZone(zone: ZoneDto, labelPrefix: string = "Copie de "): ZoneDto {
  const now = new Date().toISOString();
  return {
    ...zone,
    id: `zone-${Math.random().toString(36).substr(2, 9)}`,
    label: `${labelPrefix}${zone.label}`,
    createdAt: now,
    updatedAt: now,
    points: zone.points.map(p => ({ ...p })),
  };
}
